package acknowledgment

import (
	"github.com/abhissng/neuron/utils/types"
)

// APIResponse structure for final response to REST clients will change later on
type APIResponse[T any] struct {
	Success       bool                `json:"success"`
	CorrelationID types.CorrelationID `json:"correlation_id"`
	Result        T                   `json:"result"`
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
