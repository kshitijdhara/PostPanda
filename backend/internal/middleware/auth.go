package middleware

import (
	"context"
	"net/http"

	"github.com/kshitijdhara/blog/internal/services"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"authentication required"}}`, http.StatusUnauthorized)
				return
			}

			userID, err := authService.ValidateToken(cookie.Value)
			if err != nil {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"invalid token"}}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth extracts user ID from cookie if present but doesn't require it.
func OptionalAuth(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err == nil {
				userID, err := authService.ValidateToken(cookie.Value)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserIDKey, userID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}
