package message

import (
	"time"

	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/utils/random"
	"github.com/abhissng/neuron/utils/types"
)

// Message represents the structure of a transaction message.
// or you can use your custom Type if needed
type Message[T any] struct {
	CorrelationID  types.CorrelationID `json:"correlation_id"`
	RequestId      types.RequestID     `json:"request_id"`
	Payload        T                   `json:"payload"`
	Status         types.Status        `json:"status"` // "pending", "completed or success", "failed"
	Action         types.Action        `json:"action"` // "execute", "rollback"
	Error          blame.ErrorResponse `json:"error,omitempty"`
	Timestamp      time.Time           `json:"timestamp"`
	CurrentService string              `json:"current_service,omitempty"`
}

// NewNilMessage creates a new nil Message
func NewNilMessage[T any]() *Message[T] {
	return &Message[T]{}
}

// NewMessage creates a new Message
func NewMessage[T any](
	action types.Action,
	status types.Status,
	correlationID types.CorrelationID,
	payload T,
) *Message[T] {
	return &Message[T]{
		CorrelationID: correlationID,
		RequestId:     types.RequestID(random.GenerateUUID()),
		Payload:       payload,
		Status:        status,
		Action:        action,
		Timestamp:     time.Now(),
	}
}

// AddError adds an error to the message
func (m *Message[T]) AddError(error blame.ErrorResponse) *Message[T] {
	m.Error = error
	return m
}
