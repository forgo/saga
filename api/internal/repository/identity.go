package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// IdentityRepository handles identity data access
type IdentityRepository struct {
	db database.Database
}

// NewIdentityRepository creates a new identity repository
func NewIdentityRepository(db database.Database) *IdentityRepository {
	return &IdentityRepository{db: db}
}

// Create creates a new identity linked to a user
func (r *IdentityRepository) Create(ctx context.Context, identity *model.Identity) error {
	query := `
		CREATE identity CONTENT {
			user: type::record($user_id),
			provider: $provider,
			provider_user_id: $provider_user_id,
			provider_email: $provider_email,
			email_verified_by_provider: $email_verified,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"user_id":          identity.UserID,
		"provider":         identity.Provider,
		"provider_user_id": identity.ProviderUserID,
		"provider_email":   identity.ProviderEmail,
		"email_verified":   identity.EmailVerifiedByProvider,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("%w: identity already exists", database.ErrDuplicate)
		}
		return err
	}

	// Extract created identity ID
	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	identity.ID = created.ID
	identity.CreatedOn = created.CreatedOn
	return nil
}

// GetByProviderID retrieves an identity by provider and provider user ID
func (r *IdentityRepository) GetByProviderID(ctx context.Context, provider, providerUserID string) (*model.Identity, error) {
	query := `
		SELECT * FROM identity
		WHERE provider = $provider AND provider_user_id = $provider_user_id
		LIMIT 1
	`
	vars := map[string]interface{}{
		"provider":         provider,
		"provider_user_id": providerUserID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	identity, err := parseIdentityResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return identity, nil
}

// GetByUserID retrieves all identities for a user
func (r *IdentityRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Identity, error) {
	query := `SELECT * FROM identity WHERE user = type::record($user_id)`
	vars := map[string]interface{}{"user_id": userID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parseIdentitiesResult(results)
}

// GetByProviderEmail retrieves an identity by provider and email
func (r *IdentityRepository) GetByProviderEmail(ctx context.Context, provider, email string) (*model.Identity, error) {
	query := `
		SELECT * FROM identity
		WHERE provider = $provider AND provider_email = $email
		LIMIT 1
	`
	vars := map[string]interface{}{
		"provider": provider,
		"email":    email,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	identity, err := parseIdentityResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return identity, nil
}

// Delete deletes an identity
func (r *IdentityRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	vars := map[string]interface{}{"id": id}

	return r.db.Execute(ctx, query, vars)
}

// CountByUserID counts identities for a user
func (r *IdentityRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT count() AS count FROM identity WHERE user = type::record($user_id) GROUP ALL`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	// Parse count result
	if data, ok := result.(map[string]interface{}); ok {
		if count, ok := data["count"].(float64); ok {
			return int(count), nil
		}
	}

	return 0, nil
}

func parseIdentityResult(result interface{}) (*model.Identity, error) {
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

	// Handle SurrealDB 3 record ID format (convert object to string)
	if id, ok := data["id"]; ok {
		data["id"] = convertIdentityID(id)
	}
	if userID, ok := data["user"]; ok {
		data["user_id"] = convertIdentityID(userID)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var identity model.Identity
	if err := json.Unmarshal(jsonBytes, &identity); err != nil {
		return nil, err
	}

	return &identity, nil
}

// convertIdentityID converts a SurrealDB ID to a string
func convertIdentityID(id interface{}) string {
	if str, ok := id.(string); ok {
		return str
	}
	if m, ok := id.(map[string]interface{}); ok {
		// Handle models.RecordID format from SurrealDB Go client
		if tb, ok := m["tb"].(string); ok {
			if idVal, ok := m["id"]; ok {
				return fmt.Sprintf("%s:%v", tb, idVal)
			}
		}
		// Handle Table/ID format
		if tb, ok := m["Table"].(string); ok {
			if idVal, ok := m["ID"]; ok {
				return fmt.Sprintf("%s:%v", tb, idVal)
			}
		}
	}
	return fmt.Sprintf("%v", id)
}

func parseIdentitiesResult(results []interface{}) ([]*model.Identity, error) {
	identities := make([]*model.Identity, 0)

	for _, result := range results {
		// Navigate SurrealDB response
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						identity, err := parseIdentityResult(item)
						if err == nil {
							identities = append(identities, identity)
						}
					}
				}
			}
		}
	}

	return identities, nil
}
