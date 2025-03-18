package types

import (
	"crypto/ed25519"
	"encoding/hex"
	"strings"

	"go.uber.org/zap"
)

// StringConstant represents a constant string value.
type StringConstant string

// String returns the string representation of the StringConstant.
func (s StringConstant) String() string {
	return string(s)
}

// RequestID represents a request ID.
type RequestID string

// String returns the string representation of the RequestID.
func (r RequestID) String() string {
	return string(r)
}

// CorrelationID represents a correlation ID.
type CorrelationID string

// String returns the string representat	ion of the CorrelationID.
func (c CorrelationID) String() string {
	return string(c)
}

// ErrorCode represents an error code.
type ErrorCode string

// String returns the string representation of the ErrorCode.
func (e ErrorCode) String() string {
	return string(e)
}

// ResponseErrorType represents the type of response error.
type ResponseErrorType string

// String returns the string representation of the ResponseErrorType.
func (e ResponseErrorType) String() string {
	return string(e)
}

// ComponentErrorType represents the type of component error.
type ComponentErrorType string

// String returns the string representation of the ComponentErrorType.
func (e ComponentErrorType) String() string {
	return string(e)
}

// DBType defines the type of database (e.g., PostgreSQL, MySQL).
type DBType string

// String returns the string representation of the DBType.
func (e DBType) String() string {
	return string(e)
}

// CodecType defines the type of encoder (e.g., JSON, XML).
type CodecType string

// String returns the string representation of the CodecType.
func (e CodecType) String() string {
	return string(e)
}

// Method to convert string to uppercase
func (s CodecType) ToUpperCase() string {
	return strings.ToUpper(string(s))
}

// key defines the type for a key.
type Key string

// Method to convert Key Type to string
func (e Key) String() string {
	return string(e)
}

// Method to convert Key Type to Ed25519PrivateKey
func (e Key) ToEd25519PrivateKey() ed25519.PrivateKey {
	b, _ := hex.DecodeString(e.String())
	return ed25519.PrivateKey(b)
}

// Method to convert Key Type to Ed25519PublicKey
func (e Key) ToEd25519PublicKey() ed25519.PublicKey {
	b, _ := hex.DecodeString(e.String())
	return ed25519.PublicKey(b)
}

// ContentType defines the type for a ContentType.
type ContentType string

// Method to convert ContentType Type to string
func (c ContentType) String() string {
	return string(c)
}

// Field type to represent structured log fields
//
//nolint:gochecknoglobals
type Field = zap.Field

// BusinessID represents a business ID.
type BusinessID int64

// Int64 returns the int64 representation of the BusinessID.
func (e BusinessID) Int64() int64 {
	return int64(e)
}

// UserID represents a user ID.
type UserID int64

// Int64 returns the int64 representation of the UserID.
func (e UserID) Int64() int64 {
	return int64(e)
}

// Milliseconds represents a duration in milliseconds.
type Milliseconds int64

// Int64 returns the int64 representation of the Milliseconds.
func (e Milliseconds) Int64() int64 {
	return int64(e)
}

// Service represents a service.
type Service string

// String returns the string representation of the Service.
func (s Service) String() string {
	return string(s)
}

// Protocol represents a protocol.
type Protocol string

// String returns the string representation of the Protocol.
func (p Protocol) String() string {
	return string(p)
}

// Status represents a status.
type Status string

// String returns the string representation of the Status.
func (s Status) String() string {
	return string(s)
}

// Action represents an action.
type Action string

// String returns the string representation of the Action.
func (a Action) String() string {
	return string(a)
}

// LogMode represents the logging mode
type LogMode string

// String returns the string representation of the LogMode.
func (l LogMode) String() string {
	return string(l)
}
