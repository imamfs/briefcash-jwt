package test

import (
	dto "briefcash-jwt/internal/dto"
	service "briefcash-jwt/internal/service"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateToken_Success(t *testing.T) {
	jwtRepo := new(MockJWTRepository)
	redisRepo := new(MockRedisRepository)
	svc := service.NewMockTokenService(jwtRepo, redisRepo, "imamfahruzi")

	req := dto.JwtRequest{
		UserID: "STARK-1225",
		Type:   "access",
	}

	jwtRepo.On("WithTransaction", mock.Anything).Return(jwtRepo)
	jwtRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.JwtToken")).Return(nil)
	redisRepo.On("SetToken", mock.Anything, mock.Anything, "valid", mock.Anything).Return(nil)
	resp, err := svc.GenerateToken(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	jwtRepo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*entity.JwtToken"))

}
