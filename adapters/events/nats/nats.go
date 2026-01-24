package nats

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/idempotency"
	"github.com/abhissng/neuron/utils/types"
	"github.com/sony/gobreaker"

	"github.com/nats-io/nats.go"
)

//----------------------------------------------------------
// GENERIC NATS Manager WITH CIRCUIT BREAKER
//----------------------------------------------------------

// NATSManager encapsulates the NATS connection, JetStream context, and a circuit breaker.
type NATSManager struct {
	context.Context
	nc                 *nats.Conn
	js                 nats.JetStreamContext // JetStream context
	mu                 sync.Mutex
	logger             *log.Log
	loggerSet          bool
	idempotencyManager *idempotency.IdempotencyManager[string]
	breaker            *gobreaker.CircuitBreaker
	subjects           map[string]*nats.Subscription
	subParams          map[string]*subscriptionParams // Track subscription parameters
	done               chan struct{}                  // Channel to signal shutdown
	reconnect          bool                           // Flag to enable auto-reconnection
}

// subscriptionParams stores the parameters needed to recreate a subscription.
// This is used for automatic resubscription when connections are lost.
type subscriptionKind uint8

const (
	subscriptionKindPush subscriptionKind = iota
	subscriptionKindPull
)

type subscriptionParams struct {
	kind         subscriptionKind
	queue        string
	handler      nats.MsgHandler
	subOpts      []nats.SubOpt
	pullConsumer string
}

// SubOptBuilder builds a slice of nats.SubOpt using functional options.
type SubOptBuilder struct {
	opts []nats.SubOpt
}

// NewSubOptBuilder creates a new SubOptBuilder.
func NewSubOptBuilder() *SubOptBuilder {
	return &SubOptBuilder{opts: make([]nats.SubOpt, 0)}
}

// SubOptOption is a functional option for building SubOpts.
type SubOptOption func(*SubOptBuilder)

// BuildSubOpts creates []nats.SubOpt from functional options.
func BuildSubOpts(options ...SubOptOption) []nats.SubOpt {
	b := NewSubOptBuilder()
	for _, opt := range options {
		opt(b)
	}
	return b.opts
}

// WithDurable adds nats.Durable option.
func WithDurable(name string) SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, nats.Durable(name))
	}
}

// WithBindStream adds nats.BindStream option.
func WithBindStream(stream string) SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, nats.BindStream(stream))
	}
}

// WithBind adds nats.Bind option to bind to an existing consumer.
func WithBind(stream, consumer string) SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, nats.Bind(stream, consumer))
	}
}

// WithManualAck adds nats.ManualAck option.
func WithManualAck() SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, nats.ManualAck())
	}
}

// WithAckExplicit adds nats.AckExplicit option.
func WithAckExplicit() SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, nats.AckExplicit())
	}
}

// WithDeliverNew adds nats.DeliverNew option.
func WithDeliverNew() SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, nats.DeliverNew())
	}
}

// WithDeliverAll adds nats.DeliverAll option.
func WithDeliverAll() SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, nats.DeliverAll())
	}
}

// WithDeliverLast adds nats.DeliverLast option.
func WithDeliverLast() SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, nats.DeliverLast())
	}
}

// WithSubOpt adds a raw nats.SubOpt directly.
func WithSubOpt(opt nats.SubOpt) SubOptOption {
	return func(b *SubOptBuilder) {
		b.opts = append(b.opts, opt)
	}
}

// AddSubOpts appends additional subscription options to an existing subscription's params.
// These options will be used during resubscription.
func (w *NATSManager) AddSubOpts(subject string, opts ...nats.SubOpt) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if params, ok := w.subParams[subject]; ok {
		params.subOpts = append(params.subOpts, opts...)
		return true
	}
	return false
}

// SetSubOpts replaces the subscription options for an existing subscription's params.
// These options will be used during resubscription.
func (w *NATSManager) SetSubOpts(subject string, opts ...nats.SubOpt) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if params, ok := w.subParams[subject]; ok {
		params.subOpts = append([]nats.SubOpt(nil), opts...)
		return true
	}
	return false
}

// GetSubOpts returns a copy of the subscription options for a given subject.
func (w *NATSManager) GetSubOpts(subject string) []nats.SubOpt {
	w.mu.Lock()
	defer w.mu.Unlock()

	if params, ok := w.subParams[subject]; ok {
		return append([]nats.SubOpt(nil), params.subOpts...)
	}
	return nil
}

/*
foo.*: Matches subjects like foo.bar, foo.baz, but not foo.bar.baz.
foo.>: Matches subjects like foo.bar, foo.bar.baz, foo.baz.qux, etc.
>.foo: Matches subjects like bar.foo, baz.foo, qux.foo, etc.
>.foo.*: Matches subjects like bar.foo.baz, baz.foo.qux, but not foo.bar.
*/

