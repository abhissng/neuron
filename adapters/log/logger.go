package log

import (
	"fmt"
	"os"
	"sync"

	"github.com/abhissng/neuron/adapters/opensearch"
	"github.com/abhissng/neuron/utils/helpers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log struct holds the zap Logger instance.
type Log struct {
	*zap.Logger
	mu        sync.Mutex   // Mutex for thread-safe logging
	closeLog  func() error // Function to gracefully shut down the logger
	sanitizer *helpers.Sanitizer
}

// It creates basic logger for utilities function and by default it will carry default confinguration
func NewBasicLogger(isProd, isOpenSearchDisabled bool) *Log {
	basicLogger, _ := NewLogger(NewLoggerConfig(isProd, WithDisableOpenSearch(isOpenSearchDisabled)))
	return &Log{
		Logger: basicLogger.Logger,
		closeLog: func() error {
			return basicLogger.Sync()
		},
	}
}

// NewLogger creates a new Log instance with the specified log level and options.
// If cfg.Sanitizer is set, use l.Any(key, value) when logging to mask sensitive fields.
func NewLogger(cfg *LoggerConfig) (*Log, error) {

	// ✅ 1. Set the log level
	atomicLevel := zap.NewAtomicLevel()
	if cfg.IsProd {
		atomicLevel.SetLevel(zapcore.InfoLevel)
	} else {
		atomicLevel.SetLevel(zapcore.DebugLevel) // Debug mode for development
	}

	// ✅ 2. Configure encoder settings
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "log",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		EncodeLevel: func() zapcore.LevelEncoder {
			if cfg.IsProd {
				return zapcore.CapitalLevelEncoder
			}
			return zapcore.CapitalColorLevelEncoder
		}(), // INFO, WARN, ERROR (readable)
		EncodeTime: zapcore.ISO8601TimeEncoder, // 2025-02-22T13:43:42.977+0530
		// EncodeCaller:   zapcore.ShortCallerEncoder,       // nats/nats.go:120
		EncodeCaller:   helpers.TailCallerEncoder(cfg.EncoderTailLength),
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	defaultOptions := []zap.Option{
		zap.Fields(
			zap.String("environment", cfg.Environment),
			zap.String("service", cfg.ServiceName),
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}
	options := append(defaultOptions, cfg.ZapOptions...)

	// ✅ 3. Select the encoder based on mode
	var encoder zapcore.Encoder
	if cfg.IsProd {
		encoder = zapcore.NewJSONEncoder(encoderConfig) // JSON logs for production
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig) // Readable console logs
	}

	// ✅ 4. Setup log output (stdout by default, but can be rotated)
	logOutput := zapcore.AddSync(os.Stdout)

	// ✅ 5. Create the logger core
	core := zapcore.NewCore(encoder, logOutput, atomicLevel)

	// ✅ 6. Create OpenSearch core
	var closeFunc func() error

	// ✅ 7. Create a list of all cores. Start with the local one.
	cores := []zapcore.Core{core}

	// ✅ 8. Add OpenSearch core if enabled
	osCore, closer := opensearch.GetOpenSearchLogCore(atomicLevel, cfg.OpenSearchOptions...)
	if osCore != nil {
		cores = append(cores, osCore)
		closeFunc = closer
	}

	// ✅ 9. Combine all cores using NewTee.
	// Every log message will now be sent to every core in the 'cores' slice.
	finalCore := zapcore.NewTee(cores...)

	// ✅ 10. Build the logger with additional options
	l := zap.New(finalCore, options...)

	return &Log{Logger: l, closeLog: closeFunc, sanitizer: cfg.Sanitizer}, nil
}

// GetEncoderPool returns a sync.Pool of zapcore.Encoder instances.
func GetEncoderPool() *sync.Pool {
	// Define a sync.Pool for encoders.
	encoderPool := &sync.Pool{
		New: func() any {
			return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				LevelKey:       "level",
				TimeKey:        "time",
				NameKey:        "log",
				MessageKey:     "msg",
				CallerKey:      "caller",
				StacktraceKey:  "stacktrace",
				EncodeLevel:    zapcore.CapitalLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				// EncodeCaller : zapcore.ShortCallerEncoder
				EncodeCaller: helpers.TailCallerEncoder(4),
			})
		},
	}
	return encoderPool
}

