package nats

import (
	"fmt"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/nats-io/nats.go"
)

// Subscribe subscribes to a subject and processes messages using the provided handler.
func (w *NATSManager) Subscribe(subject string, handler nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, blame.Blame) {
	defer helpers.RecoverException(recover())
	return w.subscribeInternal(subject, handler, opts, nil)
}

// SubscribeWithMiddleware subscribes to a subject and applies middleware functions.
func (w *NATSManager) SubscribeWithMiddleware(subject string, processor NATSMsgProcessor, opts []nats.SubOpt, middlewares ...MiddlewareFunc) (*nats.Subscription, blame.Blame) {
	wrappedHandler := func(msg *nats.Msg) {
		defer helpers.RecoverException(recover())
		err := processor(msg) // Ignoring blame for now, adjust if needed
		if err != nil {
			w.logger.Error(constant.HandlerFailed, log.Any("SubscribeWithMiddleware", err))
		}

	}
	return w.subscribeInternal(subject, wrappedHandler, opts, middlewares...)
}

// Internal method to handle subscription logic
func (w *NATSManager) subscribeInternal(subject string, handler nats.MsgHandler, opts []nats.SubOpt, middlewares ...MiddlewareFunc) (*nats.Subscription, blame.Blame) {
	defer helpers.RecoverException(recover())
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.subjects[subject]; exists {
		return nil, blame.AlreadySubscribedToSubjectError(subject)
	}

	// Apply middlewares if provided
	var finalHandler nats.MsgHandler
	if len(middlewares) > 0 {
		wrappedHandler := w.WrapNATSMsgProcessor(handler)
		finalHandler = func(msg *nats.Msg) {
			messageID := w.processMessageIDHeader(msg)
			if messageID == "" {
				w.logger.Error("subscribeInternal Message ID not found in header", log.Any(constant.MessageIdHeader, messageID))
				// ACK duplicate/invalid messages to prevent redelivery
				w.ackIfJetStream(msg)
				return
			}

			// Apply middleware and get blame
			if middlewareBlame := applyMiddleware(wrappedHandler, middlewares...)(msg); middlewareBlame != nil {
				w.logger.Error(constant.MiddlewareFailed, log.Any(constant.MessageIdHeader, messageID), log.Any("subscribeInternal", middlewareBlame.FetchErrCode()))
				// NAK on middleware failure to allow redelivery
				w.nakIfJetStream(msg)
				return
			}
			// ACK successful processing
			w.ackIfJetStream(msg)
			w.logger.Info(constant.MessageProcessed, log.Any(constant.MessageIdHeader, messageID))
		}
	} else {
		finalHandler = func(msg *nats.Msg) {
			w.handleMessage(msg, handler)
		}
	}

	var sub *nats.Subscription
	var err error

	if w.js != nil {
		opts = append(opts, nats.ManualAck())
		sub, err = w.js.Subscribe(subject, finalHandler, opts...)
	} else {
		sub, err = w.nc.Subscribe(subject, finalHandler)
	}

	if err != nil {
		w.logger.Error(constant.SubjectSubscribeFailed, log.Any("nats.Subscribe", err))
		return nil, blame.SubscribeToSubjectError(subject, err)
	}

	if w.js == nil {
		// Ensure subscription is active before continuing
		if err := w.nc.Flush(); err != nil {
			w.logger.Error(constant.ConnectionClosed, log.Err(err))
			return nil, blame.SubscribeToSubjectError(subject, err)
		}
	}

	w.subjects[subject] = sub
	w.logger.Info(constant.SubjectSubscribed, log.Any("message", fmt.Sprintf("Subscribed to subject %s", subject)))

	// Start subscription monitoring
	go w.monitorSubscription(subject, sub)
	return sub, nil
}

// SubscribeQueue subscribes to a subject using a queue and processes messages using the provided handler.
func (w *NATSManager) SubscribeQueue(subject, queue string, handler nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, blame.Blame) {
	return w.subscribeQueueInternal(subject, queue, handler, opts)
}

