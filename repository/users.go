package repository

import (
	"context"
	"regexp"

	"go.ectobit.com/arc/domain"
)

var regex = regexp.MustCompile(`\s+`)

// Users abstracts users repository methods.
type Users interface {
	// Create creates new user in a repository.
	Create(ctx context.Context, email string, password []byte) (*domain.User, error)
	// FindByEmail fetches user from repository using email address.
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	// Activate activates user account in repository.
	Activate(ctx context.Context, token string) (*domain.User, error)
	// FindByEmailWithPasswordResetToken sets password reset token for a user in repository.
	FindByEmailWithPasswordResetToken(ctx context.Context, email string) (*domain.User, error)
}
