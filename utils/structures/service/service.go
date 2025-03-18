package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	ServiceNotFoundError = "service not found"
)

// Services represents a collection of services
type Services struct {
	Services              map[string]*ServiceDefinition
	LatestUpdateTimeStamp time.Time
	mu                    sync.Mutex
}

// NewServices initializes a new Services.
func NewServices() *Services {
	return &Services{
		Services:              make(map[string]*ServiceDefinition),
		LatestUpdateTimeStamp: time.Time{},
	}
}

// GetServiceDefinition safely fetches the definitionuration for a specific service.
func (sc *Services) GetServiceDefinition(serviceName string) (*ServiceDefinition, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	definition, exists := sc.Services[serviceName]
	if !exists {
		return &ServiceDefinition{}, errors.New(ServiceNotFoundError)
	}
	return definition, nil
}

// AddServiceDefinition safely adds a new service definitionuration.
func (sc *Services) AddServiceDefinition(definition *ServiceDefinition) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, exists := sc.Services[definition.ServiceName]; exists {
		return fmt.Errorf("service already exists")
	}

	sc.Services[definition.ServiceName] = definition
	return nil
}

// UpdateServiceDefinition safely updates an existing service definitionuration.
func (sc *Services) UpdateServiceDefinition(definition *ServiceDefinition) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, exists := sc.Services[definition.ServiceName]; !exists {
		return errors.New(ServiceNotFoundError)
	}

	sc.Services[definition.ServiceName] = definition
	return nil
}

// AddUpdateServiceDefinition safely add or updates an existing service definitionuration.
func (sc *Services) AddUpdateServiceDefinition(definition *ServiceDefinition) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, exists := sc.Services[definition.ServiceName]; !exists {
		sc.Services[definition.ServiceName] = definition
	}

	sc.Services[definition.ServiceName] = definition
}

// DeleteServiceDefinition safely deletes an existing service definitionuration.
func (sc *Services) DeleteServiceDefinition(serviceName string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, exists := sc.Services[serviceName]; !exists {
		return errors.New(ServiceNotFoundError)
	}

	delete(sc.Services, serviceName)
	return nil
}

// IsServiceActive simulates a database check to verify if a service is active.
func (sc *Services) IsServiceActive(service string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, definition := range sc.Services {
		if strings.EqualFold(definition.ServiceName, service) {
			return definition.Active
		}
	}
	return false
}

// GetServiceDefinitionList returns a list of service definitions
func (sc *Services) GetServiceDefinitionList() []*ServiceDefinition {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	services := make([]*ServiceDefinition, 0, len(sc.Services))
	for _, definition := range sc.Services {
		services = append(services, definition)
	}
	return services
}
