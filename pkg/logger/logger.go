package logger

import (
	"context"
	"log/slog"
	"os"
)

func New() {
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	loggerHandler := slog.NewJSONHandler(os.Stdout, handlerOpts)
	logger := slog.New(loggerHandler)
	slog.SetDefault(logger)
}

func With(args ...any) *slog.Logger {
	return slog.Default().With(args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	requestID := ctx.Value("request_id")
	slog.ErrorContext(ctx, msg, append([]any{"request_id", requestID}, args...)...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, append([]any{}, args...)...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	requestID := ctx.Value("request_id")
	slog.InfoContext(ctx, msg, append([]any{"request_id", requestID}, args...)...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, append([]any{}, args...)...)
}

func SetLogLoggerLevel(level slog.Level) {
	slog.SetLogLoggerLevel(level)
}
