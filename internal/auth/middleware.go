package auth

import (
	"context"
	"net/http"
	"partnersale/internal/httpx"
	"strings"
)

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpx.WriteError(w, http.StatusUnauthorized, "Missing token")
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			httpx.WriteError(w, http.StatusUnauthorized, "Invalid token format")
			return
		}
		claims, err := ValidateToken(parts[1])
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
