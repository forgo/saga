package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/pkg/jwt"
)

// Mock implementations for Passkey tests

type passkeyMockUserRepo struct {
	users      map[string]*model.User
	emailIndex map[string]*model.User
	getErr     error
}

func newPasskeyMockUserRepo() *passkeyMockUserRepo {
	return &passkeyMockUserRepo{
		users:      make(map[string]*model.User),
		emailIndex: make(map[string]*model.User),
	}
}

func (m *passkeyMockUserRepo) Create(ctx context.Context, user *model.User) error {
	user.ID = "user:" + user.Email
	user.CreatedOn = time.Now()
	user.UpdatedOn = time.Now()
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user
	return nil
}

func (m *passkeyMockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.users[id], nil
}

func (m *passkeyMockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.emailIndex[email], nil
}

func (m *passkeyMockUserRepo) Update(ctx context.Context, user *model.User) error {
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user
	return nil
}

func (m *passkeyMockUserRepo) UpdatePassword(ctx context.Context, userID, hash string) error {
	return nil
}

func (m *passkeyMockUserRepo) Delete(ctx context.Context, id string) error {
	if user, ok := m.users[id]; ok {
		delete(m.emailIndex, user.Email)
		delete(m.users, id)
	}
	return nil
}

func (m *passkeyMockUserRepo) SetEmailVerified(ctx context.Context, userID string, verified bool) error {
	if user, ok := m.users[userID]; ok {
		user.EmailVerified = verified
	}
	return nil
}

type passkeyMockPasskeyRepo struct {
	passkeys    map[string]*model.Passkey
	createErr   error
	getErr      error
	deleteErr   error
	updateErr   error
	countResult int
	countErr    error
}

func newPasskeyMockPasskeyRepo() *passkeyMockPasskeyRepo {
	return &passkeyMockPasskeyRepo{
		passkeys:    make(map[string]*model.Passkey),
		countResult: -1, // -1 means use actual count
	}
}

func (m *passkeyMockPasskeyRepo) Create(ctx context.Context, passkey *model.Passkey) error {
	if m.createErr != nil {
		return m.createErr
	}
	passkey.ID = "passkey:" + passkey.CredentialID
	passkey.CreatedOn = time.Now()
	m.passkeys[passkey.ID] = passkey
	return nil
}

func (m *passkeyMockPasskeyRepo) GetByCredentialID(ctx context.Context, credentialID string) (*model.Passkey, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, pk := range m.passkeys {
		if pk.CredentialID == credentialID {
			return pk, nil
		}
	}
	return nil, nil
}

func (m *passkeyMockPasskeyRepo) GetByUserID(ctx context.Context, userID string) ([]*model.Passkey, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var result []*model.Passkey
	for _, pk := range m.passkeys {
		if pk.UserID == userID {
			result = append(result, pk)
		}
	}
	return result, nil
}

func (m *passkeyMockPasskeyRepo) GetByID(ctx context.Context, id string) (*model.Passkey, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.passkeys[id], nil
}

func (m *passkeyMockPasskeyRepo) UpdateSignCount(ctx context.Context, credentialID string, signCount uint32) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for _, pk := range m.passkeys {
		if pk.CredentialID == credentialID {
			pk.SignCount = signCount
		}
	}
	return nil
}

func (m *passkeyMockPasskeyRepo) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.passkeys, id)
	return nil
}

func (m *passkeyMockPasskeyRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	if m.countResult >= 0 {
		return m.countResult, nil
	}
	count := 0
	for _, pk := range m.passkeys {
		if pk.UserID == userID {
			count++
		}
	}
	return count, nil
}

type passkeyMockTokenRepo struct {
	tokens map[string]*RefreshToken
}

func newPasskeyMockTokenRepo() *passkeyMockTokenRepo {
	return &passkeyMockTokenRepo{
		tokens: make(map[string]*RefreshToken),
	}
}

