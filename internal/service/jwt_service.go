package service

import (
	"context"
	"fmt"
	"time"

	dto "briefcash-jwt/internal/dto"
	model "briefcash-jwt/internal/entity"
	clock "briefcash-jwt/internal/helper/timehelper"
	repo "briefcash-jwt/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService interface {
	GenerateToken(ctx context.Context, dto dto.JwtRequest) (*dto.JwtResponse, error)
	ValidateToken(ctx context.Context, stringToken string) (*jwt.Token, error)
	BlacklistToken(ctx context.Context, stringToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*dto.JwtResponse, error)
}

type tokenService struct {
	jwtRepo   repo.JwtRepository
	redisRepo repo.RedisRepository
	jwtSecret string
}

func NewTokenService(jr repo.JwtRepository, rr repo.RedisRepository, secret string) TokenService {
	return &tokenService{jr, rr, secret}
}

func (ts *tokenService) GenerateToken(ctx context.Context, dto dto.JwtRequest) (*dto.JwtResponse, error) {
	expiration := time.Now().Add(15 * time.Minute)
	var accessToken, refreshToken string
	var err error

	if typ := dto.Type; typ == "access" {
		accessTokenClaims := jwt.MapClaims{
			"user_id": dto.UserID,
			"type":    dto.Type,
			"exp":     expiration.Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
		accessToken, err = token.SignedString([]byte(ts.jwtSecret))
		if err != nil {
			return nil, err
		}
	} else {
		refreshTokenClaims := jwt.MapClaims{
			"user_id": dto.UserID,
			"type":    dto.Type,
			"exp":     expiration.Unix(),
		}

		rToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
		refreshToken, err = rToken.SignedString([]byte(ts.jwtSecret))
		if err != nil {
			return nil, err
		}
	}

	tokenEntity := &model.JwtToken{
		MerchantID:   dto.UserID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiration,
	}

	if err := ts.jwtRepo.Create(tokenEntity); err != nil {
		return nil, err
	}

	ts.redisRepo.SetToken(ctx, accessToken, "valid", time.Duration(expiration.Minute()))
	return TokenResponse(tokenEntity), nil
}

func (ts *tokenService) ValidateToken(ctx context.Context, stringToken string) (*jwt.Token, error) {
	if _, err := ts.redisRepo.GetToken(ctx, stringToken); err != nil {
		return nil, fmt.Errorf("token invalid or blacklisted")
	}

	token, err := jwt.Parse(stringToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(ts.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (ts *tokenService) BlacklistToken(ctx context.Context, stringToken string) error {
	if err := ts.redisRepo.DeleteToken(ctx, stringToken); err != nil {
		return err
	}

	return ts.redisRepo.SetToken(ctx, "blacklist:"+stringToken, "true", 15*time.Minute)
}

func (ts *tokenService) RefreshToken(ctx context.Context, refreshToken string) (*dto.JwtResponse, error) {
	oldToken, err := ts.jwtRepo.FindByRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	_ = ts.BlacklistToken(ctx, oldToken.AccessToken)
	refToken := dto.JwtRequest{
		UserID: oldToken.MerchantID,
		Type:   "refresh",
	}
	return ts.GenerateToken(ctx, refToken)
}

func TokenResponse(token *model.JwtToken) *dto.JwtResponse {
	return &dto.JwtResponse{
		UserID:       token.MerchantID,
		RefreshToken: token.RefreshToken,
		AccessToken:  token.AccessToken,
		CreatedAt:    clock.FormatTimeToISO7(token.CreatedAt),
		ExpiresAt:    clock.FormatTimeToISO7(token.ExpiresAt),
	}
}
