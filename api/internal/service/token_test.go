package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/pkg/jwt"
)

// ============================================================================
// Mock Repositories
// ============================================================================

type mockTokenRepo struct {
	createRefreshTokenFunc    func(ctx context.Context, token *RefreshToken) error
	getRefreshTokenByHashFunc func(ctx context.Context, hash string) (*RefreshToken, error)
	revokeRefreshTokenFunc    func(ctx context.Context, hash string) error
	revokeAllUserTokensFunc   func(ctx context.Context, userID string) error
	deleteExpiredTokensFunc   func(ctx context.Context) error
}

func (m *mockTokenRepo) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	if m.createRefreshTokenFunc != nil {
		return m.createRefreshTokenFunc(ctx, token)
	}
	return nil
}

func (m *mockTokenRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	if m.getRefreshTokenByHashFunc != nil {
		return m.getRefreshTokenByHashFunc(ctx, hash)
	}
	return nil, nil
}

func (m *mockTokenRepo) RevokeRefreshToken(ctx context.Context, hash string) error {
	if m.revokeRefreshTokenFunc != nil {
		return m.revokeRefreshTokenFunc(ctx, hash)
	}
	return nil
}

func (m *mockTokenRepo) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if m.revokeAllUserTokensFunc != nil {
		return m.revokeAllUserTokensFunc(ctx, userID)
	}
	return nil
}

func (m *mockTokenRepo) DeleteExpiredTokens(ctx context.Context) error {
	if m.deleteExpiredTokensFunc != nil {
		return m.deleteExpiredTokensFunc(ctx)
	}
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTestJWTService(t *testing.T) *jwt.Service {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return jwt.NewTestService(privateKey, "test-issuer", time.Hour)
}

// ============================================================================
// hashToken Tests
// ============================================================================

func TestHashToken_Deterministic(t *testing.T) {
	t.Parallel()

	token := "test-refresh-token"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	if hash1 != hash2 {
		t.Error("hash should be deterministic")
	}
}

func TestHashToken_DifferentInputsDifferentHashes(t *testing.T) {
	t.Parallel()

	hash1 := hashToken("token-a")
	hash2 := hashToken("token-b")

	if hash1 == hash2 {
		t.Error("different tokens should have different hashes")
	}
}

func TestHashToken_CorrectLength(t *testing.T) {
	t.Parallel()

	hash := hashToken("test")
	// SHA-256 produces 32 bytes = 64 hex characters
	if len(hash) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash))
	}
}

// ============================================================================
// stringValue Tests
// ============================================================================

func TestStringValue_NilPointer(t *testing.T) {
	t.Parallel()

	result := stringValue(nil)
	if result != "" {
		t.Errorf("expected empty string for nil, got %q", result)
	}
}

