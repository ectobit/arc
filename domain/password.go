package domain

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Errors.
var (
	ErrEmptyPassword = errors.New("empty password")
	ErrWeakPassword  = errors.New("weak password")
	ErrAlreadyHashed = errors.New("password already hashed")
)

// Password contains user password.
type Password struct {
	inner string
}

// PasswordFromPlain creates password from plain password by hashing it into inner field.
func PasswordFromPlain(plainPassword string) (*Password, error) {
	if plainPassword == "" {
		return nil, fmt.Errorf("hash: %w", ErrEmptyPassword)
	}

	password := &Password{
		inner: plainPassword,
	}

	if err := password.hash(); err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	return password, nil
}

// PasswordFromHashed creates password from hashed password.
func PasswordFromHashed(hashedPassword string) *Password {
	return &Password{inner: hashedPassword}
}

// String implements fmt.Stringer interface.
func (p *Password) String() string {
	return p.inner
}

// IsValid checks if provided plain password matches with stored hash.
func (p *Password) IsValid(plainPassword string) bool {
	if plainPassword == "" || p.inner == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(p.inner), []byte(plainPassword)) == nil
}

func (p *Password) hash() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(p.inner), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash: %w", err)
	}

	p.inner = string(hash)

	return nil
}