// NewNATSManager creates and initializes a new NATS manager with circuit breaker support.
// It establishes a connection to NATS server and configures reliability options.
func NewNATSManager(url string, options ...Option) (*NATSManager, error) {
	defaultLog := log.NewBasicLogger(helpers.IsProdEnvironment(), true)

	// Configure NATS options for reliability
	opts := []nats.Option{
		nats.MaxReconnects(DefaultMaxReconnects),
		nats.ReconnectWait(DefaultReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			defaultLog.Error("NATS disconnected", log.Any("error", err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			defaultLog.Info("NATS reconnected", log.Any("url", nc.ConnectedUrl()))
		}),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}

	manager := &NATSManager{
		Context:            context.Background(),
		nc:                 nc,
		subjects:           make(map[string]*nats.Subscription),
		subParams:          make(map[string]*subscriptionParams),
		logger:             defaultLog,
		loggerSet:          false,
		idempotencyManager: idempotency.NewIdempotencyManager[string](idempotency.DefaultCleanupInterval),
		done:               make(chan struct{}),
		reconnect:          true,
		breaker:            nil,
	}

	for _, opt := range options {
		opt(manager)
	}

	if manager.loggerSet {
		_ = defaultLog.Sync()
	}

	return manager, nil
}

// Ping checks the health of the NATS connection.
// It returns an error if the connection is not established or has been lost.
func (w *NATSManager) Ping() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.nc != nil {
		if w.nc.IsConnected() {
			return nil
		}
	}
	return errors.New(ConnectionFailedMessage)
}

// IsClosed reports whether the underlying NATS connection has been closed.
// It is safe for concurrent use.
func (w *NATSManager) IsClosed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.nc == nil {
		return true
	}
	return w.nc.IsClosed()
}

// Close gracefully shuts down the NATS manager.
// It unsubscribes from all subjects, closes connections, and cleans up resources.
func (w *NATSManager) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	defer close(w.done)

	for subject, sub := range w.subjects {
		if err := sub.Unsubscribe(); err != nil {
			b := blame.UnsubscribeFailedError(subject, err)
			message, description := b.Translate()
			w.logger.Error(blame.ErrorUnsubscribeFailed.String(),
				log.Any("message", message), log.Any("description", description))
		}
	}
	// Clear the map to prevent double Unsubscribe
	w.subjects = make(map[string]*nats.Subscription)

	if w.nc != nil && !w.nc.IsClosed() {
		w.logger.Info(constant.ConnectionClosing, log.Any("message", "NATS connection closing"))
		_ = w.nc.Drain()
		// w.nc.Close()
	}

	if w.idempotencyManager != nil {
		w.idempotencyManager.Close()
	}
	w.logger.Info(constant.ConnectionClosed, log.Any("message", "NATS connection closed"))
}

// IsJetStreamEnabled returns true if JetStream is enabled for this manager
func (w *NATSManager) IsJetStreamEnabled() bool {
	return w.js != nil
}

// ackIfJetStream sends an ACK if using JetStream
func (w *NATSManager) ackIfJetStream(msg *nats.Msg) {
	if w.js != nil {
		if err := msg.Ack(); err != nil {
			w.logger.Error("Failed to ACK message", log.Any("error", err))
		}
	}
}

// nakIfJetStream sends a NAK if using JetStream
func (w *NATSManager) nakIfJetStream(msg *nats.Msg) {
	if w.js != nil {
		if err := msg.Nak(); err != nil {
			w.logger.Error("Failed to NAK message", log.Any("error", err))
		}
	}
}

// processMessageIDHeader process an incoming NATS message header using MessageID.
//
// 1. It checks if the message has a "Message-ID" header. If not, an error is logged and the message is discarded.
// 2. It acquires a mutex to ensure thread safety when accessing and modifying internal state.
// 3. It checks if the message has already been processed. If so, a log message is printed and the message is discarded.
// 4. If the message is not processed, it marks it as processed in the internal map.
func (w *NATSManager) processMessageIDHeader(msg *nats.Msg) string {
	messageID := msg.Header.Get(constant.MessageIdHeader)
	if messageID == "" {
		w.logger.Error(constant.MessageIdHeader + " header is missing")
		return ""
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.idempotencyManager.IsProcessed(messageID) {
		w.logger.Info("Message already processed", log.Any(constant.MessageIdHeader, messageID))
		return ""
	}

	// Mark the message as processed
	w.idempotencyManager.MarkAsProcessed(messageID)
	return messageID
}

// handleMessage handles an incoming NATS message.
//
// 1. It calls the processMessageIDHeader function to process the message id header
// 2. It calls the provided handler function to process the message.
// 3. ACKs the message on success (JetStream only)
// 4. A log message is printed indicating that the message has been successfully processed.
func (w *NATSManager) handleMessage(msg *nats.Msg, handler nats.MsgHandler) {
	messageID := w.processMessageIDHeader(msg)
	if messageID == "" {
		// Message already processed or invalid - ACK to prevent redelivery
		w.ackIfJetStream(msg)
		return
	}

	// Process the message with panic recovery
	var processingError error
	func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				processingError = fmt.Errorf("panic recovered: %v\nStack Trace:\n%s", r, string(stack))
				w.logger.Error("Panic in message handler", log.Any("error", processingError))
			}
		}()
		handler(msg)
	}()

	if processingError != nil {
		// NAK on processing failure to allow redelivery
		w.nakIfJetStream(msg)
		return
	}

	// ACK successful processing
	w.ackIfJetStream(msg)
	w.logger.Info("Message processed", log.Any("message_id", messageID))
}

