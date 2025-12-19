package redishelper

import (
	"briefcash-jwt/config"
	logs "briefcash-jwt/internal/helper/loghelper"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisAdapter struct {
	Client *redis.Client
}

func NewRedisAdapter(cfg *config.Config) (*redisAdapter, error) {

	if cfg.RedisAddress == "" || cfg.RedisPort == "" {
		logs.Logger.Error("Invalid redis config: address or port is empty")
		return nil, fmt.Errorf("invalid redis config: address or port is empty")
	}

	// Set up config
	address := fmt.Sprintf("%s:%s", cfg.RedisAddress, cfg.RedisPort)

	client := redis.NewClient(&redis.Options{
		Addr: address,
		//Password:     cfg.RedisPassword,
		DB:           0,
		PoolSize:     50,
		MinIdleConns: 10,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	// Ping Redis connection with context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logs.Logger.WithError(err).Error("Failed to connect to redis")
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logs.Logger.Info("Connected to Redis successfully")

	return &redisAdapter{Client: client}, nil
}

func (r *redisAdapter) Close() error {
	return r.Client.Close()
}
