package repository

import (
	jwt "briefcash-jwt/internal/entity"

	"gorm.io/gorm"
)

type JwtRepository interface {
	Create(jwt *jwt.JwtToken) error
	FindByToken(token string) (*jwt.JwtToken, error)
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

func (r *jwtRepository) FindByToken(token string) (*jwt.JwtToken, error) {
	var tkn jwt.JwtToken
	if err := r.db.First(&tkn, token).Error; err != nil {
		return nil, err
	}
	return &tkn, nil
}
