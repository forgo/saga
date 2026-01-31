package database

// Transaction Utilities for Saga
//
// This file provides multiple patterns for atomic database operations.
// Choose the right pattern based on your needs:
//
// # AtomicBatch (Recommended for most cases)
//
// Simple, fluent API for 2-5 statements that must succeed together:
//
//	batch := NewAtomicBatch()
//	batch.Add(query1, vars1)
//	batch.Add(query2, vars2)
//	batch.Execute(ctx, db)  // All or nothing
//
// # TxBuilder (For complex variable handling)
//
// Use when combining queries with potentially conflicting variable names.
// Variables are automatically namespaced ($email -> $1_email):
//
//	tb := NewTxBuilder()
//	tb.Add("CREATE user SET email = $email", vars1)  // $email -> $1_email
//	tb.Add("CREATE profile SET email = $email", vars2)  // $email -> $2_email
//	ExecuteTransaction(ctx, db, tb)
//
// # UnitOfWork (For service-layer transactions with cleanup)
//
// Use when you need custom rollback handlers for failed operations:
//
//	uow := NewUnitOfWork(db)
//	uow.AddWithRollback(query, vars, cleanupFunc)
//	uow.Commit(ctx)  // cleanupFunc called if commit fails
//
// # MultiStepOperation (For sequential workflows)
//
// Use for complex workflows where each step can fail and needs cleanup:
//
//	mso := NewMultiStepOperation(db)
//	mso.AddStep("step1", executeFunc, rollbackFunc)
//	mso.AddStep("step2", executeFunc, rollbackFunc)
//	mso.Execute(ctx)  // Rollbacks run in reverse order on failure
//
// IMPORTANT: All patterns are BATCH-BASED. Queries accumulate and execute
// together at commit time. There is no isolation between Add() calls.

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
)

// TxBuilder builds atomic transaction queries with automatic variable namespacing.
// This prevents variable name collisions when combining queries from different sources.
//
// Example: Two queries both using $email get namespaced to $1_email and $2_email.
type TxBuilder struct {
	statements []string
	vars       map[string]interface{}
	varCounter uint64
}

// NewTxBuilder creates a new transaction builder
func NewTxBuilder() *TxBuilder {
	return &TxBuilder{
		statements: make([]string, 0),
		vars:       make(map[string]interface{}),
	}
}

// Add adds a statement to the transaction, namespacing variables to avoid collisions
// Returns the namespaced variable map for reference
func (tb *TxBuilder) Add(query string, vars map[string]interface{}) map[string]string {
	// Create unique variable names to avoid collisions
	varMapping := make(map[string]string)
	newQuery := query

	for varName, varValue := range vars {
		counter := atomic.AddUint64(&tb.varCounter, 1)
		newVarName := fmt.Sprintf("v%d_%s", counter, varName)

		// Replace $varName with $newVarName in query
		newQuery = strings.ReplaceAll(newQuery, "$"+varName, "$"+newVarName)

		tb.vars[newVarName] = varValue
		varMapping[varName] = newVarName
	}

	tb.statements = append(tb.statements, newQuery)
	return varMapping
}

// AddRaw adds a raw statement without variable substitution
func (tb *TxBuilder) AddRaw(query string) {
	tb.statements = append(tb.statements, query)
}

