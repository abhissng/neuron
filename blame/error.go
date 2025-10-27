package blame

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Error struct holds the error information
type Error struct {
	reasonCode   string          //123
	errCode      types.ErrorCode //err-not-found
	component    types.ComponentErrorType
	responseType types.ResponseErrorType
	message      string
	description  string
	fields       map[string]any
	causes       []error
	source       string
	bundle       *i18n.Bundle
	bundleSet    bool
	language     types.LanguageTag
}

// NewError creates a new Error instance
func NewError(
	reasonCode string,
	errorCode types.ErrorCode,
	message, description string,
) *Error {
	if helpers.IsEmpty(reasonCode) {
		reasonCode = string(errorCode)
	}
	return &Error{
		reasonCode:  reasonCode,
		errCode:     errorCode,
		message:     message,
		description: description,
		language:    helpers.GetDefaultLanguageTag(),
		fields:      map[string]any{},
		causes:      make([]error, 0),
		source:      findSource(),
		bundle:      helpers.NewBundle(helpers.GetDefaultLanguageTag()),
	}
}

// NewBasicError creates a new Error instance with the given error code and status code
func NewBasicError(
	errorCode types.ErrorCode,
) *Error {
	return &Error{
		reasonCode: errorCode.String(),
		errCode:    errorCode,
		language:   helpers.GetDefaultLanguageTag(),
		fields:     map[string]any{},
		causes:     make([]error, 0),
		source:     findSource(),
		bundle:     helpers.NewBundle(helpers.GetDefaultLanguageTag()),
	}
}

// FetchReasonCode returns the reason code of the error as a string
func (e *Error) FetchReasonCode() string {
	return e.reasonCode
}

// FetchErrCode returns the error code of the error as a ErrorCode
func (e *Error) FetchErrCode() types.ErrorCode {
	return e.errCode
}

// FetchMessage returns the message of the error as a string
func (e *Error) FetchMessage() string {
	return e.message
}

// FetchDescription returns the description of the error as a string
func (e *Error) FetchDescription() string {
	return e.description
}

// WithLanguageTag sets the language tag of the error and returns the updated Error instance.
func (e *Error) WithLanguageTag(language types.LanguageTag) *Error {
	e.language = language
	return e
}

// FetchLanguageTag returns the language tag of the error as a LanguageTag
func (e *Error) FetchLanguageTag() types.LanguageTag {
	return e.language
}

// FetchBundle returns the bundle of the error as a *i18n.Bundle
func (e *Error) FetchBundle() *i18n.Bundle {
	return e.bundle
}

// WithBundle sets the bundle of the error and returns the updated Error instance.
func (e *Error) WithBundle(localBundle *i18n.Bundle) *Error {
	e.bundle = localBundle
	e.bundleSet = true
	return e
}

// WithMessage sets the message of the error and returns the updated Error instance.
func (e *Error) WithMessage(msg string) *Error {
	e.message = msg
	return e
}

// WithDescription sets the description of the error and returns the updated Error instance.
func (e *Error) WithDescription(description string) *Error {
	e.description = description
	return e
}

// WithMessageDescription sets the message and description of the error and returns the updated Error instance.
func (e *Error) WithMessageDescription(message, description string) *Error {
	e.message = message
	e.description = description
	return e
}

// FetchFields returns the fields of the error as a map[string]any
func (e *Error) FetchFields() map[string]any {
	return e.fields
}

// FetchSource returns the source of the error as a string
func (e *Error) FetchSource() string {
	return e.source
}

// FetchComponent returns the component of the error as a ComponentErrorType
func (e *Error) FetchComponent() types.ComponentErrorType {
	return e.component
}

// FetchResponseType returns the response type of the error as a ResponseErrorType
func (e *Error) FetchResponseType() types.ResponseErrorType {
	return e.responseType
}

// FetchCauses returns the causes of the error as a slice of errors
func (e *Error) FetchCauses() []error {
	return e.causes
}

// EmptyCause sets the causes of the error to an empty slice and returns the updated Error instance.
func (e *Error) EmptyCause() Blame {
	e.causes = make([]error, 0)
	return e
}

// WithField adds a field to the error and returns the updated Error instance.
func (e *Error) WithField(key string, value any) *Error {
	e.fields[key] = value
	return e
}

// WithFields adds multiple fields to the error and returns the updated Error instance.
func (e *Error) WithFields(fields map[string]any) *Error {
	for k, v := range fields {
		e.fields[k] = v
	}
	return e
}

// WithCause adds a cause to the error and returns the updated Error instance.
func (e *Error) WithCause(err error) *Error {
	if len(e.causes) <= 0 {
		e.causes = make([]error, 0)
	}
	e.causes = append(e.causes, err)
	return e
}

// WithComponent sets the component of the error and returns the updated Error instance.
func (e *Error) WithComponent(component types.ComponentErrorType) *Error {
	e.component = component
	return e
}

// WithResponseType sets the response type of the error and returns the updated Error instance.
func (e *Error) WithResponseType(responseType types.ResponseErrorType) *Error {
	e.responseType = responseType
	return e
}

// Error returns the error message with the causes as a string
func (e *Error) Error() string {
	return fmt.Sprintf("%s (causes: %v)", e.errCode.String(), e.causes)
}

// findSource captures the source of the error at the point of instantiation.
func findSource() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", strings.TrimPrefix(file, helpers.GetGoROOT()+"/src/"), line)
}

// FindErrorDefinition searches a slice of *Error instances for an instance with the provided error code.
func FindErrorDefinition(errors []*Error, errorCode string) *Error {
	for _, err := range errors {
		if err.errCode.String() == errorCode {
			return err
		}
	}
	return nil
}

