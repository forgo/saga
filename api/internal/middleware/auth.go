package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/pkg/jwt"
)

// AuthService defines the interface for token validation
type AuthService interface {
	ValidateAccessToken(token string) (*jwt.Claims, error)
}

// Auth returns a middleware that validates JWT tokens
func Auth(authService AuthService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				model.NewUnauthorizedError("missing authorization header").WriteJSON(w)
				return
			}

			// Check Bearer prefix
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				model.NewUnauthorizedError("invalid authorization header format").WriteJSON(w)
				return
			}

			token := parts[1]

			// Validate token
			claims, err := authService.ValidateAccessToken(token)
			if err != nil {
				switch err {
				case jwt.ErrTokenExpired:
					model.NewUnauthorizedError("token expired").WriteJSON(w)
				case jwt.ErrInvalidSignature:
					model.NewUnauthorizedError("invalid token signature").WriteJSON(w)
				default:
					model.NewUnauthorizedError("invalid token").WriteJSON(w)
				}
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsKey is the context key for JWT claims
const ClaimsKey contextKey = "claims"

// UserEmailKey is the context key for user email
const UserEmailKey contextKey = "userEmail"

// GetUserID extracts the user ID from context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserEmail extracts the user email from context
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(UserEmailKey).(string); ok {
		return email
	}
	return ""
}

// GetClaims extracts the JWT claims from context
func GetClaims(ctx context.Context) *jwt.Claims {
	if claims, ok := ctx.Value(ClaimsKey).(*jwt.Claims); ok {
		return claims
	}
	return nil
}

// OptionalAuth is like Auth but doesn't require authentication
// It will set user info in context if token is present and valid
func OptionalAuth(authService AuthService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				next.ServeHTTP(w, r)
				return
			}

			token := parts[1]
			claims, err := authService.ValidateAccessToken(token)
			if err != nil {
				// Invalid token, but optional so continue without auth
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
