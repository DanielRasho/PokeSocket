package utils

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Given an struct tagged with Go Validator tags, it validates it.
// If success:
//   - Returns nil, nil
//
// If a validation rule is broken it returns :
//   - A map with an entry for each rule broken, nil
func ValidateStruct(validate *validator.Validate, target any) (map[string]string, error) {
	// Validate the struct
	if err := validate.Struct(target); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorDetails := make(map[string]string)

		// Map validation errors
		for _, vErr := range validationErrors {
			// Construct a human-readable error message
			errorDetails[vErr.Field()] = formatValidationError(vErr)
		}

		return errorDetails, errors.New("validation failed")
	}

	// No validation errors
	return nil, nil
}

// Helper function to format a validation error
func formatValidationError(vErr validator.FieldError) string {
	switch vErr.Tag() {
	case "required":
		return "is required"
	case "max":
		return fmt.Sprintf("must be at most %s", vErr.Param())
	case "min":
		return fmt.Sprintf("must be at least %s", vErr.Param())
	case "email":
		return "must be a valid email address"
	case "email_domain":
		return "must be a valid email domain extension (e.g., @example.com)"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", vErr.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters long", vErr.Param())
	default:
		return fmt.Sprintf("failed validation on rule '%s'", vErr.Tag())
	}
}
