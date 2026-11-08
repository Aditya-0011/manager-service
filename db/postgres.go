package db

import (
	"context"
	"fmt"
	"time"

	"manager/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var compositeTypes = []string{"portfolio.position_create", "portfolio._position_create"}

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parsing POSTGRES_URL: %w", err)
	}

	config.MaxConns = 3
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnLifetimeJitter = 5 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = time.Minute

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, name := range compositeTypes {
			dt, err := conn.LoadType(ctx, name)
			if err != nil {
				return fmt.Errorf("loading type %s: %w", name, err)
			}
			conn.TypeMap().RegisterType(dt)
		}
		conn.TypeMap().RegisterDefaultPgType(model.PositionCreate{}, compositeTypes[0])
		conn.TypeMap().RegisterDefaultPgType([]model.PositionCreate{}, compositeTypes[1])
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return pool, nil
}
