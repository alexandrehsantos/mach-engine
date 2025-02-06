package errors

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

// Common errors
var (
	ErrBadRequest = &APIError{
		Status:  http.StatusBadRequest,
		Code:    "BAD_REQUEST",
		Message: "Invalid request",
	}

	ErrNotFound = &APIError{
		Status:  http.StatusNotFound,
		Code:    "NOT_FOUND",
		Message: "Resource not found",
	}

	ErrInternal = &APIError{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_ERROR",
		Message: "Internal server error",
	}
)

// Error constructors
func NewBadRequest(message string) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "BAD_REQUEST",
		Message: message,
	}
}

func NewNotFound(resource string) *APIError {
	return &APIError{
		Status:  http.StatusNotFound,
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
	}
}

func NewInternal(err error) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_ERROR",
		Message: "Internal server error",
	}
}
