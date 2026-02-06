package postgres_cli

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func NewPGClient(ctx context.Context, dbConfig string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(dbConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse DB config")
		return nil, err
	}

	// Pool concurrency
	poolConfig.MaxConns = 5                       // maximum number of open connections
	poolConfig.MinConns = 0                       // maintain a warm pool
	poolConfig.MaxConnIdleTime = 10 * time.Second // or e.g. 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create DB pool")
		return nil, err
	}

	log.Debug().
		Str("message", "Postgres Pool initialized.").
		Str("DB URI", dbConfig).
		Msg("Postgres Pool initialized.")

	return pool, nil
}