// **SafeLog** ensures thread-safe logging.
func (l *Log) SafeLog(level zapcore.Level, msg string, fields ...zap.Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch level {
	case zap.DebugLevel:
		l.Logger.Debug(msg, fields...)
	case zap.InfoLevel:
		l.Logger.Info(msg, fields...)
	case zap.WarnLevel:
		l.Logger.Warn(msg, fields...)
	case zap.ErrorLevel:
		l.Logger.Error(msg, fields...)
	case zap.FatalLevel:
		l.Logger.Fatal(msg, fields...)
	}
}

// Debug logs a message at the DebugLevel.
func (l *Log) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs a message at the InfoLevel.
func (l *Log) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs a message at the WarnLevel.
func (l *Log) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs a message at the ErrorLevel.
func (l *Log) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal logs a message at the FatalLevel and then exits the program.
func (l *Log) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// With creates a child Log with the specified fields.
func (l *Log) With(fields ...zap.Field) *Log {
	return &Log{Logger: l.Logger.With(fields...), sanitizer: l.sanitizer}
}

// Any returns a zap field; if this logger has a sanitizer, value is sanitized (blocked keys masked) before logging.
// Use this for request/response bodies, headers, or any struct/map that may contain secrets.
func (l *Log) Any(key string, value any) zap.Field {
	if l.sanitizer != nil {
		value = l.sanitizer.Sanitize(value)
	}
	return zap.Any(key, value)
}

// Sanitize returns a zap field with the value sanitized when the logger has a sanitizer; it simply calls l.Any(key, value).
// Use for audit logging when you want the name to express that the field is sanitized.
func (l *Log) Sanitize(key string, value any) zap.Field {
	return l.Any(key, value)
}

// SanitizeValue returns value with sensitive fields masked if this logger has a sanitizer; otherwise returns value unchanged.
// Use when building fields manually, e.g. log.Any("body", logger.SanitizeValue(body)).
func (l *Log) SanitizeValue(value any) any {
	if l.sanitizer != nil {
		return l.sanitizer.Sanitize(value)
	}
	return value
}

func (l *Log) Printf(level zapcore.Level, msg string, v ...interface{}) {
	formattedMsg := fmt.Sprintf(msg, v...)
	switch level {
	case zap.DebugLevel:
		l.Logger.Debug(formattedMsg)
	case zap.InfoLevel:
		l.Logger.Info(formattedMsg)
	case zap.WarnLevel:
		l.Logger.Warn(formattedMsg)
	case zap.ErrorLevel:
		l.Logger.Error(formattedMsg)
	case zap.FatalLevel:
		l.Logger.Fatal(formattedMsg)
	}
}

// Sync flushes any buffered log entries. Applications should take care to call
// Sync before exiting.
func (l *Log) Sync() error {
	err := l.Logger.Sync()

	// Then, gracefully close our custom async writer if it exists
	if l.closeLog != nil {
		if closeErr := l.closeLog(); closeErr != nil {
			// Combine errors if both operations fail
			if err != nil {
				return fmt.Errorf("zap sync error: %w; async close error: %v", err, closeErr)
			}
			return closeErr
		}
	}
	return err
}

// getLumberjackLogger returns a WriteSyncer for file logging if rotation is enabled
//
//lint:ignore U1000 // This function is used for logging and might be called later, so we're keeping it for now.
func getLumberjackLogger() zapcore.WriteSyncer {
	if !helpers.GetIsLogRotationEnabled() {
		return nil // No file rotation, return nil
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   helpers.CreateLogDirectory(), // Log file location
		MaxSize:    50,                           // Max size in MB before rotating
		MaxBackups: 5,                            // Max old log files
		MaxAge:     30,                           // Max days to retain logs
		Compress:   true,                         // Compress rotated files
	}

	return zapcore.AddSync(lumberjackLogger)
}
