package middleware

import (
	"briefcash-jwt/internal/dto"
	service "briefcash-jwt/internal/service"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type contextKey string

const tokenKey contextKey = "token"

func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}

type Middleware struct {
	svc service.MerchantService
}

func NewMiddleware(svc service.MerchantService) *Middleware {
	return &Middleware{svc}
}

// Middleware function to validate http header and payload
func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			Code string `json:"merchant_code"`
		}

		// Validate Authorization header
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, dto.JwtDataResponse{
				Status:  false,
				Message: "Missing authorization header",
				Data:    map[string]any{},
			})
			c.Abort()
			return
		}

		// Validate bearer variable
		if !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, dto.JwtDataResponse{
				Status:  false,
				Message: "Invalid authorization format, should be 'Bearer '",
				Data:    map[string]any{},
			})
			c.Abort()
			return
		}

		authToken := strings.TrimPrefix(auth, "Bearer ")
		if authToken == "" {
			c.JSON(http.StatusUnauthorized, dto.JwtDataResponse{
				Status:  false,
				Message: "empty bearer token",
				Data:    map[string]any{},
			})
			c.Abort()
			return
		}

		// Validate payload request
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, dto.JwtDataResponse{
				Status:  false,
				Message: "Invalid body request",
				Data:    map[string]any{},
			})
			c.Abort()
			return
		}

		// Validate merchant code
		if payload.Code == "" {
			c.JSON(http.StatusBadRequest, dto.JwtDataResponse{
				Status:  false,
				Message: "Merchant code is empty",
				Data:    map[string]any{},
			})
			c.Abort()
			return
		}

		isValid, err := m.svc.ValidateMerchantCode(c.Request.Context(), payload.Code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.JwtDataResponse{
				Status:  false,
				Message: "Failed to validate merchant code",
				Data:    map[string]any{},
			})
			c.Abort()
			return
		}

		if !isValid {
			c.JSON(http.StatusUnauthorized, dto.JwtDataResponse{
				Status:  false,
				Message: "Invalid or inactive merchant code",
				Data:    map[string]any{},
			})
			c.Abort()
			return
		}

		// Store access token to context
		c.Set("token", authToken)
		c.Next()
	}
}
