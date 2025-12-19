package service

import (
	"context"
	"fmt"
	"time"

	log "briefcash-jwt/internal/helper/loghelper"
	repo "briefcash-jwt/internal/repository"
)

type MerchantService interface {
	CachingCode(ctx context.Context) error
	ValidateMerchantCode(ctx context.Context, mCode string) (bool, error)
	AddMerchantCode(ctx context.Context, mCode string) error
	RemoveMerchantCode(ctx context.Context, mCode string) error
}

type merchantService struct {
	dbRepo    repo.MerchantRepository
	redisRepo repo.MerchantRedisRepository
}

func NewMerchantService(dbRepo repo.MerchantRepository, redisRepo repo.MerchantRedisRepository) MerchantService {
	return &merchantService{
		dbRepo:    dbRepo,
		redisRepo: redisRepo,
	}
}

func (s *merchantService) CachingCode(ctx context.Context) error {
	start := time.Now()

	defer func() {
		log.Logger.Infof("Load active merchant code finished in %s", time.Since(start))
	}()

	log.Logger.Info("Retrieve list of merchant codes from db")
	listCodes, err := s.dbRepo.GetAllActiveCode(ctx)
	if err != nil {
		log.Logger.WithError(err).Error("Failed to retrieve list of merchant code from db")
		return fmt.Errorf("failed to load merchant code: %w", err)
	}

	log.Logger.Infof("Total list of merchant codes: %d", len(listCodes))
	if len(listCodes) == 0 {
		log.Logger.Warn("No active merchant codes found in db")
		return fmt.Errorf("no active merchant code found in db")
	}

	log.Logger.Info("Load list of merchant codes to redis")
	if err := s.redisRepo.SetActiveMerchantCode(ctx, listCodes); err != nil {
		log.Logger.WithError(err).Error("Failed load merchant code list to redis")
		return fmt.Errorf("failed to load merchant code: %w", err)
	}

	log.Logger.Info("List of merchant codes successfully loaded to redis")

	return nil
}

func (s *merchantService) ValidateMerchantCode(ctx context.Context, mCode string) (bool, error) {
	rec := log.Logger.WithField("merchant_code", mCode)
	rec.Info("Checking active merchant code in redis")
	isActive, err := s.redisRepo.IsMerchantCodeActive(ctx, mCode)
	if err != nil {
		rec.WithError(err).Error("Failed validate merchant code in redis")
		return false, fmt.Errorf("failed to validate merchant code: %w", err)
	}

	status := func() string {
		if isActive {
			return "active"
		}
		return "inactive"
	}()

	rec.Infof("Merchant code is %s in redis", status)

	return isActive, nil
}

func (s *merchantService) AddMerchantCode(ctx context.Context, mCode string) error {
	rec := log.Logger.WithField("merchant_code", mCode)

	rec.Info("Check existing merchant code in redis")
	exist, err := s.redisRepo.IsMerchantCodeActive(ctx, mCode)

	if err != nil {
		rec.WithError(err).Error("Failed to check merchant code in redis")
		return fmt.Errorf("failed to check merchant code")
	}

	if exist {
		rec.Warn("Merchant code already exists in redis")
		return fmt.Errorf("merchant code already exists")
	}

	rec.Info("Add new merchant code to redis")
	if err := s.redisRepo.AddMerchantCode(ctx, mCode); err != nil {
		rec.WithError(err).Error("Failed to add merchant code to redis")
		return fmt.Errorf("failed to add merchant code: %w", err)
	}

	rec.Info("merchant code successfully added to redis")

	return nil
}

func (s *merchantService) RemoveMerchantCode(ctx context.Context, mCode string) error {
	rec := log.Logger.WithField("merchant_code", mCode)

	rec.Info("Check existing merchant code in redis")
	exist, err := s.redisRepo.IsMerchantCodeActive(ctx, mCode)

	if err != nil {
		rec.WithError(err).Error("Failed to check merchant code in redis")
		return fmt.Errorf("failed to check merchant code")
	}

	if !exist {
		rec.Warn("Merchant code does not exist in redis")
		return fmt.Errorf("merchant code does not exist")
	}

	rec.Info("Remove existing merchant code from redis")
	if err := s.redisRepo.RemoveMerchantCode(ctx, mCode); err != nil {
		rec.WithError(err).Error("Failed to remove merchant code from redis")
		return fmt.Errorf("failed to remove merchant code: %w", err)
	}

	rec.Info("Merchant code successfully removed from redis")

	return nil
}
