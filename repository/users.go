package repository

import (
	"context"
	"regexp"

	"go.ectobit.com/arc/domain"
)

var regex = regexp.MustCompile(`\s+`)

// Users abstracts users repository methods.
type Users interface {
	// Create creates new user in users repository.
	Create(ctx context.Context, email string, password []byte) (*domain.User, error)
	// FindOne fetches user from users repository using ID.
	FindOne(ctx context.Context, email string) (*domain.User, error)
	// FindOneByEmail fetches user from users repository using email address.
	FindOneByEmail(ctx context.Context, email string) (*domain.User, error)
	// Activate activates user account in users repository.
	Activate(ctx context.Context, token string) (*domain.User, error)
	// FetchPasswordResetToken sets user's password reset token in users repository.
	FetchPasswordResetToken(ctx context.Context, email string) (*domain.User, error)
	// ResetPassword sets new user's password in users repository.
	ResetPassword(ctx context.Context, passwordResetToken string, password []byte) (*domain.User, error)
}
