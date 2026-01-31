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
	"golang.org/x/crypto/bcrypt"
)

// Mock implementations

type mockUserRepo struct {
	users       map[string]*model.User
	emailIndex  map[string]*model.User
	createErr   error
	getErr      error
	updateErr   error
	passwordErr error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:      make(map[string]*model.User),
		emailIndex: make(map[string]*model.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	user.ID = "user:" + user.Email
	user.CreatedOn = time.Now()
	user.UpdatedOn = time.Now()
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.users[id], nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.emailIndex[email], nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user
	return nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID, hash string) error {
	if m.passwordErr != nil {
		return m.passwordErr
	}
	if user, ok := m.users[userID]; ok {
		user.Hash = &hash
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	if user, ok := m.users[id]; ok {
		delete(m.emailIndex, user.Email)
		delete(m.users, id)
	}
	return nil
}

func (m *mockUserRepo) SetEmailVerified(ctx context.Context, userID string, verified bool) error {
	if user, ok := m.users[userID]; ok {
		user.EmailVerified = verified
	}
	return nil
}

type mockIdentityRepo struct {
	identities map[string]*model.Identity
}

func newMockIdentityRepo() *mockIdentityRepo {
	return &mockIdentityRepo{
		identities: make(map[string]*model.Identity),
	}
}

func (m *mockIdentityRepo) Create(ctx context.Context, identity *model.Identity) error {
	identity.ID = "identity:" + identity.Provider + ":" + identity.ProviderUserID
	m.identities[identity.ID] = identity
	return nil
}

func (m *mockIdentityRepo) GetByProviderID(ctx context.Context, provider, providerUserID string) (*model.Identity, error) {
	for _, id := range m.identities {
		if id.Provider == provider && id.ProviderUserID == providerUserID {
			return id, nil
		}
	}
	return nil, nil
}

func (m *mockIdentityRepo) GetByUserID(ctx context.Context, userID string) ([]*model.Identity, error) {
	var result []*model.Identity
	for _, id := range m.identities {
		if id.UserID == userID {
			result = append(result, id)
		}
	}
	return result, nil
}

func (m *mockIdentityRepo) GetByProviderEmail(ctx context.Context, provider, email string) (*model.Identity, error) {
	for _, id := range m.identities {
		if id.Provider == provider && id.ProviderEmail != nil && *id.ProviderEmail == email {
			return id, nil
		}
	}
	return nil, nil
}

func (m *mockIdentityRepo) Delete(ctx context.Context, id string) error {
	delete(m.identities, id)
	return nil
}

func (m *mockIdentityRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	count := 0
	for _, id := range m.identities {
		if id.UserID == userID {
			count++
		}
	}
	return count, nil
}

type mockPasskeyRepo struct {
	passkeys map[string]*model.Passkey
}

func newMockPasskeyRepo() *mockPasskeyRepo {
	return &mockPasskeyRepo{
		passkeys: make(map[string]*model.Passkey),
	}
}

func (m *mockPasskeyRepo) Create(ctx context.Context, passkey *model.Passkey) error {
	passkey.ID = "passkey:" + passkey.CredentialID
	m.passkeys[passkey.ID] = passkey
	return nil
}

func (m *mockPasskeyRepo) GetByCredentialID(ctx context.Context, credentialID string) (*model.Passkey, error) {
	for _, pk := range m.passkeys {
		if pk.CredentialID == credentialID {
			return pk, nil
		}
	}
	return nil, nil
}

func (m *mockPasskeyRepo) GetByUserID(ctx context.Context, userID string) ([]*model.Passkey, error) {
	var result []*model.Passkey
	for _, pk := range m.passkeys {
		if pk.UserID == userID {
			result = append(result, pk)
		}
	}
	return result, nil
}

func (m *mockPasskeyRepo) GetByID(ctx context.Context, id string) (*model.Passkey, error) {
	return m.passkeys[id], nil
}

func (m *mockPasskeyRepo) UpdateSignCount(ctx context.Context, credentialID string, signCount uint32) error {
	for _, pk := range m.passkeys {
		if pk.CredentialID == credentialID {
			pk.SignCount = signCount
		}
	}
	return nil
}

func (m *mockPasskeyRepo) Delete(ctx context.Context, id string) error {
	delete(m.passkeys, id)
	return nil
}

func (m *mockPasskeyRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	count := 0
	for _, pk := range m.passkeys {
		if pk.UserID == userID {
			count++
		}
	}
	return count, nil
}

type authMockTokenRepo struct {
	tokens map[string]*RefreshToken
}

func newAuthMockTokenRepo() *authMockTokenRepo {
	return &authMockTokenRepo{
		tokens: make(map[string]*RefreshToken),
	}
}

func (m *authMockTokenRepo) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *authMockTokenRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	return m.tokens[hash], nil
}

func (m *authMockTokenRepo) RevokeRefreshToken(ctx context.Context, hash string) error {
	if t, ok := m.tokens[hash]; ok {
		t.Revoked = true
	}
	return nil
}

func (m *authMockTokenRepo) RevokeAllUserTokens(ctx context.Context, userID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
		}
	}
	return nil
}

