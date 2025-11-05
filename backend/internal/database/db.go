package database

import (
	"context"

	"gh-ts/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.DBURL)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, pcfg)
}
