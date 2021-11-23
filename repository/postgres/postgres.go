// Package postgres contains repository implementation for postgres database.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ectobit/arc/repository"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// Connect connects to postgres database.
func Connect(ctx context.Context, dsn string, log *zap.Logger, logLevel string) (*pgxpool.Pool, error) {
	pgxLogLevel, err := pgx.LogLevelFromString(logLevel)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres config: %w", err)
	}

	config.ConnConfig.Logger = zapadapter.NewLogger(log)
	config.ConnConfig.LogLevel = pgxLogLevel

	pool, err := pgxpool.ConnectConfig(ctx, config)
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
