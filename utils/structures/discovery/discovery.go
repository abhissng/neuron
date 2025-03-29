package discovery

import (
	"github.com/abhissng/core-hub/core"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/structures"
	"github.com/abhissng/neuron/utils/structures/message"
	"github.com/abhissng/neuron/utils/types"
)

// DiscoveryMessagepayload represents the structure of a discovery message request.
type DiscoveryMessagePayload struct {
	// TODO add additionalInformation Later also add binding to metadata
	// MetaData *structures.MetaData `json:"meta_data" binding:"required"`
	Message  *message.Message[*core.Core] `json:"message" binding:"required"`
	MetaData *structures.MetaData         `json:"meta_data"`
}

// NewDiscoveryMessagePayload creates a new DiscoveryMessagePayload
// with the given correlation ID and ("github.com/abhissng/core-hub/core") Core payload
// with action set as execute and status set as pending
func NewDiscoveryMessagePayload(
	correlationId types.CorrelationID,
	core *core.Core,
) *DiscoveryMessagePayload {
	return &DiscoveryMessagePayload{
		Message:  message.NewMessage(constant.Execute, constant.Pending, correlationId, core),
		MetaData: structures.NewMetaData(),
	}
}

// AddMetaData adds metadata to the DiscoveryMessage
func (d *DiscoveryMessagePayload) AddMetaData(metaData *structures.MetaData) *DiscoveryMessagePayload {
	d.MetaData = metaData
	return d
}

// AddAction adds action to the DiscoveryMessage
func (d *DiscoveryMessagePayload) AddAction(action types.Action) *DiscoveryMessagePayload {
	d.Message.Action = action
	return d
}

// AddStatus adds status to the DiscoveryMessage
func (d *DiscoveryMessagePayload) AddStatus(status types.Status) *DiscoveryMessagePayload {
	d.Message.Status = status
	return d
}
