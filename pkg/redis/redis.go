package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type Config struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
}

type Redis struct {
	Client *redis.Client
}

func New(cfg Config) (*Redis, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       0,
	}
	rdb := redis.NewClient(opts)

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	slog.Info("redis client started", "addr", cfg.Addr)

	return &Redis{
		Client: rdb,
	}, nil
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
