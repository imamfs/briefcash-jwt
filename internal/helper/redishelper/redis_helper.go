package redishelper

import (
	"briefcash-jwt/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisHelper(jwt *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     jwt.RedisAddress,
		Password: jwt.RedisPassword,
		DB:       0,
	})
}
