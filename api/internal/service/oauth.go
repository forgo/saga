package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

var (
	ErrInvalidAuthCode    = errors.New("invalid authorization code")
	ErrPKCEVerifyFailed   = errors.New("PKCE verification failed")
	ErrProviderError      = errors.New("OAuth provider error")
	ErrInvalidIDToken     = errors.New("invalid ID token")
	ErrEmailNotVerified   = errors.New("email not verified by provider")
	ErrAccountLinkPending = errors.New("account linking required")
)

// OAuthProvider represents supported OAuth providers
type OAuthProvider string

const (
	ProviderGoogle OAuthProvider = "google"
	ProviderApple  OAuthProvider = "apple"
)

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	Google GoogleOAuthConfig
	Apple  AppleOAuthConfig
}

// GoogleOAuthConfig holds Google OAuth settings
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// AppleOAuthConfig holds Apple OAuth settings
type AppleOAuthConfig struct {
	ClientID    string
	TeamID      string
	KeyID       string
	PrivateKey  string
	RedirectURI string
}

// OAuthService handles OAuth authentication
type OAuthService struct {
	config       OAuthConfig
	authService  *AuthService
	identityRepo IdentityRepository
	userRepo     UserRepository
	tokenService *TokenService
	httpClient   *http.Client
}

// OAuthServiceConfig holds configuration for the OAuth service
type OAuthServiceConfig struct {
	Config       OAuthConfig
	AuthService  *AuthService
	IdentityRepo IdentityRepository
	UserRepo     UserRepository
	TokenService *TokenService
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(cfg OAuthServiceConfig) *OAuthService {
	return &OAuthService{
		config:       cfg.Config,
		authService:  cfg.AuthService,
		identityRepo: cfg.IdentityRepo,
		userRepo:     cfg.UserRepo,
		tokenService: cfg.TokenService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OAuthRequest represents an OAuth callback request
type OAuthRequest struct {
	Code         string // Authorization code from OAuth provider
	CodeVerifier string // PKCE code verifier
	State        string // Optional state parameter
}

// OAuthResult represents a successful OAuth authentication
type OAuthResult struct {
	User         *model.User
	TokenPair    *TokenPair
	IsNewUser    bool
	LinkRequired bool
	LinkToken    string // Token to use for account linking
	ExistingUser *model.User
}

// GoogleTokenResponse represents Google's token endpoint response
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// GoogleUserInfo represents user info from Google
type GoogleUserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// AuthenticateGoogle handles Google OAuth callback
func (s *OAuthService) AuthenticateGoogle(ctx context.Context, req OAuthRequest) (*OAuthResult, error) {
	// Exchange code for tokens
	tokenResp, err := s.exchangeGoogleCode(ctx, req.Code, req.CodeVerifier)
	if err != nil {
		return nil, err
	}

	// Parse and validate ID token
	userInfo, err := s.parseGoogleIDToken(tokenResp.IDToken)
	if err != nil {
		return nil, err
	}

	// Require verified email
	if !userInfo.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	return s.handleOAuthUser(ctx, ProviderGoogle, userInfo.ID, userInfo.Email, userInfo.GivenName, userInfo.FamilyName)
}

// AppleTokenResponse represents Apple's token endpoint response
type AppleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// AppleUserInfo represents user info from Apple
type AppleUserInfo struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// AuthenticateApple handles Apple OAuth callback
func (s *OAuthService) AuthenticateApple(ctx context.Context, req OAuthRequest) (*OAuthResult, error) {
	// Exchange code for tokens
	tokenResp, err := s.exchangeAppleCode(ctx, req.Code, req.CodeVerifier)
	if err != nil {
		return nil, err
	}

	// Parse and validate ID token
	userInfo, err := s.parseAppleIDToken(tokenResp.IDToken)
	if err != nil {
		return nil, err
	}

	// Apple always returns verified email
	return s.handleOAuthUser(ctx, ProviderApple, userInfo.ID, userInfo.Email, "", "")
}

// handleOAuthUser processes OAuth user info and returns authentication result
func (s *OAuthService) handleOAuthUser(ctx context.Context, provider OAuthProvider, providerUserID, email, firstname, lastname string) (*OAuthResult, error) {
	// Check if identity already exists
	identity, err := s.identityRepo.GetByProviderID(ctx, string(provider), providerUserID)
	if err != nil {
		return nil, err
	}

	if identity != nil {
		// Existing identity - login
		user, err := s.userRepo.GetByID(ctx, identity.UserID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, ErrUserNotFound
		}

		tokenPair, err := s.tokenService.GenerateTokenPair(ctx, user)
		if err != nil {
			return nil, err
		}

		return &OAuthResult{
			User:      user,
			TokenPair: tokenPair,
			IsNewUser: false,
		}, nil
	}

	// Check if user with this email exists (potential account linking)
	existingUser, err := s.userRepo.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		// Account exists with this email - require linking
		// Generate a short-lived link token
		linkToken, err := s.generateLinkToken(existingUser.ID, string(provider), providerUserID, email)
		if err != nil {
			return nil, err
		}

		return &OAuthResult{
			LinkRequired: true,
			LinkToken:    linkToken,
			ExistingUser: existingUser,
		}, nil
	}

	// Create new user
	user := &model.User{
		Email:         strings.ToLower(email),
		Firstname:     stringPtr(firstname),
		Lastname:      stringPtr(lastname),
		EmailVerified: true, // OAuth providers verify email
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create identity
	identityModel := &model.Identity{
		UserID:                  user.ID,
		Provider:                string(provider),
		ProviderUserID:          providerUserID,
		ProviderEmail:           &email,
		EmailVerifiedByProvider: true,
	}

	if err := s.identityRepo.Create(ctx, identityModel); err != nil {
		return nil, err
	}

	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &OAuthResult{
		User:      user,
		TokenPair: tokenPair,
		IsNewUser: true,
	}, nil
}

// LinkAccount links a new OAuth identity to an existing account
func (s *OAuthService) LinkAccount(ctx context.Context, userID string, provider OAuthProvider, providerUserID, providerEmail string) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Check if identity already linked to another account
	existingIdentity, err := s.identityRepo.GetByProviderID(ctx, string(provider), providerUserID)
	if err != nil {
		return err
	}
	if existingIdentity != nil {
		if existingIdentity.UserID != userID {
			return fmt.Errorf("identity already linked to another account")
		}
		return nil // Already linked to this user
	}

	// Create identity
	identity := &model.Identity{
		UserID:                  userID,
		Provider:                string(provider),
		ProviderUserID:          providerUserID,
		ProviderEmail:           &providerEmail,
		EmailVerifiedByProvider: true,
	}

	return s.identityRepo.Create(ctx, identity)
}

// exchangeGoogleCode exchanges authorization code for tokens
func (s *OAuthService) exchangeGoogleCode(ctx context.Context, code, codeVerifier string) (*GoogleTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.config.Google.ClientID)
	data.Set("client_secret", s.config.Google.ClientSecret)
	data.Set("redirect_uri", s.config.Google.RedirectURI)
	data.Set("grant_type", "authorization_code")
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderError, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrProviderError, string(body))
	}

	var tokenResp GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// parseGoogleIDToken parses and validates Google ID token
