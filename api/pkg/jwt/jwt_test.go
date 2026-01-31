package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// Test Helpers
// ============================================================================

func newTestService(t *testing.T) *Service {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return NewTestService(privateKey, "test-issuer", 15*time.Minute)
}

func newTestServiceWithExpiration(t *testing.T, expiration time.Duration) *Service {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	return NewTestService(privateKey, "test-issuer", expiration)
}

// ============================================================================
// Claims.Valid() Tests
// ============================================================================

func TestClaims_Valid_NoExpiration_ReturnsNil(t *testing.T) {
	t.Parallel()
	claims := Claims{
		UserID: "user:123",
		Email:  "test@example.com",
	}

	err := claims.Valid()

	if err != nil {
		t.Errorf("expected no error for claims without expiration, got %v", err)
	}
}

func TestClaims_Valid_NotExpired_ReturnsNil(t *testing.T) {
	t.Parallel()
	claims := Claims{
		UserID:    "user:123",
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}

	err := claims.Valid()

	if err != nil {
		t.Errorf("expected no error for non-expired token, got %v", err)
	}
}

func TestClaims_Valid_Expired_ReturnsErrTokenExpired(t *testing.T) {
	t.Parallel()
	claims := Claims{
		UserID:    "user:123",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}

	err := claims.Valid()

	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestClaims_Valid_ExpiresAtBoundary_ReturnsExpired(t *testing.T) {
	t.Parallel()
	// Token that expired 1 second ago
	claims := Claims{
		UserID:    "user:123",
		ExpiresAt: time.Now().Add(-1 * time.Second).Unix(),
	}

	err := claims.Valid()

	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired for just-expired token, got %v", err)
	}
}

func TestClaims_Valid_NotYetValid_ReturnsErrTokenNotYetValid(t *testing.T) {
	t.Parallel()
	claims := Claims{
		UserID:    "user:123",
		NotBefore: time.Now().Add(1 * time.Hour).Unix(),
	}

	err := claims.Valid()

	if err != ErrTokenNotYetValid {
		t.Errorf("expected ErrTokenNotYetValid, got %v", err)
	}
}

func TestClaims_Valid_NotBeforeInPast_ReturnsNil(t *testing.T) {
	t.Parallel()
	claims := Claims{
		UserID:    "user:123",
		NotBefore: time.Now().Add(-1 * time.Hour).Unix(),
	}

	err := claims.Valid()

	if err != nil {
		t.Errorf("expected no error when NotBefore is in past, got %v", err)
	}
}

func TestClaims_Valid_ZeroNotBefore_ReturnsNil(t *testing.T) {
	t.Parallel()
	claims := Claims{
		UserID:    "user:123",
		NotBefore: 0,
	}

	err := claims.Valid()

	if err != nil {
		t.Errorf("expected no error when NotBefore is zero, got %v", err)
	}
}

// ============================================================================
// Service.Sign() Tests
// ============================================================================

func TestSign_ValidClaims_ReturnsToken(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID: "user:123",
		Email:  "test@example.com",
	}

	token, err := svc.Sign(claims)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Token should have 3 parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts in JWT, got %d", len(parts))
	}
}

