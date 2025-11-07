package controller

import (
	dto "briefcash-jwt/internal/dto"
	loghelper "briefcash-jwt/internal/helper/loghelper"
	"briefcash-jwt/internal/helper/securityhelper"
	middleware "briefcash-jwt/internal/middleware"
	service "briefcash-jwt/internal/service"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type TokenController struct {
	TokenService service.TokenService
}

func NewTokenController(s service.TokenService) *TokenController {
	return &TokenController{s}
}

func (c *TokenController) GenerateToken(ctx *gin.Context) {
	start := time.Now()
	log := loghelper.Logger.WithField("service", "jwt_controller")

	defer func() {
		log.WithField("elapsed_time", time.Since(start).Milliseconds()).Info("JWT Token successfully generated")
	}()

	log.WithField("step", "decode_payload").Info("Decoding JSON payload to Struct")
	var tokenRequest dto.JwtRequest
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&tokenRequest); err != nil {
		log.WithField("step", "decode_payload").WithError(err).Error("Failed to decode JSON payload")
		ctx.JSON(http.StatusBadRequest, dto.JwtDataResponse{
			Status:  false,
			Message: "Invalid request body",
			Data:    map[string]any{},
		})
		return
	}

	log.WithFields(logrus.Fields{
		"step":     "generate_token",
		"user_id":  tokenRequest.UserID,
		"jwt_type": tokenRequest.Type,
	}).Info("Processing generate JWT Token")
	token, err := c.TokenService.GenerateToken(ctx.Request.Context(), tokenRequest)
	if err != nil {
		log.WithField("step", "generate_token").WithError(err).Error("Failed to generate JWT Token")
		ctx.JSON(http.StatusInternalServerError, dto.JwtDataResponse{
			Status:  false,
			Message: "Failed to generate JWT Token, internal error",
			Data:    map[string]any{},
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.JwtDataResponse{
		Status:  true,
		Message: "SUCCESS",
		Data:    token,
	})
}

func (c *TokenController) ValidateToken(ctx *gin.Context) {
	start := time.Now()
	log := loghelper.Logger.WithField("service", "jwt_validation_controller")

	defer func() {
		log.WithField("elapsed_time", time.Since(start).Milliseconds()).Info("JWT Token successfully validated")
	}()

	log.WithField("step", "jwt_context").Info("Retrieving JWT Token from context")
	tokenString, ok := middleware.GetTokenFromContext(ctx.Request.Context())
	if !ok {
		log.WithField("step", "jwt_context").Error("JWT Token not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.JwtDataResponse{
			Status:  false,
			Message: "JWT Token not found",
			Data:    map[string]any{},
		})
		return
	}

	log.WithFields(logrus.Fields{
		"step":  "jwt_validation",
		"token": securityhelper.MaskToken(tokenString),
	}).Info("Validating JWT Token")
	token, err := c.TokenService.ValidateToken(ctx.Request.Context(), tokenString)

	if err != nil {
		log.WithField("step", "jwt_validation").WithError(err).Error("JWT Token failed to validate")
		ctx.JSON(http.StatusUnauthorized, dto.JwtDataResponse{
			Status:  false,
			Message: "JWT Token mismatched",
			Data:    map[string]any{},
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.JwtDataResponse{
		Status:  true,
		Message: "SUCCESS",
		Data: map[string]any{
			"valid":  true,
			"claims": token.Claims,
		},
	})
}

func (c *TokenController) Logout(ctx *gin.Context) {
	start := time.Now()
	log := loghelper.Logger.WithField("service", "jwt_logout_controller")

	defer func() {
		log.WithField("elapsed_time", time.Since(start).Milliseconds()).Info("JWT Token successfully removed from system")
	}()

	log.WithField("step", "jwt_context").Info("Retrieving JWT Token from context")
	tokenString, ok := middleware.GetTokenFromContext(ctx.Request.Context())
	if !ok {
		log.WithField("step", "jwt_context").Error("JWT Token not found in context")
		ctx.JSON(http.StatusUnauthorized, dto.JwtDataResponse{
			Status:  false,
			Message: "JWT Token not found",
			Data:    map[string]any{},
		})
		return
	}

	log.WithFields(logrus.Fields{
		"step":  "jwt_blacklist",
		"token": securityhelper.MaskToken(tokenString),
	}).Info("Validating JWT Token")
	err := c.TokenService.BlacklistToken(ctx.Request.Context(), tokenString)

	if err != nil {
		log.WithField("step", "jwt_blacklist").WithError(err).Error("JWT Token failed to removed from system")
		ctx.JSON(http.StatusInternalServerError, dto.JwtDataResponse{
			Status:  false,
			Message: "Failed to logout JWT Token, internal error",
			Data:    map[string]any{},
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.JwtDataResponse{
		Status:  true,
		Message: "SUCCESS",
		Data:    "logout success",
	})
}

func (c *TokenController) RefreshToken(ctx *gin.Context) {
	start := time.Now()
	log := loghelper.Logger.WithField("service", "jwt_refresh_controller")

	defer func() {
		log.WithField("elapsed_time", time.Since(start).Milliseconds()).Info("Refresh JWT Token successfully generated")
	}()

	log.WithField("step", "decode_payload").Info("Decoding JSON payload to Struct")

	var req dto.JwtRefreshToken
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		log.WithField("step", "decode_payload").WithError(err).Error("Failed to decode JSON payload")
		ctx.JSON(http.StatusBadRequest, dto.JwtDataResponse{
			Status:  false,
			Message: "Invalid body request",
			Data:    map[string]any{},
		})
		return
	}

	log.WithFields(logrus.Fields{
		"step":     "refresh_token",
		"user_id":  req.UserID,
		"jwt_type": req.Type,
	}).Info("Processing generate JWT Refresh Token")

	token, err := c.TokenService.RefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		log.WithField("step", "refresh_token").WithError(err).Error("Failed to generate JWT Refresh Token")
		ctx.JSON(http.StatusInternalServerError, dto.JwtDataResponse{
			Status:  false,
			Message: "Failed to retrieve refresh token, internal error",
			Data:    map[string]any{},
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.JwtDataResponse{
		Status:  true,
		Message: "SUCCESS",
		Data:    token,
	})
}
