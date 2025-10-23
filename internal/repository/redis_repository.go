package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrTokenNotFound = errors.New("token not found in redis")

type RedisRepository interface {
	SetToken(ctx context.Context, key, value string, ttl time.Duration) error
	GetToken(ctx context.Context, key string) (string, error)
	DeleteToken(ctx context.Context, key string) error
	ExistToken(ctx context.Context, key string) (bool, error)
	ExpireToken(ctx context.Context, key string, ttl time.Duration) error
}

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) RedisRepository {
	return &redisRepository{client: client}
}

func (r *redisRepository) SetToken(ctx context.Context, key, value string, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()

	if err != nil {
		return err
	}

	return nil
}

func (r *redisRepository) GetToken(ctx context.Context, key string) (string, error) {
	fmt.Println("[DEBUG]: key:", key)
	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrTokenNotFound
	}

	if err != nil {
		return "", err
	}

	return value, nil
}

func (r *redisRepository) DeleteToken(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisRepository) ExistToken(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()

	if err != nil {
		return false, err
	}

	return result > 0, nil
}

func (r *redisRepository) ExpireToken(ctx context.Context, key string, ttl time.Duration) error {
	err := r.client.Expire(ctx, key, ttl).Err()

	if err != nil {
		return err
	}

	return nil
}
