package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository interface {
	SetToken(ctx context.Context, key, value string, ttl time.Duration) error
	GetToken(ctx context.Context, key string) (string, error)
	DeleteToken(ctx context.Context, key string) error
}

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepositry(client *redis.Client) RedisRepository {
	return &redisRepository{client: client}
}

func (r *redisRepository) SetToken(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisRepository) GetToken(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisRepository) DeleteToken(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