func (m *passkeyMockTokenRepo) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	m.tokens[token.TokenHash] = token
	return nil
}

func (m *passkeyMockTokenRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error) {
	return m.tokens[hash], nil
}

func (m *passkeyMockTokenRepo) RevokeRefreshToken(ctx context.Context, hash string) error {
	if t, ok := m.tokens[hash]; ok {
		t.Revoked = true
	}
	return nil
}

func (m *passkeyMockTokenRepo) RevokeAllUserTokens(ctx context.Context, userID string) error {
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.Revoked = true
		}
	}
	return nil
}

func (m *passkeyMockTokenRepo) DeleteExpiredTokens(ctx context.Context) error {
	return nil
}

// Setup helper for Passkey service tests
func setupPasskeyService(t *testing.T) (*PasskeyService, *passkeyMockUserRepo, *passkeyMockPasskeyRepo, *passkeyMockTokenRepo) {
	t.Helper()

	userRepo := newPasskeyMockUserRepo()
	passkeyRepo := newPasskeyMockPasskeyRepo()
	tokenRepo := newPasskeyMockTokenRepo()

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

	passkeyConfig := PasskeyConfig{
		RPID:            "localhost",
		RPName:          "Test App",
		RPOrigins:       []string{"http://localhost:3000"},
		Timeout:         60 * time.Second,
		RequireUV:       false,
		AttestationType: "none",
	}

	passkeyService := NewPasskeyService(PasskeyServiceConfig{
		Config:       passkeyConfig,
		PasskeyRepo:  passkeyRepo,
		UserRepo:     userRepo,
		TokenService: tokenService,
	})

	return passkeyService, userRepo, passkeyRepo, tokenRepo
}

// Helper to create a test user
func createTestUser(t *testing.T, userRepo *passkeyMockUserRepo, email string) *model.User {
	t.Helper()
	firstname := "Test"
	lastname := "User"
	user := &model.User{
		Email:     email,
		Firstname: &firstname,
		Lastname:  &lastname,
	}
	_ = userRepo.Create(context.Background(), user)
	return user
}

// Tests for challengeStore

func TestChallengeStore_Create(t *testing.T) {
	store := newChallengeStore()

	challenge, err := store.Create("user123", false)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if len(challenge) != challengeLength {
		t.Errorf("expected challenge length %d, got %d", challengeLength, len(challenge))
	}

	// Verify challenge is stored
	key := base64.RawURLEncoding.EncodeToString(challenge)
	stored, exists := store.challenges[key]
	if !exists {
		t.Error("challenge should be stored")
	}
	if stored.UserID != "user123" {
		t.Error("stored challenge should have correct user ID")
	}
	if stored.IsLogin {
		t.Error("stored challenge should have IsLogin = false")
	}
}

func TestChallengeStore_Verify_Success(t *testing.T) {
	store := newChallengeStore()

	challenge, _ := store.Create("user123", false)
	key := base64.RawURLEncoding.EncodeToString(challenge)

	if !store.Verify(key, "user123", false) {
		t.Error("Verify should return true for valid challenge")
	}

	// Challenge should be consumed (deleted after verification)
	if store.Verify(key, "user123", false) {
		t.Error("challenge should be consumed after verification")
	}
}

func TestChallengeStore_Verify_WrongUser(t *testing.T) {
	store := newChallengeStore()

	challenge, _ := store.Create("user123", false)
	key := base64.RawURLEncoding.EncodeToString(challenge)

	if store.Verify(key, "wronguser", false) {
		t.Error("Verify should return false for wrong user")
	}
}

func TestChallengeStore_Verify_WrongMode(t *testing.T) {
	store := newChallengeStore()

	challenge, _ := store.Create("user123", false) // registration
	key := base64.RawURLEncoding.EncodeToString(challenge)

	if store.Verify(key, "user123", true) { // trying to use for login
		t.Error("Verify should return false for wrong mode")
	}
}

