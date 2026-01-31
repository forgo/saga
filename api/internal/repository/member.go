package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// MemberRepository handles member data access
type MemberRepository struct {
	db database.Database
}

// NewMemberRepository creates a new member repository
func NewMemberRepository(db database.Database) *MemberRepository {
	return &MemberRepository{db: db}
}

// Create creates a new member linked to a user
func (r *MemberRepository) Create(ctx context.Context, member *model.Member) error {
	query := `
		CREATE member CONTENT {
			name: $name,
			email: $email,
			user: type::record($user_id),
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"name":    member.Name,
		"email":   member.Email,
		"user_id": member.UserID,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	member.ID = created.ID
	member.CreatedOn = created.CreatedOn
	member.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByID retrieves a member by ID
func (r *MemberRepository) GetByID(ctx context.Context, id string) (*model.Member, error) {
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return parseMemberResult(result)
}

// GetByUserID retrieves a member by user ID
func (r *MemberRepository) GetByUserID(ctx context.Context, userID string) (*model.Member, error) {
	query := `SELECT * FROM member WHERE user = type::record($user_id) LIMIT 1`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	member, err := parseMemberResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return member, nil
}

// GetOrCreate retrieves an existing member for a user, or creates one
func (r *MemberRepository) GetOrCreate(ctx context.Context, userID, name, email string) (*model.Member, error) {
	member, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if member != nil {
		return member, nil
	}

	// Create new member
	member = &model.Member{
		Name:   name,
		Email:  email,
		UserID: userID,
	}
	if err := r.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// Update updates a member
func (r *MemberRepository) Update(ctx context.Context, member *model.Member) error {
	query := `UPDATE type::record($id) SET name = $name, email = $email, updated_on = time::now()`
	vars := map[string]interface{}{
		"id":    member.ID,
		"name":  member.Name,
		"email": member.Email,
	}

	return r.db.Execute(ctx, query, vars)
}

// Delete deletes a member
func (r *MemberRepository) Delete(ctx context.Context, id string) error {
	// Remove from all guilds first
	memberQuery := `DELETE responsible_for WHERE in = type::record($id)`
	if err := r.db.Execute(ctx, memberQuery, map[string]interface{}{"id": id}); err != nil {
		return err
	}

	// Delete member
	query := `DELETE type::record($id)`
	return r.db.Execute(ctx, query, map[string]interface{}{"id": id})
}

// Helper functions

func parseMemberResult(result interface{}) (*model.Member, error) {
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

	// Handle SurrealDB's complex ID format
	if id, ok := data["id"]; ok {
		data["id"] = convertMemberID(id)
	}
	if userID, ok := data["user"]; ok {
		data["user"] = convertMemberID(userID)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var member model.Member
	if err := json.Unmarshal(jsonBytes, &member); err != nil {
		return nil, err
	}

	return &member, nil
}

// convertMemberID converts a SurrealDB ID to a string
func convertMemberID(id interface{}) string {
	if str, ok := id.(string); ok {
		return str
	}
	if rid, ok := id.(models.RecordID); ok {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	if rid, ok := id.(*models.RecordID); ok && rid != nil {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	return fmt.Sprintf("%v", id)
}
