// Package blame provides a custom error type that adds additional information and functionality to standard errors.
package blame

import (
	"github.com/abhissng/neuron/utils/types"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Blame represents a custom error type that provides additional information and functionality.
type Blame interface {
	// error is embedded to ensure Blame implements the error interface.
	error

	// FetchReasonCode returns the reason code associated with the error.
	FetchReasonCode() string

	// FetchErrCode returns the error code associated with the error.
	FetchErrCode() types.ErrorCode

	// FetchMessage returns the error message.
	FetchMessage() string

	// WithMessage sets the error message and returns the updated Blame instance.
	WithMessage(string) *Error

	// WithDescription sets the error description and returns the updated Blame instance.
	WithDescription(string) *Error

	// FetchMessage returns the error description.
	FetchDescription() string

	// WithMessageDescription sets the error message and description and returns the updated Blame instance.
	WithMessageDescription(string, string) *Error

	// FetchFields returns a map of additional error fields.
	FetchFields() map[string]any

	// FetchSource returns the source of the error.
	FetchSource() string

	// FetchComponent returns the component associated with the error.
	FetchComponent() types.ComponentErrorType

	// FetchResponseType returns the response type associated with the error.
	FetchResponseType() types.ResponseErrorType

	// FetchCauses returns a slice of underlying errors that caused this error.
	FetchCauses() []error

	// FetchBundle returns a local bundle for internal error transformation.
	FetchBundle() *i18n.Bundle

	// WithBundle adds a new bundle to the error and returns the updated Blame instance.
	WithBundle(bundle *i18n.Bundle) *Error

	// WithLanguageTag adds a new language to the error and returns the updated Blame instance.
	WithLanguageTag(language types.LanguageTag) *Error

	// FetchLanguageTag returns a types.LanguageTag for the error instance
	FetchLanguageTag() types.LanguageTag

	// WithField adds a new field to the error and returns the updated Blame instance.
	WithField(key string, value any) *Error

	// WithCause adds a new underlying error to the error and returns the updated Blame instance.
	WithCause(err error) *Error

	// WrapToError creates a new error instance with the provided fields and cause.
	WrapToError(opts ...BlameOption) *Error

	// WithComponent sets the component associated with the error and returns the updated Blame instance.
	WithComponent(component types.ComponentErrorType) *Error

	// WithResponseType sets the response type associated with the error and returns the updated Blame instance.
	WithResponseType(responseType types.ResponseErrorType) *Error

	// Translate translates the error message and description using the provided i18n bundle and language in the error instance.
	Translate() (string, string)

	// WithFields adds multiple fields to the error and returns the updated Blame instance.
	WithFields(fields map[string]any) *Error

	// FetchErrorResponse returns a map representing the error response.
	FetchErrorResponse(options ...SendErrorResponseOption) ErrorResponse

	// Wrap wrapes the basicBlame to Blame
	Wrap(opts ...BlameOption) Blame

	// EmptyCause sets the causes of the error to an empty slice and returns the updated Error instance.
	EmptyCause() Blame

	// ErrorFromBlame creates a new error  string from a Blame instance.
	ErrorFromBlame() error
}

// NewBlame creates a new instance of Blame with the provided reason code, error code, and message. It captures the source of the error at the point of instantiation.
func NewBlame(
	reasonCode string,
	errCode types.ErrorCode,
	message, description string,
) Blame {
	return NewError(reasonCode, errCode, message, description)
}

// NewBasicBlame creates a new instance of Blame with the provided error code. It captures the source of the error at the point of instantiation.
func NewBasicBlame(
	errCode types.ErrorCode,
) Blame {
	return NewBasicError(errCode)
}

// NilBlame returns a nil blame
func NilBlame() Blame {
	return nil
}
