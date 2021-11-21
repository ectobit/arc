package public

import (
	"net/http"
	"strings"
)

// Error contains status code and error message to be returned to the client.
type Error struct {
	StatusCode int
	Message    string
}

// NewError creates error.
func NewError(statusCode int, message string) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewBadRequestError creates error with status bad request.
func NewBadRequestError(message string) *Error {
	return &Error{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

// NewInternalServerError creates error with status internal server error.
func NewInternalServerError(message string) *Error {
	return &Error{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

// ErrorFromStatusCode creates error with message from status code.
func ErrorFromStatusCode(statusCode int) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    strings.ToLower(http.StatusText(statusCode)),
	}
}

// Error implements error interface.
func (e *Error) Error() string {
	return e.Message
}
