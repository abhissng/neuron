package nats

import (
	"errors"
	"runtime/debug"

	"github.com/abhissng/core-hub/core"
	"github.com/abhissng/neuron/adapters/jwt"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/message"
	"github.com/abhissng/neuron/utils/types"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

// ----------------------
// Middleware support
// ----------------------

// NATSMsgProcessor defines the signature for a message processor.
type NATSMsgProcessor func(msg *nats.Msg) blame.Blame

// MiddlewareFunc defines the signature for a middleware function.
type MiddlewareFunc func(NATSMsgProcessor) NATSMsgProcessor

// applyMiddleware applies the middleware chain to a processor.
func applyMiddleware(processor NATSMsgProcessor, middlewares ...MiddlewareFunc) NATSMsgProcessor {
	defer helpers.RecoverException(recover())
	// Apply in reverse order so that the first middleware in the list is executed first.
	for i := len(middlewares) - 1; i >= 0; i-- {
		processor = middlewares[i](processor)
	}
	return processor
}

// AddHeaderMiddleware returns a middleware that sets a header key/value on the message.
func AddHeaderMiddleware(key, value string) MiddlewareFunc {
	return func(next NATSMsgProcessor) NATSMsgProcessor {
		return func(msg *nats.Msg) blame.Blame {
			msg.Header.Set(key, value)
			return next(msg)
		}
	}
}

// LogMiddleware returns a middleware that logs the publishing event.
func LogMiddleware(eventType string, logger *log.Log) MiddlewareFunc {
	return func(next NATSMsgProcessor) NATSMsgProcessor {
		return func(msg *nats.Msg) blame.Blame {
			defer helpers.RecoverException(recover())
			logger.Info(eventType, log.Any(constant.EventReceived, msg))
			err := next(msg)
			if err != nil {
				if logger == nil {
					helpers.Println(constant.ERROR, eventType+" failed", log.Err(err))
				} else {
					logger.Error(eventType+" failed", log.Err(err))
				}
			}
			return err
		}
	}
}

// Convert nats.MsgHandler to NATSMsgProcessor
func (w *NATSManager) WrapNATSMsgProcessor(handler nats.MsgHandler) NATSMsgProcessor {
	return func(msg *nats.Msg) blame.Blame {
		defer helpers.RecoverException(recover())
		RecoveryMiddleware(handler)(msg)
		// handler(msg) // Call the original handler
		return nil
	}
}

// sendErrorResponse sends an error response message back through NATS
func sendErrorResponse(msg *nats.Msg, err error) {
	message := message.NewMessage(
		constant.Execute,
		constant.Failed,
		types.CorrelationID(helpers.CorrelationIDFromNatsMsg(msg)),
		core.Core{},
	)
	message.Error = blame.HeadersNotFound(err).FetchErrorResponse(blame.WithTranslation())

	msgByt, encodeErr := codec.Encode(message, codec.JSON)
	if encodeErr != nil {
		return
	}
	_ = msg.Respond(msgByt)
}

// validateAuthToken validates the authorization token from headers
func validateAuthToken(msg *nats.Msg) blame.Blame {
	token := helpers.AuthorizationHeaderFromNatsMsg(msg)
	if helpers.IsEmpty(token) {
		return blame.MalformedAuthToken(errors.New("token is empty"))
	}

	_, err := jwt.ValidateJWT(token, viper.GetString(constant.JWTSecret), []string{"admin"})
	if err != nil {
		return blame.MalformedAuthToken(err)
	}
	return nil
}

// ValidateHeadersMiddleware checks for the existence and validity of required headers.
func ValidateHeadersMiddleware(next NATSMsgProcessor) MiddlewareFunc {
	defer helpers.RecoverException(recover())
	return func(next NATSMsgProcessor) NATSMsgProcessor {
		return func(msg *nats.Msg) blame.Blame {
			if msg.Header == nil {
				err := errors.New("missing headers")
				sendErrorResponse(msg, err)
				return blame.HeadersNotFound(err)
			}

			if blameErr := validateAuthToken(msg); blameErr != nil {
				sendErrorResponse(msg, blameErr.ErrorFromBlame())
				return blameErr
			}

			return next(msg)
		}
	}
}

// RecoveryMiddleware wraps the NATS message handler to recover from panics.
func RecoveryMiddleware(handler nats.MsgHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		defer func() {
			if r := recover(); r != nil {
				helpers.Println(constant.ERROR, "Recovered from panic in NATS message handler")
				helpers.Println(constant.ERROR, string(debug.Stack()))
			}
		}()
		handler(msg) // Call the actual handler
	}
}
