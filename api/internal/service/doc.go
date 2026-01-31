// Package service implements the business logic layer for the Saga API.
//
// The service package contains all domain logic, validation rules, and
// orchestration of repository operations. Services are the primary
// abstraction between HTTP handlers and data access.
//
// # Service Pattern
//
// All services follow a consistent pattern:
//
//   - Constructor function (NewXxxService) accepts a config struct with repository dependencies
//   - Methods implement business operations with proper validation
//   - Errors are returned as sentinel errors or wrapped errors for context
//   - Context is passed through for cancellation and request-scoped values
//
// # Repository Interfaces
//
// Services define their own repository interfaces, allowing:
//
//   - Easy mocking for unit tests
//   - Decoupling from specific database implementations
//   - Clear contracts for data access requirements
//
// # Error Handling
//
// Services return domain-specific errors defined as package-level variables:
//
//	var (
//	    ErrGuildNotFound   = errors.New("guild not found")
//	    ErrNotGuildMember  = errors.New("not a member of this guild")
//	)
//
// # Example Usage
//
//	service := NewGuildService(GuildServiceConfig{
//	    GuildRepo:  guildRepository,
//	    MemberRepo: memberRepository,
//	    UserRepo:   userRepository,
//	})
//	guild, err := service.CreateGuild(ctx, userID, CreateGuildRequest{
//	    Name: "My Guild",
//	})
package service
