package request

import (
	"net/http"
	"strings"
)

// Error contains status code and error message to be returned to the client.
type Error struct {
	StatusCode int
	Message    string
}

// NewBadRequestError creates error with status bad request.
func NewBadRequestError(message string) *Error {
	return &Error{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

// NewInternalServerError creates error with status internal server error.
func NewInternalServerError() *Error {
	return &Error{
		StatusCode: http.StatusInternalServerError,
		Message:    strings.ToLower(http.StatusText(http.StatusInternalServerError)),
	}
}

// Error implements error interface.
func (e *Error) Error() string {
	return e.Message
}
