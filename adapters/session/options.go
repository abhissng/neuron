package session

import "github.com/abhissng/neuron/utils/structures"

// SessionMiddlewareOptions defines options for the Paseto middleware.
type SessionMiddlewareOptions struct {
	excludedOptions *structures.ExcludedOptions // List of options to exclude from token validation.
}

func NewSessionMiddlewareOptions() *SessionMiddlewareOptions {
	return &SessionMiddlewareOptions{}
}

// ExcludedOptions returns the list of options to exclude from token validation.
func (p *SessionMiddlewareOptions) ExcludedOptions() *structures.ExcludedOptions {
	return p.excludedOptions
}

// AddExcludedService adds a service to the list of excluded services.
func (p *SessionMiddlewareOptions) AddExcludedService(service *string) {
	if p.excludedOptions == nil {
		p.excludedOptions = structures.NewExcludedOptions()
	}
	p.excludedOptions.Services = append(p.excludedOptions.Services, service)
}

// AddExcludedRecord adds a record to the list of excluded records.
func (p *SessionMiddlewareOptions) AddExcludedRecord(record *string) {
	if p.excludedOptions == nil {
		p.excludedOptions = structures.NewExcludedOptions()
	}
	p.excludedOptions.Records = append(p.excludedOptions.Records, record)
}

// AddExcludedEvent adds an event to the list of excluded events.
func (p *SessionMiddlewareOptions) AddExcludedEvent(event *string) {
	if p.excludedOptions == nil {
		p.excludedOptions = structures.NewExcludedOptions()
	}
	p.excludedOptions.Events = append(p.excludedOptions.Events, event)
}

// HasExcludedOption returns true if the list of excluded options is not empty.
func (p *SessionMiddlewareOptions) HasExcludedOption() bool {
	return p.excludedOptions != nil
}
