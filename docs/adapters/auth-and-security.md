# Auth and Security Adapters Usage

Packages covered:

- [`adapters/paseto`](../../adapters/paseto)
- [`adapters/jwt`](../../adapters/jwt)
- [`adapters/session`](../../adapters/session)
- [`adapters/vault`](../../adapters/vault)
- [`utils/structures/claims`](../../utils/structures/claims)

## PASETO

`adapters/paseto` is used for token issuance/validation and middleware verification flows.

Common usage:

- configure key material and issuer options
- issue access/refresh/internal tokens
- validate token and extract claims in middleware

```go
pm := paseto.NewPasetoManager(
	paseto.WithKeys(privateKey, publicKey),
	paseto.WithIssuer("orders-service"),
)

tokenRes := pm.FetchToken(
	claims.WithSubject("user_123"),
	claims.WithAudience("orders-service"),
)
_ = tokenRes
```

## JWT

`adapters/jwt` provides JWT support (often inter-service or legacy integration contexts), including signing and validation helpers.

## Session Management

`adapters/session` provides session manager primitives and middleware integration:

- create/store/fetch/invalidate sessions
- bind session identity to request context
- support Redis-backed persistence

```go
sm, err := session.NewSessionManager(
	session.WithRedisManager(redisMgr),
	session.WithDefaultExpiry(24*time.Hour),
)
if err != nil {
	return err
}

sessionID, err := sm.CreateSession(ctx, session.SessionData{
	UserID:          "user_123",
	BusinessID:      "biz_456",
	IsAuthenticated: true,
})
_ = sessionID
_ = err
```

## Vault

`adapters/vault` wraps secret retrieval patterns and supports secure runtime configuration.

## Claims Structures

`utils/structures/claims` contains claims payload contracts used by token and middleware layers.

## Internal Helpers (implementation detail)

Parsing and validation helper methods inside auth packages are frequently optimized around middleware internals; rely only on exported surface for cross-package contracts.

## Caveats

- Keep token/session expiry strategy explicit and documented per service.
- Avoid mixing PASETO and JWT semantics in a single request path unless required.
- Never log raw secret values or full bearer tokens.
