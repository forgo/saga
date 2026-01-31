// Package tests contains end-to-end acceptance tests for the Saga API.
package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/service"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/helpers"
	"github.com/forgo/saga/api/internal/testing/testdb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
FEATURE: Authentication
DOMAIN: Auth

ACCEPTANCE CRITERIA:
===================

AC-AUTH-001: Register with Email/Password
  GIVEN valid email and password (8+ chars)
  WHEN user submits registration
  THEN user is created with hashed password
  AND access token + refresh token returned
  AND user can authenticate with credentials

AC-AUTH-002: Register Duplicate Email
  GIVEN an existing user with email X
  WHEN new user registers with email X
  THEN request fails with email already exists error

AC-AUTH-003: Login with Valid Credentials
  GIVEN registered user with email/password
  WHEN user logs in with correct credentials
  THEN access token + refresh token returned
  AND tokens are valid for authentication

AC-AUTH-004: Login with Invalid Credentials
  GIVEN registered user
  WHEN user logs in with wrong password
  THEN request fails with invalid credentials error

AC-AUTH-005: Refresh Token
  GIVEN valid refresh token
  WHEN user requests token refresh
  THEN new access token returned
  AND old refresh token invalidated (rotation)

AC-AUTH-006: Refresh with Invalid Token
  GIVEN invalid/expired refresh token
  WHEN user requests token refresh
  THEN request fails with invalid token error

AC-AUTH-007: Logout Revokes Tokens
  GIVEN authenticated user
  WHEN user logs out
  THEN refresh token is invalidated
  AND subsequent refresh requests fail

AC-AUTH-008: OAuth Google Flow
  GIVEN valid Google OAuth identity
  WHEN user authenticates via Google
  THEN user is created/linked
  AND tokens returned

AC-AUTH-009: OAuth Identity Linking
  GIVEN existing user authenticated via email
  WHEN user links Google OAuth identity
  THEN identity is associated with account
  AND user can login via either method

AC-AUTH-010: Cannot Link Already-Used OAuth
  GIVEN OAuth identity linked to user A
  WHEN user B attempts to link same identity
  THEN request fails with conflict error

AC-AUTH-011: Passkey Registration
  GIVEN authenticated user
  WHEN user starts passkey registration
  THEN challenge is returned
  AND completing registration stores credential

AC-AUTH-012: Passkey Authentication
  GIVEN user with registered passkey
  WHEN user authenticates with passkey
  THEN authentication succeeds
  AND tokens returned

AC-AUTH-013: Passkey Deletion
  GIVEN user with multiple passkeys
  WHEN user deletes a passkey
  THEN passkey is removed
  AND cannot authenticate with deleted passkey
*/

// createAuthService creates an AuthService instance for testing
func createAuthService(t *testing.T, tdb *testdb.TestDB) *service.AuthService {
	t.Helper()

	userRepo := repository.NewUserRepository(tdb.DB)
	identityRepo := repository.NewIdentityRepository(tdb.DB)
	passkeyRepo := repository.NewPasskeyRepository(tdb.DB)
	tokenRepo := repository.NewTokenRepository(tdb.DB)

	jwtService := helpers.NewTestJWTService(t)

	tokenService := service.NewTokenService(service.TokenServiceConfig{
		JWTService:      jwtService,
		TokenRepo:       tokenRepo,
		RefreshDuration: 24 * time.Hour,
	})

	return service.NewAuthService(service.AuthServiceConfig{
		UserRepo:     userRepo,
		IdentityRepo: identityRepo,
		PasskeyRepo:  passkeyRepo,
		TokenService: tokenService,
	})
}

// createOAuthService creates an OAuthService instance for testing
func createOAuthService(t *testing.T, tdb *testdb.TestDB, authService *service.AuthService) *service.OAuthService {
	t.Helper()

	userRepo := repository.NewUserRepository(tdb.DB)
	identityRepo := repository.NewIdentityRepository(tdb.DB)
	tokenRepo := repository.NewTokenRepository(tdb.DB)

	jwtService := helpers.NewTestJWTService(t)

	tokenService := service.NewTokenService(service.TokenServiceConfig{
		JWTService:      jwtService,
		TokenRepo:       tokenRepo,
		RefreshDuration: 24 * time.Hour,
	})

	return service.NewOAuthService(service.OAuthServiceConfig{
		Config: service.OAuthConfig{
			Google: service.GoogleOAuthConfig{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
				RedirectURI:  "http://localhost:3000/callback",
			},
		},
		AuthService:  authService,
		IdentityRepo: identityRepo,
		UserRepo:     userRepo,
		TokenService: tokenService,
	})
}

