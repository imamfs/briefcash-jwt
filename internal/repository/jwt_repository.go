package repository

import (
	jwt "briefcash-jwt/internal/entity"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type JwtRepository interface {
	Save(ctx context.Context, jwt *jwt.JwtToken) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (*jwt.JwtToken, error)
	FindByAccessToken(ctx context.Context, accessToken string) (*jwt.JwtToken, error)
	DeleteByAccessToken(ctx context.Context, accessToken string) error
	WithTransaction(tx *gorm.DB) JwtRepository
}

type jwtRepository struct {
	db *gorm.DB
}

func NewJwtRepository(db *gorm.DB) JwtRepository {
	return &jwtRepository{db}
}

// Save jwt token
func (r *jwtRepository) Save(ctx context.Context, jwt *jwt.JwtToken) error {
	return r.db.WithContext(ctx).Table("jwt_token").Create(jwt).Error
}

// Get jwt token find by access token
func (r *jwtRepository) FindByAccessToken(ctx context.Context, accessToken string) (*jwt.JwtToken, error) {
	var token jwt.JwtToken
	err := r.db.WithContext(ctx).Table("jwt_token").
		Where("access_token = ?", accessToken).
		First(&token).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("access token not found")
	}

	if err != nil {
		return nil, err
	}

	return &token, nil
}

// Get jwt token find by refresh token
func (r *jwtRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*jwt.JwtToken, error) {
	var tkn jwt.JwtToken
	if err := r.db.WithContext(ctx).Table("jwt_token").
		Where("refresh_token = ?", refreshToken).First(&tkn).Error; err != nil {
		return nil, err
	}
	return &tkn, nil
}

// Delete jwt token by access token
func (r *jwtRepository) DeleteByAccessToken(ctx context.Context, accessToken string) error {
	return r.db.WithContext(ctx).Table("jwt_token").
		Where("access_token = ?", accessToken).Delete(&jwt.JwtToken{}).Error
}

func (r *jwtRepository) WithTransaction(trx *gorm.DB) JwtRepository {
	return &jwtRepository{db: trx}
}
