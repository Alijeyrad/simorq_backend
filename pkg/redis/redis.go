package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/Alijeyrad/simorq_backend/config"
)

// NewRedisFromCentral creates a new Redis client from central config
func NewRedisFromCentral(cfg config.RedisConfig) (*goredis.Client, error) {
	return NewRedis(FromCentralConfig(cfg))
}

func NewRedis(cfg Config) (*goredis.Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("redis addr is empty")
	}

	opts := &goredis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout(),
		ReadTimeout:  cfg.ReadTimeout(),
		WriteTimeout: cfg.WriteTimeout(),
	}

	if cfg.PoolSize > 0 {
		opts.PoolSize = cfg.PoolSize
	}
	if cfg.MinIdleConns > 0 {
		opts.MinIdleConns = cfg.MinIdleConns
	}

	rdb := goredis.NewClient(opts)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return rdb, nil
}
