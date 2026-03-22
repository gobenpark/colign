package middleware

//go:generate mockgen -destination mock_token_validator_test.go -package middleware github.com/gobenpark/colign/internal/middleware TokenValidator

import (
	"context"
	"net/http"
	"strings"

	"github.com/gobenpark/colign/internal/auth"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "user_id"
	ContextKeyEmail  contextKey = "email"
	ContextKeyOrgID  contextKey = "org_id"
)

// TokenValidator validates a JWT token string and returns claims.
// *auth.JWTManager satisfies this interface.
type TokenValidator interface {
	ValidateAccessToken(tokenStr string) (*auth.Claims, error)
}

func JWTAuth(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"authorization header required"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := validator.ValidateAccessToken(parts[1])
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextKeyOrgID, claims.OrgID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) int64 {
	v, _ := ctx.Value(ContextKeyUserID).(int64)
	return v
}

func GetOrgID(ctx context.Context) int64 {
	v, _ := ctx.Value(ContextKeyOrgID).(int64)
	return v
}
