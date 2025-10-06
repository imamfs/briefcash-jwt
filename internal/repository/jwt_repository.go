package repository

import (
	jwt "briefcash-jwt/internal/entity"
	logs "briefcash-jwt/internal/helper/loghelper"

	"gorm.io/gorm"
)

type JwtRepository interface {
	Create(jwt *jwt.JwtToken) error
	FindByRefreshToken(token string) (*jwt.JwtToken, error)
	DeleteAccessByToken(token string) error
}

type jwtRepository struct {
	db *gorm.DB
}

func NewJwtRepository(db *gorm.DB) JwtRepository {
	return &jwtRepository{db}
}

func (r *jwtRepository) Create(jwt *jwt.JwtToken) error {
	return r.db.Create(jwt).Error
}

func (r *jwtRepository) FindByRefreshToken(token string) (*jwt.JwtToken, error) {
	var tkn jwt.JwtToken
	if err := r.db.Where("refresh_token = ?", token).First(&tkn).Error; err != nil {
		logs.Logger.Error("refresh token not found")
		return nil, err
	}
	return &tkn, nil
}

func (r *jwtRepository) DeleteAccessByToken(token string) error {
	var tkn jwt.JwtToken
	return r.db.Where("access_token = ?", token).Delete(&tkn).Error
}
