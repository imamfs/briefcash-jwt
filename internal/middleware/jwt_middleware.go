package middleware

import (
	jsonHelper "briefcash-jwt/internal/helper/jsonhelper"
	service "briefcash-jwt/internal/service"
	"context"
	"net/http"
	"strings"
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

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			MerchantCode string `json:"merchant_code"`
		}

		auth := r.Header.Get("Authorization")

		if auth == "" {
			jsonHelper.WriteJsonError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		if !strings.HasPrefix(auth, "Bearer ") {
			jsonHelper.WriteJsonError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		authToken := strings.TrimPrefix(auth, "Bearer ")
		if authToken == "" {
			jsonHelper.WriteJsonError(w, http.StatusUnauthorized, "empty bearer token")
			return
		}

		if _, err := jsonHelper.ParseJsonBody(r, &request); err != nil {
			jsonHelper.WriteJsonError(w, http.StatusBadRequest, err.Error())
			return
		}

		if request.MerchantCode == "" {
			jsonHelper.WriteJsonError(w, http.StatusBadRequest, "merchant code is empty")
			return
		}

		isValid, err := m.svc.ValidateMerchantCode(r.Context(), request.MerchantCode)
		if err != nil {
			jsonHelper.WriteJsonError(w, http.StatusInternalServerError, "failed to validate merchant code")
			return
		}

		if !isValid {
			jsonHelper.WriteJsonError(w, http.StatusUnauthorized, "invalid or inactive merchant code")
			return
		}

		ctx := context.WithValue(r.Context(), tokenKey, authToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