// WrapToError creates a new Error instance with the current error's properties.
func (e *Error) WrapToError(opts ...BlameOption) *Error {
	// Create a new BlameOptions struct
	options := NewBlameOptions()

	// Apply existing fields and causes to the options
	if len(e.fields) > 0 {
		for k, v := range e.fields {
			options.Fields[k] = v
		}
	}

	if len(e.causes) > 0 {
		options.Causes = append(options.Causes, e.causes...)
	}

	// Apply additional options
	for _, opt := range opts {
		opt(options)
	}

	e.fields = options.Fields
	e.causes = options.Causes
	return e
}

// Wrap wraps the error with the provided options and returns the updated Blame instance.
func (e *Error) Wrap(opts ...BlameOption) Blame {
	return e.WrapToError(opts...)
}

// ErrorFromBlame creates a new error from a Blame instance.
func (e *Error) ErrorFromBlame() error {
	return errors.New(helpers.FetchErrorStack(e.FetchCauses()))
}

// func (e *Error) Translate(bundle *i18n.Bundle, lang string) string,string {
// Translate transaltes the message and description and return the localized Message and Description
func (e *Error) Translate() (string, string) {
	// Replace placeholders in message with actual values
	message := e.message
	description := e.description
	for key, value := range e.fields {
		formatedValue := "[" + fmt.Sprintf("%v", value) + "]"
		message = strings.ReplaceAll(message, "{{."+key+"}}", formatedValue)
		description = strings.ReplaceAll(description, "{{."+key+"}}", formatedValue)
	}

	// Localize the message using i18n

	if e.bundle == nil {
		_ = e.WithBundle(helpers.NewBundle(types.LanguageTag{}))
		_ = e.WithLanguageTag(types.LanguageTag(language.English))
	}
	if e.bundle != nil && !helpers.IsEmpty(e.language) {
		localizer := i18n.NewLocalizer(e.bundle, e.language.String())
		localizedMessage, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:          e.errCode.String(), // Use errCode as message ID
				Other:       message,
				Description: description, // Fallback description
			},
			TemplateData: e.fields,
		})
		if err != nil {
			helpers.Println(constant.ERROR, "Error localizing message: ", err)
			return message, description // Fallback to original message and description
		}

		localizedDescription, err := localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    e.errCode.String() + ".description", // Use errCode as message ID
				Other: description,                         // Fallback message
			},
			TemplateData: e.fields,
		})

		if err != nil {
			helpers.Println(constant.ERROR, "Error localizing description: ", err)
			return localizedMessage, description // Fallback to localized message and  original description
		}
		return localizedMessage, localizedDescription
	}

	return message, description // Return original message and description if no bundle or lang is provided
}

// ErrorResponse struct holds the error information for sending as a response
type ErrorResponse struct {
	ReasonCode   string                   `json:"reason_code,omitempty"`
	ErrorCode    types.ErrorCode          `json:"error_code,omitempty"`
	Message      string                   `json:"message,omitempty"`
	Description  string                   `json:"description,omitempty"`
	Fields       map[string]any           `json:"fields,omitempty"`
	Component    types.ComponentErrorType `json:"component,omitempty"`
	ResponseType types.ResponseErrorType  `json:"response_type,omitempty"`
	Causes       []string                 `json:"causes,omitempty"`
}

// NewErrorResponseBlame creates a new Blame instance from the ErrorResponse
func (e *ErrorResponse) NewErrorResponseBlame(bw *BlameManager) Blame {
	var blameInfo Blame
	ok := true
	islocalWrapper := true

	if bw != nil {
		var exists bool
		blameInfo, exists = bw.BlameDefinitions[e.ErrorCode]
		if exists && blameInfo.FetchErrCode() == e.ErrorCode {
			islocalWrapper = false
		}
	}

	if islocalWrapper {
		blameInfo, ok = localBlameManager.BlameDefinitions[e.ErrorCode]
	}

	if !ok {
		blameInfo = NewBlame(e.ReasonCode, e.ErrorCode, e.Message, e.Description).
			WithComponent(e.Component).
			WithResponseType(e.ResponseType).
			WithBundle(helpers.NewBundle(helpers.GetDefaultLanguageTag()))
	}
	_ = blameInfo.EmptyCause()

	err := make([]error, 0)
	for i := 0; i < len(e.Causes); i++ {
		err = append(err, errors.New(e.Causes[i]))
		_ = blameInfo.WithCause(err[i])
	}
	_ = blameInfo.WithFields(e.Fields)

	return blameInfo
}

// FetchErrorResponse sends the error as a map[string]any with translated message
func (err *Error) FetchErrorResponse(options ...SendErrorResponseOption) ErrorResponse {
	response := ErrorResponse{
		ReasonCode:   err.FetchReasonCode(),
		ErrorCode:    err.FetchErrCode(),
		Message:      err.FetchMessage(),
		Description:  err.FetchDescription(),
		Fields:       err.FetchFields(),
		Component:    err.FetchComponent(),
		ResponseType: err.FetchResponseType(),
		Causes:       helpers.FetchErrorStrings(err.FetchCauses()),
	}

	for _, opt := range options {
		opt(&response, err)
	}

	return response
}

// SendErrorResponseOption is a function that can be used to modify the error response
type SendErrorResponseOption func(*ErrorResponse, Blame)

// WithTranslation translates the error message and description
func WithTranslation() SendErrorResponseOption {
	return func(response *ErrorResponse, err Blame) {
		response.Message, response.Description = err.Translate()
	}
}

// WithCustomField adds a custom field to the error response and returns the updated SendErrorResponseOption.
func WithCustomField(key string, value any) SendErrorResponseOption {
	return func(response *ErrorResponse, _ Blame) {
		response.Fields[key] = value
	}
}
