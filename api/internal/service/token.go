package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/pkg/jwt"
)

// Error definitions moved to errors.go

// RefreshToken represents a stored refresh token
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}

// TokenRepository defines the interface for refresh token storage
type TokenRepository interface {
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, hash string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
	DeleteExpiredTokens(ctx context.Context) error
}

// TokenService handles JWT and refresh token operations
type TokenService struct {
	jwtService      *jwt.Service
	tokenRepo       TokenRepository
	refreshDuration time.Duration
}

// TokenServiceConfig holds configuration for the token service
type TokenServiceConfig struct {
	JWTService      *jwt.Service
	TokenRepo       TokenRepository
	RefreshDuration time.Duration // Default: 30 days
}

// NewTokenService creates a new token service
func NewTokenService(cfg TokenServiceConfig) *TokenService {
	if cfg.RefreshDuration == 0 {
		cfg.RefreshDuration = 30 * 24 * time.Hour // 30 days
	}

	return &TokenService{
		jwtService:      cfg.JWTService,
		tokenRepo:       cfg.TokenRepo,
		refreshDuration: cfg.RefreshDuration,
	}
}

// TokenPair represents an access token and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// GenerateTokenPair creates a new access token and refresh token for a user
func (s *TokenService) GenerateTokenPair(ctx context.Context, user *model.User) (*TokenPair, error) {
	// Generate access token (JWT)
	claims := jwt.Claims{
		Subject:  user.ID,
		UserID:   user.ID,
		Email:    user.Email,
		Username: stringValue(user.Username),
	}

	accessToken, err := s.jwtService.Sign(claims)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (opaque)
	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Hash refresh token for storage
	tokenHash := hashToken(refreshToken)

	// Store refresh token
	storedToken := &RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.refreshDuration),
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	if err := s.tokenRepo.CreateRefreshToken(ctx, storedToken); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtService.GetExpiration().Seconds()),
	}, nil
}

// RefreshTokens validates a refresh token and issues new tokens
// Implements single-use rotation: old token is revoked, new token is issued
func (s *TokenService) RefreshTokens(ctx context.Context, refreshToken string, user *model.User) (*TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	// Get stored token
	storedToken, err := s.tokenRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if storedToken == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check if revoked
	if storedToken.Revoked {
		// Token reuse detected - revoke all tokens for this user (security measure)
		_ = s.tokenRepo.RevokeAllUserTokens(ctx, storedToken.UserID)
		return nil, ErrRefreshTokenRevoked
	}

	// Check if expired
	if time.Now().After(storedToken.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	// Revoke old token (single-use)
	if err := s.tokenRepo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, err
	}

	// Generate new token pair
	return s.GenerateTokenPair(ctx, user)
}

// ValidateAccessToken validates an access token and returns the claims
func (s *TokenService) ValidateAccessToken(token string) (*jwt.Claims, error) {
	return s.jwtService.Validate(token)
}

// RevokeAllUserTokens revokes all refresh tokens for a user (logout from all devices)
func (s *TokenService) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.tokenRepo.RevokeAllUserTokens(ctx, userID)
}

// generateRefreshToken creates a cryptographically secure random token
func (s *TokenService) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA-256 hash of the token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// stringValue safely dereferences a string pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
