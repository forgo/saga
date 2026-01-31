// Package handler provides HTTP request handlers for the Saga API.
//
// The handler package contains all HTTP endpoint implementations organized by domain.
// Each handler struct encapsulates the dependencies needed to serve requests for a
// specific feature area (authentication, guilds, events, etc.).
//
// # Handler Pattern
//
// All handlers follow a consistent pattern:
//
//   - Constructor function (NewXxxHandler) accepts a config struct with dependencies
//   - Methods handle specific HTTP endpoints
//   - Response helpers from response.go standardize output format
//   - Errors are mapped to RFC 9457 Problem Details responses
//
// # Response Format
//
// Handlers use standardized response functions:
//
//   - WriteData: Single resource with optional HATEOAS links
//   - WriteCollection: Paginated list of resources
//   - WriteJSON: Raw JSON response
//   - WriteError: RFC 9457 Problem Details error response
//
// # Authentication
//
// Most handlers require authentication via JWT tokens. The auth middleware
// extracts the user ID and makes it available via GetUserID(r).
//
// # Example Usage
//
//	handler := NewGuildHandler(GuildHandlerConfig{
//	    GuildService: guildService,
//	})
//	router.Get("/v1/guilds", handler.List)
//	router.Post("/v1/guilds", handler.Create)
package handler
