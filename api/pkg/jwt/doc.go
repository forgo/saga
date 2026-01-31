// Package jwt provides JSON Web Token utilities for the Saga API.
//
// The jwt package handles token generation, validation, and claims
// extraction for authentication.
//
// # Token Generation
//
// Generate tokens for authenticated users:
//
//	service := jwt.NewService(jwt.Config{
//	    Secret:     []byte("secret-key"),
//	    Expiration: 24 * time.Hour,
//	    Issuer:     "saga-api",
//	})
//
//	token, err := service.GenerateToken(userID)
//
// # Token Validation
//
// Validate and extract claims:
//
//	claims, err := service.ValidateToken(tokenString)
//	if err != nil {
//	    // Invalid or expired token
//	}
//	userID := claims.Subject
//
// # Refresh Tokens
//
// Refresh tokens have longer expiration:
//
//	refreshToken, err := service.GenerateRefreshToken(userID)
//	newToken, err := service.RefreshToken(refreshToken)
//
// # Claims
//
// Standard JWT claims are supported:
//
//	type Claims struct {
//	    Subject   string    // User ID
//	    IssuedAt  time.Time // Token creation time
//	    ExpiresAt time.Time // Token expiration
//	    Issuer    string    // Token issuer
//	}
package jwt