func (s *OAuthService) parseGoogleIDToken(idToken string) (*GoogleUserInfo, error) {
	// Split JWT into parts
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidIDToken
	}

	// Decode payload (base64url)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidIDToken
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(payload, &userInfo); err != nil {
		return nil, ErrInvalidIDToken
	}

	// In production, you should verify the token signature using Google's public keys
	// For now, we trust the token from the exchange response

	return &userInfo, nil
}

// exchangeAppleCode exchanges authorization code for tokens
func (s *OAuthService) exchangeAppleCode(ctx context.Context, code, codeVerifier string) (*AppleTokenResponse, error) {
	// Generate client secret JWT for Apple
	clientSecret, err := s.generateAppleClientSecret()
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.config.Apple.ClientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", s.config.Apple.RedirectURI)
	data.Set("grant_type", "authorization_code")
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://appleid.apple.com/auth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderError, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrProviderError, string(body))
	}

	var tokenResp AppleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// parseAppleIDToken parses and validates Apple ID token
func (s *OAuthService) parseAppleIDToken(idToken string) (*AppleUserInfo, error) {
	// Split JWT into parts
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidIDToken
	}

	// Decode payload (base64url)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidIDToken
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified any    `json:"email_verified"` // Can be bool or string
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidIDToken
	}

	// Parse email_verified (Apple can return bool or string)
	var emailVerified bool
	switch v := claims.EmailVerified.(type) {
	case bool:
		emailVerified = v
	case string:
		emailVerified = v == "true"
	}

	return &AppleUserInfo{
		ID:            claims.Sub,
		Email:         claims.Email,
		EmailVerified: emailVerified,
	}, nil
}

// generateAppleClientSecret generates the client secret JWT for Apple
func (s *OAuthService) generateAppleClientSecret() (string, error) {
	// Apple requires a JWT signed with your private key as the client secret
	// This is a simplified implementation - in production, use proper JWT library
	// The JWT should have:
	// - iss: Team ID
	// - iat: Current timestamp
	// - exp: Expiration (max 6 months)
	// - aud: "https://appleid.apple.com"
	// - sub: Client ID (Service ID)
	// Signed with ES256 using the private key

	// For now, return a placeholder - implement full JWT signing in production
	return "", fmt.Errorf("Apple client secret generation not implemented")
}

// generateLinkToken generates a short-lived token for account linking
func (s *OAuthService) generateLinkToken(userID, provider, providerUserID, email string) (string, error) {
	// Create a simple hash-based token
	// In production, use a proper token service with expiration
	data := fmt.Sprintf("%s:%s:%s:%s:%d", userID, provider, providerUserID, email, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return base64.URLEncoding.EncodeToString(hash[:]), nil
}

// ValidatePKCE validates PKCE code verifier against code challenge
func ValidatePKCE(codeVerifier, codeChallenge string) bool {
	// Generate S256 hash of verifier
	hash := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(hash[:])
	return computed == codeChallenge
}
