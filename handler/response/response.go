package response

import (
	"encoding/json"
	"net/http"

	"go.ectobit.com/lax"
)

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