// Build returns the complete transaction query and merged variables
func (tb *TxBuilder) Build() (string, map[string]interface{}) {
	if len(tb.statements) == 0 {
		return "", nil
	}

	// Wrap in transaction block
	var sb strings.Builder
	sb.WriteString("BEGIN TRANSACTION;\n")
	for _, stmt := range tb.statements {
		sb.WriteString(stmt)
		if !strings.HasSuffix(strings.TrimSpace(stmt), ";") {
			sb.WriteString(";")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("COMMIT TRANSACTION;")

	return sb.String(), tb.vars
}

// ExecuteTransaction executes a transaction built with TxBuilder
func ExecuteTransaction(ctx context.Context, db Database, tb *TxBuilder) ([]interface{}, error) {
	query, vars := tb.Build()
	if query == "" {
		return nil, nil
	}

	return db.Query(ctx, query, vars)
}

// UnitOfWork represents a set of operations that must succeed or fail together
// This provides a higher-level abstraction for service-layer transactions
type UnitOfWork struct {
	db       Database
	builder  *TxBuilder
	rollback func(ctx context.Context) error
}

// NewUnitOfWork creates a new unit of work
func NewUnitOfWork(db Database) *UnitOfWork {
	return &UnitOfWork{
		db:      db,
		builder: NewTxBuilder(),
	}
}

// Add adds a statement to the unit of work
func (uow *UnitOfWork) Add(query string, vars map[string]interface{}) {
	uow.builder.Add(query, vars)
}

// AddWithRollback adds a statement with a custom rollback handler
// The rollback function is called if Commit fails
func (uow *UnitOfWork) AddWithRollback(query string, vars map[string]interface{}, rollback func(ctx context.Context) error) {
	uow.builder.Add(query, vars)
	if rollback != nil {
		prevRollback := uow.rollback
		uow.rollback = func(ctx context.Context) error {
			if prevRollback != nil {
				if err := prevRollback(ctx); err != nil {
					// Log but continue with other rollbacks
					fmt.Printf("Rollback error: %v\n", err)
				}
			}
			return rollback(ctx)
		}
	}
}

// Commit executes all operations atomically
func (uow *UnitOfWork) Commit(ctx context.Context) error {
	_, err := ExecuteTransaction(ctx, uow.db, uow.builder)
	if err != nil {
		// Execute rollback handlers if any
		if uow.rollback != nil {
			_ = uow.rollback(ctx) // Rollback errors are logged internally
		}
		return err
	}
	return nil
}

// MultiStepOperation executes a series of operations with automatic rollback on failure
// Each step can return an error to stop execution and trigger rollback
type MultiStepOperation struct {
	db    Database
	steps []multiStep
}

type multiStep struct {
	name     string
	execute  func(ctx context.Context, db Database) error
	rollback func(ctx context.Context, db Database) error
}

// NewMultiStepOperation creates a new multi-step operation
func NewMultiStepOperation(db Database) *MultiStepOperation {
	return &MultiStepOperation{
		db:    db,
		steps: make([]multiStep, 0),
	}
}

// AddStep adds a step with optional rollback
func (mso *MultiStepOperation) AddStep(name string, execute func(ctx context.Context, db Database) error, rollback func(ctx context.Context, db Database) error) {
	mso.steps = append(mso.steps, multiStep{
		name:     name,
		execute:  execute,
		rollback: rollback,
	})
}

// Execute runs all steps, rolling back on failure
func (mso *MultiStepOperation) Execute(ctx context.Context) error {
	completedSteps := make([]int, 0, len(mso.steps))

	for i, step := range mso.steps {
		if err := step.execute(ctx, mso.db); err != nil {
			// Rollback in reverse order
			for j := len(completedSteps) - 1; j >= 0; j-- {
				stepIdx := completedSteps[j]
				if mso.steps[stepIdx].rollback != nil {
					if rbErr := mso.steps[stepIdx].rollback(ctx, mso.db); rbErr != nil {
						// Log rollback error but continue
						fmt.Printf("Rollback failed for step %s: %v\n", mso.steps[stepIdx].name, rbErr)
					}
				}
			}
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}
		completedSteps = append(completedSteps, i)
	}

	return nil
}

// AtomicBatch provides a simpler API for batch operations that should be atomic
type AtomicBatch struct {
	queries []batchQuery
}

type batchQuery struct {
	query string
	vars  map[string]interface{}
}

// NewAtomicBatch creates a new atomic batch
func NewAtomicBatch() *AtomicBatch {
	return &AtomicBatch{
		queries: make([]batchQuery, 0),
	}
}

// Add adds a query to the batch
func (ab *AtomicBatch) Add(query string, vars map[string]interface{}) *AtomicBatch {
	ab.queries = append(ab.queries, batchQuery{query: query, vars: vars})
	return ab
}

// Execute runs all queries as a single transaction
func (ab *AtomicBatch) Execute(ctx context.Context, db Database) error {
	if len(ab.queries) == 0 {
		return nil
	}

	tb := NewTxBuilder()
	for _, q := range ab.queries {
		tb.Add(q.query, q.vars)
	}

	_, err := ExecuteTransaction(ctx, db, tb)
	return err
}

// Len returns the number of queries in the batch
func (ab *AtomicBatch) Len() int {
	return len(ab.queries)
}
