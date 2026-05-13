package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/service"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	roleKey   contextKey = "role"
)

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := resolveToken(r)
			if tokenString == "" {
				writeUnauthorized(w)
				return
			}

			claims, err := service.ParseJWT(tokenString, secret)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, roleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey).(string)
	return role, ok
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetRole(r.Context())
		if !ok || role != models.RoleAdmin {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func resolveToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	cookie, err := r.Cookie("access_token")
	if err == nil {
		return cookie.Value
	}
	return ""
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
