# Events and Messaging Usage

Packages covered:

- [`adapters/events`](../../adapters/events)
- [`adapters/events/nats`](../../adapters/events/nats)
- [`adapters/events/rabbitmq`](../../adapters/events/rabbitmq)
- [`engine`](../../engine)

Primary files:

- [`adapters/events/nats/nats.go`](../../adapters/events/nats/nats.go)
- [`adapters/events/nats/subscriber.go`](../../adapters/events/nats/subscriber.go)
- [`adapters/events/nats/publisher.go`](../../adapters/events/nats/publisher.go)
- [`engine/subscribe.go`](../../engine/subscribe.go)

## NATS Integration

`adapters/events/nats` is the primary messaging adapter:

- connection and lifecycle management
- publish/subscribe helpers
- request/reply utilities
- message header helpers for correlation and metadata

Use it as the transport for orchestration and event-driven handlers.

```go
import (
	neuronNATS "github.com/abhissng/neuron/adapters/events/nats"
	gnats "github.com/nats-io/nats.go"
)

natsMgr, err := neuronNATS.NewNATSManager("nats://localhost:4222")
if err != nil {
	return err
}
defer natsMgr.Close()

_, bl := natsMgr.Publish("orders.created", map[string]any{"id": "ord_123"})
if bl != nil {
	return bl
}
```

## Engine + Messaging

The engine package uses NATS as orchestration transport:

- publishes execution requests
- waits for state responses
- emits rollback requests for compensating actions

See [`../core/engine.md`](../core/engine.md) for full flow details.

## Event Package Notes

`adapters/events` defines shared messaging abstractions/constants.

## RabbitMQ Status

`adapters/events/rabbitmq` is currently a placeholder package. It is documented for roadmap completeness, not as a full runtime integration.

### Subscriber sample

```go
handler := func(msg *gnats.Msg) {
	// process event payload
}
_, bl := natsMgr.Subscribe("orders.created", handler)
if bl != nil {
	return bl
}
```

## Internal Helpers (implementation detail)

NATS package internals include helper methods for subscription setup, header parsing, and middleware chaining; these may change without notice.

## Caveats

- Keep subjects and durable names consistent across publisher/subscriber services.
- Design handlers to be idempotent where retries are possible.
- Use correlation IDs end-to-end for observability.
