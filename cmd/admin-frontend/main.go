package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/admin-frontend/config"
	"github.com/rusneustroevkz/courier/internal/admin-frontend/router"
	"github.com/rusneustroevkz/courier/internal/admin-frontend/telegram"
	"github.com/rusneustroevkz/courier/internal/admin-frontend/users"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"github.com/rusneustroevkz/courier/pkg/postgres"
	"github.com/rusneustroevkz/courier/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.New()

	cfg, err := config.New(os.Getenv("CONFIG_NAME"))
	if err != nil {
		logger.Error("failed to initialize config", "error", err)
		os.Exit(1)
	}

	logger.SetLogLoggerLevel(cfg.LogLevel)
	logger.Info("initializing server", "log_level", cfg.LogLevel)

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		logger.Error("failed to initialize postgres", "error", err)
		os.Exit(1)
	}

	usersRepository := users.New(db.DB)

	usersService := users.NewService(usersRepository)

	telegramBot, err := telegram.NewTelegram(cfg.TelegramBot, usersService)
	if err != nil {
		logger.Error("failed to initialize telegram bot", "error", err)
		os.Exit(1)
	}
	go func() {
		telegramBot.Start()
	}()

	publicRouter := router.NewPublic()

	publicServer := server.New(cfg.PublicServer, publicRouter.Routes())
	go func() {
		if err := publicServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start public server", "error", err)
			os.Exit(1)
		}
	}()
	logger.Info("starting public server", "port", cfg.PublicServer.Port)

	<-ctx.Done()

	shutdownCtx, timeout := context.WithTimeout(context.Background(), 15*time.Second)
	defer timeout()

	logger.Info("shutting down servers...")

	if err := publicServer.Stop(shutdownCtx); err != nil {
		logger.Error("failed to stop public server", "error", err)
	}
	if err := shutdownCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("failed to shutdown gracefully", "error", err)
	}
	if err := db.Close(); err != nil {
		logger.Error("failed to close postgres", "error", err)
	}
	telegramBot.Stop()

	logger.Info("shutdown complete")
}
