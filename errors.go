package hago

import (
	"errors"
	"fmt"
)

// Common errors returned by the client.
var (
	// ErrNoBaseURL is returned when no base URL is provided.
	ErrNoBaseURL = errors.New("base URL is required")

	// ErrNoToken is returned when no authentication token is provided.
	ErrNoToken = errors.New("authentication token is required")

	// ErrNotFound is returned when the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized is returned when authentication fails.
	ErrUnauthorized = errors.New("unauthorized: invalid or missing token")

	// ErrBadRequest is returned when the request is malformed.
	ErrBadRequest = errors.New("bad request")

	// ErrMethodNotAllowed is returned when the HTTP method is not supported.
	ErrMethodNotAllowed = errors.New("method not allowed")
)

// APIError represents an error response from the Home Assistant API.
type APIError struct {
	StatusCode int
	Status     string
	Message    string
	Body       string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Status, e.Message)
	}
	if e.Body != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Status, e.Body)
	}
	return fmt.Sprintf("API error %d (%s)", e.StatusCode, e.Status)
}

// RequestError represents an error that occurred while making a request.
type RequestError struct {
	Op  string
	Err error
}

// Error implements the error interface.
func (e *RequestError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error.
func (e *RequestError) Unwrap() error {
	return e.Err
}
