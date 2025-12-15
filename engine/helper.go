package engine

import (
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/context"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/structures/message"
	"github.com/abhissng/neuron/utils/types"
	"github.com/nats-io/nats.go"
	"github.com/spf13/viper"
)

// encodeErrorRespondMesage encodes a error response map to a message
func encodeErrorRespondMesage[T any](ctx *context.ServiceContext, action types.Action, msg *nats.Msg, cause blame.Blame) T {
	var zero T
	coreMessage := message.NewMessage(
		action,
		constant.Failed,
		types.CorrelationID(helpers.CorrelationIDFromNatsMsg(msg)),
		zero,
	)
	coreMessage.AddError(cause.FetchErrorResponse(blame.WithTranslation()))
	// Ensure coreMessage is of the correct type
	if result, ok := any(coreMessage).(T); ok {
		return result
	}

	ctx.Log.Error(constant.ServiceHandlerMessage, log.Any("helpers", "encodeErrorRespondMesage"), log.Any("type", "TypeCast error"))

	return zero
}

// FetchJWTSecret returns the JWT secret
func FetchJWTSecret(ctx *context.ServiceContext) string {
	if ctx != nil && ctx.Vault != nil {
		secret, _ := ctx.FetchVaultValue(constant.JWTSecret)
		if secret == "" {
			secret = "default-secret"
		}
		return secret
	}

	return viper.GetString(constant.JWTSecret)
}
