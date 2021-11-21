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
		ID:      *user.ID,
		Email:   *user.Email,
		Created: user.Created,
		Updated: user.Updated,
	}
}

// UserRegistration contains data to receive.
type UserRegistration struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	log      *zap.Logger
}

// UserRegistrationFromJSON parses user registration data from request body.
func UserRegistrationFromJSON(body io.Reader, log *zap.Logger) (*UserRegistration, *Error) {
	var u UserRegistration

	if err := json.NewDecoder(body).Decode(&u); err != nil {
		log.Warn("decode json: %w", zap.Error(err))

		return nil, NewBadRequestError("invalid json body")
	}

	return &u, nil
}

// DomainUser converts user registration to domain user.
func (ur *UserRegistration) DomainUser() (*domain.User, *Error) {
	if ur.Email == "" {
		return nil, NewBadRequestError("empty email")
	}

	if ur.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	if isWeakPassword(ur.Password) {
		return nil, NewBadRequestError("weak password")
	}

	hashedPassword, err := domain.PasswordFromPlain(ur.Password)
	if err != nil {
		ur.log.Warn("password from plain", zap.Error(err))

		return nil, ErrorFromStatusCode(http.StatusInternalServerError)
	}

	return &domain.User{ //nolint:exhaustivestruct
		Email:    &ur.Email,
		Password: hashedPassword,
	}, nil
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
