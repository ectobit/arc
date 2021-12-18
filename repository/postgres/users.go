package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.ectobit.com/arc/domain"
	"go.ectobit.com/arc/repository"
)

var _ repository.Users = (*UsersRepository)(nil)

// UsersRepository implements repository.Users interface using PostgreSQL database.
type UsersRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates new users repository using PostgreSQL database.
func NewUserRepository(conn *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{pool: conn}
}

// Create creates new user in PostgreSQL database.
func (repo *UsersRepository) Create(ctx context.Context, email string, password []byte) (*domain.User, error) {
	query := `INSERT INTO users (email, password) VALUES ($1, $2)
		RETURNING id, email, password, created, activation_token, active`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), email, password)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.ActivationToken,
		&user.Active); err != nil {
		return nil, repositoryError("create user", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// FindOne fetches user from PostgreSQL database using ID.
func (repo *UsersRepository) FindOne(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, password, created, updated, activation_token, password_reset_token, active
FROM users WHERE id=$1`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), id)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.Updated,
		&user.ActivationToken, &user.PasswordResetToken, &user.Active); err != nil {
		return nil, repositoryError("find one", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// FindOneByEmail fetches user from PostgreSQL database using email address.
func (repo *UsersRepository) FindOneByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password, created, updated, activation_token, password_reset_token, active
FROM users WHERE email=$1`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), email)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.Updated,
		&user.ActivationToken, &user.PasswordResetToken, &user.Active); err != nil {
		return nil, repositoryError("fetch user by email", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// Activate activates user account in PostgreSQL database.
func (repo *UsersRepository) Activate(ctx context.Context, token string) (*domain.User, error) {
	query := `UPDATE users SET updated=now(), activation_token=NULL, active=TRUE
		WHERE activation_token=$1 RETURNING id, email, password, created, updated`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), token)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.Updated); err != nil {
		return nil, repositoryError("activate user account", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// FetchPasswordResetToken sets user's password reset token in PostgreSQL repository.
func (repo *UsersRepository) FetchPasswordResetToken(ctx context.Context, email string) (*domain.User, error) {
	query := `UPDATE users SET password_reset_token=gen_random_uuid()
WHERE email=$1 AND active=TRUE RETURNING id, email, password_reset_token`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), email)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.PasswordResetToken); err != nil {
		return nil, repositoryError("fetch pasword reset token", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// ResetPassword sets new user's password in PostgreSQL database.
func (repo *UsersRepository) ResetPassword(ctx context.Context, passwordResetToken string,
	password []byte) (*domain.User, error) {
	query := `UPDATE users SET password=$1, password_reset_token=NULL
WHERE password_reset_token=$2 AND active=TRUE RETURNING id, email`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), password, passwordResetToken)

	var user User

	if err := row.Scan(&user.ID, &user.Email); err != nil {
		return nil, repositoryError("reset password", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}
