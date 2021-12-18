package render

import (
	"encoding/json"
	"net/http"

	"go.ectobit.com/lax"
)

var _ Renderer = (*JSON)(nil)

// JSON implements render.Renderer interface using JSON encoding.
type JSON struct {
	log lax.Logger
}

// NewJSON creates JSON renderer.
func NewJSON(log lax.Logger) *JSON {
	return &JSON{
		log: log,
	}
}

// Render renders HTTP response with JSON body.
// Deprecated.
func (r *JSON) Render(res http.ResponseWriter, statusCode int, body interface{}) {
	res.Header().Set("Content-Type", "application/json")

	if body == nil || body == http.NoBody {
		res.WriteHeader(statusCode)

		return
	}

	data, err := json.Marshal(body)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)

		return
	}

	res.WriteHeader(statusCode)

	if _, err := res.Write(data); err != nil {
		r.log.Warn("response write", lax.Error(err))
	}
}

// Error renders HTTP response with error in JSON body.
func (r *JSON) Error(res http.ResponseWriter, statusCode int, message string) {
	r.Render(res, statusCode, &Error{Error: message})
}

// Error is used to render JSON error.
type Error struct {
	Error string `json:"error"`
}
