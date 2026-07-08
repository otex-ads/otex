package mw

import (
	"context"
	"net/http"
	"strings"

	"adnet/internal/auth"

	"github.com/google/uuid"
)

type contextKey string

const (
	ContextKeyAccountID contextKey = "account_id"
	ContextKeyEmail      contextKey = "email"
	ContextKeyType       contextKey = "account_type"
)

type AuthMiddleware struct {
	authService *auth.AuthService
}

func NewAuthMiddleware(authService *auth.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := m.authService.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = contextWithAccountID(ctx, claims.AccountID)
		ctx = contextWithEmail(ctx, claims.Email)
		ctx = contextWithType(ctx, claims.Type)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireAccountType(accountType auth.AccountType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accountTypeFromCtx, ok := TypeFromContext(r.Context())
			if !ok {
				http.Error(w, "Account type not found in context", http.StatusUnauthorized)
				return
			}

			if !m.authService.HasPermission(accountTypeFromCtx, accountType) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func contextWithAccountID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ContextKeyAccountID, id)
}

func contextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, ContextKeyEmail, email)
}

func contextWithType(ctx context.Context, accountType auth.AccountType) context.Context {
	return context.WithValue(ctx, ContextKeyType, accountType)
}

func AccountIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ContextKeyAccountID).(uuid.UUID)
	return id, ok
}

func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(ContextKeyEmail).(string)
	return email, ok
}

func TypeFromContext(ctx context.Context) (auth.AccountType, bool) {
	accountType, ok := ctx.Value(ContextKeyType).(auth.AccountType)
	return accountType, ok
}

func AccountTypeFromContext(ctx context.Context) (string, bool) {
	accountType, ok := ctx.Value(ContextKeyType).(auth.AccountType)
	return string(accountType), ok
}
