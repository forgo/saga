// Package repository implements the data access layer for the Saga API.
//
// The repository package contains all database operations using SurrealDB.
// Each repository struct handles CRUD operations for a specific domain entity.
//
// # Repository Pattern
//
// All repositories follow a consistent pattern:
//
//   - Constructor function (NewXxxRepository) accepts a database connection
//   - Methods implement specific data operations (Create, GetByID, Update, Delete, etc.)
//   - SurrealQL queries are used for all database interactions
//   - Results are parsed and mapped to model structs
//
// # Database Connection
//
// Repositories accept a database.Database interface, allowing:
//
//   - Connection pooling and management at a higher level
//   - Transaction support when needed
//   - Easy testing with mock implementations
//
// # Query Patterns
//
// Common query patterns used:
//
//   - Parameterized queries with $variable syntax for security
//   - RELATE statements for graph relationships
//   - type::record() for safe ID handling
//   - time::now() for automatic timestamps
//
// # Example Usage
//
//	repo := NewGuildRepository(db)
//	guild, err := repo.GetByID(ctx, "guild:abc123")
//	if err != nil {
//	    if errors.Is(err, database.ErrNotFound) {
//	        // Handle not found
//	    }
//	    return err
//	}
package repository
