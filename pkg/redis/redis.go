package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/rusneustroevkz/courier/pkg/logger"
)

type Config struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
}

type Redis struct {
	client *redis.Client
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

	logger.Info("redis client started", "addr", cfg.Addr)

	return &Redis{
		client: rdb,
	}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}