func TestChallengeStore_Verify_Expired(t *testing.T) {
	store := newChallengeStore()

	challenge, _ := store.Create("user123", false)
	key := base64.RawURLEncoding.EncodeToString(challenge)

	// Manually expire the challenge
	store.challenges[key].ExpiresAt = time.Now().Add(-1 * time.Second)

	if store.Verify(key, "user123", false) {
		t.Error("Verify should return false for expired challenge")
	}
}

func TestChallengeStore_Verify_NotFound(t *testing.T) {
	store := newChallengeStore()

	if store.Verify("nonexistent-challenge", "user123", false) {
		t.Error("Verify should return false for nonexistent challenge")
	}
}

func TestChallengeStore_Cleanup(t *testing.T) {
	store := newChallengeStore()

	// Create some challenges
	challenge1, _ := store.Create("user1", false)
	challenge2, _ := store.Create("user2", false)

	key1 := base64.RawURLEncoding.EncodeToString(challenge1)
	key2 := base64.RawURLEncoding.EncodeToString(challenge2)

	// Expire one challenge
	store.challenges[key1].ExpiresAt = time.Now().Add(-1 * time.Second)

	// Cleanup is called on Create
	_, _ = store.Create("user3", false)

	// Expired challenge should be cleaned up
	if _, exists := store.challenges[key1]; exists {
		t.Error("expired challenge should be cleaned up")
	}

	// Valid challenge should still exist
	if _, exists := store.challenges[key2]; !exists {
		t.Error("valid challenge should still exist")
	}
}

// Tests for StartRegistration

