package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// isUniqueConstraintError checks if an error is a unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "unique") ||
		strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "already exists")
}

// extractRecordID extracts record ID from SurrealDB result
func extractRecordID(id interface{}) string {
	switch v := id.(type) {
	case string:
		return v
	case models.RecordID:
		return v.String()
	case *models.RecordID:
		if v != nil {
			return v.String()
		}
	case map[string]interface{}:
		// Handle {"tb": "table", "id": "xxx"} format
		if tb, ok := v["tb"].(string); ok {
			if id, ok := v["id"].(string); ok {
				return tb + ":" + id
			}
		}
	}

	// Try JSON marshaling as fallback
	if data, err := json.Marshal(id); err == nil {
		var recordID models.RecordID
		if err := json.Unmarshal(data, &recordID); err == nil {
			return recordID.String()
		}
	}

	return ""
}

// parseTime parses time from various formats
func parseTime(v interface{}) time.Time {
	switch t := v.(type) {
	case time.Time:
		return t
	case string:
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			return parsed
		}
		if parsed, err := time.Parse(time.RFC3339Nano, t); err == nil {
			return parsed
		}
	case models.CustomDateTime:
		return t.Time
	case *models.CustomDateTime:
		if t != nil {
			return t.Time
		}
	}
	return time.Time{}
}

// extractQueryResults extracts query results array from SurrealDB response
func extractQueryResults(result interface{}) ([]interface{}, bool) {
	// Handle SurrealDB response format
	if results, ok := result.([]interface{}); ok {
		if len(results) > 0 {
			if firstResult, ok := results[0].(map[string]interface{}); ok {
				if resultArray, ok := firstResult["result"].([]interface{}); ok {
					return resultArray, true
				}
			}
			// Direct array format
			return results, true
		}
	}
	return nil, false
}

// WithTransaction executes a function within a transaction context
// If the function returns an error, the transaction is rolled back
func WithTransaction(ctx context.Context, db database.Database, fn func(tx database.Transaction) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// BatchExecute executes multiple queries atomically using AtomicBatch
func BatchExecute(ctx context.Context, db database.Database, queries []struct {
	Query string
	Vars  map[string]interface{}
}) error {
	batch := database.NewAtomicBatch()
	for _, q := range queries {
		batch.Add(q.Query, q.Vars)
	}
	return batch.Execute(ctx, db)
}

// extractCount extracts count from SurrealDB count query result
func extractCount(result interface{}) int {
	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok && len(resultData) > 0 {
				if data, ok := resultData[0].(map[string]interface{}); ok {
					return extractCountValue(data["count"])
				}
			}
		}
		// Direct access
		return extractCountValue(resp["count"])
	}
	return 0
}

// extractCountValue converts various numeric types to int
func extractCountValue(v interface{}) int {
	switch c := v.(type) {
	case float64:
		return int(c)
	case float32:
		return int(c)
	case int:
		return c
	case int64:
		return int(c)
	case uint64:
		return int(c)
	}
	return 0
}

// getString extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// getStringPtr extracts an optional string value from a map
func getStringPtr(m map[string]interface{}, key string) *string {
	if v, ok := m[key].(string); ok && v != "" {
		return &v
	}
	return nil
}

// getInt extracts an int value from a map
func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(float32); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(int64); ok {
		return int(v)
	}
	if v, ok := m[key].(uint64); ok {
		return int(v)
	}
	return 0
}

// getFloat extracts a float value from a map
func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(float32); ok {
		return float64(v)
	}
	return 0
}

// getBool extracts a bool value from a map
func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

// getTime extracts a time value from a map
func getTime(m map[string]interface{}, key string) *time.Time {
	if v, ok := m[key].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return &t
		}
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return &t
		}
	}
	if t, ok := m[key].(time.Time); ok {
		return &t
	}
	// Handle SurrealDB CustomDateTime type
	if dt, ok := m[key].(models.CustomDateTime); ok {
		t := dt.Time
		return &t
	}
	if dt, ok := m[key].(*models.CustomDateTime); ok && dt != nil {
		t := dt.Time
		return &t
	}
	return nil
}

// getStringSlice extracts a string slice from a map
func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// Note: convertSurrealID and extractCreatedRecord are defined in user.go with more comprehensive handling
