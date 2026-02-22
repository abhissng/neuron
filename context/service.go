package context

import (
	natsInternal "github.com/abhissng/neuron/adapters/events/nats"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/types"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// ServiceContext embeds AppContext and Gin Context and default Context of Go
type ServiceContext struct {
	*DefaultContext
	*AppContext
	*gin.Context
}

// ServiceContextOption is a function that can be used to configure a ServiceContext.
type ServiceContextOption func(*ServiceContext)

// WithAppContext sets the AppContext for the ServiceContext.
func WithAppContext(appCtx *AppContext) ServiceContextOption {
	return func(sc *ServiceContext) {
		sc.AppContext = appCtx
	}
}

// WithGinContext sets the Gin Context for the ServiceContext.
func WithGinContext(ginCtx *gin.Context) ServiceContextOption {
	return func(sc *ServiceContext) {
		sc.Context = ginCtx
	}
}

// NewServiceContext initializes a ServiceContext with the provided options.
func NewServiceContext(opts ...ServiceContextOption) *ServiceContext {
	// Create a default ServiceContext
	sc := &ServiceContext{
		DefaultContext: NewDefaultContext(),
		AppContext:     &AppContext{},
		Context:        &gin.Context{},
	}

	// Apply all options
	for _, opt := range opts {
		opt(sc)
	}

	return sc
}

// FetchGinRequestSlogFields fetched a requestId and correlationid as slice of fields
func (ctx *ServiceContext) FetchGinRequestSlogFields() []types.Field {
	fields := make([]types.Field, 2)
	fields[0] = log.String(constant.RequestID, ctx.GetGinContextRequestID().String())
	fields[1] = log.String(constant.CorrelationIDHeader, ctx.GetGinContextCorrelationID().String())
	if ctx.GetCookieSessionID() != "" {
		fields = append(fields, log.String(constant.SessionID, ctx.GetCookieSessionID()))
	}
	return fields
}

// GetGinContextRequestID returns the request ID from the Gin context.
func (ctx *ServiceContext) GetGinContextRequestID() types.RequestID {
	if ctx.Context != nil {
		return types.RequestID(ctx.GetString(constant.RequestID))
	}
	return ""
}

// GetGinContextCorrelationID returns the correlation ID from the Gin context.
func (ctx *ServiceContext) GetGinContextCorrelationID() types.CorrelationID {
	if ctx.Context != nil {
		return types.CorrelationID(ctx.GetString(constant.CorrelationID))
	}
	return ""
}

// SlogFields returns a slice of types.Field with request and correlation fields and additional fields.
func (ctx *ServiceContext) SlogFields(withFields ...types.Field) []types.Field {
	// Start with the request and correlation fields
	fields := make([]types.Field, 0, 2+len(withFields))
	fields = append(fields, ctx.FetchGinRequestSlogFields()...)

	// Append additional fields provided as variadic arguments
	if len(withFields) > 0 {
		fields = append(fields, withFields...)
	}
	return fields
}

// SlogInfo logs a message at the InfoLevel.
func (ctx *ServiceContext) SlogInfo(message string, withFields ...types.Field) {
	// Start with the request and correlation fields and additional fields
	slogfields := ctx.SlogFields(withFields...)
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Info(message, slogfields...)
}

func (ctx *ServiceContext) SlogWarn(message string, withFields ...types.Field) {
	// Start with the request and correlation fields and additional fields
	slogfields := ctx.SlogFields(withFields...)
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Warn(message, slogfields...)
}

func (ctx *ServiceContext) SlogError(message string, withFields ...types.Field) {
	// Start with the request and correlation fields and additional fields
	slogfields := ctx.SlogFields(withFields...)
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Error(message, slogfields...)
}

func (ctx *ServiceContext) SlogFatal(message string, withFields ...types.Field) {
	// Start with the request and correlation fields and additional fields
	slogfields := ctx.SlogFields(withFields...)
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Fatal(message, slogfields...)
}

func (ctx *ServiceContext) SlogDebug(message string, withFields ...types.Field) {
	// Start with the request and correlation fields and additional fields
	slogfields := ctx.SlogFields(withFields...)
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Debug(message, slogfields...)
}

// SlogEvent returns a slice of types.Field with message and correlation fields and additional fields.
func (ctx *ServiceContext) SlogEvent(msg *nats.Msg, withFields ...types.Field) []types.Field {
	if msg == nil {
		return withFields
	}
	return natsInternal.Slog(msg, withFields...)
}

// SlogEventInfo logs a message at the InfoLevel with NATS event metadata.
func (ctx *ServiceContext) SlogEventInfo(logMessage string, natsMsg *nats.Msg, withFields ...types.Field) {
	var fields []types.Field
	if natsMsg == nil {
		fields = ctx.SlogFields(withFields...)
	} else {
		fields = ctx.SlogEvent(natsMsg, withFields...)
	}
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Info(logMessage, fields...)
}

// SlogEventWarn logs a message at the WarnLevel with NATS event metadata.
func (ctx *ServiceContext) SlogEventWarn(logMessage string, natsMsg *nats.Msg, withFields ...types.Field) {
	var fields []types.Field
	if natsMsg == nil {
		fields = ctx.SlogFields(withFields...)
	} else {
		fields = ctx.SlogEvent(natsMsg, withFields...)
	}
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Warn(logMessage, fields...)
}

// SlogEventError logs a message at the ErrorLevel with NATS event metadata.
func (ctx *ServiceContext) SlogEventError(logMessage string, natsMsg *nats.Msg, withFields ...types.Field) {
	var fields []types.Field
	if natsMsg == nil {
		fields = ctx.SlogFields(withFields...)
	} else {
		fields = ctx.SlogEvent(natsMsg, withFields...)
	}
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Error(logMessage, fields...)
}

// SlogEventDebug logs a message at the DebugLevel with NATS event metadata.
func (ctx *ServiceContext) SlogEventDebug(logMessage string, natsMsg *nats.Msg, withFields ...types.Field) {
	var fields []types.Field
	if natsMsg == nil {
		fields = ctx.SlogFields(withFields...)
	} else {
		fields = ctx.SlogEvent(natsMsg, withFields...)
	}
	logger := ctx.WithOptions(zap.AddCaller())
	logger.Debug(logMessage, fields...)
}

// GetRequestID returns the request ID from the AppContext or the Gin context.
func (ctx *ServiceContext) GetRequestID() types.RequestID {
	if requestId, ok := ctx.GetAppContextRequestID(); ok {
		return requestId
	}
	return ctx.GetGinContextRequestID()
}

// GetCorrelationID returns the correlation ID from the AppContext or the Gin context.
func (ctx *ServiceContext) GetCorrelationID() types.CorrelationID {
	if correlationID, ok := ctx.GetAppContextCorrelationID(); ok {
		return correlationID
	}
	return ctx.GetGinContextCorrelationID()
}

// GetCookieSessionID returns the session ID from the cookie.
func (ctx *ServiceContext) GetCookieSessionID() string {
	if sessionID, ok := ctx.Cookie(constant.SessionID); ok == nil {
		return sessionID
	}
	return ""
}
