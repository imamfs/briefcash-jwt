package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	dto "briefcash-jwt/internal/dto"
	model "briefcash-jwt/internal/entity"
	logs "briefcash-jwt/internal/helper/loghelper"
	mask "briefcash-jwt/internal/helper/securityhelper"
	clock "briefcash-jwt/internal/helper/timehelper"
	repo "briefcash-jwt/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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
	db        *gorm.DB
	jwtSecret string
}

func NewTokenService(jr repo.JwtRepository, rr repo.RedisRepository, db *gorm.DB, secret string) TokenService {
	return &tokenService{jr, rr, db, secret}
}

func NewMockTokenService(jr repo.JwtRepository, rr repo.RedisRepository, secret string) TokenService {
	return &tokenService{
		jwtRepo:   jr,
		redisRepo: rr,
		db:        nil,
		jwtSecret: secret,
	}
}

func (ts *tokenService) GenerateToken(ctx context.Context, req dto.JwtRequest) (*dto.JwtResponse, error) {
	log := logs.Logger.WithField("user_id", req.UserID)
	log.Infof("Generating %s token", req.Type)

	var expAccessToken, expRefreshToken time.Time
	var signedAccessToken, signedRefreshToken string
	var err error

	expAccessToken = time.Now().Add(ts.tokenTTL("access"))
	expRefreshToken = time.Now().Add(ts.tokenTTL("refresh"))

	accessClaims := ts.buildClaims(req.UserID, expAccessToken)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccessToken, err = token.SignedString([]byte(ts.jwtSecret))
	if err != nil {
		log.WithError(err).Error("Failed signing access token")
		return nil, fmt.Errorf("failed to generate access token")
	}

	refreshClaims := ts.buildRefreshClaims(req.UserID, expRefreshToken)
	rtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err = rtoken.SignedString([]byte(ts.jwtSecret))
	if err != nil {
		log.WithError(err).Error("Failed signing refresh token")
		return nil, fmt.Errorf("failed to generate refresh token")
	}

	tokenEntity := &model.JwtToken{
		MerchantID:   req.UserID,
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    expAccessToken,
		IsRevoke:     false,
	}

	log.Infof("Saving %s token to database", req.Type)
	if err := ts.saveToken(ctx, tokenEntity); err != nil {
		log.WithError(err).Errorf("Failed saving %s token to database", req.Type)
		return nil, err
	}

	log.Infof("Saving %s token to redis", req.Type)
	if err := ts.redisRepo.SetToken(ctx, signedAccessToken, "valid", time.Until(expAccessToken)); err != nil {
		log.WithError(err).Warnf("Failed to cache %s token to redis", req.Type)
	}

	log.Infof("%s token successfully generated", req.Type)

	return tokenResponse(tokenEntity), nil
}

func (ts *tokenService) ValidateToken(ctx context.Context, stringToken string) (*jwt.Token, error) {

	masked := mask.MaskToken(stringToken)
	log := logs.Logger.WithField("token", masked)

	log.Infof("Check token in redis")

	val, err := ts.redisRepo.GetToken(ctx, stringToken)

	if errors.Is(err, repo.ErrTokenNotFound) {
		log.Warn("Token not found in redis, checking in database")

		tokenData, dbErr := ts.jwtRepo.FindByAccessToken(ctx, stringToken)
		if dbErr != nil {
			log.WithError(dbErr).Error("Token not found in database")
			return nil, fmt.Errorf("token invalid or blacklisted")
		}

		go func() {
			if err := ts.redisRepo.SetToken(context.Background(), tokenData.AccessToken, "valid", time.Until(tokenData.ExpiresAt)); err != nil {
				log.WithError(err).Warn("Failed to cache token into redis")
			}
		}()

		val = "valid"
	} else if err != nil {
		log.WithError(err).Error("Redis error while checking token")
		return nil, fmt.Errorf("temporary cache issue, please retry")
	}

	if val != "valid" {
		log.Warnf("Token found in redis, but invalid state: %s", val)
		return nil, fmt.Errorf("token invalid or blacklisted")
	}

	log.Infof("Parsing JWT for token %s", masked)
	token, err := jwt.Parse(stringToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(ts.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Warn("Token has expired")
			return nil, fmt.Errorf("token expired")
		}

		log.WithError(err).Error("Unexpected error while parsing token")
		return nil, err
	}

	log.Info("Token is valid")

	return token, nil
}

