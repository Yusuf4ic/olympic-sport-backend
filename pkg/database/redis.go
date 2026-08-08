package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// NewRedisClient initializes and returns a go-redis client.
func NewRedisClient(ctx context.Context, cfg *config.RedisConfig) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	}

	// Upstash and other managed Redis services require TLS
	if cfg.TLS {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	log.Info().
		Str("addr", cfg.Addr()).
		Int("pool_size", cfg.PoolSize).
		Msg("Successfully connected to Redis")

	return client, nil
}
