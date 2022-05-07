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
	query := `SELECT id, email, password, created, updated, activation_token, recovery_token, active
FROM users WHERE id=$1`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), id)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.Updated,
		&user.ActivationToken, &user.RecoveryToken, &user.Active); err != nil {
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
	query := `SELECT id, email, password, created, updated, activation_token, recovery_token, active
FROM users WHERE email=$1`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), email)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Created, &user.Updated,
		&user.ActivationToken, &user.RecoveryToken, &user.Active); err != nil {
		return nil, repositoryError("fetch user by email", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// FindAll fetches alls users from PostgreSQL.
func (repo *UsersRepository) FindAll(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, email, password, activation_token, recovery_token, activated, active, created, updated
FROM users`

	rows, err := repo.pool.Query(ctx, repository.StripWhitespaces(query))
	if err != nil {
		return nil, repositoryError("fetch all users", err)
	}

	defer rows.Close()

	domainUsers := []domain.User{}

	for rows.Next() {
		var user User

		if err := rows.Scan(&user.ID, &user.Email, &user.Password, &user.ActivationToken, &user.RecoveryToken,
			&user.Activated, &user.Active, &user.Created, &user.Updated); err != nil {
			return nil, repositoryError("scan", err)
		}

		domainUser, err := user.DomainUser()
		if err != nil {
			return nil, fmt.Errorf("convert to domain user: %w", err)
		}

		domainUsers = append(domainUsers, *domainUser)
	}

	if err := rows.Err(); err != nil {
		return nil, repositoryError("rows err", err)
	}

	return domainUsers, nil
}

// Activate activates user account in PostgreSQL database.
func (repo *UsersRepository) Activate(ctx context.Context, token string) (*domain.User, error) {
	query := `UPDATE users SET updated=now(), activation_token=NULL, activated=current_timestamp, active=TRUE
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

// FetchRecoveryToken sets user's password reset token in PostgreSQL repository.
func (repo *UsersRepository) FetchRecoveryToken(ctx context.Context, email string) (*domain.User, error) {
	query := `UPDATE users SET recovery_token=gen_random_uuid()
WHERE email=$1 AND active RETURNING id, email, recovery_token`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), email)

	var user User

	if err := row.Scan(&user.ID, &user.Email, &user.RecoveryToken); err != nil {
		return nil, repositoryError("fetch pasword reset token", err)
	}

	domainUser, err := user.DomainUser()
	if err != nil {
		return nil, fmt.Errorf("convert to domain user: %w", err)
	}

	return domainUser, nil
}

// ResetPassword sets new user's password in PostgreSQL database.
func (repo *UsersRepository) ResetPassword(ctx context.Context, recoveryToken string,
	password []byte,
) (*domain.User, error) {
	query := `UPDATE users SET password=$1, recovery_token=NULL
WHERE recovery_token=$2 AND active RETURNING id, email`

	row := repo.pool.QueryRow(ctx, repository.StripWhitespaces(query), password, recoveryToken)

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
