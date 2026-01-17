package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// Error definitions moved to errors.go

const (
	// Challenge expiration time
	challengeExpiration = 60 * time.Second

	// Maximum passkeys per user
	maxPasskeysPerUser = 10

	// Challenge length in bytes
	challengeLength = 32
)

// PasskeyConfig holds passkey/WebAuthn configuration
type PasskeyConfig struct {
	RPID          string // Relying Party ID (e.g., "saga.forgo.software")
	RPName        string // Relying Party display name
	RPOrigins     []string
	Timeout       time.Duration
	RequireUV     bool // Require user verification
	AttestationType string
}

// PasskeyService handles WebAuthn/Passkey operations
type PasskeyService struct {
	config       PasskeyConfig
	passkeyRepo  PasskeyRepository
	userRepo     UserRepository
	tokenService *TokenService
	challenges   *challengeStore
}

// PasskeyServiceConfig holds configuration for the passkey service
type PasskeyServiceConfig struct {
	Config       PasskeyConfig
	PasskeyRepo  PasskeyRepository
	UserRepo     UserRepository
	TokenService *TokenService
}

// NewPasskeyService creates a new passkey service
func NewPasskeyService(cfg PasskeyServiceConfig) *PasskeyService {
	return &PasskeyService{
		config:       cfg.Config,
		passkeyRepo:  cfg.PasskeyRepo,
		userRepo:     cfg.UserRepo,
		tokenService: cfg.TokenService,
		challenges:   newChallengeStore(),
	}
}

// challengeStore stores pending challenges with expiration
type challengeStore struct {
	challenges map[string]*challenge
}

type challenge struct {
	Value     []byte
	UserID    string
	ExpiresAt time.Time
	IsLogin   bool // true for login, false for registration
}

func newChallengeStore() *challengeStore {
	return &challengeStore{
		challenges: make(map[string]*challenge),
	}
}

func (s *challengeStore) Create(userID string, isLogin bool) ([]byte, error) {
	value := make([]byte, challengeLength)
	if _, err := rand.Read(value); err != nil {
		return nil, err
	}

	key := base64.RawURLEncoding.EncodeToString(value)
	s.challenges[key] = &challenge{
		Value:     value,
		UserID:    userID,
		ExpiresAt: time.Now().Add(challengeExpiration),
		IsLogin:   isLogin,
	}

	// Cleanup expired challenges
	s.cleanup()

	return value, nil
}

func (s *challengeStore) Verify(challengeB64 string, userID string, isLogin bool) bool {
	c, exists := s.challenges[challengeB64]
	if !exists {
		return false
	}

	defer delete(s.challenges, challengeB64)

	if time.Now().After(c.ExpiresAt) {
		return false
	}

	if c.UserID != userID || c.IsLogin != isLogin {
		return false
	}

	return true
}

func (s *challengeStore) cleanup() {
	now := time.Now()
	for key, c := range s.challenges {
		if now.After(c.ExpiresAt) {
			delete(s.challenges, key)
		}
	}
}

// RegistrationStartRequest represents input for starting passkey registration
type RegistrationStartRequest struct {
	UserID string
}

// RegistrationStartResponse is returned to start passkey registration
type RegistrationStartResponse struct {
	Challenge        string                  `json:"challenge"`
	RP               RelyingPartyInfo        `json:"rp"`
	User             WebAuthnUserInfo        `json:"user"`
	PubKeyCredParams []PubKeyCredParam       `json:"pubKeyCredParams"`
	Timeout          int                     `json:"timeout"`
	Attestation      string                  `json:"attestation"`
	AuthenticatorSelection *AuthenticatorSelection `json:"authenticatorSelection,omitempty"`
	ExcludeCredentials []CredentialDescriptor `json:"excludeCredentials,omitempty"`
}

type RelyingPartyInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WebAuthnUserInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type PubKeyCredParam struct {
	Type string `json:"type"`
	Alg  int    `json:"alg"`
}

