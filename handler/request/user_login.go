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
	var u UserLogin

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

	return &u, nil
}
