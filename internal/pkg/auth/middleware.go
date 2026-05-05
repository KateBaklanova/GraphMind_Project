package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/kate/knowledge-graph/internal/pkg/models"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// Проверяет JWT
func AuthMiddleware(jwtManager *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization format, use 'Bearer <token>'", http.StatusUnauthorized)
				return
			}

			// Верификация
			claims, err := jwtManager.VerifyToken(parts[1])
			if err != nil {
				if err == ErrExpiredToken {
					http.Error(w, "token expired", http.StatusUnauthorized)
				} else {
					http.Error(w, "invalid token", http.StatusUnauthorized)
				}
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Получает пользователя из контекста
func GetUserFromContext(ctx context.Context) (*models.Claims, error) {
	user, ok := ctx.Value(UserContextKey).(*models.Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return user, nil
}

// Проверяет авторизацию
func RequireAuth(w http.ResponseWriter, r *http.Request) (*models.Claims, bool) {
	user, err := GetUserFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	return user, true
}
