package opensearch

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// OpenSearchWriter is an asynchronous writer that sends logs to OpenSearch.
type OpenSearchWriter struct {
	client       *opensearchapi.Client
	indexName    string
	logChannel   chan []byte   // Buffer for incoming logs
	doneChannel  chan struct{} // For signaling shutdown
	batchSize    int           // Number of logs to buffer before sending
	flushTimeout time.Duration // How often to flush logs
	wg           sync.WaitGroup
}

// Write is now non-blocking. It sends the log to a channel.
func (w *OpenSearchWriter) Write(p []byte) (n int, err error) {
	// We need to copy the byte slice because zap reuses the underlying array.
	logData := make([]byte, len(p))
	copy(logData, p)

	select {
	case w.logChannel <- logData:
		// Log successfully sent to the buffer channel
	default:
		// Channel is full, meaning we are logging faster than we can send.
		// To prevent blocking the app, we drop the log.
		helpers.Println(constant.WARN, "OpenSearch log buffer is full. Dropping log.")
	}
	return len(p), nil
}

// Sync is a no-op for this writer, as each Write is sent immediately.
func (w *OpenSearchWriter) Sync() error {
	return nil
}

// start runs the background worker goroutine.
func (w *OpenSearchWriter) start() {
	w.wg.Add(1)
	go func() {
		defer func() {
			helpers.Println(constant.INFO, "OpenSearch writer stopped")
			w.wg.Done()
		}()

		batch := make([][]byte, 0, w.batchSize)
		ticker := time.NewTicker(w.flushTimeout)
		defer ticker.Stop()

		for {
			select {
			case logData, ok := <-w.logChannel:
				if !ok {
					// Channel closed, flush remaining batch and exit
					w.flush(batch)
					helpers.Println(constant.ERROR, "channel closed existing now")
					return
				}
				batch = append(batch, logData)
				if len(batch) >= w.batchSize {
					w.flush(batch)
					batch = make([][]byte, 0, w.batchSize) // Reset batch
				}
			case <-ticker.C:
				// Timer fired, flush whatever is in the batch
				if len(batch) > 0 {
					w.flush(batch)
					batch = make([][]byte, 0, w.batchSize) // Reset batch
				}
			case <-w.doneChannel:
				// Shutdown signal received
				w.flush(batch)
				return
			}
		}
	}()
}

// flush sends a batch of logs to OpenSearch's Bulk API.
func (w *OpenSearchWriter) flush(batch [][]byte) {
	if len(batch) == 0 {
		return
	}

	var body bytes.Buffer
	for _, doc := range batch {
		// Each document in a bulk request needs a header line.
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s" } }%s`, w.indexName, "\n"))
		body.Write(meta)
		body.Write(doc)
		body.WriteByte('\n') // Bulk format requires a newline at the end of each doc line
	}

	req := opensearchapi.BulkReq{
		Body: &body,
	}

	res, err := w.client.Bulk(context.Background(), req)
	if err != nil {
		helpers.Println(constant.ERROR, "Failed to execute OpenSearch bulk request: ", err)
		return
	}
	defer func() {
		_ = res.Inspect().Response.Body.Close()
	}()

	if res.Inspect().Response.IsError() {
		helpers.Println(constant.ERROR, "OpenSearch bulk indexing failed: ", res.Inspect().Response.String())
	}
}

// close handles the graceful shutdown.
func (w *OpenSearchWriter) close() error {
	// Signal the worker to stop
	close(w.doneChannel)
	// Wait for the worker to finish flushing
	w.wg.Wait()
	return nil
}

// GetOpenSearchLogCore creates and configures a zapcore.Core for OpenSearch.
func GetOpenSearchLogCore(level zap.AtomicLevel, opts ...Option) (zapcore.Core, func() error) {
	if !helpers.GetIsOpenSearchEnabled() {
		return nil, nil // OpenSearch logging is disabled
	}

	// --- 1. Create OpenSearch Client ---
	client, options, err := NewClient(
		helpers.GetOpenSearchAddresses(),
		helpers.GetOpenSearchUsername(),
		helpers.GetOpenSearchPassword(),
		opts...,
	)
	if err != nil {
		if err.Error() == constant.OpenSearchDisabledError.String() {
			return nil, nil
		}
		helpers.Println(constant.ERROR, "Cannot initialize OpenSearch client: ", err)
		return nil, nil
	}
	helpers.Println(constant.INFO, "OpenSearch client initialized successfully")
	// client, err := opensearch.NewClient(opensearch.Config{
	// 	Addresses: helpers.GetOpenSearchAddresses(),
	// 	Username:  helpers.GetOpenSearchUsername(),
	// 	Password:  helpers.GetOpenSearchPassword(),
	// 	// For development, you might need to skip insecure cert verification
	// 	Transport: &http.Transport{
	// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// 	},
	// })

	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "ERROR: Cannot initialize OpenSearch client: %s\n", err)
	// 	return nil
	// }

	// --- 2. Create a dedicated JSON encoder for OpenSearch ---
	// This ensures logs are always JSON, without console color codes.
	osEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "log",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // "INFO", not colored
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		// EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeCaller: helpers.TailCallerEncoder(options.EncoderLength),
	}
	osEncoder := zapcore.NewJSONEncoder(osEncoderConfig)

	// --- 3. Create the custom WriteSyncer ---
	writer, err := NewOpenSearchWriter(client, helpers.GetOpenSearchIndexName())
	if err != nil {
		return nil, nil
	}
	writer.start()

	// --- 4. Return the new core ---
	return zapcore.NewCore(osEncoder, writer, level), writer.close
}
