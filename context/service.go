package context

import (
	natsInternal "github.com/abhissng/neuron/adapters/events/nats"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/types"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
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

// FetchGinRequestAndCorrelationField fetched a requestId and correlationid as slice of fields
func (ctx *ServiceContext) FetchGinRequestAndCorrelationField() []types.Field {
	fields := make([]types.Field, 2)
	fields[0] = log.String(constant.RequestID, ctx.GetGinContextRequestID().String())
	fields[1] = log.String(constant.CorrelationID, ctx.GetGinContextCorrelationID().String())
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

// Slog returns a slice of types.Field with request and correlation fields and additional fields.
func (ctx *ServiceContext) Slog(withFields ...types.Field) []types.Field {
	// Start with the request and correlation fields
	fields := make([]types.Field, 0, 2+len(withFields))
	fields = append(fields, ctx.FetchGinRequestAndCorrelationField()...)

	// Append additional fields provided as variadic arguments
	fields = append(fields, withFields...)
	return fields
}

// SlogEvent returns a slice of types.Field with message and correlation fields and additional fields.
func (ctx *ServiceContext) SlogEvent(msg *nats.Msg, withFields ...types.Field) []types.Field {
	return natsInternal.Slog(msg, withFields...)
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
