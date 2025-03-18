package structures

// import (
// 	"fmt"
// 	"sync"
// 	"time"

// 	"github.com/abhissng/neuron/utils/types"
// )

// // ServiceConfig represents the configuration of a service.
// type ServiceConfig struct {
// 	ServiceName string    `json:"serviceName"`
// 	States      *States   `json:"states"`
// 	Version     int       `json:"version"`
// 	Active      bool      `json:"active"`
// 	LastUpdated time.Time `json:"lastUpdated"`
// }

// // NewServiceConfig creates a new ServiceConfig
// func NewServiceConfig() *ServiceConfig {
// 	return &ServiceConfig{
// 		States:      NewStates(),
// 		LastUpdated: time.Time{},
// 	}
// }

// // States represents the states of a service
// type States struct {
// 	States        []*ServiceState `json:"states"`
// 	RollbackOrder []string        `json:"rollbackOrder"`
// }

// // NewStates creates a new States
// func NewStates() *States {
// 	return &States{
// 		States:        make([]*ServiceState, 0),
// 		RollbackOrder: make([]string, 0),
// 	}
// }

// // ServiceState represents a state of a service
// type ServiceState struct {
// 	Service         string `json:"service"`
// 	ExecuteSubject  string `json:"executeSubject"`
// 	RollbackSubject string `json:"rollbackSubject"`
// }

// // Services represents a collection of services
// type Services struct {
// 	Services              map[string]*ServiceConfig
// 	LatestUpdateTimeStamp time.Time
// 	mu                    sync.Mutex
// }

// // NewServices initializes a new Services.
// func NewServices() *Services {
// 	return &Services{
// 		Services:              make(map[string]*ServiceConfig),
// 		LatestUpdateTimeStamp: time.Time{},
// 	}
// }

// // GetServiceConfig safely fetches the configuration for a specific service.
// func (sc *Services) GetServiceConfig(serviceName string) (*ServiceConfig, error) {
// 	sc.mu.Lock()
// 	defer sc.mu.Unlock()

// 	config, exists := sc.Services[serviceName]
// 	if !exists {
// 		return &ServiceConfig{}, fmt.Errorf("service not found")
// 	}
// 	return config, nil
// }

// // AddServiceConfig safely adds a new service configuration.
// func (sc *Services) AddServiceConfig(config *ServiceConfig) error {
// 	sc.mu.Lock()
// 	defer sc.mu.Unlock()

// 	if _, exists := sc.Services[config.ServiceName]; exists {
// 		return fmt.Errorf("service already exists")
// 	}

// 	sc.Services[config.ServiceName] = config
// 	return nil
// }

// // UpdateServiceConfig safely updates an existing service configuration.
// func (sc *Services) UpdateServiceConfig(config *ServiceConfig) error {
// 	sc.mu.Lock()
// 	defer sc.mu.Unlock()

// 	if _, exists := sc.Services[config.ServiceName]; !exists {
// 		return fmt.Errorf("service not found")
// 	}

// 	sc.Services[config.ServiceName] = config
// 	return nil
// }

// // AddUpdateServiceConfig safely add or updates an existing service configuration.
// func (sc *Services) AddUpdateServiceConfig(config *ServiceConfig) {
// 	sc.mu.Lock()
// 	defer sc.mu.Unlock()

// 	if _, exists := sc.Services[config.ServiceName]; !exists {
// 		sc.Services[config.ServiceName] = config
// 	}

// 	sc.Services[config.ServiceName] = config
// }

// // DeleteServiceConfig safely deletes an existing service configuration.
// func (sc *Services) DeleteServiceConfig(serviceName string) error {
// 	sc.mu.Lock()
// 	defer sc.mu.Unlock()

// 	if _, exists := sc.Services[serviceName]; !exists {
// 		return fmt.Errorf("service not found")
// 	}

// 	delete(sc.Services, serviceName)
// 	return nil
// }

// // Message represents the structure of a transaction message.
// type Message[T any] struct {
// 	CorrelationID types.CorrelationID `json:"correlation_id"`
// 	RequestId     types.RequestID     `json:"request_id"`
// 	Payload       T                   `json:"payload"`
// 	Status        string              `json:"status"` // "pending", "completed", "failed"
// 	Error         string              `json:"error,omitempty"`
// 	Timestamp     time.Time           `json:"timestamp"`
// }

/////

/*
type ServicePayload[T any] struct {
	CorrelationID types.CorrelationID `json:"correlation_id"`
	RequestId     types.RequestID     `json:"request_id"`
	Payload       T                   `json:"payload"`
}

func NewServicePayload[T any](ctx *context.ServiceContext, payload T) ServicePayload[T] {
	return ServicePayload[T]{
		CorrelationID: types.CorrelationID(ctx.GetCorrelationID()),
		RequestId:     types.RequestID(helpers.GenerateUUID()),
		Payload:       payload,
	}
}

type RollbackHistory struct {
	CorrelationID types.CorrelationID `json:"correlation_id"`
	History       []string            `json:"history"`
}

func NewRollbackHistory(correlationId types.CorrelationID) *RollbackHistory {
	return &RollbackHistory{
		CorrelationID: correlationId,
		History:       make([]string, 0),
	}
}

func (r *RollbackHistory) AppendHistory(history string) *RollbackHistory {
	r.History = append(r.History, history)
	return r
}

// Message represents the structure of a transaction message.
type Message[T any] struct {
	// CorrelationID types.CorrelationID `json:"correlation_id"`
	// RequestId     types.RequestID     `json:"request_id"`
	Payload   T         `json:"payload"`
	Status    string    `json:"status"` // "pending", "completed", "failed"
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

*/

// QueueConfig represents one queue (and subject) definition in queues.json.
//
//	type QueueConfig struct {
//		Name    string `json:"name"`
//		Subject string `json:"subject"`
//	}
