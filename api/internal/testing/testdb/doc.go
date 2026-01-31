// Package testdb provides test database utilities for the Saga API.
//
// The testdb package manages test database connections with automatic
// setup, migration, and cleanup.
//
// # Test Database Setup
//
// Create a test database for each test:
//
//	func TestSomething(t *testing.T) {
//	    tdb := testdb.New(t)
//	    defer tdb.Close()
//
//	    // Use tdb.DB for database operations
//	}
//
// # Migrations
//
// Migrations are automatically applied on setup:
//
//	tdb := testdb.New(t) // Applies all migrations
//
// # Isolation
//
// Each test gets an isolated database namespace:
//
//	func TestA(t *testing.T) {
//	    tdb := testdb.New(t) // namespace: test_a_123
//	}
//
//	func TestB(t *testing.T) {
//	    tdb := testdb.New(t) // namespace: test_b_456
//	}
//
// # Shared Database
//
// For subtests that share data:
//
//	tdb := testdb.NewShared(t)
//	t.Run("create", func(t *testing.T) { ... })
//	t.Run("read", func(t *testing.T) { ... })
//
// # Timeout Context
//
// Test databases include timeout contexts:
//
//	ctx := tdb.Context() // 30 second timeout
package testdb
