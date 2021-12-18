package response

import (
	"encoding/json"
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
	reqErr := &request.Error{} //nolint:exhaustivestruct

	if errors.As(err, &reqErr) {
		Render(res, reqErr.StatusCode, &Error{Error: reqErr.Error()}, log)

		return
	}

	Render(res, http.StatusInternalServerError, nil, log)
}

func RenderErrorStatus(res http.ResponseWriter, statusCode int, message string, log lax.Logger) {
	Render(res, statusCode, &Error{Error: message}, log)
}

// Render renders HTTP response with JSON body.
func Render(res http.ResponseWriter, statusCode int, body interface{}, log lax.Logger) {
	res.Header().Set("Content-Type", "application/json")

	if body == nil || body == http.NoBody {
		res.WriteHeader(statusCode)

		return
	}

	data, err := json.Marshal(body)
	if err != nil {
		log.Warn("json marshal", lax.Error(err))
		res.WriteHeader(http.StatusInternalServerError)

		return
	}

	res.WriteHeader(statusCode)

	if _, err := res.Write(data); err != nil {
		log.Warn("response write", lax.Error(err))
	}
}