type AuthenticatorSelection struct {
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`
	ResidentKey            string `json:"residentKey,omitempty"`
	RequireResidentKey     bool   `json:"requireResidentKey,omitempty"`
	UserVerification       string `json:"userVerification,omitempty"`
}

type CredentialDescriptor struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Transports []string `json:"transports,omitempty"`
}

// StartRegistration initiates passkey registration
func (s *PasskeyService) StartRegistration(ctx context.Context, req RegistrationStartRequest) (*RegistrationStartResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check passkey limit
	count, err := s.passkeyRepo.CountByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if count >= maxPasskeysPerUser {
		return nil, ErrPasskeyLimitReached
	}

	// Generate challenge
	challengeBytes, err := s.challenges.Create(req.UserID, false)
	if err != nil {
		return nil, err
	}

	// Get existing credentials to exclude
	existingPasskeys, err := s.passkeyRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	excludeCredentials := make([]CredentialDescriptor, 0, len(existingPasskeys))
	for _, p := range existingPasskeys {
		excludeCredentials = append(excludeCredentials, CredentialDescriptor{
			Type: "public-key",
			ID:   p.CredentialID,
		})
	}

	displayName := user.Email
	if user.Firstname != nil && *user.Firstname != "" {
		displayName = *user.Firstname
		if user.Lastname != nil && *user.Lastname != "" {
			displayName += " " + *user.Lastname
		}
	}

	return &RegistrationStartResponse{
		Challenge: base64.RawURLEncoding.EncodeToString(challengeBytes),
		RP: RelyingPartyInfo{
			ID:   s.config.RPID,
			Name: s.config.RPName,
		},
		User: WebAuthnUserInfo{
			ID:          base64.RawURLEncoding.EncodeToString([]byte(user.ID)),
			Name:        user.Email,
			DisplayName: displayName,
		},
		PubKeyCredParams: []PubKeyCredParam{
			{Type: "public-key", Alg: -7},   // ES256 (ECDSA with P-256)
			{Type: "public-key", Alg: -257}, // RS256 (RSASSA-PKCS1-v1_5)
		},
		Timeout:     int(s.config.Timeout.Milliseconds()),
		Attestation: s.config.AttestationType,
		AuthenticatorSelection: &AuthenticatorSelection{
			ResidentKey:      "preferred",
			UserVerification: "preferred",
		},
		ExcludeCredentials: excludeCredentials,
	}, nil
}

// RegistrationFinishRequest represents input for completing passkey registration
type RegistrationFinishRequest struct {
	UserID     string
	Name       string // User-friendly name for this passkey
	Credential *CredentialResponse
}

type CredentialResponse struct {
	ID       string            `json:"id"`
	RawID    string            `json:"rawId"`
	Type     string            `json:"type"`
	Response AttestationResponse `json:"response"`
}

type AttestationResponse struct {
	ClientDataJSON    string `json:"clientDataJSON"`
	AttestationObject string `json:"attestationObject"`
}

// RegistrationFinishResult represents successful passkey registration
type RegistrationFinishResult struct {
	Passkey *model.Passkey
}

// FinishRegistration completes passkey registration
func (s *PasskeyService) FinishRegistration(ctx context.Context, req RegistrationFinishRequest) (*RegistrationFinishResult, error) {
	// In a full implementation, you would:
	// 1. Decode and parse clientDataJSON
	// 2. Verify challenge matches
	// 3. Verify origin matches configured origins
	// 4. Parse attestationObject (CBOR encoded)
	// 5. Extract authenticator data and public key
	// 6. Verify attestation signature
	// 7. Store credential

	// For now, we'll implement a simplified version
	// In production, use go-webauthn/webauthn library

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Create passkey record
	passkey := &model.Passkey{
		UserID:       req.UserID,
		CredentialID: req.Credential.ID,
		PublicKey:    []byte{}, // Would be extracted from attestationObject
		SignCount:    0,
		Name:         req.Name,
	}

	if err := s.passkeyRepo.Create(ctx, passkey); err != nil {
		return nil, err
	}

	return &RegistrationFinishResult{
		Passkey: passkey,
	}, nil
}

// LoginStartRequest represents input for starting passkey login
type LoginStartRequest struct {
	Email string // Optional email hint
}

// LoginStartResponse is returned to start passkey login
type LoginStartResponse struct {
	Challenge          string                 `json:"challenge"`
	Timeout            int                    `json:"timeout"`
	RPID               string                 `json:"rpId"`
	AllowCredentials   []CredentialDescriptor `json:"allowCredentials,omitempty"`
	UserVerification   string                 `json:"userVerification,omitempty"`
}

// StartLogin initiates passkey login
func (s *PasskeyService) StartLogin(ctx context.Context, req LoginStartRequest) (*LoginStartResponse, error) {
	var allowCredentials []CredentialDescriptor
	var userID string

	if req.Email != "" {
		// If email provided, get user's passkeys
		user, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if user == nil {
			// Don't reveal whether user exists
			// Return empty credentials - authenticator will show available passkeys
			userID = req.Email // Use email as placeholder for challenge
		} else {
			userID = user.ID
			passkeys, err := s.passkeyRepo.GetByUserID(ctx, user.ID)
			if err != nil {
				return nil, err
			}

			allowCredentials = make([]CredentialDescriptor, 0, len(passkeys))
			for _, p := range passkeys {
				allowCredentials = append(allowCredentials, CredentialDescriptor{
					Type: "public-key",
					ID:   p.CredentialID,
				})
			}
		}
	} else {
		// Discoverable credential flow - no allowCredentials
		userID = "discoverable"
	}

	// Generate challenge
	challengeBytes, err := s.challenges.Create(userID, true)
	if err != nil {
		return nil, err
	}

	return &LoginStartResponse{
		Challenge:        base64.RawURLEncoding.EncodeToString(challengeBytes),
		Timeout:          int(s.config.Timeout.Milliseconds()),
		RPID:             s.config.RPID,
		AllowCredentials: allowCredentials,
		UserVerification: "preferred",
	}, nil
}

// LoginFinishRequest represents input for completing passkey login
type LoginFinishRequest struct {
	Credential *AssertionResponse
}

type AssertionResponse struct {
	ID       string           `json:"id"`
	RawID    string           `json:"rawId"`
	Type     string           `json:"type"`
	Response AuthenticatorAssertionResponse `json:"response"`
}

type AuthenticatorAssertionResponse struct {
	ClientDataJSON    string `json:"clientDataJSON"`
	AuthenticatorData string `json:"authenticatorData"`
	Signature         string `json:"signature"`
	UserHandle        string `json:"userHandle,omitempty"`
}

// LoginFinishResult represents successful passkey login
type LoginFinishResult struct {
	User      *model.User
	TokenPair *TokenPair
}

// FinishLogin completes passkey login
func (s *PasskeyService) FinishLogin(ctx context.Context, req LoginFinishRequest) (*LoginFinishResult, error) {
	// Find passkey by credential ID
	passkey, err := s.passkeyRepo.GetByCredentialID(ctx, req.Credential.ID)
	if err != nil {
		return nil, err
	}
	if passkey == nil {
		return nil, ErrPasskeyNotFound
	}

	// In a full implementation, you would:
	// 1. Decode and parse clientDataJSON
	// 2. Verify challenge matches
	// 3. Verify origin matches configured origins
	// 4. Parse authenticatorData
	// 5. Verify signature using stored public key
	// 6. Verify sign count (detect cloned authenticators)

	// Get user
	user, err := s.userRepo.GetByID(ctx, passkey.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Update sign count
	newSignCount := passkey.SignCount + 1
	if err := s.passkeyRepo.UpdateSignCount(ctx, passkey.CredentialID, newSignCount); err != nil {
		return nil, err
	}

	// Generate tokens
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &LoginFinishResult{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

// DeletePasskey removes a passkey
func (s *PasskeyService) DeletePasskey(ctx context.Context, userID, passkeyID string) error {
	// Get passkey
	passkey, err := s.passkeyRepo.GetByID(ctx, passkeyID)
	if err != nil {
		return err
	}
	if passkey == nil {
		return ErrPasskeyNotFound
	}

	// Verify ownership
	if passkey.UserID != userID {
		return ErrCredentialNotAllowed
	}

	// Check that user has at least one other auth method
	// (another passkey, password, or OAuth identity)
	// This is to prevent account lockout

	return s.passkeyRepo.Delete(ctx, passkeyID)
}

// GetUserPasskeys returns all passkeys for a user
func (s *PasskeyService) GetUserPasskeys(ctx context.Context, userID string) ([]*model.Passkey, error) {
	return s.passkeyRepo.GetByUserID(ctx, userID)
}
