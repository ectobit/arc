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
	var refreshToken RefreshToken

	if err := json.NewDecoder(body).Decode(&refreshToken); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if refreshToken.RefreshToken == "" {
		return nil, NewBadRequestError("empty refresh token")
	}

	return &refreshToken, nil
}
