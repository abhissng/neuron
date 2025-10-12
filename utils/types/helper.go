package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/language"
)

// EmptyCheck defines an interface for checking if a value is empty.
type EmptyCheck interface {
	IsEmpty() bool
}

// Implement EmptyCheck for common types
type String string

// IsEmpty checks if the string is empty or contains only whitespace
func (s String) IsEmpty() bool {
	return strings.TrimSpace(string(s)) == ""
}

// Implement EmptyCheck for common types
type Int int

// IsEmpty checks if the integer is empty
func (i Int) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Int8 int8

// IsEmpty checks if the integer is empty
func (i Int8) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Int16 int16

// IsEmpty checks if the integer is empty
func (i Int16) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Int32 int32

// IsEmpty checks if the integer is empty
func (i Int32) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Int64 int64

// IsEmpty checks if the integer is empty
func (i Int64) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Uint uint

// IsEmpty checks if the unsigned integer is empty
func (i Uint) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Uint8 uint8

// IsEmpty checks if the unsigned integer is empty
func (i Uint8) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Uint16 uint16

// IsEmpty checks if the unsigned integer is empty
func (i Uint16) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Uint32 uint32

// IsEmpty checks if the unsigned integer is empty
func (i Uint32) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Uint64 uint64

// IsEmpty checks if the unsigned integer is empty
func (i Uint64) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Float32 float32

// IsEmpty checks if the float is empty
func (i Float32) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Float64 float64

// IsEmpty checks if the float is empty
func (i Float64) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Bool bool

// IsEmpty checks if the bool is empty
func (i Bool) IsEmpty() bool {
	return !bool(i)
}

// Implement EmptyCheck for common types
type Byte byte

// IsEmpty checks if the byte is empty
func (i Byte) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type Rune rune

// IsEmpty checks if the rune is empty
func (i Rune) IsEmpty() bool {
	return i == 0
}

// Implement EmptyCheck for common types
type LanguageTag language.Tag

// IsEmpty checks if the language tag is empty
func (l LanguageTag) IsEmpty() bool {
	// Compare the tag to its zero value
	return l == LanguageTag{}
}

// Convert LanguageTag to language.Tag
func ToLanguageTag(l LanguageTag) language.Tag {
	return language.Tag(l)
}

// Convert LanguageTag to string
func (l LanguageTag) String() string {
	// Convert LanguageTag to language.Tag and use its String() method
	return language.Tag(l).String()
}

// CreateRef generates a pointer for any given data.
func CreateRef[T any](value T) *T {
	return &value
}

// ConvertUUIDToBytesRef converts a UUID pointer to a byte slice pointer.
func ConvertUUIDToBytesRef(uuidRef *uuid.UUID) *[]byte {
	if uuidRef == nil {
		return nil
	}
	return CreateRef(uuidRef[:])
}

// ConvertUUIDToBytes converts a UUID to a byte slice.
func ConvertUUIDToBytes(uuidVal uuid.UUID) []byte {
	return uuidVal[:]
}

// BytesToUUID converts a byte slice to a UUID.
func BytesToUUID(byteData []byte) uuid.UUID {
	return uuid.UUID(byteData)
}

// BytesRefToUUIDRef converts a byte slice pointer to a UUID pointer.
func BytesRefToUUIDRef(byteDataRef *[]byte) *uuid.UUID {
	if byteDataRef == nil {
		return nil
	}
	return CreateRef(uuid.UUID(*byteDataRef))
}

// MillisToTime converts milliseconds to a time.Time value.
func MillisToTime(millis int64) time.Time {
	return time.Unix(0, millis*int64(time.Millisecond))
}

// MillisRefToTimeRef converts a milliseconds pointer to a time.Time pointer.
func MillisRefToTimeRef(millisRef *int64) *time.Time {
	if millisRef == nil {
		return nil
	}
	return CreateRef(time.Unix(0, *millisRef*int64(time.Millisecond)))
}

// CastTo is a generic function to safely cast a value to a specific type.
// It returns the cast value and a boolean indicating whether the cast was successful.
// CastTo tries to cast or convert any value to the given type T.
func CastTo[T any](value any) (T, bool) {
	var zero T

	// If value is already the right type
	if v, ok := value.(T); ok {
		return v, true
	}

	// Handle conversions via reflection
	val := reflect.ValueOf(value)
	targetType := reflect.TypeOf(zero)

	if !val.IsValid() {
		return zero, false
	}

	// Handle string conversions
	switch targetType.Kind() {
	case reflect.String:
		switch v := value.(type) {
		case string:
			return any(v).(T), true
		case fmt.Stringer:
			return any(v.String()).(T), true
		default:
			return any(fmt.Sprintf("%v", v)).(T), true
		}

	case reflect.Bool:
		switch v := value.(type) {
		case bool:
			return any(v).(T), true
		case string:
			b, err := strconv.ParseBool(v)
			if err == nil {
				return any(b).(T), true
			}
		case int, int64, float64:
			return any(val.Float() != 0).(T), true
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case string:
			i, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				return any(reflect.ValueOf(i).Convert(targetType).Interface()).(T), true
			}
		case float64, float32:
			return any(reflect.ValueOf(int64(val.Float())).Convert(targetType).Interface()).(T), true
		case int, int8, int16, int32, int64:
			return any(reflect.ValueOf(v).Convert(targetType).Interface()).(T), true
		}

	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case string:
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				return any(reflect.ValueOf(f).Convert(targetType).Interface()).(T), true
			}
		case int, int64:
			return any(reflect.ValueOf(float64(val.Int())).Convert(targetType).Interface()).(T), true
		case float32, float64:
			return any(reflect.ValueOf(v).Convert(targetType).Interface()).(T), true
		}
	}

	// Last resort: try JSON re-marshal
	b, err := json.Marshal(value)
	if err == nil {
		var t T
		if err := json.Unmarshal(b, &t); err == nil {
			return t, true
		}
	}

	return zero, false
}
