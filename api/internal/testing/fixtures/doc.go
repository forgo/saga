// Package fixtures provides test data factories for the Saga API.
//
// The fixtures package contains factory functions for creating test data
// with sensible defaults and optional customization.
//
// # Factory Pattern
//
// Create a factory with a database connection:
//
//	f := fixtures.New(testDB)
//
// # Creating Test Data
//
// Factory methods create domain entities:
//
//	user := f.CreateUser(t)                    // Default user
//	user := f.CreateUserWithEmail(t, "test@example.com")
//	guild := f.CreateGuild(t, user)           // Guild owned by user
//	f.AddMemberToGuild(t, otherUser, guild)   // Add member
//
// # Customization
//
// Use option functions for customization:
//
//	user := f.CreateUser(t, WithEmail("custom@example.com"))
//	guild := f.CreateGuild(t, user, WithVisibility("public"))
//
// # Random Data
//
// Unique identifiers are generated automatically:
//
//	user1 := f.CreateUser(t) // user_abc123
//	user2 := f.CreateUser(t) // user_def456
//
// # Cleanup
//
// Test data is cleaned up when the test database is closed.
package fixtures
