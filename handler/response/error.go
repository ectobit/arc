package response

import (
	"errors"
	"net/http"

	"go.ectobit.com/arc/handler/request"
	"go.ectobit.com/lax"
)

// Error is used to render JSON error.
type Error struct {
	Error string `json:"error"`
}

func RenderError(res http.ResponseWriter, err error, log lax.Logger) {
	responseErr := &request.Error{} //nolint:exhaustivestruct

	if errors.As(err, &responseErr) {
		Render(res, responseErr.StatusCode, &Error{Error: responseErr.Error()}, log)

		return
	}

	Render(res, http.StatusInternalServerError, nil, log)
}
