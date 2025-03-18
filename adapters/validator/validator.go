package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validator is a high-level wrapper for go-playground/validator.
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

// ValidateStruct validates a struct and returns a map of field names to error messages.
func (v *Validator) ValidateStruct(s interface{}) map[string]string {
	err := v.validator.Struct(s)
	if err == nil {
		return nil // No errors
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{"error": "unexpected validation error"}
	}

	errorMap := make(map[string]string)
	for _, fieldError := range validationErrors {
		fieldName := fieldError.Field()
		errorMessage := v.getErrorMessage(fieldError)
		errorMap[fieldName] = errorMessage
	}

	return errorMap
}

// ValidateField validates a single field of a struct.
func (v *Validator) ValidateField(field interface{}, tag string) string {
	err := v.validator.Var(field, tag)
	if err == nil {
		return "" // No error
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return "unexpected validation error"
	}

	if len(validationErrors) > 0 {
		return v.getErrorMessage(validationErrors[0])
	}

	return "validation error" // Generic error if no FieldError found.
}

// RegisterValidation registers a custom validation function for a specific tag.
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}

// RegisterStructValidation registers a custom struct-level validation function.
func (v *Validator) RegisterStructValidation(fn validator.StructLevelFunc, types ...interface{}) {
	v.validator.RegisterStructValidation(fn, types...)
}

// getErrorMessage generates a user-friendly error message from a FieldError.
func (v *Validator) getErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldError.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldError.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fieldError.Field(), fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fieldError.Field(), fieldError.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fieldError.Field(), fieldError.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fieldError.Field(), fieldError.Param())
	default:
		return fmt.Sprintf("invalid %s", fieldError.Field())
	}
}
