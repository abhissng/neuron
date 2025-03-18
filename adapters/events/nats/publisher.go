package nats

import (
	"strings"
	"time"

	"github.com/abhissng/neuron/adapters/log"
	"github.com/abhissng/neuron/blame"
	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/random"
	"github.com/nats-io/nats.go"
)

// Publish publishes a message to a subject.
func (w *NATSManager) Publish(subject string, payload any) (*nats.PubAck, blame.Blame) {
	return w.publishInternal(subject, payload)
}

// PublishWithMiddleware publishes a message to a subject with middleware attached.
func (w *NATSManager) PublishWithMiddleware(subject string, payload any, middlewares ...MiddlewareFunc) (*nats.PubAck, blame.Blame) {
	return w.publishInternal(subject, payload, middlewares...)
}

// publishInternal is a helper function that handles common publishing logic.
func (w *NATSManager) publishInternal(subject string, payload any, middlewares ...MiddlewareFunc) (*nats.PubAck, blame.Blame) {
	defer helpers.RecoverException(recover())
	data, err := codec.Encode(payload, codec.JSON)
	if err != nil {
		w.logger.Error(constant.EventPublishedFailed, log.Any("codec.Encode", err))
		return nil, blame.MarshalError(codec.JSON, err)
	}
	messageId := random.GenerateUUID()
	// Create the message with headers
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set(constant.MessageIdHeader, messageId)

	var pubErr error
	pubAck := &nats.PubAck{}

	// Final publish handler
	finalHandler := func(msg *nats.Msg) blame.Blame {
		if w.js != nil {
			// Use JetStream for publishing
			pubAck, pubErr = w.js.PublishMsg(msg)
		} else {
			// Fallback to core NATS
			pubErr = w.nc.PublishMsg(msg)
		}

		if pubErr != nil {
			w.logger.Error(constant.EventPublishedFailed, log.Any("nats.PublishMsg", pubErr))
			return blame.PublishMessageError(subject, string(data), pubErr)
		}
		return nil
	}

	// Apply middleware if provided
	wrappedHandler := applyMiddleware(finalHandler, middlewares...)

	// Execute the wrapped publish handler
	if err := wrappedHandler(msg); err != nil {
		w.logger.Error(constant.EventPublishedFailed, log.Any("wrappedHandler", err.FetchErrCode()))
		return nil, err
	}

	w.logger.Info(constant.EventPublished, Slog(msg, log.String("subject", subject))...)

	if w.js != nil {
		return pubAck, nil
	}
	return nil, nil
}

// PublishAndWait handles message preparation and publishing using JetStream
func (w *NATSManager) PublishAndWait(subject, queueGroup string, payload any, timeout time.Duration, middlewares ...MiddlewareFunc) (*nats.Msg, blame.Blame) {
	defer helpers.RecoverException(recover())

	data, err := codec.Encode(payload, codec.JSON)
	if err != nil {
		w.logger.Error(constant.EventPublishedFailed, log.Any("codec.Encode", err))
		return nil, blame.MarshalError(codec.JSON, err)
	}
	messageId := random.GenerateUUID()

	result, err := w.breaker.Execute(func() (interface{}, error) {
		replySubj := w.createReplySubject(subject)
		sub, blameErr := w.createSubscription(replySubj, queueGroup, messageId)
		if blameErr != nil {
			w.logger.Error(constant.EventPublishedFailed, log.Any("createSubscription", blameErr))
			return nil, blameErr.ErrorFromBlame()
		}
		defer func() { _ = sub.Unsubscribe() }()

		if blameErr := w.publishMessage(subject, replySubj, data, messageId, middlewares...); blameErr != nil {
			w.logger.Error(constant.EventPublishedFailed, log.Any("publishMessage", blameErr))
			return nil, blameErr.ErrorFromBlame()
		}

		reply, err := sub.NextMsg(timeout)
		if err != nil {
			w.logger.Error(constant.EventPublishedFailed, log.Any("nextMsg", err), log.Any(constant.MessageIdHeader, messageId), log.Any("subject", subject))
			return nil, err
		}
		return reply, nil
	})

	if err != nil {
		w.logger.Error(constant.EventPublishedFailed, log.Any("error", err), log.Any(constant.MessageIdHeader, messageId), log.Any("subject", subject))
		return nil, blame.PublishMessageError(subject, string(data), err)
	}

	reply, ok := result.(*nats.Msg)
	if !ok {
		return nil, blame.TypeConversionError("PublishAndWait circuit breaker result", "unexpected", "*nats.Msg", nil)
	}
	return reply, nil
}

// createReplySubject creates a unique reply subject based on the original subject
func (w *NATSManager) createReplySubject(subject string) string {
	parts := strings.Split(subject, ".")
	if len(parts) > 0 {
		return parts[0] + "." + strings.ToLower(strings.Join(strings.Split(nats.NewInbox(), "."), "_"))
	}
	return subject + "." + strings.ToLower(strings.Join(strings.Split(nats.NewInbox(), "."), "_"))
}