func TestSign_NilPrivateKey_ReturnsErrInvalidKey(t *testing.T) {
	t.Parallel()
	svc := &Service{
		privateKey: nil,
		issuer:     "test",
		expiration: 15 * time.Minute,
	}
	claims := Claims{
		UserID: "user:123",
	}

	_, err := svc.Sign(claims)

	if err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestSign_SetsIssuer(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Validate and check issuer
	validatedClaims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if validatedClaims.Issuer != "test-issuer" {
		t.Errorf("expected issuer 'test-issuer', got %q", validatedClaims.Issuer)
	}
}

func TestSign_SetsIssuedAt(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	before := time.Now().Unix()

	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	after := time.Now().Unix()

	validatedClaims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if validatedClaims.IssuedAt < before || validatedClaims.IssuedAt > after {
		t.Errorf("IssuedAt %d not in expected range [%d, %d]", validatedClaims.IssuedAt, before, after)
	}
}

func TestSign_SetsDefaultExpiration(t *testing.T) {
	t.Parallel()
	svc := newTestServiceWithExpiration(t, 30*time.Minute)
	now := time.Now()

	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	validatedClaims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	expectedExpiry := now.Add(30 * time.Minute).Unix()
	// Allow 5 seconds tolerance
	if validatedClaims.ExpiresAt < expectedExpiry-5 || validatedClaims.ExpiresAt > expectedExpiry+5 {
		t.Errorf("ExpiresAt %d not near expected %d", validatedClaims.ExpiresAt, expectedExpiry)
	}
}

func TestSign_PreservesCustomExpiration(t *testing.T) {
	t.Parallel()
	svc := newTestServiceWithExpiration(t, 30*time.Minute)
	customExpiry := time.Now().Add(1 * time.Hour).Unix()

	claims := Claims{
		UserID:    "user:123",
		ExpiresAt: customExpiry,
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	validatedClaims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if validatedClaims.ExpiresAt != customExpiry {
		t.Errorf("expected custom expiry %d, got %d", customExpiry, validatedClaims.ExpiresAt)
	}
}

func TestSign_PreservesAllClaimsFields(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)

	claims := Claims{
		Subject:  "sub:123",
		Audience: "test-audience",
		JWTID:    "unique-jti",
		UserID:   "user:456",
		Email:    "user@example.com",
		Username: "testuser",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	validatedClaims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if validatedClaims.Subject != claims.Subject {
		t.Errorf("Subject mismatch: expected %q, got %q", claims.Subject, validatedClaims.Subject)
	}
	if validatedClaims.Audience != claims.Audience {
		t.Errorf("Audience mismatch: expected %q, got %q", claims.Audience, validatedClaims.Audience)
	}
	if validatedClaims.JWTID != claims.JWTID {
		t.Errorf("JWTID mismatch: expected %q, got %q", claims.JWTID, validatedClaims.JWTID)
	}
	if validatedClaims.UserID != claims.UserID {
		t.Errorf("UserID mismatch: expected %q, got %q", claims.UserID, validatedClaims.UserID)
	}
	if validatedClaims.Email != claims.Email {
		t.Errorf("Email mismatch: expected %q, got %q", claims.Email, validatedClaims.Email)
	}
	if validatedClaims.Username != claims.Username {
		t.Errorf("Username mismatch: expected %q, got %q", claims.Username, validatedClaims.Username)
	}
}

// ============================================================================
// Service.Validate() Tests
// ============================================================================

func TestValidate_ValidToken_ReturnsClaims(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID: "user:123",
		Email:  "test@example.com",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	validatedClaims, err := svc.Validate(token)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if validatedClaims.UserID != "user:123" {
		t.Errorf("expected UserID 'user:123', got %q", validatedClaims.UserID)
	}
	if validatedClaims.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got %q", validatedClaims.Email)
	}
}

func TestValidate_NilPublicKey_ReturnsErrInvalidKey(t *testing.T) {
	t.Parallel()
	svc := &Service{
		publicKey: nil,
		issuer:    "test",
	}

	_, err := svc.Validate("some.token.here")

	if err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}

func TestValidate_InvalidFormat_TwoParts_ReturnsErrInvalidToken(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)

	_, err := svc.Validate("only.twoparts")

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidate_InvalidFormat_OnePart_ReturnsErrInvalidToken(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)

	_, err := svc.Validate("onlyonepart")

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidate_InvalidFormat_FourParts_ReturnsErrInvalidToken(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)

	_, err := svc.Validate("one.two.three.four")

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidate_InvalidFormat_Empty_ReturnsErrInvalidToken(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)

	_, err := svc.Validate("")

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidate_InvalidSignature_ReturnsErrInvalidSignature(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Tamper with the signature (use valid base64 but wrong content)
	parts := strings.Split(token, ".")
	// Create a valid base64 signature that's different from the original
	wrongSig := base64URLEncode([]byte("this is not a valid signature but is valid base64"))
	tamperedToken := parts[0] + "." + parts[1] + "." + wrongSig

	_, err = svc.Validate(tamperedToken)

	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestValidate_TamperedClaims_ReturnsErrInvalidSignature(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Tamper with the claims
	parts := strings.Split(token, ".")
	tamperedClaims := base64.URLEncoding.EncodeToString([]byte(`{"user_id":"hacker","iss":"test-issuer"}`))
	tamperedToken := parts[0] + "." + tamperedClaims + "." + parts[2]

	_, err = svc.Validate(tamperedToken)

	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestValidate_ExpiredToken_ReturnsErrTokenExpired(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID:    "user:123",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	_, err = svc.Validate(token)

	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestValidate_TokenNotYetValid_ReturnsErrTokenNotYetValid(t *testing.T) {
	t.Parallel()
	// Note: Sign() always sets NotBefore to now, so we can't create a "not yet valid"
	// token via Sign(). The Claims.Valid() tests cover this case directly.
	// This test verifies that if we somehow had a token with future NotBefore,
	// the validation would fail.

	// Test Claims.Valid() directly with NotBefore in the future
	claims := Claims{
		UserID:    "user:123",
		NotBefore: time.Now().Add(1 * time.Hour).Unix(),
	}

	err := claims.Valid()

	if err != ErrTokenNotYetValid {
		t.Errorf("expected ErrTokenNotYetValid, got %v", err)
	}
}

func TestValidate_WrongIssuer_ReturnsErrInvalidToken(t *testing.T) {
	t.Parallel()
	// Create two services with different issuers
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	signingService := &Service{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		issuer:     "issuer-a",
		expiration: 15 * time.Minute,
	}

	validatingService := &Service{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		issuer:     "issuer-b",
		expiration: 15 * time.Minute,
	}

	claims := Claims{
		UserID: "user:123",
	}

	token, err := signingService.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	_, err = validatingService.Validate(token)

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for wrong issuer, got %v", err)
	}
}

func TestValidate_InvalidBase64Signature_ReturnsErrInvalidToken(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Replace signature with invalid base64
	parts := strings.Split(token, ".")
	invalidToken := parts[0] + "." + parts[1] + ".!!!invalid!!!"

	_, err = svc.Validate(invalidToken)

	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken for invalid base64, got %v", err)
	}
}

func TestValidate_InvalidBase64Claims_ReturnsError(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Replace claims with invalid base64
	parts := strings.Split(token, ".")
	invalidToken := parts[0] + ".!!!invalid!!!" + "." + parts[2]

	_, err = svc.Validate(invalidToken)

	// Should return an error (either ErrInvalidToken or ErrInvalidSignature depending
	// on whether base64 decode fails first or signature verification fails)
	if err == nil {
		t.Error("expected error for invalid base64 claims, got nil")
	}
}

func TestValidate_InvalidJSONInClaims_ReturnsErrInvalidToken(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)
	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Replace claims with valid base64 of invalid JSON
	parts := strings.Split(token, ".")
	invalidJSON := base64URLEncode([]byte("not valid json"))
	invalidToken := parts[0] + "." + invalidJSON + "." + parts[2]

	_, err = svc.Validate(invalidToken)

	// Will fail with invalid signature since claims were modified
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

// ============================================================================
// GetExpiration() Tests
// ============================================================================

func TestGetExpiration_ReturnsConfiguredDuration(t *testing.T) {
	t.Parallel()
	svc := newTestServiceWithExpiration(t, 45*time.Minute)

	exp := svc.GetExpiration()

	if exp != 45*time.Minute {
		t.Errorf("expected 45m, got %v", exp)
	}
}

// ============================================================================
// base64URLEncode/Decode Tests
// ============================================================================

func TestBase64URLEncode_NoPadding(t *testing.T) {
	t.Parallel()
	data := []byte("test")

	encoded := base64URLEncode(data)

	if strings.Contains(encoded, "=") {
		t.Error("encoded string should not contain padding")
	}
}

func TestBase64URLDecode_WithPadding(t *testing.T) {
	t.Parallel()
	// "test" encoded with padding
	encoded := "dGVzdA=="

	decoded, err := base64URLDecode(encoded)

	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if string(decoded) != "test" {
		t.Errorf("expected 'test', got %q", string(decoded))
	}
}

func TestBase64URLDecode_WithoutPadding(t *testing.T) {
	t.Parallel()
	// "test" encoded without padding
	encoded := "dGVzdA"

	decoded, err := base64URLDecode(encoded)

	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if string(decoded) != "test" {
		t.Errorf("expected 'test', got %q", string(decoded))
	}
}

func TestBase64URLDecode_SinglePadding(t *testing.T) {
	t.Parallel()
	// String that needs single padding
	encoded := "dGVzdHM" // "tests" without padding

	decoded, err := base64URLDecode(encoded)

	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if string(decoded) != "tests" {
		t.Errorf("expected 'tests', got %q", string(decoded))
	}
}

func TestBase64URLEncode_Decode_RoundTrip(t *testing.T) {
	t.Parallel()
	testCases := []string{
		"",
		"a",
		"ab",
		"abc",
		"abcd",
		"Hello, World!",
		"Special chars: +/=",
		string([]byte{0, 1, 2, 255, 254, 253}),
	}

	for _, tc := range testCases {
		encoded := base64URLEncode([]byte(tc))
		decoded, err := base64URLDecode(encoded)

		if err != nil {
			t.Errorf("failed to decode %q: %v", tc, err)
			continue
		}
		if string(decoded) != tc {
			t.Errorf("round-trip failed for %q: got %q", tc, string(decoded))
		}
	}
}

// ============================================================================
// Integration/Round-Trip Tests
// ============================================================================

func TestSignAndValidate_RoundTrip(t *testing.T) {
	t.Parallel()
	svc := newTestService(t)

	originalClaims := Claims{
		Subject:  "user:abc",
		UserID:   "user:123",
		Email:    "user@test.com",
		Username: "testuser",
	}

	token, err := svc.Sign(originalClaims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	validatedClaims, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if validatedClaims.Subject != originalClaims.Subject {
		t.Errorf("Subject: expected %q, got %q", originalClaims.Subject, validatedClaims.Subject)
	}
	if validatedClaims.UserID != originalClaims.UserID {
		t.Errorf("UserID: expected %q, got %q", originalClaims.UserID, validatedClaims.UserID)
	}
	if validatedClaims.Email != originalClaims.Email {
		t.Errorf("Email: expected %q, got %q", originalClaims.Email, validatedClaims.Email)
	}
	if validatedClaims.Username != originalClaims.Username {
		t.Errorf("Username: expected %q, got %q", originalClaims.Username, validatedClaims.Username)
	}
}

func TestValidate_DifferentKey_ReturnsErrInvalidSignature(t *testing.T) {
	t.Parallel()
	// Sign with one key
	svc1 := newTestService(t)
	claims := Claims{
		UserID: "user:123",
	}

	token, err := svc1.Sign(claims)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Validate with different key
	svc2 := newTestService(t)

	_, err = svc2.Validate(token)

	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature when validating with different key, got %v", err)
	}
}

// ============================================================================
// NewService Tests
// ============================================================================

func TestNewService_NoKeys_ReturnsService(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Issuer:         "test",
		ExpirationMins: 15,
	}

	svc, err := NewService(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected service, got nil")
	}
}

func TestNewService_WithPrivateKey_LoadsKey(t *testing.T) {
	t.Parallel()

	// Create temp keys
	tempDir := t.TempDir()
	privateKeyPath := tempDir + "/private.pem"
	publicKeyPath := tempDir + "/public.pem"

	if err := GenerateKeyPair(privateKeyPath, publicKeyPath); err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	cfg := Config{
		PrivateKeyPath: privateKeyPath,
		Issuer:         "test",
		ExpirationMins: 15,
	}

	svc, err := NewService(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.privateKey == nil {
		t.Error("expected private key to be loaded")
	}
	if svc.publicKey == nil {
		t.Error("expected public key to be derived from private key")
	}
}

func TestNewService_WithPublicKeyOnly_LoadsPublicKey(t *testing.T) {
	t.Parallel()

	// Create temp keys
	tempDir := t.TempDir()
	privateKeyPath := tempDir + "/private.pem"
	publicKeyPath := tempDir + "/public.pem"

	if err := GenerateKeyPair(privateKeyPath, publicKeyPath); err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	cfg := Config{
		PublicKeyPath:  publicKeyPath,
		Issuer:         "test",
		ExpirationMins: 15,
	}

	svc, err := NewService(cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.privateKey != nil {
		t.Error("expected private key to be nil")
	}
	if svc.publicKey == nil {
		t.Error("expected public key to be loaded")
	}
}

func TestNewService_PrivateKeyNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cfg := Config{
		PrivateKeyPath: "/nonexistent/path/to/key.pem",
		Issuer:         "test",
	}

	_, err := NewService(cfg)

	if err == nil {
		t.Error("expected error for nonexistent key file")
	}
}

func TestNewService_PublicKeyNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	cfg := Config{
		PublicKeyPath: "/nonexistent/path/to/key.pem",
		Issuer:        "test",
	}

	_, err := NewService(cfg)

	if err == nil {
		t.Error("expected error for nonexistent key file")
	}
}

func TestNewService_InvalidPrivateKeyPEM_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	invalidKeyPath := tempDir + "/invalid.pem"

	// Write invalid PEM content
	if err := writeFile(invalidKeyPath, []byte("not a valid PEM file")); err != nil {
		t.Fatalf("failed to write invalid key: %v", err)
	}

	cfg := Config{
		PrivateKeyPath: invalidKeyPath,
		Issuer:         "test",
	}

	_, err := NewService(cfg)

	if err == nil {
		t.Error("expected error for invalid PEM file")
	}
}