func (m *authMockTokenRepo) DeleteExpiredTokens(ctx context.Context) error {
	now := time.Now()
	for hash, t := range m.tokens {
		if t.ExpiresAt.Before(now) {
			delete(m.tokens, hash)
		}
	}
	return nil
}

// Test helper to create auth service with mocks
func setupAuthService(t *testing.T) (*AuthService, *mockUserRepo, *mockIdentityRepo, *mockPasskeyRepo, *authMockTokenRepo) {
	t.Helper()

	userRepo := newMockUserRepo()
	identityRepo := newMockIdentityRepo()
	passkeyRepo := newMockPasskeyRepo()
	tokenRepo := newAuthMockTokenRepo()

	// Generate a test RSA key pair for the JWT service
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate test RSA key: %v", err)
	}

	jwtService := jwt.NewTestService(privateKey, "test-issuer", 15*time.Minute)

	tokenService := NewTokenService(TokenServiceConfig{
		JWTService:      jwtService,
		TokenRepo:       tokenRepo,
		RefreshDuration: 24 * time.Hour,
	})

	authService := NewAuthService(AuthServiceConfig{
		UserRepo:     userRepo,
		IdentityRepo: identityRepo,
		PasskeyRepo:  passkeyRepo,
		TokenService: tokenService,
	})

	return authService, userRepo, identityRepo, passkeyRepo, tokenRepo
}

// Tests

func TestAuthService_Register_Success(t *testing.T) {
	authService, userRepo, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	result, err := authService.Register(ctx, RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		Firstname: "John",
		Lastname:  "Doe",
	})

	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", result.User.Email)
	}
	if result.User.Hash == nil {
		t.Error("expected password hash to be set")
	}

	// Verify password was hashed correctly
	err = bcrypt.CompareHashAndPassword([]byte(*result.User.Hash), []byte("password123"))
	if err != nil {
		t.Error("password hash verification failed")
	}

	// Verify user was stored
	stored, _ := userRepo.GetByEmail(ctx, "test@example.com")
	if stored == nil {
		t.Error("user was not stored in repository")
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	tests := []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"no at sign", "testexample.com"},
		{"no domain", "test@"},
		{"no local part", "@example.com"},
		{"no TLD", "test@example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authService.Register(ctx, RegisterRequest{
				Email:    tt.email,
				Password: "password123",
			})
			if !errors.Is(err, ErrInvalidEmail) {
				t.Errorf("expected ErrInvalidEmail, got %v", err)
			}
		})
	}
}

