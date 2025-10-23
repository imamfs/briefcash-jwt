package repository

import (
	jwt "briefcash-jwt/internal/entity"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type JwtRepository interface {
	Create(ctx context.Context, jwt *jwt.JwtToken) error
	FindByRefreshToken(ctx context.Context, token string) (*jwt.JwtToken, error)
	FindByAccessToken(ctx context.Context, stringToken string) (*jwt.JwtToken, error)
	DeleteAccessByToken(ctx context.Context, token string) error
	WithTransaction(tx *gorm.DB) JwtRepository
}

type jwtRepository struct {
	db *gorm.DB
}

func NewJwtRepository(db *gorm.DB) JwtRepository {
	return &jwtRepository{db}
}

func (r *jwtRepository) Create(ctx context.Context, jwt *jwt.JwtToken) error {
	return r.db.WithContext(ctx).Table("jwt_token").Create(jwt).Error
}

func (r *jwtRepository) FindByAccessToken(ctx context.Context, tokenString string) (*jwt.JwtToken, error) {
	var token jwt.JwtToken

	err := r.db.WithContext(ctx).Table("jwt_token").Where("access_token = ?", tokenString).First(&token).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("token not found")
	}

	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *jwtRepository) FindByRefreshToken(ctx context.Context, token string) (*jwt.JwtToken, error) {
	var tkn jwt.JwtToken
	if err := r.db.WithContext(ctx).Table("jwt_token").Where("refresh_token = ?", token).First(&tkn).Error; err != nil {
		return nil, err
	}
	return &tkn, nil
}

func (r *jwtRepository) DeleteAccessByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Table("jwt_token").Where("access_token = ?", token).Delete(&jwt.JwtToken{}).Error
}

func (r *jwtRepository) WithTransaction(tx *gorm.DB) JwtRepository {
	return &jwtRepository{db: tx}
}
