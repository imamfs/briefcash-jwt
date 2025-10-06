package controller

import (
	dto "briefcash-jwt/internal/dto"
	service "briefcash-jwt/internal/service"
	"context"
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

	if err := json.NewDecoder(r.Body).Decode(&tokenRequest); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := c.TokenService.GenerateToken(context.Background(), tokenRequest)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(token)
}

func (c *TokenController) ValidateToken(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if len(auth) < 8 || auth[:7] != "Bearer " {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
	}

	tokenString := auth[7:]

	token, err := c.TokenService.ValidateToken(context.Background(), tokenString)

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"valid": true, "claims": token.Claims})
}

func (c *TokenController) Logout(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if len(auth) < 8 || auth[:7] != "Bearer " {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
	}
	tokenString := auth[7:]

	err := c.TokenService.BlacklistToken(context.Background(), tokenString)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "log out success"})
}

func (c *TokenController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, err := c.TokenService.RefreshToken(context.Background(), req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
	json.NewEncoder(w).Encode(token)
}
