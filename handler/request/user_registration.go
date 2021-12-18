package request

import (
	"encoding/json"
	"io"
	"net/http"

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
	var u UserRegistration

	var err error

	if err := json.NewDecoder(body).Decode(&u); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if u.Email == "" {
		return nil, NewBadRequestError("empty email")
	}

	if !isValidEmail(u.Email) {
		return nil, NewBadRequestError("invalid email")
	}

	if u.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	if isWeakPassword(u.Password) {
		return nil, NewBadRequestError("weak password")
	}

	u.HashedPassword, err = domain.HashPassword(u.Password)
	if err != nil {
		log.Warn("hash password", lax.Error(err))

		return nil, ErrorFromStatusCode(http.StatusInternalServerError)
	}

	return &u, nil
}
