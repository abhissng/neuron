package engine

import (
	natsInternal "github.com/abhissng/neuron/adapters/events/nats"
	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
)

type subscriberOpts struct {
	durableName string
	streamName  string
}

// SubscriberOption applies optional overrides for a single subscription (durable name, stream name).
type SubscriberOption func(o *subscriberOpts)

// WithSubscriberDurable sets the durable consumer name for this subscription.
func WithSubscriberDurable(name string) SubscriberOption {
	return func(o *subscriberOpts) { o.durableName = name }
}

// WithSubscriberStream sets the stream name for this subscription.
func WithSubscriberStream(name string) SubscriberOption {
	return func(o *subscriberOpts) { o.streamName = name }
}

// SubEntry represents one subscription (subject, processor, and optional overrides).
type SubEntry struct {
	Subject   string
	Processor natsInternal.NATSMsgProcessor
	Options   []SubscriberOption
}

// Subscriptions is a versioned list of subscriptions. Build it with NewSubscriptions (options for durable name/stream name), AddSubscriberEvent, then pass to SubscribeWithMiddleware.
type Subscriptions struct {
	version     string
	durableName string
	streamName  string
	entries     []SubEntry
}

// SubscriptionsOption configures durable name and stream name for the whole subscription set (e.g. per version).
type SubscriptionsOption func(s *Subscriptions)

// WithDurableName sets the durable consumer name for this subscription set.
func WithDurableName(name string) SubscriptionsOption {
	return func(s *Subscriptions) { s.durableName = name }
}

// WithStreamName sets the stream name for this subscription set.
func WithStreamName(name string) SubscriptionsOption {
	return func(s *Subscriptions) { s.streamName = name }
}

// NewSubscriptions starts a new subscription set. Version is optional (e.g. "v1"). Use WithDurableName/WithStreamName to set durable and stream for this set.
func NewSubscriptions(version string, options ...SubscriptionsOption) *Subscriptions {
	s := &Subscriptions{version: version, entries: make([]SubEntry, 0)}
	for _, apply := range options {
		apply(s)
	}
	return s
}

// AddSubscriberEvent adds a new subscriber event and returns the same Subscriptions for chaining.
func (s *Subscriptions) AddSubscriberEvent(subject string, processor natsInternal.NATSMsgProcessor, options ...SubscriberOption) *Subscriptions {
	s.entries = append(s.entries, SubEntry{Subject: subject, Processor: processor, Options: options})
	return s
}

// SubscribeWithMiddleware subscribes to all entries in subs using the set's durable name and stream name (from options), logger, and middlewares.
// Each entry may override durable/stream via SubscriberOption. Returns on first subscription error.
func SubscribeWithMiddleware(
	nm *natsInternal.NATSManager,
	subs *Subscriptions,
	logger *log.Log,
	middlewares ...natsInternal.MiddlewareFunc,
) error {
	defer func() {
		helpers.RecoverException(recover())
	}()
	for _, e := range subs.entries {
		opts := &subscriberOpts{durableName: subs.durableName, streamName: subs.streamName}
		for _, apply := range e.Options {
			apply(opts)
		}
		durable, stream := opts.durableName, opts.streamName
		if durable == "" {
			durable = subs.durableName
		}
		if stream == "" {
			stream = subs.streamName
		}
		_, cause := nm.SubscribeWithMiddleware(
			e.Subject,
			e.Processor,
			natsInternal.BuildSubOpts(natsInternal.WithDurable(durable), natsInternal.WithBindStream(stream)),
			append([]natsInternal.MiddlewareFunc{natsInternal.LogMiddleware(constant.Subscribe, logger)}, middlewares...)...,
		)
		if cause != nil {
			logger.Error(constant.StartServiceFailed, log.Any("Subscriber", e.Subject), log.Err(cause))
			return cause
		}
		logger.Info(constant.StartServiceSuccessful, log.Any("Subscriber", e.Subject))
	}
	return nil
}
