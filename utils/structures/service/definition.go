package service

import (
	"time"
)

// TODO change main service, microservice and states schema changes
// ServiceDefinition represents the configuration of a service.
type ServiceDefinition struct {
	ServiceName string    `json:"serviceName"`
	States      *States   `json:"states"`
	Version     string    `json:"version"`
	Active      bool      `json:"active"`
	LastUpdate  time.Time `json:"lastUpdated"`
	QueueGroup  string    `json:"queueGroup"`
}

// NewServiceDefinition creates a new ServiceDefinition
func NewServiceDefinition() *ServiceDefinition {
	return &ServiceDefinition{
		States:     NewStates(),
		LastUpdate: time.Time{},
	}
}

// States represents the states of a service
type States struct {
	States        []*ServiceState `json:"states"`
	RollbackOrder []string        `json:"rollbackOrder"`
}

// NewStates creates a new States
func NewStates() *States {
	return &States{
		States:        make([]*ServiceState, 0),
		RollbackOrder: make([]string, 0),
	}
}

// ServiceState represents a state of a service
type ServiceState struct {
	Service         string `json:"service"`
	ExecuteSubject  string `json:"executeSubject"`
	RollbackSubject string `json:"rollbackSubject"`
}

func NewServiceState(service, executeSubject, rollbackSubject string) *ServiceState {
	return &ServiceState{
		Service:         service,
		ExecuteSubject:  executeSubject,
		RollbackSubject: rollbackSubject,
	}
}
