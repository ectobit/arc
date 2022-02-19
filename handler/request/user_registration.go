package request

import (
	"encoding/json"
	"io"

	"go.ectobit.com/arc/domain"
	"go.ectobit.com/lax"
)

// UserRegistration contains data to receive.
type UserRegistration struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	HashedPassword []byte `json:"-"`
}

// UserRegistrationFromJSON parses user registration data from request body.
func UserRegistrationFromJSON(body io.Reader, log lax.Logger) (*UserRegistration, error) {
	var userRegistration UserRegistration

	var err error

	if err := json.NewDecoder(body).Decode(&userRegistration); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if userRegistration.Email == "" {
		return nil, NewBadRequestError("empty email")
	}

	if !isValidEmail(userRegistration.Email) {
		return nil, NewBadRequestError("invalid email")
	}

	if userRegistration.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	if isWeakPassword(userRegistration.Password) {
		return nil, NewBadRequestError("weak password")
	}

	userRegistration.HashedPassword, err = domain.HashPassword(userRegistration.Password)
	if err != nil {
		log.Warn("hash password", lax.Error(err))

		return nil, NewInternalServerError()
	}

	return &userRegistration, nil
}
