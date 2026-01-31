package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/service"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// TokenRepository handles refresh token data access
type TokenRepository struct {
	db database.Database
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db database.Database) *TokenRepository {
	return &TokenRepository{db: db}
}

// CreateRefreshToken stores a new refresh token
func (r *TokenRepository) CreateRefreshToken(ctx context.Context, token *service.RefreshToken) error {
	query := `
		CREATE refresh_token CONTENT {
			user: type::record($user),
			token_hash: $token_hash,
			expires_at: <datetime>$expires_at,
			created_at: time::now(),
			revoked: false
		}
	`

	vars := map[string]interface{}{
		"user":       token.UserID, // UserID is in format "user:xxx"
		"token_hash": token.TokenHash,
		"expires_at": token.ExpiresAt.Format(time.RFC3339),
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	token.ID = created.ID
	token.CreatedAt = created.CreatedOn
	return nil
}

// GetRefreshTokenByHash retrieves a refresh token by its hash
func (r *TokenRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (*service.RefreshToken, error) {
	query := `SELECT * FROM refresh_token WHERE token_hash = $hash LIMIT 1`
	vars := map[string]interface{}{"hash": hash}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	token, err := parseRefreshTokenResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return token, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, hash string) error {
	query := `UPDATE refresh_token SET revoked = true WHERE token_hash = $hash`
	vars := map[string]interface{}{"hash": hash}

	return r.db.Execute(ctx, query, vars)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_token SET revoked = true WHERE user = type::record($user)`
	vars := map[string]interface{}{"user": userID}

	return r.db.Execute(ctx, query, vars)
}

// DeleteExpiredTokens removes all expired refresh tokens
func (r *TokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE refresh_token WHERE expires_at < time::now()`

	return r.db.Execute(ctx, query, nil)
}

// CleanupRevokedTokens removes tokens that have been revoked for more than 7 days
func (r *TokenRepository) CleanupRevokedTokens(ctx context.Context) error {
	cutoff := time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	query := `DELETE refresh_token WHERE revoked = true AND created_at < <datetime>$cutoff`
	vars := map[string]interface{}{"cutoff": cutoff}

	return r.db.Execute(ctx, query, vars)
}

func parseRefreshTokenResult(result interface{}) (*service.RefreshToken, error) {
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
		data["id"] = convertTokenID(id)
	}
	if userID, ok := data["user"]; ok {
		data["user_id"] = convertTokenID(userID) // Map "user" to "user_id" for struct
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var token service.RefreshToken
	if err := json.Unmarshal(jsonBytes, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// convertTokenID converts a SurrealDB ID to a string
func convertTokenID(id interface{}) string {
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