// createPasskeyService creates a PasskeyService instance for testing
func createPasskeyService(t *testing.T, tdb *testdb.TestDB) *service.PasskeyService {
	t.Helper()

	userRepo := repository.NewUserRepository(tdb.DB)
	passkeyRepo := repository.NewPasskeyRepository(tdb.DB)
	tokenRepo := repository.NewTokenRepository(tdb.DB)

	jwtService := helpers.NewTestJWTService(t)

	tokenService := service.NewTokenService(service.TokenServiceConfig{
		JWTService:      jwtService,
		TokenRepo:       tokenRepo,
		RefreshDuration: 24 * time.Hour,
	})

	return service.NewPasskeyService(service.PasskeyServiceConfig{
		Config: service.PasskeyConfig{
			RPID:            "saga.test",
			RPName:          "Saga Test",
			RPOrigins:       []string{"http://localhost:3000"},
			Timeout:         60 * time.Second,
			RequireUV:       false,
			AttestationType: "none",
		},
		PasskeyRepo:  passkeyRepo,
		UserRepo:     userRepo,
		TokenService: tokenService,
	})
}

func TestAuth_RegisterWithEmailPassword(t *testing.T) {
	// AC-AUTH-001: Register with Email/Password
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Register new user
	result, err := authService.Register(ctx, service.RegisterRequest{
		Email:     "newuser@test.local",
		Password:  "password123",
		Firstname: "Test",
		Lastname:  "User",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.User)
	require.NotNil(t, result.TokenPair)

	// Verify user was created
	assert.NotEmpty(t, result.User.ID)
	assert.Equal(t, "newuser@test.local", result.User.Email)
	assert.False(t, result.User.EmailVerified) // Not verified until email confirmation

	// Verify tokens were generated
	assert.NotEmpty(t, result.TokenPair.AccessToken)
	assert.NotEmpty(t, result.TokenPair.RefreshToken)
	assert.Equal(t, "Bearer", result.TokenPair.TokenType)

	// Verify user can authenticate
	claims, err := authService.ValidateAccessToken(result.TokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, result.User.ID, claims.UserID)
}

func TestAuth_RegisterPasswordValidation(t *testing.T) {
	// AC-AUTH-001 (validation): Password must be 8+ characters
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "empty password",
			password: "",
			wantErr:  service.ErrPasswordRequired,
		},
		{
			name:     "too short password",
			password: "1234567",
			wantErr:  service.ErrPasswordTooShort,
		},
		{
			name:     "exactly 8 chars is valid",
			password: "12345678",
			wantErr:  nil,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use index for unique email to avoid invalid chars from test name
			_, err := authService.Register(ctx, service.RegisterRequest{
				Email:    fmt.Sprintf("passtest_%d@test.local", i),
				Password: tt.password,
			})

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	// AC-AUTH-002: Register Duplicate Email
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Create existing user
	existingUser := f.CreateUser(t, func(o *fixtures.UserOpts) {
		o.Email = "existing@test.local"
	})
	require.NotEmpty(t, existingUser.ID)

	// Try to register with same email
	_, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "existing@test.local",
		Password: "password123",
	})

	require.ErrorIs(t, err, service.ErrEmailAlreadyExists)
}

func TestAuth_RegisterDuplicateEmailCaseInsensitive(t *testing.T) {
	// AC-AUTH-002 (variation): Email comparison should be case-insensitive
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Register with lowercase email
	_, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "test@test.local",
		Password: "password123",
	})
	require.NoError(t, err)

	// Try to register with uppercase version
	_, err = authService.Register(ctx, service.RegisterRequest{
		Email:    "TEST@TEST.LOCAL",
		Password: "password456",
	})

	require.ErrorIs(t, err, service.ErrEmailAlreadyExists)
}

func TestAuth_LoginWithValidCredentials(t *testing.T) {
	// AC-AUTH-003: Login with Valid Credentials
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Register user first
	regResult, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "logintest@test.local",
		Password: "correctpassword",
	})
	require.NoError(t, err)

	// Login with correct credentials
	loginResult, err := authService.Login(ctx, service.LoginRequest{
		Email:    "logintest@test.local",
		Password: "correctpassword",
	})

	require.NoError(t, err)
	require.NotNil(t, loginResult)
	require.NotNil(t, loginResult.User)
	require.NotNil(t, loginResult.TokenPair)

	// Verify user matches
	assert.Equal(t, regResult.User.ID, loginResult.User.ID)

	// Verify tokens are valid
	claims, err := authService.ValidateAccessToken(loginResult.TokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, loginResult.User.ID, claims.UserID)
}

