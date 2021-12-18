package request

import (
	"encoding/json"
	"io"

	"go.ectobit.com/lax"
)

// RefreshToken request body.
type RefreshToken struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenFromBody parses RefreshToken from request body.
func RefreshTokenFromBody(body io.Reader, log lax.Logger) (*RefreshToken, error) {
	var rt RefreshToken

	if err := json.NewDecoder(body).Decode(&rt); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if rt.RefreshToken == "" {
		return nil, NewBadRequestError("empty refresh token")
	}

	return &rt, nil
}