func TestPasskeyService_StartRegistration_Success(t *testing.T) {
	passkeyService, userRepo, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	result, err := passkeyService.StartRegistration(ctx, RegistrationStartRequest{
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("StartRegistration failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Challenge == "" {
		t.Error("expected Challenge to be set")
	}
	if result.RP.ID != "localhost" {
		t.Errorf("expected RP.ID localhost, got %s", result.RP.ID)
	}
	if result.RP.Name != "Test App" {
		t.Errorf("expected RP.Name Test App, got %s", result.RP.Name)
	}
	if result.User.Name != "test@example.com" {
		t.Errorf("expected User.Name to be email, got %s", result.User.Name)
	}
	if result.User.DisplayName != "Test User" {
		t.Errorf("expected User.DisplayName 'Test User', got %s", result.User.DisplayName)
	}
	if len(result.PubKeyCredParams) != 2 {
		t.Errorf("expected 2 PubKeyCredParams, got %d", len(result.PubKeyCredParams))
	}
	if result.Timeout == 0 {
		t.Error("expected Timeout to be set")
	}
}

func TestPasskeyService_StartRegistration_UserNotFound(t *testing.T) {
	passkeyService, _, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	_, err := passkeyService.StartRegistration(ctx, RegistrationStartRequest{
		UserID: "nonexistent-user",
	})

	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

func TestPasskeyService_StartRegistration_PasskeyLimitReached(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	// Simulate max passkeys reached
	passkeyRepo.countResult = maxPasskeysPerUser

	_, err := passkeyService.StartRegistration(ctx, RegistrationStartRequest{
		UserID: user.ID,
	})

	if err != ErrPasskeyLimitReached {
		t.Errorf("expected ErrPasskeyLimitReached, got %v", err)
	}
}

func TestPasskeyService_StartRegistration_ExcludesExistingCredentials(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	// Add existing passkey
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user.ID,
		CredentialID: "existing-cred-123",
		PublicKey:    []byte("pubkey"),
	})

	result, err := passkeyService.StartRegistration(ctx, RegistrationStartRequest{
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("StartRegistration failed: %v", err)
	}

	if len(result.ExcludeCredentials) != 1 {
		t.Errorf("expected 1 excluded credential, got %d", len(result.ExcludeCredentials))
	}
	if result.ExcludeCredentials[0].ID != "existing-cred-123" {
		t.Error("excluded credential should have correct ID")
	}
}

func TestPasskeyService_StartRegistration_DisplayNameFallback(t *testing.T) {
	passkeyService, userRepo, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	// User without firstname/lastname
	user := &model.User{
		Email: "noname@example.com",
	}
	_ = userRepo.Create(ctx, user)

	result, err := passkeyService.StartRegistration(ctx, RegistrationStartRequest{
		UserID: user.ID,
	})
	if err != nil {
		t.Fatalf("StartRegistration failed: %v", err)
	}

	// Should fall back to email
	if result.User.DisplayName != "noname@example.com" {
		t.Errorf("expected DisplayName to fall back to email, got %s", result.User.DisplayName)
	}
}

// Tests for FinishRegistration

func TestPasskeyService_FinishRegistration_Success(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	result, err := passkeyService.FinishRegistration(ctx, RegistrationFinishRequest{
		UserID: user.ID,
		Name:   "My Passkey",
		Credential: &CredentialResponse{
			ID:    "new-cred-456",
			RawID: "new-cred-456",
			Type:  "public-key",
			Response: AttestationResponse{
				ClientDataJSON:    "mock-client-data",
				AttestationObject: "mock-attestation",
			},
		},
	})
	if err != nil {
		t.Fatalf("FinishRegistration failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Passkey == nil {
		t.Fatal("expected Passkey to be set")
	}
	if result.Passkey.CredentialID != "new-cred-456" {
		t.Errorf("expected CredentialID new-cred-456, got %s", result.Passkey.CredentialID)
	}
	if result.Passkey.Name != "My Passkey" {
		t.Errorf("expected Name 'My Passkey', got %s", result.Passkey.Name)
	}
	if result.Passkey.SignCount != 0 {
		t.Error("initial SignCount should be 0")
	}

	// Verify passkey was stored
	stored, _ := passkeyRepo.GetByCredentialID(ctx, "new-cred-456")
	if stored == nil {
		t.Error("passkey should be stored in repository")
	}
}

func TestPasskeyService_FinishRegistration_UserNotFound(t *testing.T) {
	passkeyService, _, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	_, err := passkeyService.FinishRegistration(ctx, RegistrationFinishRequest{
		UserID: "nonexistent-user",
		Name:   "My Passkey",
		Credential: &CredentialResponse{
			ID:   "cred-123",
			Type: "public-key",
		},
	})

	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

// Tests for StartLogin

func TestPasskeyService_StartLogin_WithEmail(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	// Add passkeys for user
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user.ID,
		CredentialID: "cred-1",
		PublicKey:    []byte("pubkey"),
	})
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user.ID,
		CredentialID: "cred-2",
		PublicKey:    []byte("pubkey"),
	})

	result, err := passkeyService.StartLogin(ctx, LoginStartRequest{
		Email: "test@example.com",
	})
	if err != nil {
		t.Fatalf("StartLogin failed: %v", err)
	}

	if result.Challenge == "" {
		t.Error("expected Challenge to be set")
	}
	if result.RPID != "localhost" {
		t.Errorf("expected RPID localhost, got %s", result.RPID)
	}
	if len(result.AllowCredentials) != 2 {
		t.Errorf("expected 2 allowed credentials, got %d", len(result.AllowCredentials))
	}
}

func TestPasskeyService_StartLogin_WithoutEmail_DiscoverableFlow(t *testing.T) {
	passkeyService, _, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	result, err := passkeyService.StartLogin(ctx, LoginStartRequest{
		Email: "", // No email - discoverable credential flow
	})
	if err != nil {
		t.Fatalf("StartLogin failed: %v", err)
	}

	if result.Challenge == "" {
		t.Error("expected Challenge to be set")
	}
	if len(result.AllowCredentials) != 0 {
		t.Error("discoverable flow should have empty AllowCredentials")
	}
}