func TestStringValue_NonNilPointer(t *testing.T) {
	t.Parallel()

	s := "hello"
	result := stringValue(&s)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestStringValue_EmptyString(t *testing.T) {
	t.Parallel()

	s := ""
	result := stringValue(&s)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

// ============================================================================
// NewTokenService Tests
// ============================================================================

func TestNewTokenService_DefaultDuration(t *testing.T) {
	t.Parallel()

	svc := NewTokenService(TokenServiceConfig{
		JWTService:      nil,
		TokenRepo:       nil,
		RefreshDuration: 0, // Should default to 30 days
	})

	expected := 30 * 24 * time.Hour
	if svc.refreshDuration != expected {
		t.Errorf("expected default duration %v, got %v", expected, svc.refreshDuration)
	}
}

func TestNewTokenService_CustomDuration(t *testing.T) {
	t.Parallel()

	customDuration := 7 * 24 * time.Hour
	svc := NewTokenService(TokenServiceConfig{
		JWTService:      nil,
		TokenRepo:       nil,
		RefreshDuration: customDuration,
	})

	if svc.refreshDuration != customDuration {
		t.Errorf("expected custom duration %v, got %v", customDuration, svc.refreshDuration)
	}
}

// ============================================================================
// GenerateTokenPair Tests
// ============================================================================

func TestGenerateTokenPair_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	tokenRepo := &mockTokenRepo{
		createRefreshTokenFunc: func(ctx context.Context, token *RefreshToken) error {
			return nil
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  tokenRepo,
	})

	user := &model.User{
		ID:    "user-123",
		Email: "test@example.com",
	}

	pair, err := svc.GenerateTokenPair(ctx, user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pair == nil {
		t.Fatal("expected token pair, got nil")
	}
	if pair.AccessToken == "" {
		t.Error("expected access token")
	}
	if pair.RefreshToken == "" {
		t.Error("expected refresh token")
	}
	if pair.TokenType != "Bearer" {
		t.Errorf("expected token type 'Bearer', got %q", pair.TokenType)
	}
	if pair.ExpiresIn <= 0 {
		t.Error("expected positive expires_in")
	}
}

func TestGenerateTokenPair_StoresHashedToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	var storedToken *RefreshToken
	tokenRepo := &mockTokenRepo{
		createRefreshTokenFunc: func(ctx context.Context, token *RefreshToken) error {
			storedToken = token
			return nil
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  tokenRepo,
	})

	user := &model.User{ID: "user-123", Email: "test@example.com"}
	pair, _ := svc.GenerateTokenPair(ctx, user)

	// The stored token hash should NOT equal the raw refresh token
	rawHash := hashToken(pair.RefreshToken)
	if storedToken.TokenHash != rawHash {
		t.Error("stored hash should match hashed refresh token")
	}
}

func TestGenerateTokenPair_SetsExpiry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	var storedToken *RefreshToken
	tokenRepo := &mockTokenRepo{
		createRefreshTokenFunc: func(ctx context.Context, token *RefreshToken) error {
			storedToken = token
			return nil
		},
	}

	refreshDuration := 7 * 24 * time.Hour
	svc := NewTokenService(TokenServiceConfig{
		JWTService:      jwtSvc,
		TokenRepo:       tokenRepo,
		RefreshDuration: refreshDuration,
	})

	user := &model.User{ID: "user-123", Email: "test@example.com"}
	_, _ = svc.GenerateTokenPair(ctx, user)

	// Expiry should be approximately 7 days from now
	expectedExpiry := time.Now().Add(refreshDuration)
	timeDiff := storedToken.ExpiresAt.Sub(expectedExpiry)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("expiry time not set correctly")
	}
}

func TestGenerateTokenPair_RepoError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	tokenRepo := &mockTokenRepo{
		createRefreshTokenFunc: func(ctx context.Context, token *RefreshToken) error {
			return errors.New("database error")
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  tokenRepo,
	})

	user := &model.User{ID: "user-123", Email: "test@example.com"}
	_, err := svc.GenerateTokenPair(ctx, user)

	if err == nil || err.Error() != "database error" {
		t.Errorf("expected database error, got %v", err)
	}
}

// ============================================================================
// RefreshTokens Tests
// ============================================================================

func TestRefreshTokens_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	refreshToken := "valid-refresh-token"
	tokenHash := hashToken(refreshToken)

	tokenRepo := &mockTokenRepo{
		getRefreshTokenByHashFunc: func(ctx context.Context, hash string) (*RefreshToken, error) {
			if hash == tokenHash {
				return &RefreshToken{
					UserID:    "user-123",
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(24 * time.Hour),
					Revoked:   false,
				}, nil
			}
			return nil, nil
		},
		revokeRefreshTokenFunc: func(ctx context.Context, hash string) error {
			return nil
		},
		createRefreshTokenFunc: func(ctx context.Context, token *RefreshToken) error {
			return nil
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  tokenRepo,
	})

	user := &model.User{ID: "user-123", Email: "test@example.com"}
	pair, err := svc.RefreshTokens(ctx, refreshToken, user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pair == nil {
		t.Fatal("expected token pair")
	}
}

func TestRefreshTokens_InvalidToken_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	tokenRepo := &mockTokenRepo{
		getRefreshTokenByHashFunc: func(ctx context.Context, hash string) (*RefreshToken, error) {
			return nil, nil // Token not found
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  tokenRepo,
	})

	user := &model.User{ID: "user-123", Email: "test@example.com"}
	_, err := svc.RefreshTokens(ctx, "invalid-token", user)

	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshTokens_RevokedToken_ReturnsErrorAndRevokesAll(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	refreshToken := "revoked-token"
	tokenHash := hashToken(refreshToken)
	revokeAllCalled := false

	tokenRepo := &mockTokenRepo{
		getRefreshTokenByHashFunc: func(ctx context.Context, hash string) (*RefreshToken, error) {
			return &RefreshToken{
				UserID:    "user-123",
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				Revoked:   true, // Already revoked - reuse detected!
			}, nil
		},
		revokeAllUserTokensFunc: func(ctx context.Context, userID string) error {
			revokeAllCalled = true
			return nil
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  tokenRepo,
	})

	user := &model.User{ID: "user-123", Email: "test@example.com"}
	_, err := svc.RefreshTokens(ctx, refreshToken, user)

	if !errors.Is(err, ErrRefreshTokenRevoked) {
		t.Errorf("expected ErrRefreshTokenRevoked, got %v", err)
	}
	if !revokeAllCalled {
		t.Error("expected all tokens to be revoked on reuse detection")
	}
}

