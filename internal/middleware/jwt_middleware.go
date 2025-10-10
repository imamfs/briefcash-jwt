package middleware

import (
	jsonHelper "briefcash-jwt/internal/helper/jsonhelper"
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

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		ctx := context.WithValue(r.Context(), tokenKey, authToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