func TestPasskeyService_StartLogin_UserNotFound_NoLeak(t *testing.T) {
	passkeyService, _, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	// Request login for nonexistent user
	result, err := passkeyService.StartLogin(ctx, LoginStartRequest{
		Email: "nonexistent@example.com",
	})

	// Should not error - don't reveal whether user exists
	if err != nil {
		t.Fatalf("StartLogin should not error for nonexistent user: %v", err)
	}

	if result.Challenge == "" {
		t.Error("expected Challenge even for nonexistent user")
	}

	// AllowCredentials should be empty (no passkeys to allow)
	if len(result.AllowCredentials) != 0 {
		t.Error("should have empty AllowCredentials for nonexistent user")
	}
}

// Tests for FinishLogin

func TestPasskeyService_FinishLogin_Success(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	// Add passkey for user
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user.ID,
		CredentialID: "cred-login",
		PublicKey:    []byte("pubkey"),
		SignCount:    5,
	})

	result, err := passkeyService.FinishLogin(ctx, LoginFinishRequest{
		Credential: &AssertionResponse{
			ID:    "cred-login",
			RawID: "cred-login",
			Type:  "public-key",
			Response: AuthenticatorAssertionResponse{
				ClientDataJSON:    "mock-client-data",
				AuthenticatorData: "mock-auth-data",
				Signature:         "mock-signature",
			},
		},
	})
	if err != nil {
		t.Fatalf("FinishLogin failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.User == nil {
		t.Fatal("expected User to be set")
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", result.User.Email)
	}
	if result.TokenPair == nil {
		t.Error("expected TokenPair to be set")
	}

	// Verify sign count was updated
	passkey, _ := passkeyRepo.GetByCredentialID(ctx, "cred-login")
	if passkey.SignCount != 6 {
		t.Errorf("expected SignCount to be incremented to 6, got %d", passkey.SignCount)
	}
}

func TestPasskeyService_FinishLogin_PasskeyNotFound(t *testing.T) {
	passkeyService, _, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	_, err := passkeyService.FinishLogin(ctx, LoginFinishRequest{
		Credential: &AssertionResponse{
			ID:   "nonexistent-cred",
			Type: "public-key",
		},
	})

	if err != ErrPasskeyNotFound {
		t.Errorf("expected ErrPasskeyNotFound, got %v", err)
	}
}

func TestPasskeyService_FinishLogin_UserDeleted(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	// Add passkey for user
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user.ID,
		CredentialID: "orphan-cred",
		PublicKey:    []byte("pubkey"),
	})

	// Delete user but keep passkey (orphaned state)
	_ = userRepo.Delete(ctx, user.ID)

	_, err := passkeyService.FinishLogin(ctx, LoginFinishRequest{
		Credential: &AssertionResponse{
			ID:   "orphan-cred",
			Type: "public-key",
		},
	})

	if err == nil {
		t.Error("expected error when user is deleted")
	}
}

// Tests for DeletePasskey

func TestPasskeyService_DeletePasskey_Success(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	// Add passkey
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user.ID,
		CredentialID: "to-delete",
		PublicKey:    []byte("pubkey"),
	})

	passkey, _ := passkeyRepo.GetByCredentialID(ctx, "to-delete")

	err := passkeyService.DeletePasskey(ctx, user.ID, passkey.ID)
	if err != nil {
		t.Fatalf("DeletePasskey failed: %v", err)
	}

	// Verify passkey was deleted
	deleted, _ := passkeyRepo.GetByID(ctx, passkey.ID)
	if deleted != nil {
		t.Error("passkey should be deleted")
	}
}

func TestPasskeyService_DeletePasskey_NotFound(t *testing.T) {
	passkeyService, userRepo, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	err := passkeyService.DeletePasskey(ctx, user.ID, "nonexistent-passkey")

	if err != ErrPasskeyNotFound {
		t.Errorf("expected ErrPasskeyNotFound, got %v", err)
	}
}

