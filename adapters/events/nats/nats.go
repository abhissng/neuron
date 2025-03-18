package nats

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/idempotency"
	"github.com/abhissng/neuron/utils/types"
	"github.com/sony/gobreaker"

	"github.com/nats-io/nats.go"
)

//----------------------------------------------------------
// GENERIC NATS WRAPPER WITH CIRCUIT BREAKER
//----------------------------------------------------------

// NATSManager encapsulates the NATS connection, JetStream context, and a circuit breaker.
type NATSManager struct {
	context.Context
	nc                 *nats.Conn
	js                 nats.JetStreamContext // JetStream context
	mu                 sync.Mutex
	logger             *log.Log
	loggerSet          bool
	idempotencyManager *idempotency.IdempotencyManager[string]
	breaker            *gobreaker.CircuitBreaker
	subjects           map[string]*nats.Subscription
	subParams          map[string]*subscriptionParams // Track subscription parameters
	done               chan struct{}                  // Channel to signal shutdown
	reconnect          bool                           // Flag to enable auto-reconnection
}

// subscriptionParams stores the parameters for a subscription
type subscriptionParams struct {
	queue   string
	handler nats.MsgHandler
}

/*
foo.*: Matches subjects like foo.bar, foo.baz, but not foo.bar.baz.
foo.>: Matches subjects like foo.bar, foo.bar.baz, foo.baz.qux, etc.
>.foo: Matches subjects like bar.foo, baz.foo, qux.foo, etc.
>.foo.*: Matches subjects like bar.foo.baz, baz.foo.qux, but not foo.bar.
*/

// NewNATSManager initializes a new generic NATS wrapper
func NewNATSManager(url string, options ...Option) (*NATSManager, error) {
	defaultLog := log.NewBasicLogger(helpers.IsProdEnvironment())

	// Configure NATS options for reliability
	opts := []nats.Option{
		nats.MaxReconnects(DefautMaxReconnects),
		nats.ReconnectWait(DefaultReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			defaultLog.Error("NATS disconnected", log.Any("error", err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			defaultLog.Info("NATS reconnected", log.Any("url", nc.ConnectedUrl()))
		}),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}

	wrapper := &NATSManager{
		Context:            context.Background(),
		nc:                 nc,
		subjects:           make(map[string]*nats.Subscription),
		subParams:          make(map[string]*subscriptionParams),
		logger:             defaultLog,
		loggerSet:          false,
		idempotencyManager: idempotency.NewIdempotencyManager[string](idempotency.DefaultCleanupInterval),
		done:               make(chan struct{}),
		reconnect:          true,
		breaker:            nil,
	}

	for _, opt := range options {
		opt(wrapper)
	}

	if wrapper.loggerSet {
		_ = defaultLog.Sync()
	}

	return wrapper, nil
}

// Ping checks the connection to the nats
func (w *NATSManager) Ping() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.nc != nil {
		if w.nc.IsConnected() {
			return nil
		}
	}
	return errors.New(ConnectionFailedMessage)
}

// Close gracefully shuts down the wrapper
func (w *NATSManager) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	defer close(w.done)

	for subject, sub := range w.subjects {
		if err := sub.Unsubscribe(); err != nil {
			b := blame.UnsubscribeFailedError(subject, err)
			message, description := b.Translate()
			w.logger.Error(blame.ErrorUnsubscribeFailed.String(),
				log.Any("message", message), log.Any("description", description))
		}
	}
	// Clear the map to prevent double Unsubscribe
	w.subjects = make(map[string]*nats.Subscription)

	if w.nc != nil && !w.nc.IsClosed() {
		w.logger.Info(constant.ConnectionClosing, log.Any("message", "NATS connection closing"))
		_ = w.nc.Drain()
		// w.nc.Close()
	}

	if w.idempotencyManager != nil {
		w.idempotencyManager.Close()
	}
	w.logger.Info(constant.ConnectionClosed, log.Any("message", "NATS connection closed"))
}

// ackIfJetStream sends an ACK if using JetStream
//
//lint:ignore U1000 // This function is used for JetStream acking and might be called later, so we're keeping it for now.
func (w *NATSManager) ackIfJetStream(msg *nats.Msg) {
	if w.js != nil {
		if err := msg.Ack(); err != nil {
			w.logger.Error("Failed to ACK message", log.Any("error", err))
		}
	}
}

