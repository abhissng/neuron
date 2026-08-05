package helpers

import (
	"github.com/abhissng/neuron/utils/types"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"time"
)

// isEmptyPrimitive checks if a primitive type value is empty.
// It returns (isEmpty, wasHandled) where wasHandled indicates if the type was recognized.
func isEmptyPrimitive(v reflect.Value) (bool, bool) {
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == "", true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0, true
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0, true
	case reflect.Bool:
		return !v.Bool(), true
	}
	return false, false
}

// isEmptyCollection checks if a collection type (slice, map, array, func) is empty.
// It returns (isEmpty, wasHandled) where wasHandled indicates if the type was recognized.
func isEmptyCollection(v reflect.Value) (bool, bool) {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil() || v.Len() == 0, true
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !IsEmpty(v.Index(i).Interface()) {
				return false, true
			}
		}
		return true, true
	}
	return false, false
}

// isEmptyStruct checks if a struct type is empty by recursively checking all fields.
// It handles time.Time as a special case and returns (isEmpty, wasHandled).
func isEmptyStruct(v reflect.Value) (bool, bool) {
	if v.Kind() != reflect.Struct {
		return false, false
	}

	// Check all struct fields recursively
	for i := 0; i < v.NumField(); i++ {
		// Skip unexported fields to avoid panics if necessary,
		// though IsEmpty generally handles interface conversion safely.
		if !IsEmpty(v.Field(i).Interface()) {
			return false, true
		}
	}
	return true, true
}

// isEmptyKnownType checks for specific named types that have well-defined empty states.
// It handles time.Time and uuid.UUID regardless of their underlying structure (Struct vs Array).
func isEmptyKnownType(v reflect.Value) (bool, bool) {
	if !v.CanInterface() {
		return false, false
	}

	switch val := v.Interface().(type) {
	case time.Time:
		return val.IsZero(), true
	case uuid.UUID:
		return val == uuid.Nil, true
	}
	return false, false
}

// IsEmpty checks if the given interface value represents an empty or zero value.
// It supports custom EmptyCheck interface and handles all Go types recursively.
func IsEmpty[T any](value T) bool {
	// Check if value implements EmptyCheck interface
	if v, ok := any(value).(types.EmptyCheck); ok {
		return v.IsEmpty()
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return true
	}

	// Handle pointer and interface types first
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return true
		}
		return IsEmpty(v.Elem().Interface())
	}

	// 1. Check Known Types (UUID, Time) - Added this priority check
	if isEmpty, ok := isEmptyKnownType(v); ok {
		return isEmpty
	}

	// 2. Check primitive types
	if isEmpty, ok := isEmptyPrimitive(v); ok {
		return isEmpty
	}

	// 3. Check collection types
	if isEmpty, ok := isEmptyCollection(v); ok {
		return isEmpty
	}

	// 4. Check struct types
	if isEmpty, ok := isEmptyStruct(v); ok {
		return isEmpty
	}

	// Default: Compare with zero value
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}
