package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rusneustroevkz/courier/internal/backend/config"
	"github.com/rusneustroevkz/courier/internal/backend/router"
	"github.com/rusneustroevkz/courier/internal/backend/telegram"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/logger"
	"github.com/rusneustroevkz/courier/pkg/postgres"
	"github.com/rusneustroevkz/courier/pkg/redis"
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

	redisClient, err := redis.New(cfg.Redis)
	if err != nil {
		logger.Error("failed to initialize redis", "error", err)
		os.Exit(1)
	}

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

	privateRouter := router.NewPrivate()

	privateServer := server.New(cfg.PrivateServer, privateRouter.Routes())
	go func() {
		if err := privateServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start private server", "error", err)
			os.Exit(1)
		}
	}()
	logger.Info("starting private server", "port", cfg.PrivateServer.Port)

	<-ctx.Done()

	shutdownCtx, timeout := context.WithTimeout(context.Background(), 15*time.Second)
	defer timeout()

	logger.Info("shutting down servers...")

	if err := privateServer.Stop(shutdownCtx); err != nil {
		logger.Error("failed to stop private server", "error", err)
	}
	if err := shutdownCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("failed to shutdown gracefully", "error", err)
	}
	if err := redisClient.Close(); err != nil {
		logger.Error("failed to close redis", "error", err)
	}
	if err := db.Close(); err != nil {
		logger.Error("failed to close postgres", "error", err)
	}
	telegramBot.Stop()

	logger.Info("shutdown complete")
}
