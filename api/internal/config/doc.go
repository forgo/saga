// Package config manages application configuration for the Saga API.
//
// The config package loads and validates configuration from environment variables.
// All configuration is centralized here to provide a single source of truth.
//
// # Configuration Loading
//
// Configuration is loaded from environment variables:
//
//	cfg := config.Load()
//
// # Configuration Groups
//
// Configuration is organized into logical groups:
//
//   - ServerConfig: HTTP server settings (port, timeouts)
//   - DatabaseConfig: SurrealDB connection settings
//   - JWTConfig: JWT signing and validation settings
//   - PushConfig: Push notification settings
//
// # Environment Variables
//
// Key environment variables:
//
//	PORT              - HTTP server port (default: 8080)
//	DATABASE_URL      - SurrealDB connection URL
//	DATABASE_USER     - Database username
//	DATABASE_PASS     - Database password
//	DATABASE_NS       - Database namespace
//	DATABASE_DB       - Database name
//	JWT_SECRET        - JWT signing secret
//	JWT_EXPIRATION    - Token expiration duration
//
// # Default Values
//
// Sensible defaults are provided for development:
//
//	func getEnv(key, defaultValue string) string {
//	    if value := os.Getenv(key); value != "" {
//	        return value
//	    }
//	    return defaultValue
//	}
package config