func TestNewService_InvalidPublicKeyPEM_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	invalidKeyPath := tempDir + "/invalid.pem"

	// Write invalid PEM content
	if err := writeFile(invalidKeyPath, []byte("not a valid PEM file")); err != nil {
		t.Fatalf("failed to write invalid key: %v", err)
	}

	cfg := Config{
		PublicKeyPath: invalidKeyPath,
		Issuer:        "test",
	}

	_, err := NewService(cfg)

	if err == nil {
		t.Error("expected error for invalid PEM file")
	}
}

func TestNewService_WrongKeyType_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wrongKeyPath := tempDir + "/wrong.pem"

	// Write a valid PEM but with wrong key type content (DSA instead of RSA)
	wrongPEM := `-----BEGIN PUBLIC KEY-----
MIIBtzCCASsGByqGSM44BAEwggEeAoGBAP1/U4EddRIpUt9KnC7s5Of2EbdSPO9E
Y7ksX5QTk8uP11Qn5EwDhS9k8KY6W4XPAE+8g+dDANpF9b2VfB+A/gFqMFpD5cXL
CQXGfYE7Zp4PbFHjOvW2A3bYPH1VRUZqVnKNmV8GZZM+wz8B8YtlVbC3Cxw8vTOW
Ih4B5hAT+nyhAhUAl2BQjxUjC8yykrmCouuEC/BYHPUCgYEA9+GghdabPd7LvKtc
NrhXuXmUr7v6OuqC+VdMCz0HgmdRWVeOutRZT+ZxBxCBgLRJFnEj6EwoFhO3zwky
jMim4TwWeotUfI0o4KOuHiuzpnWRbqN/C/ohNWLx+2J6ASQ7zKTxvqhRkImog9/h
WuWfBpKLZl6Ae1UlZAFMO/7PSSoDgYQAAoGAf6ThwULnz1Y0LiON+NV5oDCj/lIF
dvLMYSJxfVx9vGU8jT+d3OQQ+1M6x0L/u9FY3Y2wZ7HnJPB4x5y1vNhO9u2FbADF
LQMB1cFh7PEcChR9T0o+Zv9X8UYDkw5lEQA7y8TN6L2F5rR4J0Y7Iy6QAz6/E4u8
D2Y5CTLZ4T1B5nU=
-----END PUBLIC KEY-----`

	if err := writeFile(wrongKeyPath, []byte(wrongPEM)); err != nil {
		t.Fatalf("failed to write wrong key: %v", err)
	}

	cfg := Config{
		PublicKeyPath: wrongKeyPath,
		Issuer:        "test",
	}

	_, err := NewService(cfg)

	if err == nil {
		t.Error("expected error for non-RSA key")
	}
}

