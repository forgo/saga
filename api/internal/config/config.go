package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Push     PushConfig
	OAuth    OAuthConfig
	Passkey  PasskeyConfig
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port           string
	Env            string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	AllowedOrigins []string
}

// DatabaseConfig holds SurrealDB connection settings
type DatabaseConfig struct {
	Host      string
	Port      string
	Namespace string
	Database  string
	User      string
	Password  string
}

// JWTConfig holds JWT signing settings
type JWTConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	ExpirationMins int
	Issuer         string
}

// PushConfig holds push notification settings
type PushConfig struct {
	Enabled            bool
	FCMCredentialsPath string
}

// OAuthConfig holds OAuth provider settings
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

// AppleOAuthConfig holds Apple Sign In settings
type AppleOAuthConfig struct {
	ClientID    string
	TeamID      string
	KeyID       string
	PrivateKey  string
	RedirectURI string
}

// PasskeyConfig holds WebAuthn/Passkey settings
type PasskeyConfig struct {
	RPID            string
	RPName          string
	RPOrigins       []string
	Timeout         time.Duration
	RequireUV       bool
	AttestationType string
}

// Load reads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:           getEnv("SERVER_PORT", "8080"),
			Env:            getEnv("SERVER_ENV", "development"),
			ReadTimeout:    getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:   getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			AllowedOrigins: getSliceEnv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
		Database: DatabaseConfig{
			Host:      getEnv("DB_HOST", "localhost"),
			Port:      getEnv("DB_PORT", "8000"),
			Namespace: getEnv("DB_NAMESPACE", "saga"),
			Database:  getEnv("DB_DATABASE", "main"),
			User:      getEnv("DB_USER", "root"),
			Password:  getEnv("DB_PASSWORD", "root"),
		},
		JWT: JWTConfig{
			PrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
			PublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
			ExpirationMins: getIntEnv("JWT_EXPIRATION_MINS", 15),
			Issuer:         getEnv("JWT_ISSUER", "saga.forgo.software"),
		},
		Push: PushConfig{
			Enabled:            getBoolEnv("PUSH_ENABLED", false),
			FCMCredentialsPath: getEnv("FCM_CREDENTIALS_PATH", ""),
		},
		OAuth: OAuthConfig{
			Google: GoogleOAuthConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURI:  getEnv("GOOGLE_REDIRECT_URI", ""),
			},
			Apple: AppleOAuthConfig{
				ClientID:    getEnv("APPLE_CLIENT_ID", ""),
				TeamID:      getEnv("APPLE_TEAM_ID", ""),
				KeyID:       getEnv("APPLE_KEY_ID", ""),
				PrivateKey:  getEnv("APPLE_PRIVATE_KEY", ""),
				RedirectURI: getEnv("APPLE_REDIRECT_URI", ""),
			},
		},
		Passkey: PasskeyConfig{
			RPID:            getEnv("PASSKEY_RP_ID", "localhost"),
			RPName:          getEnv("PASSKEY_RP_NAME", "Saga"),
			RPOrigins:       getSliceEnv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			Timeout:         getDurationEnv("PASSKEY_TIMEOUT", 60*time.Second),
			RequireUV:       getBoolEnv("PASSKEY_REQUIRE_UV", false),
			AttestationType: getEnv("PASSKEY_ATTESTATION_TYPE", "none"),
		},
	}, nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// Validate checks that all required configuration values are present and valid.
// It returns an error describing all validation failures, or nil if valid.
func (c *Config) Validate() error {
	var errs []error

	// Server validation
	if c.Server.Port == "" {
		errs = append(errs, errors.New("SERVER_PORT is required"))
	}
	if c.Server.Env != "development" && c.Server.Env != "production" && c.Server.Env != "test" {
		errs = append(errs, fmt.Errorf("SERVER_ENV must be 'development', 'production', or 'test', got '%s'", c.Server.Env))
	}
	if len(c.Server.AllowedOrigins) == 0 {
		errs = append(errs, errors.New("CORS_ALLOWED_ORIGINS must have at least one origin"))
	}

	// Database validation
	if c.Database.Host == "" {
		errs = append(errs, errors.New("DB_HOST is required"))
	}
	if c.Database.Port == "" {
		errs = append(errs, errors.New("DB_PORT is required"))
	}
	if c.Database.Namespace == "" {
		errs = append(errs, errors.New("DB_NAMESPACE is required"))
	}
	if c.Database.Database == "" {
		errs = append(errs, errors.New("DB_DATABASE is required"))
	}

	// JWT validation - critical for production
	if c.IsProduction() {
		if c.JWT.PrivateKeyPath == "" {
			errs = append(errs, errors.New("JWT_PRIVATE_KEY_PATH is required in production"))
		}
		if c.JWT.PublicKeyPath == "" {
			errs = append(errs, errors.New("JWT_PUBLIC_KEY_PATH is required in production"))
		}
	}
	if c.JWT.ExpirationMins <= 0 {
		errs = append(errs, errors.New("JWT_EXPIRATION_MINS must be positive"))
	}

	// Push notification validation
	if c.Push.Enabled && c.Push.FCMCredentialsPath == "" {
		errs = append(errs, errors.New("FCM_CREDENTIALS_PATH is required when PUSH_ENABLED is true"))
	}

	// OAuth validation - if any provider field is set, validate required fields
	if c.OAuth.Google.IsConfigured() {
		if err := c.OAuth.Google.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("Google OAuth: %w", err))
		}
	}
	if c.OAuth.Apple.IsConfigured() {
		if err := c.OAuth.Apple.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("Apple OAuth: %w", err))
		}
	}

	// Passkey validation
	if c.Passkey.RPID == "" {
		errs = append(errs, errors.New("PASSKEY_RP_ID is required"))
	}
	if c.Passkey.RPName == "" {
		errs = append(errs, errors.New("PASSKEY_RP_NAME is required"))
	}
	if len(c.Passkey.RPOrigins) == 0 {
		errs = append(errs, errors.New("PASSKEY_RP_ORIGINS must have at least one origin"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// IsConfigured returns true if any Google OAuth field is set
func (g GoogleOAuthConfig) IsConfigured() bool {
	return g.ClientID != "" || g.ClientSecret != "" || g.RedirectURI != ""
}

// Validate checks that all required Google OAuth fields are present
func (g GoogleOAuthConfig) Validate() error {
	var missing []string
	if g.ClientID == "" {
		missing = append(missing, "GOOGLE_CLIENT_ID")
	}
	if g.ClientSecret == "" {
		missing = append(missing, "GOOGLE_CLIENT_SECRET")
	}
	if g.RedirectURI == "" {
		missing = append(missing, "GOOGLE_REDIRECT_URI")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// IsConfigured returns true if any Apple OAuth field is set
func (a AppleOAuthConfig) IsConfigured() bool {
	return a.ClientID != "" || a.TeamID != "" || a.KeyID != "" || a.PrivateKey != ""
}

// Validate checks that all required Apple OAuth fields are present
func (a AppleOAuthConfig) Validate() error {
	var missing []string
	if a.ClientID == "" {
		missing = append(missing, "APPLE_CLIENT_ID")
	}
	if a.TeamID == "" {
		missing = append(missing, "APPLE_TEAM_ID")
	}
	if a.KeyID == "" {
		missing = append(missing, "APPLE_KEY_ID")
	}
	if a.PrivateKey == "" {
		missing = append(missing, "APPLE_PRIVATE_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// Helper functions for reading environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
