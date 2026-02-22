package structures

import (
	"fmt"
	"time"

	"github.com/abhissng/neuron/utils/types"
	"github.com/google/uuid"
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
	result := make([]*string, len(e.Services))
	copy(result, e.Services)
	return result
}

// ExcludedRecords returns the list of excluded records.
func (e *ExcludedOptions) ExcludedRecords() []*string {
	result := make([]*string, len(e.Records))
	copy(result, e.Records)
	return result
}

// ExcludedEvents returns the list of excluded events.
func (e *ExcludedOptions) ExcludedEvents() []*string {
	result := make([]*string, len(e.Events))
	copy(result, e.Events)
	return result
}

// PhoneNumberInfo holds the parsed phone number details.
type PhoneNumberInfo struct {
	E164Format     string // The standardized international format (e.g., +919876543210)
	CountryCode    int32  // The country code (e.g., 91)
	RegionCode     string // The two-letter (ISO 3166-1) region code (e.g., "IN", "US")
	CountryName    string // The full country name (e.g., "India", "United States of America")
	IsValid        bool   // Whether the library considers this a valid number
	NationalNumber uint64 // The number without the country code
}

type EssentialHeaders struct {
	OrgId        types.OrgID  `json:"org_id"`
	UserId       types.UserID `json:"user_id"`
	LocationId   uuid.UUID    `json:"location_id"`
	UserRole     string       `json:"user_role"`
	FeatureFlags string       `json:"feature_flags"`
}

// NewEssentialHeaders creates and returns an empty EssentialHeaders value.
func NewEssentialHeaders() *EssentialHeaders {
	return &EssentialHeaders{}
}

// config for behaviour
type EssentialHeadersConfig struct {
	RequireFeatureFlags bool
	RequireLocationID   bool
}

func NewEssentialHeadersConfig() *EssentialHeadersConfig {
	return &EssentialHeadersConfig{
		RequireFeatureFlags: false,
		RequireLocationID:   false,
	}
}

// Option type
type EssentialHeadersOption func(*EssentialHeadersConfig)

// require X-Feature-Flags header
func WithFeatureFlagRequired() EssentialHeadersOption {
	return func(c *EssentialHeadersConfig) {
		c.RequireFeatureFlags = true
	}
}

// require X-Location-Id header
func WithLocationIdRequired() EssentialHeadersOption {
	return func(c *EssentialHeadersConfig) {
		c.RequireLocationID = true
	}
}

type RequestAuthValues struct {
	Token         string
	CorrelationID types.CorrelationID
	XSubject      string
}

type RequestAuthConfig struct {
	RequireToken         bool
	RequireCorrelationID bool
	RequireXSubject      bool
}

type RequestAuthOption func(*RequestAuthConfig)

// WithRequireToken marks the auth token as required when fetching request auth values.
func WithRequireToken() RequestAuthOption {
	return func(cfg *RequestAuthConfig) {
		cfg.RequireToken = true
	}
}

// WithRequireCorrelationID marks the correlation ID header as required when fetching request auth values.
func WithRequireCorrelationID() RequestAuthOption {
	return func(cfg *RequestAuthConfig) {
		cfg.RequireCorrelationID = true
	}
}

// WithRequireXSubject marks the X-Subject header as required when fetching request auth values.
func WithRequireXSubject() RequestAuthOption {
	return func(cfg *RequestAuthConfig) {
		cfg.RequireXSubject = true
	}
}
