package helpers

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// RoundToIntAmount rounds a float64 amount to the nearest integer.
// For example, 69900 translates to 699.
func RoundToIntAmount(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

// ToBytes converts any supported value into a []byte payload suitable for encryption.
//
// Conversion rules:
//   - nil           → error
//   - []byte        → returned as-is
//   - string        → []byte(s)
//   - bool          → "true" / "false"
//   - integers      → decimal text (strconv)
//   - floats        → shortest round-trip text (strconv)
//   - map / slice / struct / other → JSON encoding
func ToBytes(v any) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot convert nil value to bytes")
	}

	switch t := v.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	case bool:
		return []byte(strconv.FormatBool(t)), nil
	case int:
		return []byte(strconv.Itoa(t)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(t), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(t), 10)), nil
	case int32:
		return []byte(strconv.FormatInt(int64(t), 10)), nil
	case int64:
		return []byte(strconv.FormatInt(t, 10)), nil
	case uint:
		return []byte(strconv.FormatUint(uint64(t), 10)), nil
	case uint8:
		return []byte(strconv.FormatUint(uint64(t), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(t), 10)), nil
	case uint32:
		return []byte(strconv.FormatUint(uint64(t), 10)), nil
	case uint64:
		return []byte(strconv.FormatUint(t, 10)), nil
	case float32:
		return []byte(strconv.FormatFloat(float64(t), 'g', -1, 32)), nil
	case float64:
		return []byte(strconv.FormatFloat(t, 'g', -1, 64)), nil
	case json.Number:
		return []byte(t.String()), nil
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %T to bytes: %w", v, err)
		}
		return b, nil
	}
}

// FromBytes reverses ToBytes for decrypted payloads.
//
// Restoration rules:
//   - JSON object / array → map[string]any / []any (not escaped string)
//   - integer text       → int64
//   - float text         → float64
//   - bool text          → bool
//   - everything else    → string
func FromBytes(b []byte) any {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return ""
	}

	first, last := s[0], s[len(s)-1]
	if (first == '{' && last == '}') || (first == '[' && last == ']') {
		var v any
		if err := json.Unmarshal([]byte(s), &v); err == nil {
			return v
		}
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	if bv, err := strconv.ParseBool(s); err == nil {
		return bv
	}

	return s
}