// ============================================================================
// GenerateKeyPair Tests
// ============================================================================

func TestGenerateKeyPair_CreatesValidKeys(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	privateKeyPath := tempDir + "/private.pem"
	publicKeyPath := tempDir + "/public.pem"

	err := GenerateKeyPair(privateKeyPath, publicKeyPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify keys can be loaded
	cfg := Config{
		PrivateKeyPath: privateKeyPath,
		Issuer:         "test",
		ExpirationMins: 15,
	}

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("failed to load generated keys: %v", err)
	}

	// Sign and validate to ensure keys work
	claims := Claims{UserID: "test"}
	token, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("failed to sign with generated key: %v", err)
	}

	_, err = svc.Validate(token)
	if err != nil {
		t.Fatalf("failed to validate with generated key: %v", err)
	}
}

func TestGenerateKeyPair_InvalidPrivatePath_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	err := GenerateKeyPair("/nonexistent/dir/private.pem", tempDir+"/public.pem")

	if err == nil {
		t.Error("expected error for invalid private key path")
	}
}

func TestGenerateKeyPair_InvalidPublicPath_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	err := GenerateKeyPair(tempDir+"/private.pem", "/nonexistent/dir/public.pem")

	if err == nil {
		t.Error("expected error for invalid public key path")
	}
}

