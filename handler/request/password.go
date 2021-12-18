package request

import (
	"encoding/json"
	"io"

	"go.ectobit.com/lax"
)

// Password contains user password.
type Password struct {
	Password string `json:"password"`
}

// PasswordFromJSON parses password from request body.
func PasswordFromJSON(body io.Reader, log lax.Logger) (*Password, error) {
	var cps Password

	if err := json.NewDecoder(body).Decode(&cps); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if cps.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	return &cps, nil
}
