package response

import (
	"time"

	"go.ectobit.com/arc/domain"
)

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
