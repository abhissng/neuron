package validator

import (
	"fmt"

	"github.com/asaskevich/govalidator"
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

// ValidateStruct checks the struct against its 'valid' tags.
//
// AVAILABLE GOVALIDATOR TAGS CHEAT SHEET:
//
// --- Logic & Structure ---
// required            : Fails if the field is the zero value (e.g., "", 0, nil).
// optional            : Skips validation if the field is empty/zero value.
// in(val1|val2)       : Value must be exactly one of the listed options.
// type(string)        : Checks if the variable is of the specified type.
//
// --- Strings & Length ---
// length(min|max)     : Validates length of strings, slices, maps, or arrays.
// stringlength(min|max): Strictly validates the length of a string.
// alpha               : Contains only letters (a-zA-Z).
// alphanum            : Contains only letters and numbers.
// numeric             : Contains only numbers (0-9).
// lowercase           : Contains only lowercase letters.
// uppercase           : Contains only uppercase letters.
// ascii               : Contains only ASCII characters.
// printableascii      : Contains only printable ASCII characters.
// hexadecimal         : Contains only hexadecimal values.
// json                : Validates if the string is valid JSON.
//
// --- Network & Web ---
// email               : Validates a standard email address.
// url                 : Validates a URL (http/https not strictly required).
// requrl              : Validates a URL and requires the scheme (http/https).
// ip                  : Validates an IP address (v4 or v6).
// ipv4 / ipv6         : Strictly validates IPv4 or IPv6.
// mac                 : Validates a MAC address.
// host                : Validates a host (either a valid IP or valid DNS).
// dns                 : Validates a DNS name.
// port                : Validates a valid port number (0-65535).
//
// --- Math & Numbers ---
// int                 : Checks if the value is an integer.
// float               : Checks if the value is a floating-point number.
// range(min|max)      : Ensures a numeric value falls within a range.
//
// --- Identifiers & Misc ---
// uuid                : Validates a UUID (v1 to v5).
// uuidv4              : Strictly validates a UUIDv4.
// hexcolor            : Validates a hex color code (e.g., #fff, #000000).
// rgbcolor            : Validates an RGB color (e.g., rgb(255,255,255)).
// creditcard          : Validates standard credit card numbers.
// base64              : Validates if a string is Base64 encoded.
// datauri             : Validates a Base64 data URI.
//
// Example usage: `json:"type" valid:"required,in(new|upgrade|add_ons)"`
func ValidateStruct[T any](data T) error {
	// ValidateStruct returns a boolean and an error.
	// We can ignore the boolean since the error being nil means it's valid.
	_, err := govalidator.ValidateStruct(data)
	return err
}
