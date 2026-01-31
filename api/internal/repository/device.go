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

// DeviceTokenRepository handles device token data access
type DeviceTokenRepository struct {
	db database.Database
}

// NewDeviceTokenRepository creates a new device token repository
func NewDeviceTokenRepository(db database.Database) *DeviceTokenRepository {
	return &DeviceTokenRepository{db: db}
}

// Create creates a new device token
func (r *DeviceTokenRepository) Create(ctx context.Context, token *model.DeviceToken) error {
	query := `
		CREATE device_token CONTENT {
			user_id: $user_id,
			platform: $platform,
			token: $device_token,
			name: IF $name IS NOT NULL THEN $name ELSE NONE END,
			active: true,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"user_id":      token.UserID,
		"platform":     string(token.Platform),
		"device_token": token.Token,
		"name":         nilIfEmpty(token.Name),
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("%w: device token already registered", database.ErrDuplicate)
		}
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	token.ID = created.ID
	token.Active = true
	token.CreatedOn = created.CreatedOn
	token.UpdatedOn = created.UpdatedOn
	return nil
}

// GetByUserID retrieves all device tokens for a user
func (r *DeviceTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*model.DeviceToken, error) {
	query := `SELECT * FROM device_token WHERE user_id = $user_id AND active = true ORDER BY created_on DESC`
	vars := map[string]interface{}{"user_id": userID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseDeviceTokenResults(results)
}

// GetByToken retrieves a device token by its token value
func (r *DeviceTokenRepository) GetByToken(ctx context.Context, token string) (*model.DeviceToken, error) {
	query := `SELECT * FROM device_token WHERE token = $device_token LIMIT 1`
	vars := map[string]interface{}{"device_token": token}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseDeviceTokenResult(result)
}

// GetByID retrieves a device token by ID
func (r *DeviceTokenRepository) GetByID(ctx context.Context, id string) (*model.DeviceToken, error) {
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseDeviceTokenResult(result)
}

// Delete deletes a device token by ID
func (r *DeviceTokenRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	return r.db.Execute(ctx, query, map[string]interface{}{"id": id})
}

// DeleteByToken deletes a device token by its token value
func (r *DeviceTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE device_token WHERE token = $device_token`
	return r.db.Execute(ctx, query, map[string]interface{}{"device_token": token})
}

// MarkInactive marks a device token as inactive (for failed push attempts)
func (r *DeviceTokenRepository) MarkInactive(ctx context.Context, token string) error {
	query := `UPDATE device_token SET active = false, updated_on = time::now() WHERE token = $device_token`
	return r.db.Execute(ctx, query, map[string]interface{}{"device_token": token})
}

// UpdateLastUsed updates the last_used timestamp for a device token
func (r *DeviceTokenRepository) UpdateLastUsed(ctx context.Context, id string) error {
	query := `UPDATE type::record($id) SET last_used = time::now(), updated_on = time::now()`
	return r.db.Execute(ctx, query, map[string]interface{}{"id": id})
}

// CountByUserID counts active device tokens for a user
func (r *DeviceTokenRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT count() AS count FROM device_token WHERE user_id = $user_id AND active = true GROUP ALL`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return extractCount(result), nil
}

// UpsertByToken creates or updates a device token (for re-registration)
func (r *DeviceTokenRepository) UpsertByToken(ctx context.Context, token *model.DeviceToken) error {
	// Try to find existing token
	existing, err := r.GetByToken(ctx, token.Token)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing token
		query := `UPDATE type::record($id) SET
			user_id = $user_id,
			platform = $platform,
			name = IF $name IS NOT NULL THEN $name ELSE NONE END,
			active = true,
			updated_on = time::now()`
		vars := map[string]interface{}{
			"id":       existing.ID,
			"user_id":  token.UserID,
			"platform": string(token.Platform),
			"name":     nilIfEmpty(token.Name),
		}
		if err := r.db.Execute(ctx, query, vars); err != nil {
			return err
		}
		token.ID = existing.ID
		token.Active = true
		return nil
	}

	// Create new token
	return r.Create(ctx, token)
}

// parseDeviceTokenResult parses a single device token result
func (r *DeviceTokenRepository) parseDeviceTokenResult(result interface{}) (*model.DeviceToken, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	// Handle wrapped response from QueryOne
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

	// Handle record ID
	if id, ok := data["id"]; ok {
		data["id"] = convertDeviceTokenID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var token model.DeviceToken
	if err := json.Unmarshal(jsonBytes, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// parseDeviceTokenResults parses multiple device token results
func (r *DeviceTokenRepository) parseDeviceTokenResults(results []interface{}) ([]*model.DeviceToken, error) {
	tokens := make([]*model.DeviceToken, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						token, err := r.parseDeviceTokenFromData(item)
						if err == nil && token != nil {
							tokens = append(tokens, token)
						}
					}
				}
			}
		}
	}

	return tokens, nil
}

// parseDeviceTokenFromData parses a device token from map data
func (r *DeviceTokenRepository) parseDeviceTokenFromData(data interface{}) (*model.DeviceToken, error) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected data format")
	}

	if id, ok := m["id"]; ok {
		m["id"] = convertDeviceTokenID(id)
	}

	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var token model.DeviceToken
	if err := json.Unmarshal(jsonBytes, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// convertDeviceTokenID converts a SurrealDB ID to a string
func convertDeviceTokenID(id interface{}) string {
	if str, ok := id.(string); ok {
		return str
	}
	if rid, ok := id.(models.RecordID); ok {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	if rid, ok := id.(*models.RecordID); ok && rid != nil {
		return fmt.Sprintf("%s:%v", rid.Table, rid.ID)
	}
	if m, ok := id.(map[string]interface{}); ok {
		if tb, ok := m["tb"].(string); ok {
			if idVal, ok := m["id"]; ok {
				return fmt.Sprintf("%s:%v", tb, idVal)
			}
		}
		if tb, ok := m["Table"].(string); ok {
			if idVal, ok := m["ID"]; ok {
				return fmt.Sprintf("%s:%v", tb, idVal)
			}
		}
	}
	return fmt.Sprintf("%v", id)
}

