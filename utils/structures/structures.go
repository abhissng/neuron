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

type ExcludedOptions struct {
	Services []*string `json:"services"`
	Records  []*string `json:"records"`
	Events   []*string `json:"events"`
}

func NewExcludedOptions() *ExcludedOptions {
	return &ExcludedOptions{
		Services: make([]*string, 0),
		Records:  make([]*string, 0),
		Events:   make([]*string, 0),
	}
}

// HasExcludedService returns true if the list of excluded services is not empty.
func (e *ExcludedOptions) HasExcludedService() bool {
	return len(e.Services) > 0
}

// HasExcludedRecords returns true if the list of excluded records is not empty.
func (e *ExcludedOptions) HasExcludedRecords() bool {
	return len(e.Records) > 0
}

// HasExcludedEvent returns true if the list of excluded events is not empty.
func (e *ExcludedOptions) HasExcludedEvent() bool {
	return len(e.Events) > 0
}

// ExcludedServices returns the list of excluded services.
func (e *ExcludedOptions) ExcludedServices() []*string {
	return e.Services
}

// ExcludedRecords returns the list of excluded records.
func (e *ExcludedOptions) ExcludedRecords() []*string {
	return e.Records
}

// ExcludedEvents returns the list of excluded events.
func (e *ExcludedOptions) ExcludedEvents() []*string {
	return e.Events
}
