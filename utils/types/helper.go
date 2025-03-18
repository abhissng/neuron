package types

import (
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
func CastTo[T any](value any) (T, bool) {
	castValue, ok := value.(T)
	return castValue, ok
}
