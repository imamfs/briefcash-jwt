package service

import (
	dto "briefcash-jwt/internal/dto"
	model "briefcash-jwt/internal/entity"
	repo "briefcash-jwt/internal/repository"
	"fmt"

	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type MockJWTRepository struct {
	CreateErr                error
	FindByAccessTokenResult  *model.JwtToken
	FindByAccessTokenErr     error
	FindByRefreshTokenResult *model.JwtToken
	FindByRefreshTokenErr    error
	DeleteByAccessTokenErr   error
	WithTransactionCancelled bool
}

func (m *MockJWTRepository) Create(ctx context.Context, token *model.JwtToken) error {
	return m.CreateErr
}

func (m *MockJWTRepository) WithTransaction(tx *gorm.DB) repo.JwtRepository {
	m.WithTransactionCancelled = true
	return m
}

func (m *MockJWTRepository) FindByAccessToken(ctx context.Context, token string) (*model.JwtToken, error) {
	return m.FindByAccessTokenResult, m.FindByAccessTokenErr
}

func (m *MockJWTRepository) FindByRefreshToken(ctx context.Context, token string) (*model.JwtToken, error) {
	return m.FindByRefreshTokenResult, m.FindByRefreshTokenErr
}

func (m *MockJWTRepository) DeleteAccessByToken(ctx context.Context, token string) error {
	return m.DeleteByAccessTokenErr
}

type MockRedisRepository struct {
	Store map[string]string
	Err   error
}

func (r *MockRedisRepository) SetToken(ctx context.Context, key, value string, ttl time.Duration) error {
	if r.Store == nil {
		r.Store = make(map[string]string)
	}
	if r.Err != nil {
		return r.Err
	}
	r.Store[key] = value
	return nil
}

func (r *MockRedisRepository) GetToken(ctx context.Context, key string) (string, error) {
	if r.Err != nil {
		return "", r.Err
	}

	val, ok := r.Store[key]
	if !ok {
		return "", errors.New("token not found")
	}

	if val == "nil" {
		return "", fmt.Errorf("token not found in redis")
	}

	return val, nil
}

func (r *MockRedisRepository) DeleteToken(ctx context.Context, key string) error {
	if r.Err != nil {
		return r.Err
	}
	delete(r.Store, key)
	return nil
}

func (r *MockRedisRepository) ExistToken(ctx context.Context, key string) (bool, error) {
	_, ok := r.Store[key]

	if !ok {
		return false, r.Err
	}

	return true, nil
}

func (r *MockRedisRepository) ExpireToken(ctx context.Context, key string, ttl time.Duration) error {
	if r.Err != nil {
		return r.Err
	}
	delete(r.Store, key)
	return nil
}

func TestGenerateToken(t *testing.T) {
	jr := &MockJWTRepository{}
	rr := &MockRedisRepository{Store: make(map[string]string)}
	svc := NewMockTokenService(jr, rr, "imamfahruzi")
	request := dto.JwtRequest{UserID: "STARK-1225", Type: "access"}

	resp, err := svc.GenerateToken(context.Background(), request)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.AccessToken == "" {
		t.Fatal("expected access token not empty")
	}
}

func TestGenerateToken_FailedUserId(t *testing.T) {
	jr := &MockJWTRepository{}
	rr := &MockRedisRepository{Store: make(map[string]string)}
	svc := NewMockTokenService(jr, rr, "imamfahruzi")
	request := dto.JwtRequest{UserID: "", Type: "access"}

	resp, err := svc.GenerateToken(context.Background(), request)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	if err.Error() != "user id not found" {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Fatal("expected access token empty")
	}
}

func TestValidateToken_Success(t *testing.T) {

	jr := &MockJWTRepository{}
	rr := &MockRedisRepository{Store: make(map[string]string)}
	svc := NewMockTokenService(jr, rr, "imamfahruzi")

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "STARK-1225",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString([]byte("imamfahruzi"))

	rr.Store[token] = "valid"

	resp, err := svc.ValidateToken(context.Background(), token)

	if err != nil {
		t.Fatalf("Expected no error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response value")
	}

}

func TestValidateToken_Failed(t *testing.T) {

	jr := &MockJWTRepository{}
	rr := &MockRedisRepository{Store: make(map[string]string)}
	svc := NewMockTokenService(jr, rr, "imamfahruzi")

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "STARK-1225",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}).SignedString([]byte("imamfahruzi"))
	rr.Store[token] = "invalid"

	resp, err := svc.ValidateToken(context.Background(), token)
	fmt.Printf("error: %v \n", err)

	if err == nil {
		t.Fatalf("Expected error but got nil")
	}

	if resp != nil {
		t.Fatal("expected no return value")
	}

}

func TestBlacklistToken_Success(t *testing.T) {

	exp := time.Now().Add(15 * time.Minute)
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "STARK-1225",
		"exp":     exp.Unix(),
	}).SignedString([]byte("imamfahruzi"))
	jr := &MockJWTRepository{
		FindByAccessTokenResult: &model.JwtToken{
			AccessToken: token,
			ExpiresAt:   exp,
		},
	}
	rr := &MockRedisRepository{Store: make(map[string]string)}
	svc := NewMockTokenService(jr, rr, "imamfahruzi")
	err := svc.BlacklistToken(context.Background(), token)

	if err != nil {
		t.Fatalf("Expected no error: %v", err)
	}

}

func TestBlacklistToken_Failed(t *testing.T) {

	exp := time.Now().Add(15 * time.Minute)
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "STARK-1225",
		"exp":     exp.Unix(),
	}).SignedString([]byte("imamfahruzi"))
	jr := &MockJWTRepository{
		FindByAccessTokenResult: nil,
		FindByAccessTokenErr:    fmt.Errorf("token not found"),
	}
	rr := &MockRedisRepository{Store: make(map[string]string)}
	svc := NewMockTokenService(jr, rr, "svc")

	err := svc.BlacklistToken(context.Background(), token)
	fmt.Printf("error:%v", err)

	if err == nil {
		t.Fatalf("expected error occured, but received empty error instead: %v", err)
	}
}

func TestRefreshToken_Success(t *testing.T) {

	exp := time.Now().Add(15 * time.Minute)
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "STARK-1225",
		"exp":     exp.Unix(),
	}).SignedString([]byte("imamfahruzi"))
	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "STARK-1225",
		"exp":     exp.Unix(),
	}).SignedString([]byte("imamfahruzi"))
	jr := &MockJWTRepository{
		FindByRefreshTokenResult: &model.JwtToken{
			AccessToken:  token,
			RefreshToken: refreshToken,
			MerchantID:   "STARK-1225",
			ExpiresAt:    exp,
		},
		FindByAccessTokenResult: &model.JwtToken{
			AccessToken: token,
			ExpiresAt:   exp,
		},
	}
	rr := &MockRedisRepository{Store: make(map[string]string)}
	svc := NewMockTokenService(jr, rr, "imamfahruzi")
	resp, err := svc.RefreshToken(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("Expected no error occured: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected no empty result")
	}

}
