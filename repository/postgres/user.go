package postgres

import (
	"fmt"

	"github.com/jackc/pgtype"
	"go.ectobit.com/arc/domain"
)

// User entity.
type User struct {
	ID              string
	Email           string
	Password        pgtype.Bytea
	Activated       pgtype.Timestamptz
	Created         pgtype.Timestamptz
	Updated         pgtype.Timestamptz
	ActivationToken pgtype.UUID
	RecoveryToken   pgtype.UUID
	Active          bool
}

// DomainUser converts user entity to domain user.
func (u *User) DomainUser() (*domain.User, error) {
	domainUser := &domain.User{ //nolint:exhaustruct
		ID:     u.ID,
		Email:  u.Email,
		Active: &u.Active,
	}

	if u.Password.Status == pgtype.Present {
		domainUser.Password = u.Password.Bytes
	}

	if u.Activated.Status == pgtype.Present {
		domainUser.Activated = &u.Activated.Time
	}

	if u.Created.Status == pgtype.Present {
		domainUser.Created = &u.Created.Time
	}

	if u.Updated.Status == pgtype.Present {
		domainUser.Updated = &u.Updated.Time
	}

	if u.ActivationToken.Status == pgtype.Present {
		if err := u.ActivationToken.AssignTo(&domainUser.ActivationToken); err != nil {
			return nil, fmt.Errorf("assign activation token: %w", err)
		}
	}

	if u.RecoveryToken.Status == pgtype.Present {
		if err := u.RecoveryToken.AssignTo(&domainUser.RecoveryToken); err != nil {
			return nil, fmt.Errorf("assign password reset token: %w", err)
		}
	}

	return domainUser, nil
}
