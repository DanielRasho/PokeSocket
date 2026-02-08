package utils

import "net/http"

type VerificationError struct {
	Err       error
	Code      *DefaultMsg
	UserError map[string]string
}

func (e *VerificationError) Error() string {
	return e.Err.Error()
}

// This file provides Constant definition to API RESPONSES

// Structure of API Messages responses
type DefaultMsg struct {
	Message    string
	StatusCode int
}

func (e *DefaultMsg) Error() string {
	return e.Message
}

// List of Common Response messages for different kinf of errors.
// IS PREFERABLE that ALL common response messages reside here.
var (
	// 200: Success
	RequestSuccess = &DefaultMsg{"Request successful.", http.StatusOK}

	// 400: Missing data on request
	BadRequest      = &DefaultMsg{"Bad request", http.StatusBadRequest}
	WrongFieldsType = &DefaultMsg{"Fields of wrong type", http.StatusBadRequest}
	FailedToDecode  = &DefaultMsg{"Failed to decode request", http.StatusBadRequest}
	FailedToEncode  = &DefaultMsg{"Failed to encode request", http.StatusBadRequest}
	InvalidFields   = &DefaultMsg{"Request contains invalid fields.", http.StatusBadRequest}

	NotFound = &DefaultMsg{"Forbidden", http.StatusNotFound}

	// 404: Resource not found
	ResourceNotFound = &DefaultMsg{"Resource not found for given parameters", http.StatusNotFound}

	// 500: Internal Server Error
	BadDatabaseOperation = &DefaultMsg{"Failed to execute DB operations.", http.StatusInternalServerError}
	InternalServerError  = &DefaultMsg{"Internal server error", http.StatusInternalServerError}
)
