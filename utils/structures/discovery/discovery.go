package discovery

import (
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/structures"
	"github.com/abhissng/neuron/utils/structures/message"
	"github.com/abhissng/neuron/utils/types"
)

// DiscoveryMessagepayload represents the structure of a discovery message request.
type DiscoveryMessagePayload[T any] struct {
	// TODO add additionalInformation Later also add binding to metadata
	// MetaData *structures.MetaData `json:"meta_data" binding:"required"`
	// Message  *message.Message[*core.Core] `json:"message" binding:"required"`
	Message  *message.Message[T]  `json:"message" binding:"required"`
	MetaData *structures.MetaData `json:"meta_data"`
}

// NewDiscoveryMessagePayload creates a new DiscoveryMessagePayload
// with action set as execute and status set as pending
func NewDiscoveryMessagePayload[T any](
	correlationId types.CorrelationID,
	core T,
) *DiscoveryMessagePayload[T] {
	return &DiscoveryMessagePayload[T]{
		Message:  message.NewMessage(constant.Execute, constant.Pending, correlationId, core),
		MetaData: structures.NewMetaData(),
	}
}

// AddMetaData adds metadata to the DiscoveryMessage
func (d *DiscoveryMessagePayload[T]) AddMetaData(metaData *structures.MetaData) *DiscoveryMessagePayload[T] {
	d.MetaData = metaData
	return d
}

// AddAction adds action to the DiscoveryMessage
func (d *DiscoveryMessagePayload[T]) AddAction(action types.Action) *DiscoveryMessagePayload[T] {
	d.Message.Action = action
	return d
}

// AddStatus adds status to the DiscoveryMessage
func (d *DiscoveryMessagePayload[T]) AddStatus(status types.Status) *DiscoveryMessagePayload[T] {
	d.Message.Status = status
	return d
}
