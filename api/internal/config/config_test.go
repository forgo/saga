package config

import (
	"strings"
	"testing"
	"time"
)

func TestConfig_Validate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:           "8080",
			Env:            "development",
			AllowedOrigins: []string{"http://localhost:3000"},
		},
		Database: DatabaseConfig{
			Host:      "localhost",
			Port:      "8000",
			Namespace: "saga",
			Database:  "main",
		},
		JWT: JWTConfig{
			PrivateKeyPath: "./keys/private.pem",
			PublicKeyPath:  "./keys/public.pem",
			ExpirationMins: 15,
			Issuer:         "saga.forgo.software",
		},
		Passkey: PasskeyConfig{
			RPID:      "localhost",
			RPName:    "Saga",
			RPOrigins: []string{"http://localhost:3000"},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestConfig_Validate_InvalidServerEnv(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Server.Env = "invalid"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for invalid SERVER_ENV")
	}
	if !strings.Contains(err.Error(), "SERVER_ENV") {
		t.Errorf("expected error to mention SERVER_ENV, got: %v", err)
	}
}

func TestConfig_Validate_MissingPort(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Server.Port = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing SERVER_PORT")
	}
	if !strings.Contains(err.Error(), "SERVER_PORT") {
		t.Errorf("expected error to mention SERVER_PORT, got: %v", err)
	}
}

func TestConfig_Validate_EmptyAllowedOrigins(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Server.AllowedOrigins = []string{}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for empty CORS_ALLOWED_ORIGINS")
	}
	if !strings.Contains(err.Error(), "CORS_ALLOWED_ORIGINS") {
		t.Errorf("expected error to mention CORS_ALLOWED_ORIGINS, got: %v", err)
	}
}

func TestConfig_Validate_MissingDatabaseHost(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Database.Host = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing DB_HOST")
	}
	if !strings.Contains(err.Error(), "DB_HOST") {
		t.Errorf("expected error to mention DB_HOST, got: %v", err)
	}
}

func TestConfig_Validate_InvalidJWTExpiration(t *testing.T) {
	cfg := validBaseConfig()
	cfg.JWT.ExpirationMins = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for zero JWT_EXPIRATION_MINS")
	}
	if !strings.Contains(err.Error(), "JWT_EXPIRATION_MINS") {
		t.Errorf("expected error to mention JWT_EXPIRATION_MINS, got: %v", err)
	}
}

func TestConfig_Validate_ProductionRequiresJWTKeys(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Server.Env = "production"
	cfg.JWT.PrivateKeyPath = ""
	cfg.JWT.PublicKeyPath = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing JWT keys in production")
	}
	if !strings.Contains(err.Error(), "JWT_PRIVATE_KEY_PATH") {
		t.Errorf("expected error to mention JWT_PRIVATE_KEY_PATH, got: %v", err)
	}
	if !strings.Contains(err.Error(), "JWT_PUBLIC_KEY_PATH") {
		t.Errorf("expected error to mention JWT_PUBLIC_KEY_PATH, got: %v", err)
	}
}

func TestConfig_Validate_PushEnabledRequiresFCMPath(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Push.Enabled = true
	cfg.Push.FCMCredentialsPath = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error when push enabled without FCM path")
	}
	if !strings.Contains(err.Error(), "FCM_CREDENTIALS_PATH") {
		t.Errorf("expected error to mention FCM_CREDENTIALS_PATH, got: %v", err)
	}
}

func TestConfig_Validate_PushDisabledNoFCMRequired(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Push.Enabled = false
	cfg.Push.FCMCredentialsPath = ""

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error when push disabled, got: %v", err)
	}
}

func TestConfig_Validate_MissingPasskeyRPID(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Passkey.RPID = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for missing PASSKEY_RP_ID")
	}
	if !strings.Contains(err.Error(), "PASSKEY_RP_ID") {
		t.Errorf("expected error to mention PASSKEY_RP_ID, got: %v", err)
	}
}

func TestGoogleOAuthConfig_Validate_Complete(t *testing.T) {
	cfg := GoogleOAuthConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURI:  "http://localhost/callback",
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid Google OAuth config, got: %v", err)
	}
}

func TestGoogleOAuthConfig_Validate_MissingFields(t *testing.T) {
	cfg := GoogleOAuthConfig{
		ClientID: "client-id",
		// Missing ClientSecret and RedirectURI
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for incomplete Google OAuth config")
	}
	if !strings.Contains(err.Error(), "GOOGLE_CLIENT_SECRET") {
		t.Errorf("expected error to mention GOOGLE_CLIENT_SECRET, got: %v", err)
	}
}

