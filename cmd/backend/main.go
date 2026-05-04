package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rusneustroevkz/courier/internal/backend/router"
	"github.com/rusneustroevkz/courier/internal/config"
	"github.com/rusneustroevkz/courier/pkg/server"
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
		if err := publicServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "failed to start public server", "error", err)
			os.Exit(1)
		}
	}()
	slog.InfoContext(ctx, "starting public server", "port", cfg.PublicServer.Port)

	privateServer := server.New(cfg.PrivateServer, privateRouter.Routes())
	go func() {
		if err := privateServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "failed to start private server", "error", err)
			os.Exit(1)
		}
	}()
	slog.InfoContext(ctx, "starting private server", "port", cfg.PrivateServer.Port)

	<-ctx.Done()
	slog.Info("shutting down servers...")

	shutdownCtx, timeout := context.WithTimeout(context.Background(), 15*time.Second)
	defer timeout()

	if err := publicServer.Stop(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "failed to stop public server", "error", err)
	}
	if err := privateServer.Stop(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "failed to stop private server", "error", err)
	}
	if err := shutdownCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		slog.ErrorContext(shutdownCtx, "failed to shutdown gracefully", "error", err)
	}

	slog.Info("shutdown complete")
}