func TestRefreshTokens_ExpiredToken_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	refreshToken := "expired-token"
	tokenHash := hashToken(refreshToken)

	tokenRepo := &mockTokenRepo{
		getRefreshTokenByHashFunc: func(ctx context.Context, hash string) (*RefreshToken, error) {
			return &RefreshToken{
				UserID:    "user-123",
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired
				Revoked:   false,
			}, nil
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  tokenRepo,
	})

	user := &model.User{ID: "user-123", Email: "test@example.com"}
	_, err := svc.RefreshTokens(ctx, refreshToken, user)

	if !errors.Is(err, ErrRefreshTokenExpired) {
		t.Errorf("expected ErrRefreshTokenExpired, got %v", err)
	}
}

func TestRefreshTokens_RevokesOldToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	jwtSvc := createTestJWTService(t)
	refreshToken := "valid-token"
	tokenHash := hashToken(refreshToken)
	revokedHash := ""

	tokenRepo := &mockTokenRepo{
		getRefreshTokenByHashFunc: func(ctx context.Context, hash string) (*RefreshToken, error) {
			return &RefreshToken{
				UserID:    "user-123",
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(24 * time.Hour),
				Revoked:   false,
			}, nil
		},
		revokeRefreshTokenFunc: func(ctx context.Context, hash string) error {
			revokedHash = hash
			return nil
		},
		createRefreshTokenFunc: func(ctx context.Context, token *RefreshToken) error {
			return nil
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  tokenRepo,
	})

	user := &model.User{ID: "user-123", Email: "test@example.com"}
	_, _ = svc.RefreshTokens(ctx, refreshToken, user)

	if revokedHash != tokenHash {
		t.Error("expected old token to be revoked")
	}
}

// ============================================================================
// ValidateAccessToken Tests
// ============================================================================

func TestValidateAccessToken_ValidToken(t *testing.T) {
	t.Parallel()

	jwtSvc := createTestJWTService(t)
	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  &mockTokenRepo{},
	})

	// Generate a valid token
	claims := jwt.Claims{
		Subject: "user-123",
		UserID:  "user-123",
		Email:   "test@example.com",
	}
	token, _ := jwtSvc.Sign(claims)

	validatedClaims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if validatedClaims.UserID != "user-123" {
		t.Errorf("expected user ID 'user-123', got %q", validatedClaims.UserID)
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	t.Parallel()

	jwtSvc := createTestJWTService(t)
	svc := NewTokenService(TokenServiceConfig{
		JWTService: jwtSvc,
		TokenRepo:  &mockTokenRepo{},
	})

	_, err := svc.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

// ============================================================================
// RevokeAllUserTokens Tests
// ============================================================================

func TestRevokeAllUserTokens_CallsRepo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	revokedUserID := ""
	tokenRepo := &mockTokenRepo{
		revokeAllUserTokensFunc: func(ctx context.Context, userID string) error {
			revokedUserID = userID
			return nil
		},
	}

	svc := NewTokenService(TokenServiceConfig{
		TokenRepo: tokenRepo,
	})

	err := svc.RevokeAllUserTokens(ctx, "user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revokedUserID != "user-123" {
		t.Error("expected repo to be called with correct user ID")
	}
}

// ============================================================================
// generateRefreshToken Tests
// ============================================================================

func TestGenerateRefreshToken_UniqueTokens(t *testing.T) {
	t.Parallel()

	svc := &TokenService{}

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := svc.generateRefreshToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tokens[token] {
			t.Fatal("generated duplicate token")
		}
		tokens[token] = true
	}
}

func TestGenerateRefreshToken_CorrectLength(t *testing.T) {
	t.Parallel()

	svc := &TokenService{}

	token, err := svc.generateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 32 bytes = 64 hex characters
	if len(token) != 64 {
		t.Errorf("expected token length 64, got %d", len(token))
	}
}
