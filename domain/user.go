package domain

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User contains user data.
type User struct {
	ID                 string
	Email              string
	Password           []byte
	Created            *time.Time
	Updated            *time.Time
	ActivationToken    string
	PasswordResetToken string
	Active             *bool
}

// IsActive checks if user account is activated.
func (u *User) IsActive() bool {
	return *u.Active
}

// IsPasswordValid checks if provided plain password matched hashed password.
func (u *User) IsPasswordValid(plainPassword string) bool {
	if plainPassword == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword(u.Password, []byte(plainPassword)) == nil
}

// HashPassword hashes provided plain password using bcrypt hasher.
func HashPassword(plainPassword string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt: %w", err)
	}

	return hash, nil
}
