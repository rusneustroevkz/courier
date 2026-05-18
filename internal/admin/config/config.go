package config

import (
	"log/slog"
	"os"

	"github.com/rusneustroevkz/courier/internal/admin/telegram"
	"github.com/rusneustroevkz/courier/pkg/middlewares"
	"github.com/rusneustroevkz/courier/pkg/postgres"
	"github.com/rusneustroevkz/courier/pkg/server"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// LogLevel LevelDebug -4 | LevelInfo 0 | LevelWarn 4 | LevelError 8
	LogLevel slog.Level `yaml:"log_level"`
	// Env local | stage | prd
	ENV           string             `yaml:"env"`
	PrivateServer server.Config      `yaml:"private_server"`
	PublicServer  server.Config      `yaml:"public_server"`
	RenderServer  server.Config      `yaml:"render_server"`
	TelegramBot   telegram.Config    `yaml:"telegram_bot"`
	Postgres      postgres.Config    `yaml:"postgres"`
	Middleware    middlewares.Config `yaml:"middleware"`
}

func New(configName string) (*Config, error) {
	var cfg Config

	data, err := os.ReadFile(configName)
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
