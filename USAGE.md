# Neuron Usage Guide

This guide is the task-oriented entry point for using `github.com/abhissng/neuron`.

Use this file to choose the right package path, then jump to deep package documentation.

Deep docs root: [`docs/README.md`](docs/README.md)

## Start Here

- Bootstrap app and shared managers: [`docs/core/context.md`](docs/core/context.md)
- Run service workflows over NATS: [`docs/core/engine.md`](docs/core/engine.md)
- Handle errors/results consistently: [`docs/core/blame-and-result.md`](docs/core/blame-and-result.md)
- Use database abstractions and adapters: [`docs/core/database.md`](docs/core/database.md)

## Quick Samples

### Bootstrap service context stack

```go
appCtx := context.NewAppContext(
	context.WithServiceID("orders"),
	context.WithLogger(logger),
	context.WithNATSManager("nats://localhost:4222"),
)
_ = appCtx
```

### Build a typed result flow

```go
order := &Order{ID: "ord_123"}
ok := result.NewSuccess(order)
_ = ok

fail := result.NewFailure[Order](blame.InternalServerError(err))
_ = fail
```

### Generate UUID v7 with fallback

```go
id := random.GenerateUUID(random.WithUUIDVersion7())
_ = id
```

## By Use Case

### Build HTTP APIs

- Gin server options and lifecycle: [`docs/adapters/http-and-gin.md`](docs/adapters/http-and-gin.md)
- Request parsing and validation helpers: [`docs/adapters/http-and-gin.md`](docs/adapters/http-and-gin.md)
- HTTP client wrappers and middleware: [`docs/adapters/http-and-gin.md`](docs/adapters/http-and-gin.md)

### Build Event-Driven Services

- NATS manager setup and publish/subscribe: [`docs/adapters/events-and-messaging.md`](docs/adapters/events-and-messaging.md)
- Engine orchestration and rollback flow: [`docs/core/engine.md`](docs/core/engine.md)
- Event package notes and stubs: [`docs/adapters/events-and-messaging.md`](docs/adapters/events-and-messaging.md)

### Add Authentication and Sessions

- PASETO manager and token flows: [`docs/adapters/auth-and-security.md`](docs/adapters/auth-and-security.md)
- JWT middleware usage: [`docs/adapters/auth-and-security.md`](docs/adapters/auth-and-security.md)
- Session manager and middleware: [`docs/adapters/auth-and-security.md`](docs/adapters/auth-and-security.md)
- Vault integration for secrets: [`docs/adapters/auth-and-security.md`](docs/adapters/auth-and-security.md)

### Use Data Stores and Search

- Postgres/MySQL generic database patterns: [`docs/core/database.md`](docs/core/database.md)
- Mongo/Redis/OpenSearch/Cosmos adapters: [`docs/adapters/data-and-storage.md`](docs/adapters/data-and-storage.md)
- Store and regex utilities: [`docs/adapters/data-and-storage.md`](docs/adapters/data-and-storage.md)

### Integrate Cloud, Files, and Payments

- AWS/OCI/Cloud manager patterns: [`docs/adapters/cloud-and-file.md`](docs/adapters/cloud-and-file.md)
- File upload and validation flow: [`docs/adapters/cloud-and-file.md`](docs/adapters/cloud-and-file.md)
- Payment manager and Razorpay service: [`docs/adapters/payment-and-email.md`](docs/adapters/payment-and-email.md)
- Email adapter usage: [`docs/adapters/payment-and-email.md`](docs/adapters/payment-and-email.md)

### Observability and Ops

- Logging wrapper usage: [`docs/adapters/ops-and-observability.md`](docs/adapters/ops-and-observability.md)
- Prometheus adapter patterns: [`docs/adapters/ops-and-observability.md`](docs/adapters/ops-and-observability.md)
- Viper configuration and validator helpers: [`docs/adapters/ops-and-observability.md`](docs/adapters/ops-and-observability.md)

### Shared Utilities

- Helpers, types, and conversion patterns: [`docs/utils/helpers-and-types.md`](docs/utils/helpers-and-types.md)
- Structures and message contracts: [`docs/utils/structures.md`](docs/utils/structures.md)
- Cache, crypto, scheduling, worker pools, and resilience tools: [`docs/utils/infra-utilities.md`](docs/utils/infra-utilities.md)

## API Surface Coverage Policy

All package docs include:

- exported/public functions and methods
- internal/unexported helper notes where behavior matters
- caveats for unstable internal APIs

Internal sections are explicitly marked as implementation detail and may change.

## Package Inventory Quick Links

### Core

- [`context`](context)
- [`engine`](engine)
- [`blame`](blame)
- [`result`](result)
- [`database`](database)

### Adapters

- [`adapters/aws`](adapters/aws), [`adapters/cloud`](adapters/cloud), [`adapters/oci`](adapters/oci)
- [`adapters/events`](adapters/events), [`adapters/events/nats`](adapters/events/nats), [`adapters/events/rabbitmq`](adapters/events/rabbitmq)
- [`adapters/file`](adapters/file), [`adapters/file/uploadFile`](adapters/file/uploadFile)
- [`adapters/gin/handler`](adapters/gin/handler), [`adapters/gin/middleware`](adapters/gin/middleware), [`adapters/gin/middleware/redis_rate_limiter`](adapters/gin/middleware/redis_rate_limiter), [`adapters/gin/request`](adapters/gin/request), [`adapters/gin/server`](adapters/gin/server)
- [`adapters/grpcserver`](adapters/grpcserver), [`adapters/grpcserver/example`](adapters/grpcserver/example)
- [`adapters/http`](adapters/http), [`adapters/jwt`](adapters/jwt), [`adapters/log`](adapters/log)
- [`adapters/mongo`](adapters/mongo), [`adapters/mysql`](adapters/mysql), [`adapters/postgres`](adapters/postgres), [`adapters/redis`](adapters/redis), [`adapters/opensearch`](adapters/opensearch), [`adapters/cosmos`](adapters/cosmos)
- [`adapters/paseto`](adapters/paseto), [`adapters/session`](adapters/session), [`adapters/vault`](adapters/vault)
- [`adapters/payment`](adapters/payment), [`adapters/payment/razorpay`](adapters/payment/razorpay)
- [`adapters/prometheus`](adapters/prometheus), [`adapters/store`](adapters/store), [`adapters/store/regex`](adapters/store/regex), [`adapters/validator`](adapters/validator), [`adapters/viper`](adapters/viper), [`adapters/email`](adapters/email)

### Utils

- [`utils/cache`](utils/cache), [`utils/circuitBreaker`](utils/circuitBreaker), [`utils/codec`](utils/codec)
- [`utils/concurrent/concurrentMap`](utils/concurrent/concurrentMap), [`utils/constant`](utils/constant), [`utils/cryptography`](utils/cryptography), [`utils/graceful`](utils/graceful), [`utils/helpers`](utils/helpers), [`utils/idempotency`](utils/idempotency), [`utils/random`](utils/random), [`utils/schedule`](utils/schedule)
- [`utils/structures`](utils/structures), [`utils/structures/acknowledgment`](utils/structures/acknowledgment), [`utils/structures/claims`](utils/structures/claims), [`utils/structures/discovery`](utils/structures/discovery), [`utils/structures/message`](utils/structures/message), [`utils/structures/service`](utils/structures/service)
- [`utils/timeutil`](utils/timeutil), [`utils/types`](utils/types), [`utils/workerpool`](utils/workerpool)
