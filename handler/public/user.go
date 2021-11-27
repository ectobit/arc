package public

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ectobit/arc/domain"
	"github.com/nbutton23/zxcvbn-go"
	"go.uber.org/zap"
)

const minPasswordStrength = 3

// User contains user data to send out.
type User struct {
	ID           string     `json:"id,omitempty"`
	Email        string     `json:"email"`
	Created      *time.Time `json:"created"`
	Updated      *time.Time `json:"updated,omitempty"`
	AuthToken    string     `json:"authToken,omitempty"`
	RefreshToken string     `json:"refreshToken,omitempty"`
}

// FromDomainUser converts domain user to public user.
func FromDomainUser(user *domain.User) *User {
	return &User{ //nolint:exhaustivestruct
		ID:      user.ID,
		Email:   user.Email,
		Created: user.Created,
		Updated: user.Updated,
	}
}

// UserRegistration contains data to receive.
type UserRegistration struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	HashedPassword []byte `json:"-"`
	log            *zap.Logger
}

// UserRegistrationFromJSON parses user registration data from request body.
func UserRegistrationFromJSON(body io.Reader, log *zap.Logger) (*UserRegistration, *Error) {
	var u UserRegistration

	var err error

	if err := json.NewDecoder(body).Decode(&u); err != nil {
		log.Warn("decode json: %w", zap.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	if u.Email == "" {
		return nil, NewBadRequestError("empty email")
	}

	if u.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	if isWeakPassword(u.Password) {
		return nil, NewBadRequestError("weak password")
	}

	u.HashedPassword, err = domain.HashPassword(u.Password)
	if err != nil {
		u.log.Warn("hash password", zap.Error(err))

		return nil, ErrorFromStatusCode(http.StatusInternalServerError)
	}

	return &u, nil
}

// UserLogin contains user login data to receive.
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserLoginFromJSON parses user login data from request body.
func UserLoginFromJSON(body io.Reader) (*UserLogin, error) {
	var u UserLogin

	if err := json.NewDecoder(body).Decode(&u); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return &u, nil
}

func isWeakPassword(plainPassword string) bool {
	return zxcvbn.PasswordStrength(plainPassword, nil).Score < minPasswordStrength
}
