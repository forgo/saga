package database

import (
	"context"
	"fmt"

	"github.com/surrealdb/surrealdb.go"
)

// SurrealDB implements the Database interface for SurrealDB
type SurrealDB struct {
	db     *surrealdb.DB
	config Config
}

// NewSurrealDB creates a new SurrealDB instance
func NewSurrealDB(cfg Config) *SurrealDB {
	return &SurrealDB{
		config: cfg,
	}
}

// Connect establishes a connection to SurrealDB
func (s *SurrealDB) Connect(ctx context.Context) error {
	endpoint := fmt.Sprintf("ws://%s:%s", s.config.Host, s.config.Port)

	db, err := surrealdb.FromEndpointURLString(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnection, err)
	}

	// Sign in as root user
	_, err = db.SignIn(ctx, &surrealdb.Auth{
		Username: s.config.User,
		Password: s.config.Password,
	})
	if err != nil {
		_ = db.Close(ctx)
		return fmt.Errorf("%w: signin failed: %v", ErrConnection, err)
	}

	// Use namespace and database
	if err := db.Use(ctx, s.config.Namespace, s.config.Database); err != nil {
		_ = db.Close(ctx)
		return fmt.Errorf("%w: use failed: %v", ErrConnection, err)
	}

	s.db = db
	return nil
}

// Close closes the database connection
func (s *SurrealDB) Close() error {
	if s.db != nil {
		return s.db.Close(context.Background())
	}
	return nil
}

// Ping checks the database connection
func (s *SurrealDB) Ping(ctx context.Context) error {
	if s.db == nil {
		return ErrConnection
	}
	// Execute a simple query to verify connection
	_, err := s.db.Version(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnection, err)
	}
	return nil
}

// Query executes a query and returns results
func (s *SurrealDB) Query(ctx context.Context, query string, vars map[string]interface{}) ([]interface{}, error) {
	if s.db == nil {
		return nil, ErrConnection
	}

	results, err := surrealdb.Query[interface{}](ctx, s.db, query, vars)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQuery, err)
	}

	// Convert QueryResult to []interface{}
	if results == nil {
		return nil, nil
	}

	output := make([]interface{}, 0, len(*results))
	for _, r := range *results {
		if r.Status != "OK" {
			if r.Error != nil {
				return nil, fmt.Errorf("%w: %s", ErrQuery, r.Error.Message)
			}
			return nil, ErrQuery
		}
		output = append(output, map[string]interface{}{
			"status": r.Status,
			"result": r.Result,
		})
	}

	return output, nil
}

// QueryOne executes a query and returns a single result
func (s *SurrealDB) QueryOne(ctx context.Context, query string, vars map[string]interface{}) (interface{}, error) {
	results, err := s.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ErrNotFound
	}

	// Unwrap the response wrapper {status: "OK", result: [...]}
	first := results[0]
	if resp, ok := first.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok {
				if len(resultData) == 0 {
					return nil, ErrNotFound
				}
				// Return the first record from the result array
				return resultData[0], nil
			}
			// Result is not an array, return as-is (e.g., scalar values)
			return resp["result"], nil
		}
	}

	return first, nil
}

// Execute runs a query without returning results
func (s *SurrealDB) Execute(ctx context.Context, query string, vars map[string]interface{}) error {
	_, err := s.Query(ctx, query, vars)
	return err
}

// BeginTx starts a new transaction
func (s *SurrealDB) BeginTx(ctx context.Context) (Transaction, error) {
	if s.db == nil {
		return nil, ErrConnection
	}

	// SurrealDB transactions are handled via BEGIN TRANSACTION / COMMIT / CANCEL
	// We wrap this in a transaction object
	return &SurrealTransaction{
		db:      s.db,
		ctx:     ctx,
		queries: make([]txQuery, 0),
	}, nil
}

// SurrealTransaction implements Transaction for SurrealDB
type SurrealTransaction struct {
	db        *surrealdb.DB
	ctx       context.Context
	queries   []txQuery
	committed bool
}

type txQuery struct {
	query string
	vars  map[string]interface{}
}

func (t *SurrealTransaction) Query(ctx context.Context, query string, vars map[string]interface{}) ([]interface{}, error) {
	t.queries = append(t.queries, txQuery{query: query, vars: vars})
	// In SurrealDB, we batch execute on commit
	return nil, nil
}

func (t *SurrealTransaction) QueryOne(ctx context.Context, query string, vars map[string]interface{}) (interface{}, error) {
	t.queries = append(t.queries, txQuery{query: query, vars: vars})
	return nil, nil
}

func (t *SurrealTransaction) Execute(ctx context.Context, query string, vars map[string]interface{}) error {
	t.queries = append(t.queries, txQuery{query: query, vars: vars})
	return nil
}

func (t *SurrealTransaction) Commit() error {
	if t.committed {
		return nil
	}

	// Build transaction query
	txQueryStr := "BEGIN TRANSACTION;\n"
	for _, q := range t.queries {
		txQueryStr += q.query + ";\n"
	}
	txQueryStr += "COMMIT TRANSACTION;"

	// Merge all vars
	allVars := make(map[string]interface{})
	for _, q := range t.queries {
		for k, v := range q.vars {
			allVars[k] = v
		}
	}

	_, err := surrealdb.Query[interface{}](t.ctx, t.db, txQueryStr, allVars)
	if err != nil {
		return fmt.Errorf("%w: commit failed: %v", ErrQuery, err)
	}

	t.committed = true
	return nil
}

func (t *SurrealTransaction) Rollback() error {
	// Clear pending queries
	t.queries = nil
	return nil
}

// Helper function to unmarshal SurrealDB results
func UnmarshalResult[T any](result interface{}) (T, error) {
	var zero T

	// SurrealDB returns results wrapped in response objects
	if results, ok := result.([]interface{}); ok {
		if len(results) == 0 {
			return zero, ErrNotFound
		}
		// Try to get the first result
		result = results[0]
	}

	// Handle the surrealdb response wrapper
	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"]; ok {
				result = resultData
			}
		}
	}

	// Handle array of results
	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			return zero, ErrNotFound
		}
		result = arr[0]
	}

	// Now try to convert to the target type
	if typed, ok := result.(T); ok {
		return typed, nil
	}

	return zero, fmt.Errorf("failed to unmarshal result to type %T", zero)
}
