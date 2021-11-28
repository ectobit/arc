package public

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/nbutton23/zxcvbn-go"
	"go.ectobit.com/arc/domain"
	"go.uber.org/zap"
)

const minPasswordStrength = 3

var emailRegex = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$") //nolint:lll

// User contains user data to send out.
type User struct {
	ID           string     `json:"id,omitempty" format:"uuid"`
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

	if !isValidEmail(u.Email) {
		return nil, NewBadRequestError("invalid email")
	}

	if u.Password == "" {
		return nil, NewBadRequestError("empty password")
	}

	return &u, nil
}

func isWeakPassword(plainPassword string) bool {
	return zxcvbn.PasswordStrength(plainPassword, nil).Score < minPasswordStrength
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