func TestGoogleOAuthConfig_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		cfg      GoogleOAuthConfig
		expected bool
	}{
		{"empty", GoogleOAuthConfig{}, false},
		{"client_id_only", GoogleOAuthConfig{ClientID: "id"}, true},
		{"full", GoogleOAuthConfig{ClientID: "id", ClientSecret: "secret", RedirectURI: "uri"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsConfigured(); got != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppleOAuthConfig_Validate_Complete(t *testing.T) {
	cfg := AppleOAuthConfig{
		ClientID:   "client-id",
		TeamID:     "team-id",
		KeyID:      "key-id",
		PrivateKey: "private-key",
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid Apple OAuth config, got: %v", err)
	}
}

func TestAppleOAuthConfig_Validate_MissingFields(t *testing.T) {
	cfg := AppleOAuthConfig{
		ClientID: "client-id",
		// Missing TeamID, KeyID, PrivateKey
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for incomplete Apple OAuth config")
	}
	if !strings.Contains(err.Error(), "APPLE_TEAM_ID") {
		t.Errorf("expected error to mention APPLE_TEAM_ID, got: %v", err)
	}
}

func TestAppleOAuthConfig_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		cfg      AppleOAuthConfig
		expected bool
	}{
		{"empty", AppleOAuthConfig{}, false},
		{"client_id_only", AppleOAuthConfig{ClientID: "id"}, true},
		{"full", AppleOAuthConfig{ClientID: "id", TeamID: "t", KeyID: "k", PrivateKey: "p"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsConfigured(); got != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_Validate_PartialGoogleOAuth(t *testing.T) {
	cfg := validBaseConfig()
	cfg.OAuth.Google.ClientID = "only-client-id"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for partial Google OAuth config")
	}
	if !strings.Contains(err.Error(), "Google OAuth") {
		t.Errorf("expected error to mention Google OAuth, got: %v", err)
	}
}

func TestConfig_Validate_PartialAppleOAuth(t *testing.T) {
	cfg := validBaseConfig()
	cfg.OAuth.Apple.ClientID = "only-client-id"

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for partial Apple OAuth config")
	}
	if !strings.Contains(err.Error(), "Apple OAuth") {
		t.Errorf("expected error to mention Apple OAuth, got: %v", err)
	}
}

func TestConfig_Validate_MultipleErrors(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:           "",
			Env:            "invalid",
			AllowedOrigins: []string{},
		},
		Database: DatabaseConfig{
			Host: "",
		},
		JWT: JWTConfig{
			ExpirationMins: 0,
		},
		Passkey: PasskeyConfig{
			RPID:      "",
			RPName:    "",
			RPOrigins: []string{},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected multiple validation errors")
	}

	errStr := err.Error()
	expectedFields := []string{"SERVER_PORT", "SERVER_ENV", "CORS_ALLOWED_ORIGINS", "DB_HOST", "JWT_EXPIRATION_MINS", "PASSKEY_RP_ID"}
	for _, field := range expectedFields {
		if !strings.Contains(errStr, field) {
			t.Errorf("expected error to mention %s, got: %v", field, err)
		}
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Env: "development"}}
	if !cfg.IsDevelopment() {
		t.Error("expected IsDevelopment() to return true")
	}

	cfg.Server.Env = "production"
	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment() to return false in production")
	}
}

func TestConfig_IsProduction(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Env: "production"}}
	if !cfg.IsProduction() {
		t.Error("expected IsProduction() to return true")
	}

	cfg.Server.Env = "development"
	if cfg.IsProduction() {
		t.Error("expected IsProduction() to return false in development")
	}
}

// validBaseConfig returns a minimal valid configuration for testing
func validBaseConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:           "8080",
			Env:            "development",
			ReadTimeout:    15 * time.Second,
			WriteTimeout:   15 * time.Second,
			AllowedOrigins: []string{"http://localhost:3000"},
		},
		Database: DatabaseConfig{
			Host:      "localhost",
			Port:      "8000",
			Namespace: "saga",
			Database:  "main",
			User:      "root",
			Password:  "root",
		},
		JWT: JWTConfig{
			PrivateKeyPath: "./keys/private.pem",
			PublicKeyPath:  "./keys/public.pem",
			ExpirationMins: 15,
			Issuer:         "saga.forgo.software",
		},
		Push: PushConfig{
			Enabled: false,
		},
		OAuth: OAuthConfig{},
		Passkey: PasskeyConfig{
			RPID:            "localhost",
			RPName:          "Saga",
			RPOrigins:       []string{"http://localhost:3000"},
			Timeout:         60 * time.Second,
			RequireUV:       false,
			AttestationType: "none",
		},
	}
}
