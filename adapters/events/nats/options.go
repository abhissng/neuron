package nats

import (
	"errors"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/circuitBreaker"
	"github.com/abhissng/neuron/utils/idempotency"
	"github.com/nats-io/nats.go"
	"github.com/sony/gobreaker"
)

// Option defines a functional option for configuring NATSManager.
type Option func(*NATSManager)

// NewStreamConfig creates a new StreamConfig
func NewStreamConfig(name string, subjects []string) *nats.StreamConfig {
	return &nats.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Retention: nats.WorkQueuePolicy,
	}
}

// JetStreamOptions is a slice of JetStreamOption pointers
type JetStreamOptions []*nats.StreamConfig

// NewJetStreamOptions creates a new JetStreamOptions slice
func NewJetStreamOptions() JetStreamOptions {
	return make(JetStreamOptions, 0)
}

// AttachJetStreamOption adds a JetStreamOption to the slice
func (options *JetStreamOptions) AttachJetStreamOption(streamConfig *nats.StreamConfig) {
	*options = append(*options, streamConfig)
}

// WithJetStream enables JetStream and configures the stream
func WithJetStream(cfgs JetStreamOptions, opts ...nats.JSOpt) Option {
	return func(w *NATSManager) {
		js, err := w.nc.JetStream()
		if err != nil {
			w.logger.Error("Failed to initialize JetStream", log.Any("error", err))
			return
		}
		w.js = js

		for _, cfg := range cfgs {
			// Create stream if it doesn't exist
			_, err = js.AddStream(cfg, opts...)
			if err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
				w.logger.Error("Failed to create stream", log.Any("error", err))
				continue
			}
			w.logger.Info("Stream created or exists", log.Any("stream", cfg.Name), log.Any("subjects", cfg.Subjects))
		}

	}
}

// WithLogger sets the logger  for the manager.
func WithLogger(log *log.Log) Option {
	return func(w *NATSManager) {
		w.logger = log
		w.loggerSet = true
	}
}

// WithCircuitBreaker enables Circuit Breaks for the NATS
func WithCircuitBreaker(options ...circuitBreaker.CircuitBreakerOption) Option {
	if len(options) <= 0 {
		options = append(options, circuitBreaker.WithName(BreakerName))
	}

	return func(w *NATSManager) {
		w.breaker = &gobreaker.CircuitBreaker{}
		w.breaker = circuitBreaker.NewCircuitBreaker(
			options...)
	}
}

// WithLogger sets the logger  for the manager.
func WithIdempotencyManager(cleanUpInterval time.Duration) Option {
	return func(w *NATSManager) {
		w.idempotencyManager = idempotency.NewIdempotencyManager[string](cleanUpInterval)
	}
}
