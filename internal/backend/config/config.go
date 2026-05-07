package config

import (
	"log/slog"
	"os"

	"github.com/rusneustroevkz/courier/internal/backend/telegram"
	"github.com/rusneustroevkz/courier/pkg/postgres"
	"github.com/rusneustroevkz/courier/pkg/redis"
	"github.com/rusneustroevkz/courier/pkg/server"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// LogLevel LevelDebug -4 | LevelInfo 0 | LevelWarn 4 | LevelError 8
	LogLevel slog.Level `yaml:"log_level"`
	// Env local | stage | prd
	ENV           string          `yaml:"env"`
	PrivateServer server.Config   `yaml:"private_server"`
	PublicServer  server.Config   `yaml:"public_server"`
	TelegramBot   telegram.Config `yaml:"telegram_bot"`
	Redis         redis.Config    `yaml:"redis"`
	Postgres      postgres.Config `yaml:"postgres"`
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

	return &cfg, nil
}
