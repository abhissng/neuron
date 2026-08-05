# Structures Usage

Packages covered:

- [`utils/structures`](../../utils/structures)
- [`utils/structures/acknowledgment`](../../utils/structures/acknowledgment)
- [`utils/structures/claims`](../../utils/structures/claims)
- [`utils/structures/discovery`](../../utils/structures/discovery)
- [`utils/structures/message`](../../utils/structures/message)
- [`utils/structures/service`](../../utils/structures/service)

## Purpose

These packages provide cross-service payload and contract structures:

- API acknowledgment/response envelopes
- auth/token claims
- message envelopes for engine/event flows
- service definition and discovery contracts

## Message and Service Contracts

- `utils/structures/message` contains canonical message models for transaction/service flow.
- `utils/structures/service` contains service definition/state models used by orchestration.

```go
msg := message.NewMessage(
	types.ActionExecute,
	types.StatusPending,
	types.CorrelationID("corr-123"),
	payload,
)
_ = msg
```

## Acknowledgment

`utils/structures/acknowledgment` defines standardized API response shapes used by HTTP handler wrappers.

```go
resp := acknowledgment.NewAPIResponse(
	true,
	types.CorrelationID("corr-123"),
	data,
)
_ = resp
```

## Claims

`utils/structures/claims` defines claims payload schemas used by session/PASETO/JWT layers.

## Discovery

`utils/structures/discovery` includes discovery-oriented structural definitions and legacy/draft references where noted.

## Internal Helpers (implementation detail)

Some files preserve legacy drafts/comments for historical context. Treat active exported contracts as source of truth for current integrations.

## Caveats

- Keep wire contracts backward-compatible where they cross service boundaries.
- Avoid duplicating message payload structures in downstream services.
