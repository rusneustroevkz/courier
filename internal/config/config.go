package config

import (
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
)

type Config struct {
	// LogLevel LevelDebug -4 | LevelInfo 0 | LevelWarn 4 | LevelError 8
	LogLevel slog.Level `yaml:"log_level"`
	// Env local | stage | prd
	ENV           string `yaml:"env"`
	PrivateServer Server `yaml:"private_server"`
	PublicServer  Server `yaml:"public_server"`
}

type Server struct {
	Port int `yaml:"port"`
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
