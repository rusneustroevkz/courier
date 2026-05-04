package main

import (
	"context"
	"github.com/rusneustroevkz/courier/internal/backend/router"
	"github.com/rusneustroevkz/courier/internal/config"
	"github.com/rusneustroevkz/courier/pkg/server"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(loggerHandler)
	slog.SetDefault(logger)

	cfg, err := config.New(os.Getenv("CONFIG_NAME"))
	if err != nil {
		slog.ErrorContext(ctx, "failed to initialize config", "error", err)
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)
	slog.InfoContext(ctx, "initializing server", "log_level", cfg.LogLevel)

	publicRouter := router.NewPublic()
	privateRouter := router.NewPrivate()

	publicServer := server.New(cfg.PublicServer, publicRouter.Routes())
	go func() {
		if err := publicServer.Start(); err != nil {
			slog.ErrorContext(ctx, "failed to start public server", "error", err)
			os.Exit(1)
		}
	}()
	slog.InfoContext(ctx, "starting public server", "port", cfg.PublicServer.Port)

	privateServer := server.New(cfg.PrivateServer, privateRouter.Routes())
	go func() {
		if err := privateServer.Start(); err != nil {
			slog.ErrorContext(ctx, "failed to start private server", "error", err)
			os.Exit(1)
		}
	}()
	slog.InfoContext(ctx, "starting private server", "port", cfg.PrivateServer.Port)

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			slog.ErrorContext(ctx, "failed to shutdown gracefully", "error", err)
		}
		stop()
	}

	if err := publicServer.Stop(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to stop public server", "error", err)
	}
}