// nakIfJetStream sends a NAK if using JetStream
//
//lint:ignore U1000 // This function is used for JetStream acking and might be called later, so we're keeping it for now.
func (w *NATSManager) nakIfJetStream(msg *nats.Msg) {
	if w.js != nil {
		if err := msg.Nak(); err != nil {
			w.logger.Error("Failed to NAK message", log.Any("error", err))
		}
	}
}

// processMessageIDHeader process an incoming NATS message header using MessageID.
//
// 1. It checks if the message has a "Message-ID" header. If not, an error is logged and the message is discarded.
// 2. It acquires a mutex to ensure thread safety when accessing and modifying internal state.
// 3. It checks if the message has already been processed. If so, a log message is printed and the message is discarded.
// 4. If the message is not processed, it marks it as processed in the internal map.
func (w *NATSManager) processMessageIDHeader(msg *nats.Msg) string {
	messageID := msg.Header.Get(constant.MessageIdHeader)
	if messageID == "" {
		w.logger.Error(constant.MessageIdHeader + " header is missing")
		return ""
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.idempotencyManager.IsProcessed(messageID) {
		w.logger.Info("Message already processed", log.Any(constant.MessageIdHeader, messageID))
		return ""
	}

	// Mark the message as processed
	w.idempotencyManager.MarkAsProcessed(messageID)
	return messageID
}

// handleMessage handles an incoming NATS message.
//
// 1. It calls the processMessageIDHeader function to process the message id header
// 2. It calls the provided handler function to process the message.
// 3. A log message is printed indicating that the message has been successfully processed.
func (w *NATSManager) handleMessage(msg *nats.Msg, handler nats.MsgHandler) {

	messageID := w.processMessageIDHeader(msg)

	// Process the message
	handler(msg)

	w.logger.Info("Message processed", log.Any("message_id", messageID))
}

// monitorSubscription ensures subscription stays active
func (w *NATSManager) monitorSubscription(subject string, sub *nats.Subscription) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			if !sub.IsValid() && w.reconnect {
				w.logger.Warn("Subscription invalid, attempting to resubscribe",
					log.Any("subject", subject))
				w.resubscribe(subject)
			}
		}
	}
}

// resubscribe attempts to reestablish an invalid subscription
// In resubscribe, re-subscribe using stored parameters:
func (w *NATSManager) resubscribe(subject string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if sub, exists := w.subjects[subject]; exists {
		_ = sub.Unsubscribe()
		delete(w.subjects, subject)
	}

	if params, ok := w.subParams[subject]; ok {

		var sub *nats.Subscription
		var err error

		if w.js != nil {
			sub, err = w.js.QueueSubscribe(
				subject,
				params.queue,
				params.handler,
				nats.ManualAck(),
				nats.Durable(params.queue),
			)
		} else {
			sub, err = w.nc.QueueSubscribe(subject, params.queue, params.handler)
		}
		if err != nil {
			w.logger.Error("Failed to resubscribe:", log.Err(err))
			return
		}

		if w.js == nil {
			// Ensure subscription is active before continuing
			err = w.nc.Flush()
			if err != nil {
				w.logger.Error("Failed to flush subscriptions:", log.Err(err))
				return
			}
		}

		w.subjects[subject] = sub
	}
}

// FetchMessageAndCorrelationField returns a slice of types.Field with request and correlation fields.
func FetchMessageAndCorrelationField(msg *nats.Msg) []types.Field {
	fields := make([]types.Field, 2)
	fields[0] = log.String(constant.MessageIdHeader, helpers.MessageIDFromNatsMsg(msg))
	fields[1] = log.String(constant.CorrelationIDHeader, helpers.CorrelationIDFromNatsMsg(msg))
	return fields
}

// Slog returns a slice of types.Field with message and correlation fields and additional fields.
func Slog(msg *nats.Msg, withFields ...types.Field) []types.Field {
	// Start with the message and correlation fields
	fields := make([]types.Field, 0, 3+len(withFields))
	fields = append(fields, FetchMessageAndCorrelationField(msg)...)
	fields = append(fields, log.String(constant.IPHeader, helpers.IPHeaderFromNatsMsg(msg)))

	// Append additional fields provided as variadic arguments
	fields = append(fields, withFields...)
	return fields
}

func (w *NATSManager) RunSafely(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			msg := fmt.Sprintf("Panic recovered: %v\nStack Trace:\n%s", r, string(stack))
			// Use the service logger if available, otherwise use standard log
			if w.logger != nil {
				w.logger.Error("Panic recovered", log.Any("error", msg))
			} else {
				helpers.Println(constant.INFO, msg)
			}

		}
	}()
	fn()
}
