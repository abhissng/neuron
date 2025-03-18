package log

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/abhissng/neuron/utils/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents the severity level of a log message.
type LogLevel string

const (
	// DebugLevel is the lowest severity level, used for detailed debugging information.
	DebugLevel LogLevel = "debug"
	// InfoLevel is used for general informational messages.
	InfoLevel LogLevel = "info"
	// WarnLevel is used for warnings and potential problems.
	WarnLevel LogLevel = "warn"
	// ErrorLevel is used for errors that have occurred.
	ErrorLevel LogLevel = "error"
	// FatalLevel is the highest severity level, used for critical errors that result in program termination.
	FatalLevel LogLevel = "fatal"
)

// Helper functions to create fields without directly using zap

// String creates a single types.Field (string) for a given key-value pair.
func String(key string, value string) types.Field {
	return zap.String(key, value)
}

// Int creates a single types.Field (int) for a given key-value pair.
func Int(key string, value int) types.Field {
	return zap.Int(key, value)
}

// Int64 creates a single types.Field (int64) for a given key-value pair.
func Int64(key string, value int64) types.Field {
	return zap.Int64(key, value)
}

// Float64 creates a single types.Field (float64) for a given key-value pair.
func Float64(key string, value float64) types.Field {
	return zap.Float64(key, value)
}

// Bool creates a single types.Field (bool) for a given key-value pair.
func Bool(key string, value bool) types.Field {
	return zap.Bool(key, value)
}

// Time creates a single types.Field (time.Time) for a given key-value pair.
func Time(key string, value time.Time) types.Field {
	return zap.Time(key, value)
}

// Duration creates a single types.Field (time.Duration) for a given key-value pair.
func Duration(key string, value time.Duration) types.Field {
	return zap.Duration(key, value)
}

// Any creates a single types.Field (any) for a given key-value pair.
func Any(key string, value any) types.Field {
	return zap.Any(key, value)
}

// Err creates a single types.Field (error) for a given error.
func Err(err error) types.Field {
	return zap.Error(err)
}

// Stringer creates a single types.Field (fmt.Stringer) for a given key-value pair.
func Stringer(key string, value fmt.Stringer) types.Field {
	return zap.Stringer(key, value)
}

// WithField creates a single types.Field (any) for a given key-value pair.
func WithField(key string, value any) types.Field {
	return zap.Any(key, value)
}

// WithFields creates a slice of types.Field (any) for given slice of fields.
func WithFields(fields ...types.Field) []types.Field {
	return fields
}

// GetLogLevelForEnvironment returns the appropriate log level based on environment
func GetLogLevelForEnvironment(isProd bool) LogLevel {
	if isProd {
		return InfoLevel
	}
	return DebugLevel
}

// getZapLevel converts our LogLevel to zap.Level
//
//lint:ignore U1000 ignore unused function will use later
func getZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// GetCallerFunctionName returns the name of the function that called it
func GetCallerFunctionName(skip int) string {
	// Use `skip + 1` to get the caller of the caller
	pc, _, _, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	// Extract the function name (e.g., "github.com/yourpackage.YourFunction")
	fullName := fn.Name()

	// Simplify the function name to just the last part (e.g., "YourFunction")
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

// Sprintf is a wrapper around fmt.Sprintf
func Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}
