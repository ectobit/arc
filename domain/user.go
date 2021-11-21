package domain

import (
	"time"
)

// User contains user data.
type User struct {
	ID                 *string
	Email              *string
	Password           *Password
	Created            *time.Time
	Updated            *time.Time
	ActivationToken    *string
	PasswordResetToken *string
	Active             *bool
}

// IsActive check if user account is activated.
func (u *User) IsActive() bool {
	return *u.Active
}
