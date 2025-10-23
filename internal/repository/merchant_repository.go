package repository

import (
	jwt "briefcash-jwt/internal/entity"
	"context"
	"errors"

	"gorm.io/gorm"
)

type MerchantRepository interface {
	GetAllActiveCode(ctx context.Context) ([]string, error)
	GetByCode(ctx context.Context, userID string) (*jwt.Merchant, error)
}

type merchantRepository struct {
	db *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) MerchantRepository {
	return &merchantRepository{db}
}

func (m *merchantRepository) GetAllActiveCode(ctx context.Context) ([]string, error) {
	var merchant jwt.Merchant
	var listOfCodes []string

	if err := m.db.WithContext(ctx).Model(&merchant).Table("merchant").Where("is_active = ?", true).Order("code ASC").Pluck("code", &listOfCodes).Error; err != nil {
		return nil, err
	}

	return listOfCodes, nil
}

func (m *merchantRepository) GetByCode(ctx context.Context, code string) (*jwt.Merchant, error) {
	var merchant jwt.Merchant

	if err := m.db.Table("merchant").WithContext(ctx).Where("code = ?", code).First(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &merchant, nil
}
