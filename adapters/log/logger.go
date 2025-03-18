package log

import (
	"fmt"
	"os"
	"sync"

	"github.com/abhissng/neuron/utils/helpers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log struct holds the zap Logger instance.
type Log struct {
	*zap.Logger
	mu sync.Mutex // Mutex for thread-safe logging
}

// It creates basic logger for utilities function and by default it will carry default confinguration
func NewBasicLogger(isProd bool) *Log {
	basicLogger, _ := NewLogger(isProd)
	return &Log{
		Logger: basicLogger.Logger,
	}
}

// NewLogger creates a new Log instance with the specified log level and options.
func NewLogger(isProd bool, options ...zap.Option) (*Log, error) {

	// ✅ 1. Set the log level
	atomicLevel := zap.NewAtomicLevel()
	if isProd {
		atomicLevel.SetLevel(zapcore.InfoLevel)
	} else {
		atomicLevel.SetLevel(zapcore.DebugLevel) // Debug mode for development
	}

	// ✅ 2. Configure encoder settings
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "log",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // INFO, WARN, ERROR (readable)
		EncodeTime:     zapcore.ISO8601TimeEncoder,       // 2025-02-22T13:43:42.977+0530
		EncodeCaller:   zapcore.ShortCallerEncoder,       // nats/nats.go:120
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	defaultOptions := []zap.Option{
		zap.Fields(
			zap.String("environment", "production"),
			zap.String("service", helpers.GetServiceName()),
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}
	options = append(defaultOptions, options...)

	// ✅ 3. Select the encoder based on mode
	var encoder zapcore.Encoder
	if isProd {
		encoder = zapcore.NewJSONEncoder(encoderConfig) // JSON logs for production
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig) // Readable console logs
	}

	// ✅ 4. Setup log output (stdout by default, but can be rotated)
	logOutput := zapcore.AddSync(os.Stdout)

	// ✅ 5. Create the logger core
	core := zapcore.NewCore(encoder, logOutput, atomicLevel)

	// ✅ 6. Build the logger with additional options
	l := zap.New(core, options...)

	return &Log{Logger: l}, nil
}

// GetEncoderPool returns a sync.Pool of zapcore.Encoder instances.
func GetEncoderPool() *sync.Pool {
	// Define a sync.Pool for encoders.
	encoderPool := &sync.Pool{
		New: func() any {
			return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				LevelKey:      "level",
				TimeKey:       "time",
				NameKey:       "log",
				MessageKey:    "msg",
				CallerKey:     "caller",
				StacktraceKey: "stacktrace",
				EncodeLevel:   zapcore.CapitalLevelEncoder,
				EncodeTime:    zapcore.ISO8601TimeEncoder,
				EncodeCaller:  zapcore.ShortCallerEncoder,
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
	return &Log{Logger: l.Logger.With(fields...)}
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
	return l.Logger.Sync()
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