// SubscribeQueueWithMiddleware subscribes to a subject using a queue and processes messages using the provided handler and attached middlewares.
func (w *NATSManager) SubscribeQueueWithMiddleware(subject, queue string, processor NATSMsgProcessor, opts []nats.SubOpt, middlewares ...MiddlewareFunc) (*nats.Subscription, blame.Blame) {
	defer helpers.RecoverException(recover())
	// Wrap the NATSMsgProcessor into a nats.MsgHandler
	wrappedHandler := func(msg *nats.Msg) {
		err := processor(msg) // Ignoring blame for now, adjust if needed
		if err != nil {
			w.logger.Error(constant.HandlerFailed, log.Any("SubscribeQueueWithMiddleware", err))
		}

	}
	return w.subscribeQueueInternal(subject, queue, wrappedHandler, opts, middlewares...)
}

// subscribeQueueInternal is a helper function that handles the common logic for queue subscriptions.
func (w *NATSManager) subscribeQueueInternal(subject, queue string, handler nats.MsgHandler, opts []nats.SubOpt, middlewares ...MiddlewareFunc) (*nats.Subscription, blame.Blame) {
	defer helpers.RecoverException(recover())
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.subjects[subject]; exists {
		return nil, blame.AlreadySubscribedToSubjectError(subject)
	}

	// Apply middlewares if provided
	var finalHandler nats.MsgHandler
	if len(middlewares) > 0 {
		wrappedHandler := w.WrapNATSMsgProcessor(handler)
		finalHandler = func(msg *nats.Msg) {
			messageID := w.processMessageIDHeader(msg)
			if messageID == "" {
				w.logger.Error("subscribeQueueInternal Message ID not found in header", log.Any(constant.MessageIdHeader, messageID))
				// ACK duplicate/invalid messages to prevent redelivery
				w.ackIfJetStream(msg)
				return
			}

			// Apply middleware and get blame
			if middlewareBlame := applyMiddleware(wrappedHandler, middlewares...)(msg); middlewareBlame != nil {
				w.logger.Error(constant.MiddlewareFailed, log.Any(constant.MessageIdHeader, messageID), log.Any("subscribeQueueInternal", middlewareBlame))
				// NAK on middleware failure to allow redelivery
				w.nakIfJetStream(msg)
				return
			}
			// ACK successful processing
			w.ackIfJetStream(msg)
			w.logger.Info(constant.MessageProcessed, log.Any(constant.MessageIdHeader, messageID))
		}
	} else {
		finalHandler = func(msg *nats.Msg) {
			w.handleMessage(msg, handler)
		}
	}

	var sub *nats.Subscription
	var err error

	if w.js != nil {
		// JetStream subscription with manual ACK and durable queue
		opts = append(opts, nats.ManualAck())
		opts = append(opts, nats.Durable(queue)) // Optional: Persistent consumer
		opts = append(opts, nats.DeliverNew())

		// IMPORTANT: Add these for proper acknowledgment
		opts = append(opts, nats.AckExplicit())
		// opts = append(opts, nats.MaxDeliver(10)) // Max redelivery attempts
		// opts = append(opts, nats.AckWait(30*time.Second))

		sub, err = w.js.QueueSubscribe(
			subject,
			queue,
			finalHandler,
			opts...,
		)
	} else {
		// Core NATS subscription
		sub, err = w.nc.QueueSubscribe(subject, queue, finalHandler)
	}

	if err != nil {
		w.logger.Error(constant.SubjectWithQueueSubscribedFailed, log.Any("nats.QueueSubscribe", err))
		return nil, blame.SubscribeToSubjectError(subject, err)
	}

	if w.js == nil {
		// Ensure subscription is active before continuing
		if err := w.nc.Flush(); err != nil {
			w.logger.Error(constant.ConnectionClosed, log.Err(err))
			return nil, blame.SubscribeToSubjectError(subject, err)
		}
	}

	w.subjects[subject] = sub
	w.subParams[subject] = &subscriptionParams{
		queue,
		handler,
	}

	w.logger.Info(constant.SubjectWithQueueSubscribed,
		log.Any("message", fmt.Sprintf("Subscribed to subject %s with queue %s", subject, queue)))

	// Start subscription monitoring
	go w.monitorSubscription(subject, sub)

	return sub, nil
}
