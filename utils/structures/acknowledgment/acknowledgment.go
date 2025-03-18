package acknowledgment

import (
	"github.com/abhissng/neuron/utils/types"
)

// APIResponse structure for final response to REST clients will change later on
type APIResponse[T any] struct {
	Success       bool                `json:"success"`
	CorrelationID types.CorrelationID `json:"correlation_id"`
	Result        T                   `json:"result"`
	// Error         string              `json:"error,omitempty"`
	// ProcessingTime time.Duration       `json:"processing_time"`
	// CompletedSteps []string            `json:"completed_steps,omitempty"`
	// RollbackSteps  []string            `json:"rollback_steps,omitempty"`
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

/*
// Message represents the structure of a transaction message.
// uses core.Core from "github.com/abhissng/core-structures/core" for payload Type
type Message[T any] struct {
	CorrelationID  types.CorrelationID `json:"correlation_id"`
	RequestId      types.RequestID     `json:"request_id"`
	Payload        T                   `json:"payload"`
	Status         types.Status        `json:"status"` // "pending", "completed or success", "failed"
	Action         types.Action        `json:"action"` // "execute", "rollback"
	Error          error               `json:"error,omitempty"`
	Timestamp      time.Time           `json:"timestamp"`
	CurrentService string              `json:"current_service,omitempty"`
}

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

// DiscoveryMessageRequest represents the structure of a discovery message request.
type DiscoveryMessageRequest struct {
	// TODO add additionalInformation Later also add binding to metadata
	// MetaData *structures.MetaData `json:"meta_data" binding:"required"`
	Message  *Message[*core.Core] `json:"message" binding:"required"`
	MetaData *structures.MetaData `json:"meta_data"`
}

// NewDiscoveryMessageRequest creates a new DiscoveryMessageRequest with the given correlation ID and core.Core payload
// with action set as execute and status set as pending
func NewDiscoveryMessageRequest(
	correlationId types.CorrelationID,
	core *core.Core,
) *DiscoveryMessageRequest {
	return &DiscoveryMessageRequest{
		Message:  NewMessage(constant.Execute, constant.Pending, correlationId, core),
		MetaData: structures.NewMetaData(),
	}
}

// AddMetaData adds metadata to the DiscoveryMessage
func (d *DiscoveryMessageRequest) AddMetaData(metaData *structures.MetaData) *DiscoveryMessageRequest {
	d.MetaData = metaData
	return d
}

// DiscoveryResponse represents the structure of a discovery response.
type DiscoveryMessageResponse struct {
	// TODO add additionalInformation Later also add binding to metadata
	// MetaData *structures.MetaData `json:"meta_data" binding:"required"`
	Message  *Message[*core.Core] `json:"message"`
	MetaData *structures.MetaData `json:"meta_data"`
}

// NewDiscoveryMessageResponse creates a new DiscoveryMessageResponse
func NewDiscoveryMessageResponse(
	metaData *structures.MetaData,
	action types.Action,
	status types.Status,
	correlationId types.CorrelationID,
	core *core.Core) *DiscoveryMessageResponse {
	return &DiscoveryMessageResponse{
		Message:  NewMessage(action, status, correlationId, core),
		MetaData: metaData.UpdateMetaData(),
	}
}
*/