func TestPasskeyService_DeletePasskey_WrongUser(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user1 := createTestUser(t, userRepo, "user1@example.com")
	user2 := createTestUser(t, userRepo, "user2@example.com")

	// Add passkey for user1
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user1.ID,
		CredentialID: "user1-cred",
		PublicKey:    []byte("pubkey"),
	})

	passkey, _ := passkeyRepo.GetByCredentialID(ctx, "user1-cred")

	// Try to delete as user2
	err := passkeyService.DeletePasskey(ctx, user2.ID, passkey.ID)

	if err != ErrCredentialNotAllowed {
		t.Errorf("expected ErrCredentialNotAllowed, got %v", err)
	}
}

// Tests for GetUserPasskeys

func TestPasskeyService_GetUserPasskeys_Success(t *testing.T) {
	passkeyService, userRepo, passkeyRepo, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	// Add multiple passkeys
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user.ID,
		CredentialID: "cred-1",
		PublicKey:    []byte("pubkey"),
		Name:         "Passkey 1",
	})
	_ = passkeyRepo.Create(ctx, &model.Passkey{
		UserID:       user.ID,
		CredentialID: "cred-2",
		PublicKey:    []byte("pubkey"),
		Name:         "Passkey 2",
	})

	passkeys, err := passkeyService.GetUserPasskeys(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserPasskeys failed: %v", err)
	}

	if len(passkeys) != 2 {
		t.Errorf("expected 2 passkeys, got %d", len(passkeys))
	}
}

func TestPasskeyService_GetUserPasskeys_Empty(t *testing.T) {
	passkeyService, userRepo, _, _ := setupPasskeyService(t)
	ctx := context.Background()

	user := createTestUser(t, userRepo, "test@example.com")

	passkeys, err := passkeyService.GetUserPasskeys(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserPasskeys failed: %v", err)
	}

	if len(passkeys) != 0 {
		t.Errorf("expected 0 passkeys, got %d", len(passkeys))
	}
}

// Tests for error variables

func TestPasskeyErrorVariables(t *testing.T) {
	errors := []error{
		ErrPasskeyNotFound,
		ErrInvalidChallenge,
		ErrInvalidCredential,
		ErrCredentialNotAllowed,
		ErrSignCountMismatch,
		ErrPasskeyLimitReached,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("error variable should not be nil")
		}
		if err.Error() == "" {
			t.Error("error message should not be empty")
		}
	}
}

// Tests for constants

func TestPasskeyConstants(t *testing.T) {
	if challengeExpiration != 60*time.Second {
		t.Errorf("unexpected challengeExpiration: %v", challengeExpiration)
	}
	if maxPasskeysPerUser != 10 {
		t.Errorf("unexpected maxPasskeysPerUser: %d", maxPasskeysPerUser)
	}
	if challengeLength != 32 {
		t.Errorf("unexpected challengeLength: %d", challengeLength)
	}
}

// Tests for WebAuthn data structures

func TestWebAuthnStructures(t *testing.T) {
	// Test RelyingPartyInfo
	rp := RelyingPartyInfo{
		ID:   "example.com",
		Name: "Example App",
	}
	if rp.ID != "example.com" {
		t.Error("RelyingPartyInfo.ID mismatch")
	}

	// Test WebAuthnUserInfo
	user := WebAuthnUserInfo{
		ID:          "user-123",
		Name:        "user@example.com",
		DisplayName: "Test User",
	}
	if user.Name != "user@example.com" {
		t.Error("WebAuthnUserInfo.Name mismatch")
	}

	// Test PubKeyCredParam
	param := PubKeyCredParam{
		Type: "public-key",
		Alg:  -7, // ES256
	}
	if param.Alg != -7 {
		t.Error("PubKeyCredParam.Alg mismatch")
	}

	// Test CredentialDescriptor
	cred := CredentialDescriptor{
		Type:       "public-key",
		ID:         "cred-123",
		Transports: []string{"usb", "internal"},
	}
	if len(cred.Transports) != 2 {
		t.Error("CredentialDescriptor.Transports mismatch")
	}
}
