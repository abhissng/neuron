package helpers

import (
	"encoding/json"
	"fmt"
	"math"
)

// RoundToIntAmount converts a currency amount in major units to smallest currency
// units by multiplying by 100 and rounding (for example, 699.99 → 69999).
func RoundToIntAmount(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

const (
	bytesKindString  = "string"
	bytesKindBool    = "bool"
	bytesKindInt64   = "int64"
	bytesKindUint64  = "uint64"
	bytesKindFloat64 = "float64"
	bytesKindBytes   = "bytes"
	bytesKindJSON    = "json"
)

type bytesEnvelope struct {
	Kind string          `json:"k"`
	V    json.RawMessage `json:"v"`
}

func marshalEnvelope(kind string, v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(bytesEnvelope{Kind: kind, V: raw})
}

// ToBytes converts v into a typed byte envelope suitable for encryption.
// Strings, numeric types, bools, and byte slices round-trip through FromBytes
// without trimming or type inference. Other values are JSON-encoded under kind "json".
func ToBytes(v any) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot convert nil value to bytes")
	}

	switch t := v.(type) {
	case []byte:
		return marshalEnvelope(bytesKindBytes, t)
	case string:
		return marshalEnvelope(bytesKindString, t)
	case bool:
		return marshalEnvelope(bytesKindBool, t)
	case int:
		return marshalEnvelope(bytesKindInt64, int64(t))
	case int8:
		return marshalEnvelope(bytesKindInt64, int64(t))
	case int16:
		return marshalEnvelope(bytesKindInt64, int64(t))
	case int32:
		return marshalEnvelope(bytesKindInt64, int64(t))
	case int64:
		return marshalEnvelope(bytesKindInt64, t)
	case uint:
		return marshalEnvelope(bytesKindUint64, uint64(t))
	case uint8:
		return marshalEnvelope(bytesKindUint64, uint64(t))
	case uint16:
		return marshalEnvelope(bytesKindUint64, uint64(t))
	case uint32:
		return marshalEnvelope(bytesKindUint64, uint64(t))
	case uint64:
		return marshalEnvelope(bytesKindUint64, t)
	case float32:
		return marshalEnvelope(bytesKindFloat64, float64(t))
	case float64:
		return marshalEnvelope(bytesKindFloat64, t)
	case json.Number:
		return marshalEnvelope(bytesKindString, t.String())
	default:
		return marshalEnvelope(bytesKindJSON, t)
	}
}

// FromBytes restores a value produced by ToBytes using the embedded type tag.
// It returns an error when the payload is not a valid envelope.
func FromBytes(b []byte) (any, error) {
	var env bytesEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("invalid bytes envelope: %w", err)
	}

	switch env.Kind {
	case bytesKindString:
		var s string
		if err := json.Unmarshal(env.V, &s); err != nil {
			return nil, err
		}
		return s, nil
	case bytesKindBool:
		var v bool
		if err := json.Unmarshal(env.V, &v); err != nil {
			return nil, err
		}
		return v, nil
	case bytesKindInt64:
		var v int64
		if err := json.Unmarshal(env.V, &v); err != nil {
			return nil, err
		}
		return v, nil
	case bytesKindUint64:
		var v uint64
		if err := json.Unmarshal(env.V, &v); err != nil {
			return nil, err
		}
		return v, nil
	case bytesKindFloat64:
		var v float64
		if err := json.Unmarshal(env.V, &v); err != nil {
			return nil, err
		}
		return v, nil
	case bytesKindBytes:
		var v []byte
		if err := json.Unmarshal(env.V, &v); err != nil {
			return nil, err
		}
		return v, nil
	case bytesKindJSON:
		var v any
		if err := json.Unmarshal(env.V, &v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unknown bytes envelope kind: %q", env.Kind)
	}
}
