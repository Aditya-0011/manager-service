package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresParams struct {
	Pool *pgxpool.Pool
}

func connectPostgres(url string) (*PostgresParams, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		dt, err := conn.LoadType(ctx, "portfolio.position_create")
		if err != nil {
			return err
		}
		conn.TypeMap().RegisterType(dt)
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to Postgres pool")

	return &PostgresParams{Pool: pool}, nil
}

func (p *PostgresParams) disconnectPostgres() error {
	log.Println("Disconnecting from Postgres...")
	p.Pool.Close()
	return nil
}
