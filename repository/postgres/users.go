package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.ectobit.com/arc/domain"
	"go.ectobit.com/arc/repository"
	"go.ectobit.com/lax"
)

var _ repository.Users = (*UsersRepository)(nil)

// UsersRepository implements repository.Users interface using postgres backend.
type UsersRepository struct {
	pool *pgxpool.Pool
	log  lax.Logger
}

// NewUserRepository creates new user repository in postgres database.
func NewUserRepository(conn *pgxpool.Pool, log lax.Logger) *UsersRepository {
	return &UsersRepository{pool: conn, log: log}
}

// Create creates new user in postgres repository.
func (repo *UsersRepository) Create(ctx context.Context, email string, password []byte) (*domain.User, error) {
	query := `INSERT INTO users (email, password) VALUES ($1, $2)
		RETURNING id, email, password, created, activation_token, active`

	row := repo.pool.QueryRow(ctx, query, email, password)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.ActivationToken,
		&user.Active); err != nil {
		return nil, fmt.Errorf("create user: %w", toRepositoryError(err))
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// FindByEmail fetches user from postgres repository using email address.
func (repo *UsersRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password, created, updated, activation_token, password_reset_token, active
FROM users WHERE email=$1`

	row := repo.pool.QueryRow(ctx, query, email)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.Updated,
		&user.ActivationToken, &user.PasswordResetToken, &user.Active); err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// Activate activates user account in postgres repository.
func (repo *UsersRepository) Activate(ctx context.Context, token string) (*domain.User, error) {
	query := `UPDATE users SET updated=now(), activation_token=NULL, active=TRUE WHERE active=FALSE AND activation_token=$1
RETURNING id, email, password, created, updated`

	row := repo.pool.QueryRow(ctx, query, token)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.Updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrInvalidActivation
		}

		return nil, fmt.Errorf("activate user account: %w", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// PasswordResetToken sets password reset token for a user in postgres repository.
func (repo *UsersRepository) PasswordResetToken(ctx context.Context, email string) (*domain.User, error) {
	query := `UPDATE users SET password_reset_token=gen_random_uuid() WHERE active=TRUE AND email=$1
RETURNING id, email, password_reset_token`

	row := repo.pool.QueryRow(ctx, query, email)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.PasswordResetToken); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrInvalidAccount
		}

		return nil, fmt.Errorf("reset password: %w", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}