func TestAuth_LoginWithInvalidCredentials(t *testing.T) {
	// AC-AUTH-004: Login with Invalid Credentials
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Register user first
	_, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "invalidtest@test.local",
		Password: "correctpassword",
	})
	require.NoError(t, err)

	// Try login with wrong password
	_, err = authService.Login(ctx, service.LoginRequest{
		Email:    "invalidtest@test.local",
		Password: "wrongpassword",
	})

	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuth_LoginNonexistentUser(t *testing.T) {
	// AC-AUTH-004 (variation): Login with non-existent email
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	_, err := authService.Login(ctx, service.LoginRequest{
		Email:    "nonexistent@test.local",
		Password: "anypassword",
	})

	// Should return same error as wrong password to prevent user enumeration
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}

func TestAuth_RefreshToken(t *testing.T) {
	// AC-AUTH-005: Refresh Token
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Register and get initial tokens
	regResult, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "refreshtest@test.local",
		Password: "password123",
	})
	require.NoError(t, err)

	originalRefreshToken := regResult.TokenPair.RefreshToken

	// Refresh tokens
	newTokenPair, err := authService.RefreshTokens(ctx, originalRefreshToken)

	require.NoError(t, err)
	require.NotNil(t, newTokenPair)

	// New tokens should be different (rotation)
	assert.NotEqual(t, originalRefreshToken, newTokenPair.RefreshToken)
	assert.NotEmpty(t, newTokenPair.AccessToken)

	// New access token should be valid
	claims, err := authService.ValidateAccessToken(newTokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, regResult.User.ID, claims.UserID)

	// Old refresh token should be invalidated (single-use)
	_, err = authService.RefreshTokens(ctx, originalRefreshToken)
	require.Error(t, err)
}

func TestAuth_RefreshWithInvalidToken(t *testing.T) {
	// AC-AUTH-006: Refresh with Invalid Token
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Try to refresh with invalid token
	_, err := authService.RefreshTokens(ctx, "invalid-refresh-token")

	require.ErrorIs(t, err, service.ErrInvalidRefreshToken)
}

func TestAuth_LogoutRevokesTokens(t *testing.T) {
	// AC-AUTH-007: Logout Revokes Tokens
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Register and get tokens
	regResult, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "logouttest@test.local",
		Password: "password123",
	})
	require.NoError(t, err)

	refreshToken := regResult.TokenPair.RefreshToken

	// Verify refresh token works before logout
	_, err = authService.RefreshTokens(ctx, refreshToken)
	require.NoError(t, err)

	// Get new tokens after refresh (since we used the old one)
	loginResult, err := authService.Login(ctx, service.LoginRequest{
		Email:    "logouttest@test.local",
		Password: "password123",
	})
	require.NoError(t, err)
	refreshToken = loginResult.TokenPair.RefreshToken

	// Logout
	err = authService.Logout(ctx, regResult.User.ID)
	require.NoError(t, err)

	// Verify refresh token is now invalid
	_, err = authService.RefreshTokens(ctx, refreshToken)
	require.Error(t, err)
}

func TestAuth_OAuthIdentityLinking(t *testing.T) {
	// AC-AUTH-009: OAuth Identity Linking
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	oauthService := createOAuthService(t, tdb, authService)
	ctx := context.Background()

	// Create user with email/password
	regResult, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "linktest@test.local",
		Password: "password123",
	})
	require.NoError(t, err)

	// Link OAuth identity
	err = oauthService.LinkAccount(ctx, regResult.User.ID, service.ProviderGoogle, "google-user-123", "linktest@test.local")
	require.NoError(t, err)

	// Verify user still has identities (would need to check via authService.GetUserWithIdentities)
	userWithIdentities, err := authService.GetUserWithIdentities(ctx, regResult.User.ID)
	require.NoError(t, err)
	assert.Len(t, userWithIdentities.Identities, 1)
	assert.Equal(t, "google", userWithIdentities.Identities[0].Provider)
}

func TestAuth_CannotLinkAlreadyUsedOAuth(t *testing.T) {
	// AC-AUTH-010: Cannot Link Already-Used OAuth
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	oauthService := createOAuthService(t, tdb, authService)
	ctx := context.Background()

	// Create user A and link OAuth identity
	regResultA, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "usera@test.local",
		Password: "password123",
	})
	require.NoError(t, err)

	err = oauthService.LinkAccount(ctx, regResultA.User.ID, service.ProviderGoogle, "google-shared-id", "usera@test.local")
	require.NoError(t, err)

	// Create user B
	regResultB, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "userb@test.local",
		Password: "password456",
	})
	require.NoError(t, err)

	// Try to link same OAuth identity to user B
	err = oauthService.LinkAccount(ctx, regResultB.User.ID, service.ProviderGoogle, "google-shared-id", "usera@test.local")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already linked to another account")
}

