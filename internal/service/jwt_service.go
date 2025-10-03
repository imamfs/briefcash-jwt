package service

import (
	model "briefcash-jwt/internal/entity"
	repo "briefcash-jwt/internal/repository"
)

type JwtService interface {
	GenerateToken(merchantCode string) (string, error)
	GetToken(token string) (*model.JwtToken, error)
}

type jwtService struct {
	repo repo.JwtRepository
}

func NewJwtService(r repo.JwtRepository) JwtService {
	return &jwtService{r}
}

func (r *jwtService) GenerateToken(merchantCode string) (string, error) {
	return "nil", nil
}

func (r *jwtService) GetToken(token string) (*model.JwtToken, error) {
	return r.repo.FindByToken(token)
}