func (ts *tokenService) BlacklistToken(ctx context.Context, stringToken string) error {
	masked := mask.MaskToken(stringToken)
	log := logs.Logger.WithField("token", masked)

	log.Info("Blacklisting token")

	if err := ts.redisRepo.DeleteToken(ctx, stringToken); err != nil {
		log.WithError(err).Warn("Failed to delete token from Redis (possibly already removed)")
	}

	tokenData, err := ts.jwtRepo.FindByAccessToken(ctx, stringToken)
	if err != nil {
		log.WithError(err).Warn("Token not found in database, skipping database deletion")
		return err
	} else {
		if err := ts.deleteToken(ctx, stringToken); err != nil {
			log.WithError(err).Error("Failed to delete token from database")
			return err
		}
	}

	ttl := 15 * time.Minute
	if tokenData != nil {
		remaining := time.Until(tokenData.ExpiresAt)
		if remaining > ttl {
			ttl = remaining
		}
	}

	if err := ts.redisRepo.SetToken(ctx, "blacklist:"+stringToken, "true", ttl); err != nil {
		log.WithError(err).Error("Failed to store blacklist token in redis")
		return err
	}

	log.Infof("Token successfully blacklisted (TTL: %s)", ttl)

	return nil
}

func (ts *tokenService) RefreshToken(ctx context.Context, refreshToken string) (*dto.JwtResponse, error) {
	masked := mask.MaskToken(refreshToken)
	log := logs.Logger.WithField("token", masked)

	log.Info("Check refresh token in database")

	oldToken, err := ts.jwtRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	log.Info("Blacklist old token")
	if err := ts.BlacklistToken(ctx, oldToken.AccessToken); err != nil {
		logs.Logger.WithError(err).Warnf("Failed to blacklist old token")
	}

	refToken := dto.JwtRequest{
		UserID: oldToken.MerchantID,
		Type:   "refresh",
	}

	return ts.GenerateToken(ctx, refToken)
}

func tokenResponse(token *model.JwtToken) *dto.JwtResponse {
	return &dto.JwtResponse{
		UserID:       token.MerchantID,
		RefreshToken: token.RefreshToken,
		AccessToken:  token.AccessToken,
		CreatedAt:    clock.FormatTimeToISO7(token.CreatedAt),
		ExpiresAt:    clock.FormatTimeToISO7(token.ExpiresAt),
	}
}

func (ts *tokenService) saveToken(ctx context.Context, token *model.JwtToken) error {

	if ts.db == nil {
		return ts.jwtRepo.Create(ctx, token)
	}

	return ts.db.Transaction(func(tx *gorm.DB) error {
		repoTx := ts.jwtRepo.WithTransaction(tx)
		return repoTx.Create(ctx, token)
	})
}

func (ts *tokenService) deleteToken(ctx context.Context, stringToken string) error {
	if ts.db == nil {
		return ts.jwtRepo.DeleteAccessByToken(ctx, stringToken)
	}

	return ts.db.Transaction(func(tx *gorm.DB) error {
		repoTx := ts.jwtRepo.WithTransaction(tx)
		return repoTx.DeleteAccessByToken(ctx, stringToken)
	})
}

func (ts *tokenService) buildClaims(userID string, exp time.Time) jwt.MapClaims {
	return jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"exp":     exp.Unix(),
	}
}

func (ts *tokenService) buildRefreshClaims(userID string, exp time.Time) jwt.MapClaims {
	return jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     exp.Unix(),
	}
}

func (ts *tokenService) tokenTTL(tokenType string) time.Duration {
	switch tokenType {
	case "refresh":
		return 7 * 24 * time.Hour
	default:
		return 15 * time.Minute
	}
}
