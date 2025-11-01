package engine

import (
	"errors"

	natsInternal "github.com/abhissng/neuron/adapters/events/nats"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/message"
	"github.com/abhissng/neuron/utils/types"
	"github.com/nats-io/nats.go"
)

// ServiceHandler defines the signature for a service handler.
type ServiceHandler[T any] func(*context.ServiceContext, *nats.Msg) result.Result[T]

// WrapServiceWithNATSHandler wraps a handler logic with a nats.MsgHandler (func (msg *nats.Msg))
func WrapServiceWithNATSHandler[T any](ctx *context.ServiceContext, handler ServiceHandler[T]) nats.MsgHandler {
	return func(msg *nats.Msg) {
		defer helpers.RecoverException(recover())
		blameInfo := blame.NilBlame()

		// Add standard middleware headers
		middlewareFunc := []natsInternal.MiddlewareFunc{
			natsInternal.AddHeaderMiddleware(constant.IPHeader, helpers.IPHeaderFromNatsMsg(msg)),
			natsInternal.AddHeaderMiddleware(constant.CorrelationIDHeader, helpers.CorrelationIDFromNatsMsg(msg)),
		}
		var response any
		defer func() {

			if blameInfo != nil {
				response = encodeErrorRespondMesage[any](ctx, constant.Execute, msg, blameInfo)
			}
			ctx.Info(constant.ServiceHandlerMessage, ctx.SlogEvent(msg, log.String("handlers", "WrapServiceWithNATSHandler"), log.Any("response", response))...)
			// Use the NATS connection to send a response
			_, err := ctx.PublishWithMiddleware(
				msg.Reply, response,
				middlewareFunc...,
			)
			if err != nil {
				ctx.Log.Error(constant.ServiceHandlerMessage, ctx.SlogEvent(msg, log.String("function", "ctx.NATSManager.PublishWithMiddleware"), log.Err(err))...)
				return
			}
		}()

		// Execute your handler with the received message
		responseResult := handler(ctx, msg)
		if !responseResult.IsSuccess() {
			_, blameInfo = responseResult.Value()
			ctx.Log.Error(constant.ServiceHandlerMessage, ctx.SlogEvent(msg, log.String("handlers", "WrapServiceWithNATSHandler"), log.String("error", blameInfo.FetchErrCode().String()))...)
			middlewareFunc = append(middlewareFunc, natsInternal.AddHeaderMiddleware(constant.ErrorHeader, blameInfo.FetchErrCode().String()))
			return
		}

		res := responseResult.ToValue()
		response = *res
	}
}

// WrapServiceWithNatsProcessor wraps a handler logic with a natsInternal.NATSMsgProcessor
func WrapServiceWithNatsProcessor[T any](ctx *context.ServiceContext, handler ServiceHandler[T]) natsInternal.NATSMsgProcessor {
	defer helpers.RecoverException(recover())
	wrappedHandler := WrapServiceWithNATSHandler(ctx, handler)
	return natsInternal.NATSMsgProcessor(func(msg *nats.Msg) blame.Blame {
		defer helpers.RecoverException(recover())
		wrappedHandler(msg)
		return nil
	})
}

// GetPayload is a helper function to convert a response map to a Message
func GetPayload[T any](response map[string]any) result.Result[message.Message[T]] {
	// Convert the response to a Message
	msg, ok := types.CastTo[*message.Message[T]](response[Payload])
	if !ok {
		return result.NewFailure[message.Message[T]](
			blame.TypeConversionError(Payload, "", "*message.Message[T]", errors.New("unable to convert response to Message")))
	}

	return result.NewSuccess(msg)
}
