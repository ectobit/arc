// Package postgres contains repository implementation for postgres database.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ectobit/arc/repository"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// Connect connects to postgres database.
func Connect(ctx context.Context, dsn string, log *zap.Logger) (*pgxpool.Pool, error) {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	return pool, nil
}

func toRepositoryError(err error) error {
	pgErr := &pgconn.PgError{} //nolint:exhaustivestruct

	if errors.As(err, &pgErr) {
		switch pgErr.Code { //nolint:gocritic
		case "23505":
			return repository.ErrDuplicateKey
		}
	}

	return err
}
