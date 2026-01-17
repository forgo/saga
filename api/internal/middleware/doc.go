// Package middleware provides HTTP middleware for the Saga API.
//
// The middleware package contains reusable middleware components for
// authentication, authorization, rate limiting, and request processing.
//
// # Available Middleware
//
// Core middleware components:
//
//   - AuthMiddleware: JWT token validation and user extraction
//   - RateLimitMiddleware: Request rate limiting per user/IP
//   - IdempotencyMiddleware: Idempotent request handling
//   - GuildAccessMiddleware: Guild membership verification
//
// # Authentication
//
// The auth middleware validates JWT tokens and extracts user information:
//
//	router.Use(authMiddleware.Authenticate)
//
// After authentication, handlers can access user info:
//
//	userID := middleware.GetUserID(r)
//
// # Rate Limiting
//
// Rate limiting protects against abuse:
//
//	router.Use(rateLimiter.Limit)
//
// Configurable limits per endpoint and user tier.
//
// # Context Values
//
// Middleware sets context values accessible via helper functions:
//
//   - GetUserID(r): Returns authenticated user ID
//   - GetGuildID(r): Returns guild ID from path
//   - GetRequestID(r): Returns unique request identifier
package middleware
