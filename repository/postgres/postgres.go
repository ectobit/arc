// Package postgres contains repository implementation for postgres database.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.ectobit.com/arc/repository"
)

// Connect connects to postgres database.
func Connect(ctx context.Context, dsn string, log pgx.Logger, logLevel string) (*pgxpool.Pool, error) {
	pgxLogLevel, err := pgx.LogLevelFromString(logLevel)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres config: %w", err)
	}

	config.ConnConfig.Logger = log
	config.ConnConfig.LogLevel = pgxLogLevel
	config.LazyConnect = true

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	return pool, nil
}

func repositoryError(description string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrResourceNotFound
	}

	pgErr := &pgconn.PgError{} //nolint:exhaustivestruct

	if errors.As(err, &pgErr) {
		switch pgErr.Code { //nolint:gocritic
		case pgerrcode.UniqueViolation:
			// pgErr.ConstraintName may also be checked
			return repository.ErrUniqueViolation
		}
	}

	return fmt.Errorf("%s: %w", description, err)
}
