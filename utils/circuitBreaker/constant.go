package circuitBreaker

import "time"

const (
	DefaultCircuitBreakerName = "default-circuit-breaker"
	DefaultBreakerTimeout     = 10 * time.Second
	DefaultBreakerInterval    = 30 * time.Second
	DefaultBreakerMaxRequests = 5
)
