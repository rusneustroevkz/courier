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

	"github.com/rusneustroevkz/courier/internal/backend/config"
	"github.com/rusneustroevkz/courier/internal/backend/orders"
	"github.com/rusneustroevkz/courier/internal/backend/router"
	"github.com/rusneustroevkz/courier/internal/backend/telegram"
	"github.com/rusneustroevkz/courier/internal/backend/users"
	"github.com/rusneustroevkz/courier/pkg/postgres"
	"github.com/rusneustroevkz/courier/pkg/redis"
	"github.com/rusneustroevkz/courier/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	loggerHandler := slog.NewJSONHandler(os.Stdout, handlerOpts)
	logger := slog.New(loggerHandler)
	slog.SetDefault(logger)

	cfg, err := config.New()
	if err != nil {
		slog.Error("failed to initialize config", "error", err)
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)
	slog.Info("initializing server", "log_level", cfg.LogLevel)

	redisClient, err := redis.New(cfg.Redis)
	if err != nil {
		slog.Error("failed to initialize redis", "error", err)
		os.Exit(1)
	}

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		slog.Error("failed to initialize postgres", "error", err)
		os.Exit(1)
	}

	usersRepository := users.New(db.DB)
	ordersRepository := orders.New(db.DB)

	usersService := users.NewService(usersRepository)
	ordersService := orders.NewService(ordersRepository)

	go func() {
		t := time.NewTicker(1 * time.Minute)

		for {
			select {
			case <-t.C:
				if err := usersService.WorkerSetShareLocationAfterTTL(ctx); err != nil {
					slog.Error("failed to set worker set share location after 5 seconds", "err", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	telegramBot, err := telegram.NewTelegram(cfg.TelegramBot, usersService, ordersService, redisClient)
	if err != nil {
		slog.Error("failed to initialize telegram bot", "error", err)
		os.Exit(1)
	}
	go func() {
		telegramBot.Start()
	}()

	privateRouter := router.NewPrivate()

	privateServer := server.New(cfg.PrivateServer, privateRouter.Routes())
	go func() {
		if err := privateServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start private server", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("starting private server", "port", cfg.PrivateServer.Port)

	<-ctx.Done()

	shutdownCtx, timeout := context.WithTimeout(context.Background(), 15*time.Second)
	defer timeout()

	slog.Info("shutting down servers...")

	if err := privateServer.Stop(shutdownCtx); err != nil {
		slog.Error("failed to stop private server", "error", err)
	}
	if err := shutdownCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("failed to shutdown gracefully", "error", err)
	}
	if err := redisClient.Close(); err != nil {
		slog.Error("failed to close redis", "error", err)
	}
	if err := db.Close(); err != nil {
		slog.Error("failed to close postgres", "error", err)
	}
	telegramBot.Stop()

	slog.Info("shutdown complete")
}
