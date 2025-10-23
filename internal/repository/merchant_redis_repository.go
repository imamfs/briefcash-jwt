package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type MerchantRedisRepository interface {
	SetActiveMerchantCode(ctx context.Context, mCodes []string) error
	IsMerchantCodeActive(ctx context.Context, mCode string) (bool, error)
	AddMerchantCode(ctx context.Context, mCode string) error
	RemoveMerchantCode(ctx context.Context, mCode string) error
	GetAllMerchantCode(ctx context.Context) ([]string, error)
}

type merchantRedisRepository struct {
	client     *redis.Client
	keyPrefix  string
	expiration time.Duration
}

func NewMerchantRedisRepository(client *redis.Client) MerchantRedisRepository {
	return &merchantRedisRepository{
		client:     client,
		keyPrefix:  "active_merchants",
		expiration: 0,
	}
}

func (r *merchantRedisRepository) SetActiveMerchantCode(ctx context.Context, mCodes []string) error {
	if len(mCodes) == 0 {
		return fmt.Errorf("list of merchant code is empty")
	}

	key := r.keyPrefix
	members := make([]any, len(mCodes))
	for index, value := range mCodes {
		members[index] = value
	}

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to clear previous merchant codes in redis: %w", err)
	}

	if err := r.client.SAdd(ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("failed to set active merchant codes: %w", err)
	}

	return nil
}

func (r *merchantRedisRepository) IsMerchantCodeActive(ctx context.Context, mCode string) (bool, error) {
	key := r.keyPrefix
	exists, err := r.client.SIsMember(ctx, key, mCode).Result()

	if err != nil {
		return false, fmt.Errorf("failed to check merchant code in redis: %w", err)
	}

	return exists, nil
}

func (r *merchantRedisRepository) AddMerchantCode(ctx context.Context, mCode string) error {
	key := r.keyPrefix

	if err := r.client.SAdd(ctx, key, mCode).Err(); err != nil {
		return fmt.Errorf("failed to add merchant code to redis: %w", err)
	}

	return nil
}

func (r *merchantRedisRepository) RemoveMerchantCode(ctx context.Context, mCode string) error {
	key := r.keyPrefix
	if err := r.client.SRem(ctx, key, mCode).Err(); err != nil {
		return fmt.Errorf("failed to remove merchant code from redis: %w", err)
	}
	return nil
}

func (r *merchantRedisRepository) GetAllMerchantCode(ctx context.Context) ([]string, error) {
	key := r.keyPrefix
	codes, err := r.client.SMembers(ctx, key).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get all active merchant codes from redis: %w", err)
	}

	return codes, nil
}
