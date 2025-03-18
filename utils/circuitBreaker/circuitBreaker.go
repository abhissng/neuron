package circuitBreaker

import (
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerOption is a functional option for configuring the circuit breaker.
type CircuitBreakerOption func(*gobreaker.Settings)

// WithName sets the name of the circuit breaker.
func WithName(name string) CircuitBreakerOption {
	return func(s *gobreaker.Settings) {
		s.Name = name
	}
}

// WithTimeout sets the timeout for the circuit breaker.
func WithTimeout(timeout time.Duration) CircuitBreakerOption {
	return func(s *gobreaker.Settings) {
		s.Timeout = timeout
	}
}

// WithMaxRequests sets the maximum number of requests allowed to pass through
// the circuit breaker in a given interval.
func WithMaxRequests(maxRequests uint32) CircuitBreakerOption {
	return func(s *gobreaker.Settings) {
		s.MaxRequests = maxRequests
	}
}

// WithInterval sets the interval for the circuit breaker.
func WithInterval(interval time.Duration) CircuitBreakerOption {
	return func(s *gobreaker.Settings) {
		s.Interval = interval
	}
}

// WithReadyToTrip sets the ReadyToTrip function for the circuit breaker.
func WithReadyToTrip(readyToTrip func(gobreaker.Counts) bool) CircuitBreakerOption {
	return func(s *gobreaker.Settings) {
		s.ReadyToTrip = readyToTrip
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given options.
func NewCircuitBreaker(options ...CircuitBreakerOption) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        DefaultCircuitBreakerName, // Default name
		Timeout:     DefaultBreakerTimeout,     // Default timeout
		MaxRequests: DefaultBreakerMaxRequests, // Default max requests
		Interval:    DefaultBreakerInterval,    // Default interval
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > ((DefaultBreakerMaxRequests / 2) + 1) // Default ReadyToTrip
		},
	}

	for _, option := range options {
		option(&settings)
	}

	return gobreaker.NewCircuitBreaker(settings)
}
