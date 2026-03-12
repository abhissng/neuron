package acknowledgment

import (
	"github.com/abhissng/neuron/utils/types"
)

// APIResponse structure for final response to REST clients will change later on
type APIResponse[T any] struct {
	Success       bool                `json:"success"`
	CorrelationID types.CorrelationID `json:"correlation_id"`
	Result        T                   `json:"result"`
	Error         *string             `json:"error"`
}

func NewAPIResponse[T any](
	success bool,
	correlationID types.CorrelationID,
	result T,
) APIResponse[T] {
	return APIResponse[T]{
		Success:       success,
		CorrelationID: correlationID,
		Result:        result,
	}
}

// NewNilAPIResponse creates a new APIResponse with nil result.
func NewNilAPIResponse[T any]() APIResponse[T] {
	return APIResponse[T]{}
}

// ToValue converts the APIResponse to a value.
func (resp *APIResponse[T]) ToValue() T {
	if resp == nil {
		var zero T
		return zero
	}
	return resp.Result
}

// CastToResult casts the APIResponse to a value.
func CastToResult[R any](resp *APIResponse[any]) (R, bool) {
	if resp == nil {
		var zero R
		return zero, false
	}

	if resp.Result == nil {
		var zero R
		return zero, false
	}

	result, ok := resp.Result.(R)
	if !ok {
		var zero R
		return zero, false
	}
	return result, true
}

// WithError sets the error in the APIResponse.
func (resp *APIResponse[T]) WithError(err error) *APIResponse[T] {
	if err == nil {
		resp.Error = nil
		return resp
	}
	errorMsg := err.Error()
	resp.Error = &errorMsg
	return resp
}
