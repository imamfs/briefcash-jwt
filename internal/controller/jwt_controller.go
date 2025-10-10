package controller

import (
	dto "briefcash-jwt/internal/dto"
	jsonHelper "briefcash-jwt/internal/helper/jsonhelper"
	middleware "briefcash-jwt/internal/middleware"
	service "briefcash-jwt/internal/service"
	"encoding/json"
	"net/http"
)

type TokenController struct {
	TokenService service.TokenService
}

func NewTokenController(s service.TokenService) *TokenController {
	return &TokenController{s}
}

func (c *TokenController) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var tokenRequest dto.JwtRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&tokenRequest); err != nil {
		jsonHelper.WriteJsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := c.TokenService.GenerateToken(r.Context(), tokenRequest)
	if err != nil {
		jsonHelper.WriteJsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, dto.JwtDataResponse{
		Status:  true,
		Message: "SUCCESS",
		Data:    token,
	})
}

func (c *TokenController) ValidateToken(w http.ResponseWriter, r *http.Request) {
	tokenString, ok := middleware.GetTokenFromContext(r.Context())
	if !ok {
		jsonHelper.WriteJsonError(w, http.StatusUnauthorized, "token not found in context")
		return
	}

	token, err := c.TokenService.ValidateToken(r.Context(), tokenString)

	if err != nil {
		jsonHelper.WriteJsonError(w, http.StatusUnauthorized, err.Error())
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, dto.JwtDataResponse{
		Status:  true,
		Message: "SUCCESS",
		Data: map[string]any{
			"valid":  true,
			"claims": token.Claims,
		},
	})
}

func (c *TokenController) Logout(w http.ResponseWriter, r *http.Request) {
	tokenString, ok := middleware.GetTokenFromContext(r.Context())
	if !ok {
		jsonHelper.WriteJsonError(w, http.StatusUnauthorized, "token not found in context")
		return
	}

	err := c.TokenService.BlacklistToken(r.Context(), tokenString)

	if err != nil {
		jsonHelper.WriteJsonError(w, http.StatusInternalServerError, "failed to logout token")
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, dto.JwtDataResponse{
		Status:  true,
		Message: "SUCCESS",
		Data:    "logout success",
	})
}

func (c *TokenController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.JwtRefreshToken
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		jsonHelper.WriteJsonError(w, http.StatusBadRequest, "invalid body request")
		return
	}

	token, err := c.TokenService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		jsonHelper.WriteJsonError(w, http.StatusUnauthorized, err.Error())
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, dto.JwtDataResponse{
		Status:  true,
		Message: "SUCCESS",
		Data:    token,
	})
}
