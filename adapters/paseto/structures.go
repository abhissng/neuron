package paseto

import (
	"time"

	"github.com/o1egl/paseto"
)

// GetPasetoObj creates a new Paseto V2 instance.
func GetPasetoObj() *paseto.V2 {
	// Comment: Returns a new instance of the Paseto V2 library for token operations.
	return paseto.NewV2()
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

// PasetoMiddlewareOptions defines options for the Paseto middleware.
type PasetoMiddlewareOptions struct {
	isAutoRefresh    bool          // Indicates whether auto-refresh is enabled.
	authHeader       string        // The name of the authorization header (e.g., "Authorization").
	newAuthHeader    string        // The name of the header for the new token (e.g., "New-Authorization").
	refreshThreshold time.Duration // Time before token expiration to trigger refresh.
	// excludedServices []string         // List of services to exclude from token validation.
	excludedOptions *ExcludedOptions // List of options to exclude from token validation.
}

// Getters for PasetoMiddlewareOptions fields.
func (p *PasetoMiddlewareOptions) IsAutoRefresh() bool {
	// Comment: Returns whether auto-refresh is enabled.
	return p.isAutoRefresh
}

// AuthHeader returns the name of the authorization header.
func (p *PasetoMiddlewareOptions) AuthHeader() string {
	// Comment: Returns the name of the authorization header.
	return p.authHeader
}

// NewAuthHeader returns the name of the header for the new token (if auto-refresh is enabled).
func (p *PasetoMiddlewareOptions) NewAuthHeader() string {
	// Comment: Returns the name of the header for the new token (if auto-refresh is enabled).
	return p.newAuthHeader
}

// RefreshThreshold returns the time threshold before token expiration to trigger refresh.
func (p *PasetoMiddlewareOptions) RefreshThreshold() time.Duration {
	// Comment: Returns the time threshold before token expiration to trigger refresh.
	return p.refreshThreshold
}

/*
// ExcludedServices returns the list of services to exclude from token validation.

	func (p *PasetoMiddlewareOptions) ExcludedServices() []string {
		// Comment: Returns the list of services to exclude from token validation.
		return p.excludedServices
	}

// HasExcludedService returns true if the list of excluded services is not empty.

	func (p *PasetoMiddlewareOptions) HasExcludedService() bool {
		return len(p.excludedServices) > 0
	}
*/

// ExcludedOptions returns the list of options to exclude from token validation.
func (p *PasetoMiddlewareOptions) ExcludedOptions() *ExcludedOptions {
	return p.excludedOptions
}

// AddExcludedService adds a service to the list of excluded services.
func (p *PasetoMiddlewareOptions) AddExcludedService(service *string) {
	if p.excludedOptions == nil {
		p.excludedOptions = NewExcludedOptions()
	}
	p.excludedOptions.Services = append(p.excludedOptions.Services, service)
}

// AddExcludedRecord adds a record to the list of excluded records.
func (p *PasetoMiddlewareOptions) AddExcludedRecord(record *string) {
	if p.excludedOptions == nil {
		p.excludedOptions = NewExcludedOptions()
	}
	p.excludedOptions.Records = append(p.excludedOptions.Records, record)
}

// AddExcludedEvent adds an event to the list of excluded events.
func (p *PasetoMiddlewareOptions) AddExcludedEvent(event *string) {
	if p.excludedOptions == nil {
		p.excludedOptions = NewExcludedOptions()
	}
	p.excludedOptions.Events = append(p.excludedOptions.Events, event)
}

// HasExcludedOption returns true if the list of excluded options is not empty.
func (p *PasetoMiddlewareOptions) HasExcludedOption() bool {
	return p.excludedOptions != nil
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

// TokenDetails holds information about a token.
type TokenDetails struct {
	Token     string
	ExpiresAt time.Time
	ID        string // Unique identifier for the token
}
