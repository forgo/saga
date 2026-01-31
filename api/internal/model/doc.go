// Package model defines domain entities and data structures for the Saga API.
//
// The model package contains all struct definitions for domain objects, request/response
// types, and error definitions. Models are used across all layers of the application.
//
// # Domain Entities
//
// Core domain entities include:
//
//   - User: Application user with authentication credentials
//   - Guild: Community group with members and shared resources
//   - Member: Guild membership linking users to guilds with roles
//   - Event: Scheduled activities within guilds
//   - Adventure: Special event type with admission management
//
// # JSON Serialization
//
// All models use json struct tags for API serialization:
//
//	type Guild struct {
//	    ID          string `json:"id"`
//	    Name        string `json:"name"`
//	    Description string `json:"description,omitempty"`
//	}
//
// # Validation Constants
//
// The package defines validation constants:
//
//	const (
//	    MaxGuildNameLength = 100
//	    MaxGuildsPerUser   = 10
//	    MaxMembersPerGuild = 1000
//	)
//
// # Error Types
//
// RFC 9457 Problem Details errors are defined in errors.go:
//
//	type ProblemDetails struct {
//	    Type    string    `json:"type"`
//	    Title   string    `json:"title"`
//	    Status  int       `json:"status"`
//	    Detail  string    `json:"detail"`
//	}
package model
