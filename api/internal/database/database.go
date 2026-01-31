// Package database provides the database abstraction layer for Saga.
//
// This package defines the Database interface that abstracts SurrealDB operations,
// allowing for clean separation between business logic and data access.
//
// # Interface Design
//
// The Database interface provides three query methods:
//   - Query: Returns multiple results (for SELECT queries returning lists)
//   - QueryOne: Returns a single result (for SELECT by ID)
//   - Execute: No return value (for CREATE/UPDATE/DELETE mutations)
//
// # Transaction Support
//
// IMPORTANT: Transactions in this package are BATCH-BASED, not connection-level.
// When you call BeginTx(), queries are accumulated in memory until Commit() is called.
// At commit time, all queries are wrapped in BEGIN TRANSACTION / COMMIT TRANSACTION
// and executed atomically. This means:
//   - No isolation between Add() calls until Commit()
//   - Rollback() simply discards accumulated queries (nothing to undo)
//   - All queries succeed or fail together at commit time
//
// For most use cases, prefer AtomicBatch over BeginTx() for clarity.
// See transaction.go for advanced transaction utilities.
//
// # Error Handling
//
// Standard errors are defined for common failure cases:
//   - ErrNotFound: Record does not exist
//   - ErrDuplicate: Unique constraint violation
//   - ErrConnection: Database connection issues
//   - ErrQuery: Query execution failures
//
// Use errors.Is() to check error types:
//
//	if errors.Is(err, database.ErrNotFound) {
//	    // Handle missing record
//	}
//
// # Usage Example
//
//	db := database.NewSurrealDB(cfg)
//	db.Connect(ctx)
//	defer db.Close()
//
//	result, err := db.QueryOne(ctx, "SELECT * FROM user WHERE id = $id", map[string]interface{}{"id": userID})
package database

import (
	"context"
	"errors"
)

// Standard errors for database operations.
// Use errors.Is() to check these error types in calling code.
var (
	// ErrNotFound indicates the requested record does not exist.
	ErrNotFound = errors.New("record not found")

	// ErrDuplicate indicates a unique constraint violation (e.g., duplicate email).
	ErrDuplicate = errors.New("duplicate record")

	// ErrConnection indicates a failure to connect to or communicate with the database.
	ErrConnection = errors.New("database connection error")

	// ErrQuery indicates a query execution failure (syntax error, invalid reference, etc.).
	ErrQuery = errors.New("query error")

	// ErrLimitExceeded indicates a result set exceeded the maximum allowed size.
	ErrLimitExceeded = errors.New("limit exceeded")
)

// Database defines the interface for database operations
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// Query executes a query and returns results
	Query(ctx context.Context, query string, vars map[string]interface{}) ([]interface{}, error)

	// QueryOne executes a query and returns a single result
	QueryOne(ctx context.Context, query string, vars map[string]interface{}) (interface{}, error)

	// Execute runs a query without returning results (for mutations)
	Execute(ctx context.Context, query string, vars map[string]interface{}) error

	// Transaction support
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction represents a database transaction
type Transaction interface {
	Query(ctx context.Context, query string, vars map[string]interface{}) ([]interface{}, error)
	QueryOne(ctx context.Context, query string, vars map[string]interface{}) (interface{}, error)
	Execute(ctx context.Context, query string, vars map[string]interface{}) error
	Commit() error
	Rollback() error
}

// Config holds database configuration
type Config struct {
	Host      string
	Port      string
	User      string
	Password  string
	Namespace string
	Database  string
}
