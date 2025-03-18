package structures

import (
	"fmt"
	"time"
)

// MetaData represents the metadata of a service
type MetaData struct {
	// Unique identifier for the metadata
	ID string `json:"id"`

	// Timestamp of metadata creation
	CreatedAt time.Time `json:"created_at,omitempty"`

	// Timestamp of the last update to the metadata
	UpdatedAt time.Time `json:"updated_at,omitempty"`

	// Optional: User who created or last updated the metadata
	CreatedBy string `json:"created_by,omitempty"`

	// Optional: User who created or last updated the metadata
	UpdatedBy string `json:"updated_by,omitempty"`

	// Optional: Version of the metadata
	Version string `json:"version,omitempty"`

	// App version
	AppVersion string `json:"app_version"`

	// Channel name
	ChannelName string `json:"channel_name"`

	// IMEI (International Mobile Equipment Identity)
	IMEI string `json:"imei,omitempty"`

	// Operating System (e.g., "Android", "iOS")
	OS string `json:"os"`

	// Optional: Additional metadata-specific fields
	// Example:
	// Name string `json:"name"`
	// Description string `json:"description"`
	// ... other custom fields
}

// NewMetaData creates a new MetaData
func NewMetaData() *MetaData {
	return &MetaData{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// UpdateMetaData updates the MetaData
func (m *MetaData) UpdateMetaData() *MetaData {
	m.UpdatedAt = time.Now()
	return m
}

// String returns a string representation of the MetaData
func (m *MetaData) String() string {
	return fmt.Sprintf("ID: %s, CreatedAt: %s, UpdatedAt: %s", m.ID, m.CreatedAt, m.UpdatedAt)
}

// ValidateMetaData validates the MetaData
func (m *MetaData) ValidateMetaData() error {
	// Implement validation logic here
	return nil
}
