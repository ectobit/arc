package request

import (
	"encoding/json"
	"io"

	"go.ectobit.com/arc/domain"
	"go.ectobit.com/lax"
)

// ResetPassword contains user's password reset token and new password to be set.
type ResetPassword struct {
	RecoveryToken  string `json:"recoveryToken"`
	Password       string `json:"password"`
	HashedPassword []byte `json:"-"`
}

// ResetPasswordFromJSON parses ResetPassword from request body.
func ResetPasswordFromJSON(body io.Reader, log lax.Logger) (*ResetPassword, error) {
	var resetPassword ResetPassword

	var err error

	if err := json.NewDecoder(body).Decode(&resetPassword); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if resetPassword.RecoveryToken == "" {
		return nil, NewBadRequestError("empty password reset token")
	}

	if resetPassword.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	if isWeakPassword(resetPassword.Password) {
		return nil, NewBadRequestError("weak password")
	}

	resetPassword.HashedPassword, err = domain.HashPassword(resetPassword.Password)
	if err != nil {
		log.Warn("hash password", lax.Error(err))

		return nil, NewInternalServerError()
	}

	return &resetPassword, nil
}
