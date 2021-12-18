package request

import (
	"encoding/json"
	"io"

	"go.ectobit.com/arc/domain"
	"go.ectobit.com/lax"
)

// ResetPassword contains user's password reset token and new password to be set.
type ResetPassword struct {
	PasswordResetToken string `json:"passwordResetToken"`
	Password           string `json:"password"`
	HashedPassword     []byte `json:"-"`
}

// ResetPasswordFromJSON parses ResetPassword from request body.
func ResetPasswordFromJSON(body io.Reader, log lax.Logger) (*ResetPassword, error) {
	var rp ResetPassword

	var err error

	if err := json.NewDecoder(body).Decode(&rp); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if rp.PasswordResetToken == "" {
		return nil, NewBadRequestError("empty password reset token")
	}

	if rp.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	if isWeakPassword(rp.Password) {
		return nil, NewBadRequestError("weak password")
	}

	rp.HashedPassword, err = domain.HashPassword(rp.Password)
	if err != nil {
		log.Warn("hash password", lax.Error(err))

		return nil, NewInternalServerError()
	}

	return &rp, nil
}
