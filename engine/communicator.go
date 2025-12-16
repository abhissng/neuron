package engine

import (
	"errors"

	"github.com/abhissng/neuron/adapters/http"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/result"
	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/structures/acknowledgment"
	"github.com/abhissng/neuron/utils/structures/discovery"
)

type CommunicateResult[T any] struct {
	DiscoveryResult *discovery.DiscoveryMessagePayload[T]
	ErrorResult     blame.ErrorResponse
	Err             error
}

func NewCommunicateResult[T any]() CommunicateResult[T] {
	return CommunicateResult[T]{
		DiscoveryResult: nil,
		ErrorResult:     blame.ErrorResponse{},
		Err:             nil,
	}
}

// CommunicateWithDiscovery handles communication with the discovery service.
func CommunicateWithDiscovery[T any](ctx *http.HttpClientManager, payload *discovery.DiscoveryMessagePayload[T]) result.Result[CommunicateResult[T]] {

	communicateResult := NewCommunicateResult[T]()

	// Make the request
	res := http.DoRequest[acknowledgment.APIResponse[any]](payload, ctx)

	// If request fails, return an error
	if !res.IsSuccess() {
		_, cause := res.Value()
		ctx.Log.Error(constant.APICallMessage, log.Any("error", cause.FetchCauses()))
		return result.NewFailure[CommunicateResult[T]](cause)
	}

	if res.ToValue().Result == nil {
		ctx.Log.Error(constant.APICallMessage, log.Any("error", "discovery response result is nil"))
		return result.NewFailure[CommunicateResult[T]](blame.ResponseResultError(errors.New("discovery response result is nil")))
	}

	response, err := codec.Encode(res.ToValue().Result, codec.JSON)
	if err != nil {
		ctx.Log.Error(constant.AdaptersMessage, log.Err(err))
		return result.NewFailure[CommunicateResult[T]](blame.UnMarshalError(codec.JSON, err))
	}

	ctx.Log.Info(constant.CommunicatorMessage, log.Any("response", string(response)))

	if resMap, err := codec.Decode[blame.ErrorResponse](response, codec.JSON); err == nil {
		ctx.Log.Info(constant.AdaptersMessage, log.Any("message", "succesfully decoded to *blame.ErrorResponse"))
		errMsg := resMap.Message
		if len(resMap.Causes) > 0 {
			errMsg = resMap.Causes[0]
		}
		communicateResult.Err = errors.New(errMsg)
		communicateResult.ErrorResult = resMap
		return result.NewSuccess(&communicateResult)
	}

	if resMap, err := codec.Decode[*discovery.DiscoveryMessagePayload[T]](response, codec.JSON); err == nil {
		ctx.Log.Info(constant.AdaptersMessage, log.Any("message", "succesfully decoded to *discovery.DiscoveryMessagePayload"))
		communicateResult.DiscoveryResult = resMap
		return result.NewSuccess(&communicateResult)
	}

	// If casting fails, return an error
	ctx.Log.Error(constant.CommunicatorMessage, log.Any("error", "unable to determine communicate discovery response type"))
	return result.NewFailure[CommunicateResult[T]](blame.ResponseResultError(errors.New("unable to determine communicate discovery response type")))
}
