# Helpers and Types Usage

Packages covered:

- [`utils/helpers`](../../utils/helpers)
- [`utils/types`](../../utils/types)
- [`utils/random`](../../utils/random)
- [`utils/constant`](../../utils/constant)
- [`utils/codec`](../../utils/codec)

Primary files:

- [`utils/helpers/helpers.go`](../../utils/helpers/helpers.go)
- [`utils/helpers/convert.go`](../../utils/helpers/convert.go)
- [`utils/helpers/reflect.go`](../../utils/helpers/reflect.go)
- [`utils/types/types.go`](../../utils/types/types.go)
- [`utils/types/helper.go`](../../utils/types/helper.go)
- [`utils/types/pgtype.go`](../../utils/types/pgtype.go)
- [`utils/random/random.go`](../../utils/random/random.go)

## Helpers

`utils/helpers` contains cross-cutting helper sets:

- environment/config value resolution
- conversion and map/struct transformations
- reflection-based emptiness checks
- network and URL helpers
- i18n utility functions

Recent structure is split by theme (`helpers.go`, `env.go`, `convert.go`, `reflect.go`, `network.go`, `i18n.go`) while preserving package APIs.

```go
env := helpers.GetEnvironment()
slug := helpers.GetEnvironmentSlug(env)

m, _ := helpers.StructToMap(payload)
decoded, _ := helpers.MapToStruct[MyType](m)
_ = decoded

empty := helpers.IsEmpty(payload)
_ = slug
_ = empty
```

## Types

`utils/types` defines shared domain types and conversion helpers:

- typed identifiers and enums
- status/action/environment value contracts
- pg-type and generic helper conversions

This is one of the highest API-density packages and should be treated as foundational contract surface.

## Random

`utils/random` provides:

- UUID helpers (including option-based UUID v7 fallback patterns)
- random alphanumeric/number/token generation

```go
uuidV4OrV7 := random.GenerateUUID(random.WithUUIDVersion7())
requestID := random.GenerateUUIDString()
token, _ := random.GenerateTokenID()
_ = uuidV4OrV7
_ = requestID
_ = token
```

## Constants and Codec

- `utils/constant` centralizes string/enum constants.
- `utils/codec` contains encoding/decoding related primitives.

## Internal Helpers (implementation detail)

Some helper functions are intentionally unexported and serve package-internal decomposition; do not depend on them externally.

## Caveats

- Prefer typed wrappers from `utils/types` over raw strings in service boundaries.
- Keep helper usage explicit in critical paths to avoid hidden conversions.
