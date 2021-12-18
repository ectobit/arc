package request

import (
	"encoding/json"
	"io"

	"go.ectobit.com/lax"
)

// Email contains user email.
type Email struct {
	Email string `json:"email"`
}

// EmailFromJSON parses email from request body.
func EmailFromJSON(body io.Reader, log lax.Logger) (*Email, error) {
	var rpr Email

	if err := json.NewDecoder(body).Decode(&rpr); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if rpr.Email == "" {
		return nil, NewBadRequestError("empty email")
	}

	if !isValidEmail(rpr.Email) {
		return nil, NewBadRequestError("invalid email")
	}

	return &rpr, nil
}
