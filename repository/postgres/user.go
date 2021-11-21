package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ectobit/arc/domain"
	"github.com/jackc/pgtype"
)

// User entity.
type User struct {
	ID                 string
	Email              string
	Password           string
	Created            *time.Time
	Updated            pgtype.Timestamptz
	ActivationToken    sql.NullString
	PasswordResetToken sql.NullString
	Active             bool
}

// DomainUser converts user entity to domain user.
func (u *User) DomainUser() (*domain.User, error) {
	domainUser := &domain.User{ //nolint:exhaustivestruct
		ID:       &u.ID,
		Email:    &u.Email,
		Password: domain.PasswordFromHashed(u.Password),
		Created:  u.Created,
		Active:   &u.Active,
	}

	if err := u.Updated.AssignTo(domainUser.Updated); err != nil {
		return nil, fmt.Errorf("assign updated: %w", err)
	}

	if u.ActivationToken.Valid {
		*domainUser.ActivationToken = u.ActivationToken.String
	}

	if u.PasswordResetToken.Valid {
		*domainUser.PasswordResetToken = u.PasswordResetToken.String
	}

	return domainUser, nil
}