// ============================================================================
// loadPrivateKey Tests
// ============================================================================

func TestLoadPrivateKey_ValidKey_ReturnsKey(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	privateKeyPath := tempDir + "/private.pem"
	publicKeyPath := tempDir + "/public.pem"

	if err := GenerateKeyPair(privateKeyPath, publicKeyPath); err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	key, err := loadPrivateKey(privateKeyPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Error("expected key, got nil")
	}
}

func TestLoadPrivateKey_FileNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := loadPrivateKey("/nonexistent/key.pem")

	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadPrivateKey_InvalidPEM_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	invalidPath := tempDir + "/invalid.pem"

	if err := writeFile(invalidPath, []byte("not valid PEM")); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := loadPrivateKey(invalidPath)

	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestLoadPrivateKey_InvalidKeyData_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	invalidPath := tempDir + "/invalid.pem"

	// Valid PEM structure but garbage key data
	invalidPEM := `-----BEGIN RSA PRIVATE KEY-----
bm90IGEgdmFsaWQga2V5
-----END RSA PRIVATE KEY-----`

	if err := writeFile(invalidPath, []byte(invalidPEM)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := loadPrivateKey(invalidPath)

	if err == nil {
		t.Error("expected error for invalid key data")
	}
}

// ============================================================================
// loadPublicKey Tests
// ============================================================================

