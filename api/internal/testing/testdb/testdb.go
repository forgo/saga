// Package testdb provides test database utilities for e2e testing.
//
// This package creates isolated SurrealDB test environments that run real
// queries against a real database instance, ensuring tests validate actual
// database behavior including triggers, constraints, and functions.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    tdb := testdb.New(t)
//	    defer tdb.Close()
//
//	    // Use tdb.DB for database operations
//	    result, err := tdb.DB.Query(ctx, "SELECT * FROM user", nil)
//	}
package testdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/database"
)

// TestDB provides an isolated database environment for testing.
// Each TestDB instance gets a unique namespace to ensure test isolation.
type TestDB struct {
	DB        database.Database
	Namespace string
	Database  string
	t         *testing.T
}

var (
	// migrationOnce ensures migrations are only loaded once
	migrationOnce sync.Once
	migrations    []string
	migrationErr  error

	// counterMu protects the namespace counter
	counterMu sync.Mutex
	counter   int64
)

// getTestConfig returns database config from environment or defaults
func getTestConfig() database.Config {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "8000"
	}

	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "root"
	}

	return database.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}
}

// uniqueNamespace generates a unique namespace for test isolation
func uniqueNamespace() string {
	counterMu.Lock()
	defer counterMu.Unlock()
	counter++
	return fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), counter)
}

// loadMigrations reads all migration files in order
func loadMigrations() ([]string, error) {
	migrationOnce.Do(func() {
		// Find migrations directory - try multiple paths for flexibility
		paths := []string{
			"migrations",
			"../migrations",
			"../../migrations",
			"../../../migrations",
			"../../../../migrations",
		}

		var migrationDir string
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				migrationDir = p
				break
			}
		}

		if migrationDir == "" {
			// Try from SAGA_ROOT if set
			if root := os.Getenv("SAGA_ROOT"); root != "" {
				migrationDir = filepath.Join(root, "api", "migrations")
			}
		}

		if migrationDir == "" {
			migrationErr = fmt.Errorf("could not find migrations directory")
			return
		}

		// Read migration files
		entries, err := os.ReadDir(migrationDir)
		if err != nil {
			migrationErr = fmt.Errorf("reading migrations dir: %w", err)
			return
		}

		// Filter and sort .surql files (excluding seed.surql)
		var files []string
		for _, e := range entries {
			name := e.Name()
			if strings.HasSuffix(name, ".surql") && name != "seed.surql" {
				files = append(files, name)
			}
		}
		sort.Strings(files)

		// Read each file
		for _, name := range files {
			content, err := os.ReadFile(filepath.Join(migrationDir, name))
			if err != nil {
				migrationErr = fmt.Errorf("reading %s: %w", name, err)
				return
			}
			migrations = append(migrations, string(content))
		}
	})

	return migrations, migrationErr
}

// New creates a new isolated test database with migrations applied.
// The database uses a unique namespace to ensure test isolation.
// Call Close() when done to clean up the namespace.
func New(t *testing.T) *TestDB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := getTestConfig()
	namespace := uniqueNamespace()
	dbName := "test"

	cfg.Namespace = namespace
	cfg.Database = dbName

	// Create database connection
	db := database.NewSurrealDB(cfg)
	if err := db.Connect(ctx); err != nil {
		t.Fatalf("testdb: failed to connect: %v", err)
	}

	tdb := &TestDB{
		DB:        db,
		Namespace: namespace,
		Database:  dbName,
		t:         t,
	}

	// Apply migrations
	migs, err := loadMigrations()
	if err != nil {
		db.Close()
		t.Fatalf("testdb: failed to load migrations: %v", err)
	}

	for i, mig := range migs {
		if err := db.Execute(ctx, mig, nil); err != nil {
			db.Close()
			t.Fatalf("testdb: migration %d failed: %v", i+1, err)
		}
	}

	return tdb
}

// Close cleans up the test database by removing the namespace.
func (tdb *TestDB) Close() {
	if tdb.DB == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Remove the test namespace to clean up
	query := fmt.Sprintf("REMOVE NAMESPACE %s", tdb.Namespace)
	_ = tdb.DB.Execute(ctx, query, nil) // Ignore errors on cleanup

	tdb.DB.Close()
}

// Reset clears all data from tables while preserving schema.
// This is faster than creating a new TestDB for tests that need fresh data.
func (tdb *TestDB) Reset(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all tables
	results, err := tdb.DB.Query(ctx, "INFO FOR DB", nil)
	if err != nil {
		t.Fatalf("testdb: failed to get db info: %v", err)
	}

	// Extract table names from INFO result and delete data
	// SurrealDB INFO returns structured data about tables
	if len(results) > 0 {
		if resp, ok := results[0].(map[string]interface{}); ok {
			if result, ok := resp["result"].(map[string]interface{}); ok {
				if tables, ok := result["tables"].(map[string]interface{}); ok {
					for tableName := range tables {
						deleteQuery := fmt.Sprintf("DELETE FROM %s", tableName)
						if err := tdb.DB.Execute(ctx, deleteQuery, nil); err != nil {
							t.Logf("testdb: warning - failed to clear table %s: %v", tableName, err)
						}
					}
				}
			}
		}
	}
}

// Ctx returns a context with a reasonable timeout for test operations.
// Note: The cancel function is intentionally not returned as tests should
// complete within the timeout and the context will be garbage collected.
func (tdb *TestDB) Ctx() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// Store cancel to prevent leak warning, but we don't need to call it
	// as test contexts should complete within timeout
	_ = cancel
	return ctx
}

// MustExec executes a query and fails the test on error.
func (tdb *TestDB) MustExec(query string, vars map[string]interface{}) {
	tdb.t.Helper()
	if err := tdb.DB.Execute(tdb.Ctx(), query, vars); err != nil {
		tdb.t.Fatalf("testdb: exec failed: %v\nQuery: %s", err, query)
	}
}

// MustQuery executes a query and returns results, failing the test on error.
func (tdb *TestDB) MustQuery(query string, vars map[string]interface{}) []interface{} {
	tdb.t.Helper()
	results, err := tdb.DB.Query(tdb.Ctx(), query, vars)
	if err != nil {
		tdb.t.Fatalf("testdb: query failed: %v\nQuery: %s", err, query)
	}
	return results
}

// Shared creates a TestDB that can be shared across subtests.
// It provides a SetupSubtest method for per-subtest isolation.
type Shared struct {
	*TestDB
}

// NewShared creates a shared test database for use across multiple subtests.
// Use this when migration overhead is significant and tests can share schema.
func NewShared(t *testing.T) *Shared {
	return &Shared{TestDB: New(t)}
}

// SetupSubtest resets the database and returns the TestDB for use in a subtest.
// Call this at the start of each t.Run() block.
func (s *Shared) SetupSubtest(t *testing.T) *TestDB {
	t.Helper()
	s.TestDB.t = t
	s.TestDB.Reset(t)
	return s.TestDB
}
