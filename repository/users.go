package repository

import (
	"context"
	"errors"

	"github.com/ectobit/arc/domain"
)

// Errors.
var (
	ErrDuplicateKey      = errors.New("duplicate key")
	ErrInvalidActivation = errors.New("account already activated or invalid activation token")
	ErrInvalidAccount    = errors.New("email not found or account not activated")
)

// Users abstracts users repository methods.
type Users interface {
	// Create creates new user in a repository.
	Create(ctx context.Context, email string, password []byte) (*domain.User, error)
	// FindByEmail fetches user from repository using email address.
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	// Activate activates user account in repository.
	Activate(ctx context.Context, token string) (*domain.User, error)
	// PasswordResetToken sets password reset token for a user in repository.
	PasswordResetToken(ctx context.Context, email string) (*domain.User, error)
}
