package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"go.ectobit.com/arc/domain"
)

// User entity.
type User struct {
	ID                 string
	Email              string
	Password           pgtype.Bytea
	Created            *time.Time
	Updated            pgtype.Timestamptz
	ActivationToken    pgtype.UUID
	PasswordResetToken pgtype.UUID
	Active             bool
}

// DomainUser converts user entity to domain user.
func (u *User) DomainUser() (*domain.User, error) {
	domainUser := &domain.User{ //nolint:exhaustivestruct
		ID:      u.ID,
		Email:   u.Email,
		Created: u.Created,
		Active:  &u.Active,
	}

	if err := u.Password.AssignTo(&domainUser.Password); err != nil {
		return nil, fmt.Errorf("assign password: %w", err)
	}

	if err := u.Updated.AssignTo(domainUser.Updated); err != nil {
		return nil, fmt.Errorf("assign updated: %w", err)
	}

	if err := u.ActivationToken.AssignTo(&domainUser.ActivationToken); err != nil {
		return nil, fmt.Errorf("assign activation token: %w", err)
	}

	if err := u.PasswordResetToken.AssignTo(&domainUser.PasswordResetToken); err != nil {
		return nil, fmt.Errorf("assign activation token: %w", err)
	}

	return domainUser, nil
}