// createSubscription creates appropriate subscription based on queue group
func (w *NATSManager) createSubscription(replySubj, queueGroup string, messageId string) (*nats.Subscription, blame.Blame) {
	var sub *nats.Subscription
	var err error

	if helpers.IsEmpty(queueGroup) {
		sub, err = w.nc.SubscribeSync(replySubj)
		if err != nil {
			w.logger.Error(constant.SubscribeSyncFailed, log.Any("SubscribeSync", err), log.Any(constant.MessageIdHeader, messageId), log.Any("ReplySubject", replySubj))
			return nil, blame.PublishMessageError(replySubj, "", err)
		}
		return sub, nil
	}

	sub, err = w.nc.QueueSubscribeSync(replySubj, queueGroup)
	if err != nil {
		w.logger.Error(constant.SubscribeSyncFailed, log.Any("QueueSubscribeSync", err), log.Any(constant.MessageIdHeader, messageId), log.Any("ReplySubject", replySubj))
		return nil, blame.PublishMessageError(replySubj, "", err)
	}
	return sub, nil
}

// publishMessage handles message preparation and publishing
func (w *NATSManager) publishMessage(subject, replySubj string, data []byte, messageId string, middlewares ...MiddlewareFunc) blame.Blame {
	msg := &nats.Msg{
		Subject: subject,
		Reply:   replySubj,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set(constant.MessageIdHeader, messageId)

	finalHandler := func(msg *nats.Msg) blame.Blame {
		if err := w.nc.PublishMsg(msg); err != nil {
			w.logger.Error(constant.EventPublishedFailed, Slog(msg, log.Any("PublishMsg", err))...)
			return blame.PublishMessageError(subject, string(data), err)
		}
		w.logger.Info(constant.EventPublished, Slog(msg, log.String("subject", subject))...)
		return nil
	}

	wrappedHandler := applyMiddleware(finalHandler, middlewares...)
	return wrappedHandler(msg)
}

/*
* STREAM Logic
 */

// createStreamSubscription creates appropriate subscription based on queue group using JetStream
func (w *NATSManager) createStreamSubscription(replySubj, queueGroup string, messageId string) (*nats.Subscription, blame.Blame) {
	var sub *nats.Subscription
	var err error

	if helpers.IsEmpty(queueGroup) {
		sub, err = w.js.SubscribeSync(replySubj, nats.ManualAck())
		if err != nil {
			w.logger.Error(constant.EventPublishedFailed, log.Any("SubscribeSync", err), log.Any(constant.MessageIdHeader, messageId), log.Any("ReplySubject", replySubj))
			return nil, blame.PublishMessageError(replySubj, "", err)
		}
		return sub, nil
	}

	sub, err = w.js.QueueSubscribeSync(replySubj, queueGroup, nats.ManualAck())
	if err != nil {
		w.logger.Error(constant.EventPublishedFailed, log.Any("QueueSubscribeSync", err), log.Any(constant.MessageIdHeader, messageId), log.Any("ReplySubject", replySubj))
		return nil, blame.PublishMessageError(replySubj, "", err)
	}
	return sub, nil
}

// publishStreamMessage handles message preparation and publishing using JetStream
func (w *NATSManager) publishStreamMessage(subject, replySubj string, data []byte, messageId string, middlewares ...MiddlewareFunc) blame.Blame {
	msg := &nats.Msg{
		Subject: subject,
		Reply:   replySubj,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set(constant.MessageIdHeader, messageId)

	finalHandler := func(msg *nats.Msg) blame.Blame {
		if _, err := w.js.PublishMsg(msg); err != nil {
			w.logger.Error(constant.EventPublishedFailed, log.Any(constant.MessageIdHeader, messageId), log.Any("PublishMsg", err))
			return blame.PublishMessageError(subject, string(data), err)
		}
		w.logger.Info(constant.EventPublished, log.String(constant.MessageIdHeader, messageId), log.String("subject", subject))
		return nil
	}

	wrappedHandler := applyMiddleware(finalHandler, middlewares...)
	return wrappedHandler(msg)
}

// PublishAndWaitUsingStream handles message preparation and publishing using JetStream
func (w *NATSManager) PublishAndWaitUsingStream(subject, queueGroup string, payload any, timeout time.Duration, middlewares ...MiddlewareFunc) (*nats.Msg, blame.Blame) {
	defer helpers.RecoverException(recover())

	data, err := codec.Encode(payload, codec.JSON)
	if err != nil {
		return nil, blame.MarshalError(codec.JSON, err)
	}
	messageId := random.GenerateUUID()

	result, err := w.breaker.Execute(func() (interface{}, error) {
		replySubj := w.createReplySubject(subject)
		w.logger.Info("ReplySubject", log.Any("ReplySubject", replySubj))

		sub, blameErr := w.createStreamSubscription(replySubj, queueGroup, messageId)
		if blameErr != nil {
			return nil, blameErr.ErrorFromBlame()
		}
		defer func() { _ = sub.Unsubscribe() }()

		if blameErr := w.publishStreamMessage(subject, replySubj, data, messageId, middlewares...); blameErr != nil {
			return nil, blameErr.ErrorFromBlame()
		}

		reply, err := sub.NextMsg(timeout)
		if err != nil {
			w.logger.Error(constant.EventPublishedFailed, log.Any(constant.MessageIdHeader, messageId), log.Any("subject", subject), log.Any("nextMsg", err))
			return nil, err
		}
		return reply, nil
	})

	if err != nil {
		w.logger.Error(constant.EventPublishedFailed, log.Any(constant.MessageIdHeader, messageId), log.Any("subject", subject), log.Any("error", err))
		return nil, blame.PublishMessageError(subject, string(data), err)
	}

	reply, ok := result.(*nats.Msg)
	if !ok {
		return nil, blame.TypeConversionError("PublishAndWaitUsingStream circuit breaker result", "unexpected", "*nats.Msg", nil)
	}
	return reply, nil
}