func TestAuth_PasskeyRegistration(t *testing.T) {
	// AC-AUTH-011: Passkey Registration
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	passkeyService := createPasskeyService(t, tdb)
	ctx := context.Background()

	// Create user
	user := f.CreateUser(t)

	// Start registration
	startResp, err := passkeyService.StartRegistration(ctx, service.RegistrationStartRequest{
		UserID: user.ID,
	})

	require.NoError(t, err)
	require.NotNil(t, startResp)

	// Verify challenge was generated
	assert.NotEmpty(t, startResp.Challenge)
	assert.Equal(t, "saga.test", startResp.RP.ID)
	assert.NotEmpty(t, startResp.User.ID)
	assert.Equal(t, user.Email, startResp.User.Name)

	// Complete registration (simulated - normally this comes from browser)
	finishResult, err := passkeyService.FinishRegistration(ctx, service.RegistrationFinishRequest{
		UserID: user.ID,
		Name:   "My Passkey",
		Credential: &service.CredentialResponse{
			ID:    "test-credential-id",
			RawID: "test-credential-raw-id",
			Type:  "public-key",
			Response: service.AttestationResponse{
				ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiIn0=",
				AttestationObject: "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YQ==",
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, finishResult)
	assert.NotEmpty(t, finishResult.Passkey.ID)
	assert.Equal(t, "My Passkey", finishResult.Passkey.Name)
}

func TestAuth_PasskeyAuthentication(t *testing.T) {
	// AC-AUTH-012: Passkey Authentication
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	passkeyService := createPasskeyService(t, tdb)
	ctx := context.Background()

	// Create user and register passkey
	user := f.CreateUser(t)

	_, err := passkeyService.FinishRegistration(ctx, service.RegistrationFinishRequest{
		UserID: user.ID,
		Name:   "Test Passkey",
		Credential: &service.CredentialResponse{
			ID:    "auth-test-credential-id",
			RawID: "auth-test-credential-raw-id",
			Type:  "public-key",
			Response: service.AttestationResponse{
				ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiIn0=",
				AttestationObject: "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YQ==",
			},
		},
	})
	require.NoError(t, err)

	// Start login
	loginStartResp, err := passkeyService.StartLogin(ctx, service.LoginStartRequest{
		Email: user.Email,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginStartResp.Challenge)

	// Complete login (simulated)
	loginResult, err := passkeyService.FinishLogin(ctx, service.LoginFinishRequest{
		Credential: &service.AssertionResponse{
			ID:    "auth-test-credential-id",
			RawID: "auth-test-credential-raw-id",
			Type:  "public-key",
			Response: service.AuthenticatorAssertionResponse{
				ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiIn0=",
				AuthenticatorData: "SZYN5YgOjGh0NBcPZHZgW4/krrmihjLHmVzzuoMdl2MBAAAABQ==",
				Signature:         "MEYCIQDKMvqzR9bI...",
				UserHandle:        "",
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, loginResult)
	assert.Equal(t, user.ID, loginResult.User.ID)
	assert.NotEmpty(t, loginResult.TokenPair.AccessToken)
}

func TestAuth_PasskeyDeletion(t *testing.T) {
	// AC-AUTH-013: Passkey Deletion
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	passkeyService := createPasskeyService(t, tdb)
	ctx := context.Background()

	// Create user and register passkey
	user := f.CreateUser(t)

	regResult, err := passkeyService.FinishRegistration(ctx, service.RegistrationFinishRequest{
		UserID: user.ID,
		Name:   "Passkey To Delete",
		Credential: &service.CredentialResponse{
			ID:    "delete-test-credential-id",
			RawID: "delete-test-credential-raw-id",
			Type:  "public-key",
			Response: service.AttestationResponse{
				ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiIn0=",
				AttestationObject: "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YQ==",
			},
		},
	})
	require.NoError(t, err)

	passkeyID := regResult.Passkey.ID

	// Delete passkey
	err = passkeyService.DeletePasskey(ctx, user.ID, passkeyID)
	require.NoError(t, err)

	// Verify passkey no longer exists
	passkeys, err := passkeyService.GetUserPasskeys(ctx, user.ID)
	require.NoError(t, err)
	assert.Empty(t, passkeys)

	// Try to authenticate with deleted passkey
	_, err = passkeyService.FinishLogin(ctx, service.LoginFinishRequest{
		Credential: &service.AssertionResponse{
			ID:    "delete-test-credential-id",
			RawID: "delete-test-credential-raw-id",
			Type:  "public-key",
			Response: service.AuthenticatorAssertionResponse{
				ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiIn0=",
				AuthenticatorData: "SZYN5YgOjGh0NBcPZHZgW4/krrmihjLHmVzzuoMdl2MBAAAABQ==",
				Signature:         "MEYCIQDKMvqzR9bI...",
			},
		},
	})
	require.ErrorIs(t, err, service.ErrPasskeyNotFound)
}

func TestAuth_PasskeyCannotDeleteOthersPasskey(t *testing.T) {
	// AC-AUTH-013 (security): Cannot delete another user's passkey
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	passkeyService := createPasskeyService(t, tdb)
	ctx := context.Background()

	// Create two users
	userA := f.CreateUser(t)
	userB := f.CreateUser(t)

	// Register passkey for user A
	regResult, err := passkeyService.FinishRegistration(ctx, service.RegistrationFinishRequest{
		UserID: userA.ID,
		Name:   "User A's Passkey",
		Credential: &service.CredentialResponse{
			ID:    "usera-credential-id",
			RawID: "usera-credential-raw-id",
			Type:  "public-key",
			Response: service.AttestationResponse{
				ClientDataJSON:    "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiIn0=",
				AttestationObject: "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YQ==",
			},
		},
	})
	require.NoError(t, err)

	// User B tries to delete User A's passkey
	err = passkeyService.DeletePasskey(ctx, userB.ID, regResult.Passkey.ID)
	require.ErrorIs(t, err, service.ErrCredentialNotAllowed)
}

func TestAuth_EmailValidation(t *testing.T) {
	// AC-AUTH-001 (validation): Email must be valid format
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "missing @",
			email:   "testtest.local",
			wantErr: true,
		},
		{
			name:    "missing domain",
			email:   "test@",
			wantErr: true,
		},
		{
			name:    "missing local part",
			email:   "@test.local",
			wantErr: true,
		},
		{
			name:    "valid email",
			email:   "valid@test.local",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authService.Register(ctx, service.RegisterRequest{
				Email:    tt.email,
				Password: "password123",
			})

			if tt.wantErr {
				require.ErrorIs(t, err, service.ErrInvalidEmail)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuth_GetCurrentUser(t *testing.T) {
	// Verify GetUserWithIdentities returns correct user info
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	authService := createAuthService(t, tdb)
	ctx := context.Background()

	user := f.CreateUser(t, func(o *fixtures.UserOpts) {
		o.Email = "currentuser@test.local"
		o.Username = "currentuser"
	})

	userWithIdentities, err := authService.GetUserWithIdentities(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, userWithIdentities)

	assert.Equal(t, user.ID, userWithIdentities.User.ID)
	assert.Equal(t, "currentuser@test.local", userWithIdentities.User.Email)
	assert.Empty(t, userWithIdentities.Identities) // No OAuth linked
	assert.Empty(t, userWithIdentities.Passkeys)   // No passkeys registered
}

func TestAuth_ChangePassword(t *testing.T) {
	// Password change functionality
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Register user
	regResult, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "changepass@test.local",
		Password: "oldpassword123",
	})
	require.NoError(t, err)

	// Change password
	err = authService.ChangePassword(ctx, regResult.User.ID, "oldpassword123", "newpassword456")
	require.NoError(t, err)

	// Old password should no longer work
	_, err = authService.Login(ctx, service.LoginRequest{
		Email:    "changepass@test.local",
		Password: "oldpassword123",
	})
	require.ErrorIs(t, err, service.ErrInvalidCredentials)

	// New password should work
	loginResult, err := authService.Login(ctx, service.LoginRequest{
		Email:    "changepass@test.local",
		Password: "newpassword456",
	})
	require.NoError(t, err)
	assert.Equal(t, regResult.User.ID, loginResult.User.ID)
}

func TestAuth_ChangePasswordWrongOldPassword(t *testing.T) {
	// Cannot change password with wrong old password
	tdb := testdb.New(t)
	defer tdb.Close()

	authService := createAuthService(t, tdb)
	ctx := context.Background()

	// Register user
	regResult, err := authService.Register(ctx, service.RegisterRequest{
		Email:    "wrongold@test.local",
		Password: "correctoldpass",
	})
	require.NoError(t, err)

	// Try to change password with wrong old password
	err = authService.ChangePassword(ctx, regResult.User.ID, "wrongoldpass", "newpassword456")
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
}
