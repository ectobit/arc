package request

import (
	"encoding/json"
	"io"

	"go.ectobit.com/lax"
)

// UserLogin contains user login data to receive.
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserLoginFromJSON parses user login data from request body.
func UserLoginFromJSON(body io.Reader, log lax.Logger) (*UserLogin, error) {
	var userLogin UserLogin

	if err := json.NewDecoder(body).Decode(&userLogin); err != nil {
		log.Warn("decode json: %w", lax.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if userLogin.Email == "" {
		return nil, NewBadRequestError("empty email")
	}

	if !isValidEmail(userLogin.Email) {
		return nil, NewBadRequestError("invalid email")
	}

	if userLogin.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	return &userLogin, nil
}
