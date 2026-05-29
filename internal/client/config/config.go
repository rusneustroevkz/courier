package config

import (
	"github.com/rusneustroevkz/courier/pkg/redis"
	"log/slog"
	"os"

	"github.com/rusneustroevkz/courier/internal/client/telegram"
	"github.com/rusneustroevkz/courier/pkg/middlewares"
	"github.com/rusneustroevkz/courier/pkg/postgres"
	"github.com/rusneustroevkz/courier/pkg/server"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// LogLevel LevelDebug -4 | LevelInfo 0 | LevelWarn 4 | LevelError 8
	LogLevel slog.Level `yaml:"log_level"`
	// Env local | stage | prd
	ENV                   string             `yaml:"env"`
	PrivateServer         server.Config      `yaml:"private_server"`
	PublicServer          server.Config      `yaml:"public_server"`
	RenderServer          server.Config      `yaml:"render_server"`
	TelegramBot           telegram.Config    `yaml:"telegram_bot"`
	Postgres              postgres.Config    `yaml:"postgres"`
	JWTAccessTokenSecret  string             `yaml:"jwt_access_token_secret"`
	JWTRefreshTokenSecret string             `yaml:"jwt_refresh_token_secret"`
	Middleware            middlewares.Config `yaml:"middleware"`
	Redis                 redis.Config       `yaml:"redis"`
}

func New() (*Config, error) {
	var cfg Config

	data, err := os.ReadFile("./config/config.yml")
	if err != nil {
		return nil, err
	}

	expanded := os.ExpandEnv(string(data))

	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}

	cfg.Middleware.Env = cfg.ENV

	return &cfg, nil
}
