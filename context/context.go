package context

///
import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/abhissng/neuron/adapters/events/nats"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/random"
	"github.com/abhissng/neuron/utils/types"
)

// DefaultContext is a default implementation of the Context interface.
type DefaultContext struct {
	context.Context
}

// NewDefaultContext creates a new DefaultContext.
func NewDefaultContext() *DefaultContext {
	return &DefaultContext{Context: context.Background()}
}

// Background creates a new context with a new background context.
func (s *ServiceContext) Background() context.Context {
	return context.Background()
}
func (s *ServiceContext) GetPreField() *ServiceContext {
	return &ServiceContext{
		// These are unaffected fields
		DefaultContext: s.DefaultContext,
		AppContext:     s.AppContext,
		Context:        s.Context,
	}
}

// WithCancel creates a new ServiceContext with a cancel function.
func (s *ServiceContext) WithCancel() (*ServiceContext, context.CancelFunc) {
	ctx, cancel := context.WithCancel(s.DefaultContext)
	return &ServiceContext{
		DefaultContext: &DefaultContext{Context: ctx},

		// These are unaffected fields
		AppContext: s.AppContext,
		Context:    s.Context,
	}, cancel
}

// WithTimeout creates a new ServiceContext with a timeout.
func (s *ServiceContext) WithTimeout(timeout time.Duration) (*ServiceContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(s.DefaultContext, timeout)
	return &ServiceContext{
		DefaultContext: &DefaultContext{Context: ctx},

		// These are unaffected fields
		AppContext: s.AppContext,
		Context:    s.Context,
	}, cancel
}

// WithDeadline creates a new ServiceContext with a deadline.
func (s *ServiceContext) WithDeadline(deadline time.Time) (*ServiceContext, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(s.DefaultContext, deadline)
	return &ServiceContext{
		DefaultContext: &DefaultContext{Context: ctx},

		// These are unaffected fields
		AppContext: s.AppContext,
		Context:    s.Context,
	}, cancel
}

// WithValue adds a key-value pair to the AppContext.
func (s *ServiceContext) WithValue(key, val any) *ServiceContext {
	return &ServiceContext{
		DefaultContext: &DefaultContext{Context: context.WithValue(s.DefaultContext, key, val)},

		// These are unaffected fields
		AppContext: s.AppContext,
		Context:    s.Context,
	}
}

// WithRequestID adds a request ID to the AppContext.
func (s *ServiceContext) WithRequestID(requestID string) *ServiceContext {
	return s.WithValue(constant.RequestID, requestID)
}

// GetAppContextRequestID retrieves the request ID from the AppContext.
func (s *ServiceContext) GetAppContextRequestID() (types.RequestID, bool) {
	requestID, ok := s.Value(constant.RequestID).(string)
	return types.RequestID(requestID), ok
}

// GetAppContextCorrelationID retrieves the correlation ID from the AppContext.
func (s *ServiceContext) GetAppContextCorrelationID() (types.CorrelationID, bool) {
	correlationId, ok := s.Value(constant.CorrelationID).(string)
	return types.CorrelationID(correlationId), ok
}

// WithLogger adds a logger to the AppContext.
func (s *ServiceContext) WithLogger(logger *log.Log) *ServiceContext {
	return s.WithValue(constant.Logger, logger)
}

// GetLogger retrieves the logger from the AppContext.
func (s *ServiceContext) GetLogger() (*log.Log, bool) {
	logger, ok := s.Value(constant.Logger).(*log.Log)
	return logger, ok
}

// WithTraceID adds a trace ID to the AppContext for distributed tracing.
func (s *ServiceContext) WithTraceID(traceID string) *ServiceContext {
	return s.WithValue(constant.TraceID, traceID)
}

// GetTraceID retrieves the trace ID from the AppContext.
func (s *ServiceContext) GetTraceID() (string, bool) {
	traceID, ok := s.Value(constant.TraceID).(string)
	return traceID, ok
}

// WithMetadata adds metadata (as a map) to the AppContext.
func (s *ServiceContext) WithMetadata(metadata map[string]string) *ServiceContext {
	return s.WithValue(constant.MetaData, metadata)
}

// GetMetadata retrieves metadata from the AppContext.
func (s *ServiceContext) GetMetadata() (map[string]string, bool) {
	metadata, ok := s.Value(constant.MetaData).(map[string]string)
	return metadata, ok
}

// GetServiceID retrieves the service ID from the AppContext.
func (s *ServiceContext) GetServiceID() string {
	return s.serviceId
}

// WithGeneratedRequestID adds a generated request ID to the AppContext.
func (s *ServiceContext) WithGeneratedRequestID() *ServiceContext {
	return s.WithValue(constant.RequestID, random.GenerateUUID())
}

// WithGeneratedCorrelationID adds a generated correlation ID to the AppContext.
func (s *ServiceContext) WithGeneratedCorrelationID() *ServiceContext {
	return s.WithValue(constant.CorrelationID, random.GenerateUUID())
}

// RecoverFromException recovers from panics and logs the stack trace
func (s *ServiceContext) RecoverFromException() {
	if r := recover(); r != nil {
		stack := debug.Stack()
		errorMsg := fmt.Sprintf("Panic recovered: %v\nStack Trace:\n%s", r, string(stack))

		// Use the service logger if available, otherwise use standard log
		if s.Log != nil {
			s.Log.Error(errorMsg)
		} else {
			helpers.Println(constant.ERROR, errorMsg)
		}
	}
}

func (s *ServiceContext) RunSafely(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			msg := fmt.Sprintf("Panic recovered: %v\nStack Trace:\n%s", r, string(stack))
			// Use the service logger if available, otherwise use standard log
			if s.Log != nil {
				s.Log.Error(msg)
			} else {
				helpers.Println(constant.INFO, msg)
			}

		}
	}()
	fn()
}

// GetNATSManager retrieves the NATSManager from the App context.
func (ctx *ServiceContext) GetNATSManager() *nats.NATSManager {
	return ctx.NATSManager
}

// GetGinCtxServiceName retrieves the service name from the gin.Context.
func (ctx *ServiceContext) GetGinCtxServiceName() (*types.Service, error) {

	serviceName, exists := ctx.Get(constant.Service)
	if !exists {
		return nil, errors.New("service name not found in gin.Context")
	}

	return serviceName.(*types.Service), nil
}

// DefaultContextWithTimeout creates a new default context with a timeout.
func DefaultContextWithTimeout(timeout time.Duration) (DefaultContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return DefaultContext{
		Context: ctx,
	}, cancel
}

func (ctx *ServiceContext) GetGinCtxRecordsName() (*string, error) {

	records, exists := ctx.Get(constant.Records)
	if !exists {
		return nil, errors.New("records name not found in gin.Context")
	}

	return records.(*string), nil
}