func TestLoadPublicKey_ValidKey_ReturnsKey(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	privateKeyPath := tempDir + "/private.pem"
	publicKeyPath := tempDir + "/public.pem"

	if err := GenerateKeyPair(privateKeyPath, publicKeyPath); err != nil {
		t.Fatalf("failed to generate keys: %v", err)
	}

	key, err := loadPublicKey(publicKeyPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Error("expected key, got nil")
	}
}

func TestLoadPublicKey_FileNotFound_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := loadPublicKey("/nonexistent/key.pem")

	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadPublicKey_InvalidPEM_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	invalidPath := tempDir + "/invalid.pem"

	if err := writeFile(invalidPath, []byte("not valid PEM")); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := loadPublicKey(invalidPath)

	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestLoadPublicKey_InvalidKeyData_ReturnsError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	invalidPath := tempDir + "/invalid.pem"

	// Valid PEM structure but garbage key data
	invalidPEM := `-----BEGIN PUBLIC KEY-----
bm90IGEgdmFsaWQga2V5
-----END PUBLIC KEY-----`

	if err := writeFile(invalidPath, []byte(invalidPEM)); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := loadPublicKey(invalidPath)

	if err == nil {
		t.Error("expected error for invalid key data")
	}
}

// ============================================================================
// Test Utilities
// ============================================================================

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
