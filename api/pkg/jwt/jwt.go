package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"crypto"
	"crypto/rand"
	"crypto/sha256"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotYetValid = errors.New("token not yet valid")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidKey       = errors.New("invalid key")
)

// Claims represents JWT claims
type Claims struct {
	// Standard claims
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	JWTID     string `json:"jti,omitempty"`

	// Custom claims
	Email    string `json:"email,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Role     string `json:"role,omitempty"` // user, moderator, admin
}

// IsAdmin returns true if the claims indicate admin role
func (c *Claims) IsAdmin() bool {
	return c.Role == "admin"
}

// Valid checks if the claims are valid
func (c *Claims) Valid() error {
	now := time.Now().Unix()

	if c.ExpiresAt != 0 && now > c.ExpiresAt {
		return ErrTokenExpired
	}

	if c.NotBefore != 0 && now < c.NotBefore {
		return ErrTokenNotYetValid
	}

	return nil
}

// Service handles JWT operations
type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	expiration time.Duration
}

// Config holds JWT service configuration
type Config struct {
	PrivateKeyPath string
	PublicKeyPath  string
	Issuer         string
	ExpirationMins int
}

// NewService creates a new JWT service
func NewService(cfg Config) (*Service, error) {
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	var err error

	// Load private key if provided
	if cfg.PrivateKeyPath != "" {
		privateKey, err = loadPrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
		publicKey = &privateKey.PublicKey
	}

	// Load public key if provided (for validation-only scenarios)
	if cfg.PublicKeyPath != "" && publicKey == nil {
		publicKey, err = loadPublicKey(cfg.PublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load public key: %w", err)
		}
	}

	return &Service{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     cfg.Issuer,
		expiration: time.Duration(cfg.ExpirationMins) * time.Minute,
	}, nil
}

// GenerateKeyPair generates a new RSA key pair and saves to files
func GenerateKeyPair(privateKeyPath, publicKeyPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	// Save private key
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err := os.WriteFile(privateKeyPath, privateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	// Save public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	if err := os.WriteFile(publicKeyPath, publicKeyPEM, 0644); err != nil {
		return fmt.Errorf("failed to write public key: %w", err)
	}

	return nil
}

// Sign creates a signed JWT token
func (s *Service) Sign(claims Claims) (string, error) {
	if s.privateKey == nil {
		return "", ErrInvalidKey
	}

	now := time.Now()

	// Set standard claims
	claims.Issuer = s.issuer
	claims.IssuedAt = now.Unix()
	claims.NotBefore = now.Unix()
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = now.Add(s.expiration).Unix()
	}

	// Create header
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	}

	// Encode header and claims
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	headerB64 := base64URLEncode(headerJSON)
	claimsB64 := base64URLEncode(claimsJSON)

	// Create signature
	message := headerB64 + "." + claimsB64
	hash := sha256.Sum256([]byte(message))

	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	signatureB64 := base64URLEncode(signature)

	return message + "." + signatureB64, nil
}

// Validate validates a JWT token and returns the claims
func (s *Service) Validate(tokenString string) (*Claims, error) {
	if s.publicKey == nil {
		return nil, ErrInvalidKey
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerB64, claimsB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	message := headerB64 + "." + claimsB64
	hash := sha256.Sum256([]byte(message))

	signature, err := base64URLDecode(signatureB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if err := rsa.VerifyPKCS1v15(s.publicKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, ErrInvalidSignature
	}

	// Decode claims
	claimsJSON, err := base64URLDecode(claimsB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Validate claims
	if err := claims.Valid(); err != nil {
		return nil, err
	}

	// Verify issuer
	if claims.Issuer != s.issuer {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

// GetExpiration returns the token expiration duration
func (s *Service) GetExpiration() time.Duration {
	return s.expiration
}

// NewTestService creates a JWT service with in-memory keys for testing
// This should only be used in tests, not in production code
func NewTestService(privateKey *rsa.PrivateKey, issuer string, expiration time.Duration) *Service {
	return &Service{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		issuer:     issuer,
		expiration: expiration,
	}
}

// Helper functions

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}

func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64URLDecode(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
