// Package database provides database connectivity for the Saga API.
//
// The database package abstracts SurrealDB operations and provides
// a consistent interface for data access across the application.
//
// # Database Interface
//
// The Database interface defines core operations:
//
//	type Database interface {
//	    Query(ctx context.Context, query string, vars map[string]interface{}) ([]interface{}, error)
//	    QueryOne(ctx context.Context, query string, vars map[string]interface{}) (interface{}, error)
//	    Execute(ctx context.Context, query string, vars map[string]interface{}) error
//	    Close() error
//	}
//
// # Connection Management
//
// Connect to SurrealDB:
//
//	db, err := database.Connect(database.Config{
//	    URL:       "ws://localhost:8000/rpc",
//	    Namespace: "saga",
//	    Database:  "production",
//	    Username:  "root",
//	    Password:  "secret",
//	})
//
// # Error Types
//
// Standard error types for data operations:
//
//   - ErrNotFound: Record does not exist
//   - ErrDuplicate: Unique constraint violation
//   - ErrConnection: Database connection failed
//
// # Query Helpers
//
// Helper functions for common query patterns:
//
//   - Query: Execute query returning multiple results
//   - QueryOne: Execute query expecting single result
//   - Execute: Execute query with no return value
package database
