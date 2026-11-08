package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"manager/controller"
	"manager/db"

	"buf.build/go/protovalidate"
	"github.com/Aditya-0011/common/contracts/go/grpcx"
)

const (
	defaultPort     = "7296"
	startupTimeout  = 10 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("manager service failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	slog.Info("runtime config",
		slog.Int("GOMAXPROCS", runtime.GOMAXPROCS(0)),
		slog.Int64("GOMEMLIMIT_MiB", debug.SetMemoryLimit(-1)/1024/1024),
	)

	serverConfig, err := grpcx.ServerConfigFromEnv(defaultPort)
	if err != nil {
		return err
	}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		return errors.New("POSTGRES_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	pool, err := db.Connect(ctx, postgresURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	validator, err := protovalidate.New()
	if err != nil {
		return err
	}

	srv, err := grpcx.NewServer(serverConfig, validator)
	if err != nil {
		return err
	}
	controller.Register(srv, pool)

	return srv.Run(shutdownTimeout)
}
