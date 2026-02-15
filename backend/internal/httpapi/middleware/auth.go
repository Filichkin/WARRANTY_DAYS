package middleware

import (
	"context"
	"net/http"
	"strings"

	"warranty_days/internal/auth"
)

type userContextKey struct{}

type UserContext struct {
	UserID int64
	Email  string
}

type AccessTokenValidator interface {
	ParseAndValidate(tokenStr string, expectedType string) (*auth.Claims, error)
}

func Auth(jwtSvc AccessTokenValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if jwtSvc == nil {
			http.Error(w, "auth is not configured", http.StatusInternalServerError)
			return
		}

		token, ok := extractBearerToken(r.Header.Get("Authorization"))
		if !ok {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := jwtSvc.ParseAndValidate(token, auth.TokenTypeAccess)
		if err != nil {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, UserContext{
			UserID: claims.UserID,
			Email:  claims.Email,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) (UserContext, bool) {
	v := ctx.Value(userContextKey{})
	user, ok := v.(UserContext)
	return user, ok
}

func extractBearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", false
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}

	return token, true
}
