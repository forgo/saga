package service

import (
	"context"
	"strings"

	"github.com/forgo/saga/api/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// Error definitions moved to errors.go

const (
	// bcrypt cost factor (10-14 recommended for production)
	bcryptCost = 12

	// Password constraints
	minPasswordLength = 8
	maxPasswordLength = 128
)

// UserRepository defines the interface for user storage
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdatePassword(ctx context.Context, userID, hash string) error
	Delete(ctx context.Context, id string) error
	SetEmailVerified(ctx context.Context, userID string, verified bool) error
}

// IdentityRepository defines the interface for identity storage
type IdentityRepository interface {
	Create(ctx context.Context, identity *model.Identity) error
	GetByProviderID(ctx context.Context, provider, providerUserID string) (*model.Identity, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Identity, error)
	GetByProviderEmail(ctx context.Context, provider, email string) (*model.Identity, error)
	Delete(ctx context.Context, id string) error
	CountByUserID(ctx context.Context, userID string) (int, error)
}

// PasskeyRepository defines the interface for passkey storage
type PasskeyRepository interface {
	Create(ctx context.Context, passkey *model.Passkey) error
	GetByCredentialID(ctx context.Context, credentialID string) (*model.Passkey, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.Passkey, error)
	GetByID(ctx context.Context, id string) (*model.Passkey, error)
	UpdateSignCount(ctx context.Context, credentialID string, signCount uint32) error
	Delete(ctx context.Context, id string) error
	CountByUserID(ctx context.Context, userID string) (int, error)
}

// AuthService handles authentication operations
type AuthService struct {
	userRepo     UserRepository
	identityRepo IdentityRepository
	passkeyRepo  PasskeyRepository
	tokenService *TokenService
}

// AuthServiceConfig holds configuration for the auth service
type AuthServiceConfig struct {
	UserRepo     UserRepository
	IdentityRepo IdentityRepository
	PasskeyRepo  PasskeyRepository
	TokenService *TokenService
}

// NewAuthService creates a new auth service
func NewAuthService(cfg AuthServiceConfig) *AuthService {
	return &AuthService{
		userRepo:     cfg.UserRepo,
		identityRepo: cfg.IdentityRepo,
		passkeyRepo:  cfg.PasskeyRepo,
		tokenService: cfg.TokenService,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string
	Password  string
	Firstname string
	Lastname  string
}

// RegisterResult represents a successful registration
type RegisterResult struct {
	User      *model.User
	TokenPair *TokenPair
}

// Register creates a new user account with email/password
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	// Validate email
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	// Validate password
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hash, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		Email:         email,
		Hash:          &hash,
		Firstname:     stringPtr(strings.TrimSpace(req.Firstname)),
		Lastname:      stringPtr(strings.TrimSpace(req.Lastname)),
		EmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &RegisterResult{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResult represents a successful login
type LoginResult struct {
	User      *model.User
	TokenPair *TokenPair
}

// Login authenticates a user with email/password
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user has a password (might be OAuth-only)
	if user.Hash == nil || *user.Hash == "" {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !checkPassword(req.Password, *user.Hash) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserWithIdentities retrieves a user with all linked identities and passkeys
func (s *AuthService) GetUserWithIdentities(ctx context.Context, userID string) (*model.UserWithIdentities, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	identities, err := s.identityRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	passkeys, err := s.passkeyRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &model.UserWithIdentities{
		User:       user,
		Identities: identities,
		Passkeys:   passkeys,
	}, nil
}

// RefreshTokens validates a refresh token and issues new tokens
func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Get stored token to find user ID
	tokenHash := hashToken(refreshToken)
	storedToken, err := s.tokenService.tokenRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	if storedToken == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Refresh tokens (handles validation and rotation)
	return s.tokenService.RefreshTokens(ctx, refreshToken, user)
}

// Logout revokes the user's refresh tokens
func (s *AuthService) Logout(ctx context.Context, userID string) error {
	return s.tokenService.RevokeAllUserTokens(ctx, userID)
}

// ValidateAccessToken validates an access token and returns the claims
func (s *AuthService) ValidateAccessToken(token string) (*model.TokenClaims, error) {
	claims, err := s.tokenService.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	return &model.TokenClaims{
		UserID:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
	}, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify old password if user has one
	if user.Hash != nil && *user.Hash != "" {
		if !checkPassword(oldPassword, *user.Hash) {
			return ErrInvalidCredentials
		}
	}

	// Validate new password
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password and revoke all tokens (force re-login)
	if err := s.userRepo.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}

	return s.tokenService.RevokeAllUserTokens(ctx, userID)
}

// Helper functions

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func validatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > maxPasswordLength {
		return ErrPasswordTooLong
	}
	return nil
}

func isValidEmail(email string) bool {
	// Basic email validation
	if email == "" {
		return false
	}
	if len(email) > 254 {
		return false
	}
	atIndex := strings.Index(email, "@")
	if atIndex < 1 {
		return false
	}
	dotIndex := strings.LastIndex(email, ".")
	if dotIndex < atIndex+2 {
		return false
	}
	if dotIndex >= len(email)-1 {
		return false
	}
	return true
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