func TestAuthService_Register_InvalidPassword(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"empty password", "", ErrPasswordRequired},
		{"too short", "short", ErrPasswordTooShort},
		{"exactly 7 chars", "1234567", ErrPasswordTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authService.Register(ctx, RegisterRequest{
				Email:    "test@example.com",
				Password: tt.password,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Register first user
	_, err := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Try to register with same email
	_, err = authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "different123",
	})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestAuthService_Register_EmailNormalization(t *testing.T) {
	authService, userRepo, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	_, err := authService.Register(ctx, RegisterRequest{
		Email:    "  TEST@EXAMPLE.COM  ",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Should be stored lowercase and trimmed
	user, _ := userRepo.GetByEmail(ctx, "test@example.com")
	if user == nil {
		t.Error("user should be findable by normalized email")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Register user first
	_, err := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Login
	result, err := authService.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", result.User.Email)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Register user first
	_, err := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Login with wrong password
	_, err = authService.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	_, err := authService.Login(ctx, LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_OAuthOnlyUser(t *testing.T) {
	authService, userRepo, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Create OAuth-only user (no password)
	user := &model.User{
		Email: "oauth@example.com",
		Hash:  nil, // No password
	}
	_ = userRepo.Create(ctx, user)

	// Try to login with password
	_, err := authService.Login(ctx, LoginRequest{
		Email:    "oauth@example.com",
		Password: "anypassword",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for OAuth-only user, got %v", err)
	}
}

func TestAuthService_GetUserByID_Success(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Register user first
	regResult, _ := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	// Get by ID
	user, err := authService.GetUserByID(ctx, regResult.User.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	_, err := authService.GetUserByID(ctx, "nonexistent")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestAuthService_GetUserWithIdentities(t *testing.T) {
	authService, _, identityRepo, passkeyRepo, _ := setupAuthService(t)
	ctx := context.Background()

	// Register user
	regResult, _ := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	// Add an identity
	email := "test@gmail.com"
	_ = identityRepo.Create(ctx, &model.Identity{
		UserID:         regResult.User.ID,
		Provider:       "google",
		ProviderUserID: "google123",
		ProviderEmail:  &email,
	})

	// Add a passkey
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       regResult.User.ID,
		CredentialID: "cred123",
		PublicKey:    []byte("pubkey"),
	})

	// Get with identities
	result, err := authService.GetUserWithIdentities(ctx, regResult.User.ID)
	if err != nil {
		t.Fatalf("GetUserWithIdentities failed: %v", err)
	}
	if len(result.Identities) != 1 {
		t.Errorf("expected 1 identity, got %d", len(result.Identities))
	}
	if len(result.Passkeys) != 1 {
		t.Errorf("expected 1 passkey, got %d", len(result.Passkeys))
	}
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Register user
	regResult, _ := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "oldpassword123",
	})

	// Change password
	err := authService.ChangePassword(ctx, regResult.User.ID, "oldpassword123", "newpassword456")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Old password should no longer work
	_, err = authService.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "oldpassword123",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Error("old password should no longer work")
	}

	// New password should work
	_, err = authService.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "newpassword456",
	})
	if err != nil {
		t.Errorf("new password should work: %v", err)
	}
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Register user
	regResult, _ := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "oldpassword123",
	})

	// Try to change with wrong old password
	err := authService.ChangePassword(ctx, regResult.User.ID, "wrongoldpassword", "newpassword456")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_ChangePassword_InvalidNewPassword(t *testing.T) {
	authService, _, _, _, _ := setupAuthService(t)
	ctx := context.Background()

	// Register user
	regResult, _ := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "oldpassword123",
	})

	// Try to change to invalid password
	err := authService.ChangePassword(ctx, regResult.User.ID, "oldpassword123", "short")
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestAuthService_Logout(t *testing.T) {
	authService, _, _, _, tokenRepo := setupAuthService(t)
	ctx := context.Background()

	// Register user
	regResult, _ := authService.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	// Logout
	err := authService.Logout(ctx, regResult.User.ID)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Verify tokens are revoked
	for _, token := range tokenRepo.tokens {
		if token.UserID == regResult.User.ID && !token.Revoked {
			t.Error("expected all user tokens to be revoked")
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"valid 8 chars", "12345678", nil},
		{"valid long", "this is a valid long password", nil},
		{"empty", "", ErrPasswordRequired},
		{"too short 1", "1", ErrPasswordTooShort},
		{"too short 7", "1234567", ErrPasswordTooShort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validatePassword(%q) = %v, want %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"user+tag@example.org", true},
		{"", false},
		{"noatsign", false},
		{"@nodomain.com", false},
		{"nolocal@", false},
		{"nodot@domain", false},
		{"test@.com", false},
		{"test@domain.", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := isValidEmail(tt.email)
			if got != tt.valid {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}
