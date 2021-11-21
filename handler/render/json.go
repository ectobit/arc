package render

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

var _ Renderer = (*JSON)(nil)

// JSON implements render.Renderer interface using JSON encoding.
type JSON struct {
	log *zap.Logger
}

// NewJSON creates JSON renderer.
func NewJSON(log *zap.Logger) *JSON {
	return &JSON{
		log: log,
	}
}

// Render renders HTTP response with JSON body.
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
		r.log.Warn("response write", zap.Error(err))
	}
}

// Error renders HTTP response with error in JSON body.
func (r *JSON) Error(res http.ResponseWriter, statusCode int, message string) {
	r.Render(res, statusCode, &err{Error: message})
}

type err struct {
	Error string `json:"error"`
}
