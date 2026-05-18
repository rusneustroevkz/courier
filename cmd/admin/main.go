package main

import (
	"context"
	"github.com/rusneustroevkz/courier/pkg/middlewares"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rusneustroevkz/courier/internal/admin/config"
	"github.com/rusneustroevkz/courier/internal/admin/router"
	"github.com/rusneustroevkz/courier/internal/admin/telegram"
	"github.com/rusneustroevkz/courier/internal/admin/users"
	"github.com/rusneustroevkz/courier/pkg/postgres"
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

	cfg, err := config.New(os.Getenv("CONFIG_NAME"))
	if err != nil {
		slog.Error("failed to initialize config", "error", err)
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(cfg.LogLevel)
	slog.Info("initializing server", "log_level", cfg.LogLevel)

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		slog.Error("failed to initialize postgres", "error", err)
		os.Exit(1)
	}

	telegramBot, err := telegram.NewTelegram(cfg.TelegramBot)
	if err != nil {
		slog.Error("failed to initialize telegram bot", "error", err)
		os.Exit(1)
	}
	go func() {
		telegramBot.Start()
	}()

	mw := middlewares.NewMiddleware(cfg.Middleware, nil)

	usersRepository := users.New(db.DB)

	usersService := users.NewService(usersRepository, telegramBot)
	usersController := users.NewController(usersService)

	privateRouter := router.NewPrivate()
	privateServer := server.New(cfg.PrivateServer, privateRouter.Routes())
	go func() {
		if err := privateServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start private server", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("starting private server", "port", cfg.PrivateServer.Port)

	publicRouter := router.NewPublic(mw, usersController)
	publicServer := server.New(cfg.PublicServer, publicRouter.Routes())
	go func() {
		if err := publicServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start public server", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("starting public server", "port", cfg.PublicServer.Port)

	renderRouter := router.NewRender()
	renderServer := server.New(cfg.RenderServer, renderRouter.Routes())
	go func() {
		if err := renderServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start render server", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("starting render server", "port", cfg.RenderServer.Port)

	<-ctx.Done()

	shutdownCtx, timeout := context.WithTimeout(context.Background(), 15*time.Second)
	defer timeout()

	slog.Info("shutting down servers...")

	if err := privateServer.Stop(shutdownCtx); err != nil {
		slog.Error("failed to stop private server", "error", err)
	}
	if err := publicServer.Stop(shutdownCtx); err != nil {
		slog.Error("failed to stop public server", "error", err)
	}
	if err := shutdownCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("failed to shutdown gracefully", "error", err)
	}
	if err := db.Close(); err != nil {
		slog.Error("failed to close postgres", "error", err)
	}
	telegramBot.Stop()

	slog.Info("shutdown complete")
}
