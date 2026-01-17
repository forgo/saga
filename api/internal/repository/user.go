package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// UserRepository handles user data access
type UserRepository struct {
	db database.Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db database.Database) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	// Default to user role if not specified
	role := user.Role
	if role == "" {
		role = model.UserRoleUser
	}

	query := `
		CREATE user CONTENT {
			email: $email,
			username: IF $username IS NOT NULL THEN $username ELSE NONE END,
			hash: IF $hash IS NOT NULL THEN $hash ELSE NONE END,
			firstname: IF $firstname IS NOT NULL THEN $firstname ELSE NONE END,
			lastname: IF $lastname IS NOT NULL THEN $lastname ELSE NONE END,
			role: $role,
			email_verified: $email_verified,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"email":          user.Email,
		"username":       ptrToNone(user.Username),
		"hash":           ptrToNone(user.Hash),
		"firstname":      ptrToNone(user.Firstname),
		"lastname":       ptrToNone(user.Lastname),
		"role":           role,
		"email_verified": user.EmailVerified,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("%w: email already exists", database.ErrDuplicate)
		}
		return err
	}

	// Extract created user ID
	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	user.ID = created.ID
	user.CreatedOn = created.CreatedOn
	user.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	user, err := parseUserResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT * FROM user WHERE email = $email LIMIT 1`
	vars := map[string]interface{}{"email": email}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	user, err := parseUserResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE type::record($id) SET
			email = $email,
			username = $username,
			firstname = $firstname,
			lastname = $lastname,
			email_verified = $email_verified,
			updated_on = time::now()
	`

	vars := map[string]interface{}{
		"id":             user.ID,
		"email":          user.Email,
		"username":       user.Username,
		"firstname":      user.Firstname,
		"lastname":       user.Lastname,
		"email_verified": user.EmailVerified,
	}

	return r.db.Execute(ctx, query, vars)
}

// UpdatePassword updates a user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, hash string) error {
	query := `UPDATE type::record($id) SET hash = $hash, updated_on = time::now()`
	vars := map[string]interface{}{
		"id":   userID,
		"hash": hash,
	}

	return r.db.Execute(ctx, query, vars)
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	vars := map[string]interface{}{"id": id}

	return r.db.Execute(ctx, query, vars)
}

// SetEmailVerified marks a user's email as verified
func (r *UserRepository) SetEmailVerified(ctx context.Context, userID string, verified bool) error {
	query := `UPDATE type::record($id) SET email_verified = $verified, updated_on = time::now()`
	vars := map[string]interface{}{
		"id":       userID,
		"verified": verified,
	}

	return r.db.Execute(ctx, query, vars)
}

// SetRole updates a user's role
func (r *UserRepository) SetRole(ctx context.Context, userID string, role model.UserRole) error {
	query := `UPDATE type::record($id) SET role = $role, updated_on = time::now()`
	vars := map[string]interface{}{
		"id":   userID,
		"role": role,
	}

	return r.db.Execute(ctx, query, vars)
}

// Helper functions

type createdRecord struct {
	ID        string
	CreatedOn time.Time
	UpdatedOn time.Time
}

func extractCreatedRecord(result []interface{}) (*createdRecord, error) {
	if len(result) == 0 {
		return nil, errors.New("no result returned")
	}

	// Navigate through SurrealDB response structure
	first := result[0]
	if resp, ok := first.(map[string]interface{}); ok {
		if resultData, ok := resp["result"].([]interface{}); ok && len(resultData) > 0 {
			first = resultData[0]
		}
	}

	data, ok := first.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	record := &createdRecord{}

	// Handle SurrealDB's complex ID format
	if id, ok := data["id"]; ok {
		record.ID = convertSurrealID(id)
	}
	if createdOn, ok := data["created_on"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdOn); err == nil {
			record.CreatedOn = t
		} else if t, err := time.Parse(time.RFC3339Nano, createdOn); err == nil {
			record.CreatedOn = t
		}
	} else if createdOn, ok := data["created_on"].(time.Time); ok {
		record.CreatedOn = createdOn
	} else if dt, ok := data["created_on"].(models.CustomDateTime); ok {
		record.CreatedOn = dt.Time
	} else if dt, ok := data["created_on"].(*models.CustomDateTime); ok && dt != nil {
		record.CreatedOn = dt.Time
	}
	if updatedOn, ok := data["updated_on"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedOn); err == nil {
			record.UpdatedOn = t
		} else if t, err := time.Parse(time.RFC3339Nano, updatedOn); err == nil {
			record.UpdatedOn = t
		}
	} else if updatedOn, ok := data["updated_on"].(time.Time); ok {
		record.UpdatedOn = updatedOn
	} else if dt, ok := data["updated_on"].(models.CustomDateTime); ok {
		record.UpdatedOn = dt.Time
	} else if dt, ok := data["updated_on"].(*models.CustomDateTime); ok && dt != nil {
		record.UpdatedOn = dt.Time
	}

	return record, nil
}

func parseUserResult(result interface{}) (*model.User, error) {
	// Handle nil result
	if result == nil {
		return nil, database.ErrNotFound
	}

	// Navigate through SurrealDB response structure
	if resp, ok := result.(map[string]interface{}); ok {
		if status, ok := resp["status"].(string); ok && status == "OK" {
			if resultData, ok := resp["result"].([]interface{}); ok {
				if len(resultData) == 0 {
					return nil, database.ErrNotFound
				}
				result = resultData[0]
			}
		}
	}

	// Handle array wrapper
	if arr, ok := result.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, database.ErrNotFound
		}
		result = arr[0]
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	// Handle SurrealDB's complex ID format (Thing type)
	// The Go client returns ID as an object, need to convert to string
	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}

	// Extract hash before JSON marshal/unmarshal (since User.Hash has json:"-")
	var hash *string
	if h, ok := data["hash"].(string); ok {
		hash = &h
	}

	// Convert to JSON and back to struct for proper parsing
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal(jsonBytes, &user); err != nil {
		return nil, err
	}

	// Set the hash field manually (skipped by json:"-")
	user.Hash = hash

	return &user, nil
}

// convertSurrealID converts a SurrealDB ID (which may be a complex object) to a string
func convertSurrealID(id interface{}) string {
	// Already a string
	if str, ok := id.(string); ok {
		return str
	}

	// Handle models.RecordID from SurrealDB Go client
	if rid, ok := id.(models.RecordID); ok {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	if rid, ok := id.(*models.RecordID); ok && rid != nil {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}

	// Handle map format: {"tb": "user", "id": {"String": "demo"}} or similar
	if m, ok := id.(map[string]interface{}); ok {
		tb := ""
		idPart := ""

		// Get table name
		if t, ok := m["tb"].(string); ok {
			tb = t
		} else if t, ok := m["TB"].(string); ok {
			tb = t
		} else if t, ok := m["Table"].(string); ok {
			tb = t
		}

		// Get ID part - could be nested
		if idVal, ok := m["id"]; ok {
			idPart = extractIDValue(idVal)
		} else if idVal, ok := m["ID"]; ok {
			idPart = extractIDValue(idVal)
		}

		if tb != "" && idPart != "" {
			return tb + ":" + idPart
		}
		if idPart != "" {
			return idPart
		}
	}

	// Fallback: use fmt.Sprintf
	return fmt.Sprintf("%v", id)
}

// extractIDValue extracts the ID value which may be nested
func extractIDValue(val interface{}) string {
	if str, ok := val.(string); ok {
		return str
	}
	if m, ok := val.(map[string]interface{}); ok {
		// Check for {"String": "value"} format
		if s, ok := m["String"].(string); ok {
			return s
		}
		// Check for other common formats
		if s, ok := m["string"].(string); ok {
			return s
		}
	}
	return fmt.Sprintf("%v", val)
}

// isUniqueConstraintError is defined in helpers.go

// ptrToNone converts a string pointer to either the string value or an empty string marker.
// When used with SurrealDB queries that check for NONE, this allows proper handling of optional fields.
func ptrToNone(s *string) interface{} {
	if s == nil {
		return nil // Will be checked with != NONE in query
	}
	return *s
}