// monitorSubscription continuously monitors a subscription's health.
// It automatically attempts to resubscribe if the subscription becomes invalid.
func (w *NATSManager) monitorSubscription(subject string, sub *nats.Subscription) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			w.mu.Lock()
			currentSub := w.subjects[subject]
			w.mu.Unlock()
			if currentSub == nil {
				return
			}
			if !currentSub.IsValid() && w.reconnect {
				w.logger.Warn("Subscription invalid, attempting to resubscribe",
					log.Any("subject", subject))
				w.resubscribe(subject)
			}
		}
	}
}

// resubscribe attempts to reestablish an invalid subscription using stored parameters.
// It handles both regular NATS and JetStream subscriptions with appropriate configurations.
func (w *NATSManager) resubscribe(subject string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if sub, exists := w.subjects[subject]; exists {
		_ = sub.Unsubscribe()
		delete(w.subjects, subject)
	}

	if params, ok := w.subParams[subject]; ok {

		var sub *nats.Subscription
		var err error

		if w.js != nil {
			if params.kind == subscriptionKindPull {
				sub, err = w.js.PullSubscribe(subject, params.pullConsumer, params.subOpts...)
			} else if params.queue != "" {
				sub, err = w.js.QueueSubscribe(subject, params.queue, params.handler, params.subOpts...)
			} else {
				sub, err = w.js.Subscribe(subject, params.handler, params.subOpts...)
			}
		} else {
			if params.kind == subscriptionKindPull {
				w.logger.Error("Failed to resubscribe:", log.Any("error", "pull subscription requires jetstream"))
				return
			}
			if params.queue != "" {
				sub, err = w.nc.QueueSubscribe(subject, params.queue, params.handler)
			} else {
				sub, err = w.nc.Subscribe(subject, params.handler)
			}
		}
		if err != nil {
			w.logger.Error("Failed to resubscribe:", log.Err(err))
			return
		}

		if w.js == nil {
			// Ensure subscription is active before continuing
			err = w.nc.Flush()
			if err != nil {
				w.logger.Error("Failed to flush subscriptions:", log.Err(err))
				return
			}
		}

		w.subjects[subject] = sub
	}
}

// FetchMessageAndCorrelationField extracts message ID and correlation ID from NATS message headers.
// It returns a slice of log fields for structured logging.
func FetchMessageAndCorrelationField(msg *nats.Msg) []types.Field {
	fields := make([]types.Field, 2)
	fields[0] = log.String(constant.MessageIdHeader, helpers.MessageIDFromNatsMsg(msg))
	fields[1] = log.String(constant.CorrelationIDHeader, helpers.CorrelationIDFromNatsMsg(msg))
	return fields
}

// Slog creates a structured log entry with message metadata and optional additional fields.
// It combines message ID, correlation ID, IP header, and any provided fields.
func Slog(msg *nats.Msg, withFields ...types.Field) []types.Field {
	// Start with the message and correlation fields
	fields := make([]types.Field, 0, 3+len(withFields))
	fields = append(fields, FetchMessageAndCorrelationField(msg)...)
	fields = append(fields, log.String(constant.IPHeader, helpers.IPHeaderFromNatsMsg(msg)))

	// Append additional fields provided as variadic arguments
	fields = append(fields, withFields...)
	return fields
}

// RunSafely executes a function with panic recovery.
// It logs any panics that occur during execution and prevents the application from crashing.
func (w *NATSManager) RunSafely(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			msg := fmt.Sprintf("Panic recovered: %v\nStack Trace:\n%s", r, string(stack))
			// Use the service logger if available, otherwise use standard log
			if w.logger != nil {
				w.logger.Error("Panic recovered", log.Any("error", msg))
			} else {
				helpers.Println(constant.INFO, msg)
			}

		}
	}()
	fn()
}
