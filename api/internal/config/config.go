package config

import (
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
