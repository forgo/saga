package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// PasskeyRepository handles passkey data access
type PasskeyRepository struct {
	db database.Database
}

// NewPasskeyRepository creates a new passkey repository
func NewPasskeyRepository(db database.Database) *PasskeyRepository {
	return &PasskeyRepository{db: db}
}

// Create creates a new passkey
func (r *PasskeyRepository) Create(ctx context.Context, passkey *model.Passkey) error {
	query := `
		CREATE passkey CONTENT {
			user: type::record($user_id),
			credential_id: $credential_id,
			public_key: $public_key,
			sign_count: $sign_count,
			name: $name,
			created_on: time::now(),
			last_used_on: NONE
		}
	`

	vars := map[string]interface{}{
		"user_id":       passkey.UserID,
		"credential_id": passkey.CredentialID,
		"public_key":    passkey.PublicKey,
		"sign_count":    passkey.SignCount,
		"name":          passkey.Name,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		if isUniqueConstraintError(err) {
			return database.ErrDuplicate
		}
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	passkey.ID = created.ID
	passkey.CreatedOn = created.CreatedOn
	return nil
}

// GetByCredentialID retrieves a passkey by credential ID
func (r *PasskeyRepository) GetByCredentialID(ctx context.Context, credentialID string) (*model.Passkey, error) {
	query := `SELECT * FROM passkey WHERE credential_id = $credential_id LIMIT 1`
	vars := map[string]interface{}{"credential_id": credentialID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	passkey, err := parsePasskeyResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return passkey, nil
}

// GetByUserID retrieves all passkeys for a user
func (r *PasskeyRepository) GetByUserID(ctx context.Context, userID string) ([]*model.Passkey, error) {
	query := `SELECT * FROM passkey WHERE user = type::record($user_id) ORDER BY created_on DESC`
	vars := map[string]interface{}{"user_id": userID}

	results, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return parsePasskeysResult(results)
}

// GetByID retrieves a passkey by ID
func (r *PasskeyRepository) GetByID(ctx context.Context, id string) (*model.Passkey, error) {
	query := `SELECT * FROM type::record($id)`
	vars := map[string]interface{}{"id": id}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	passkey, err := parsePasskeyResult(result)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return passkey, nil
}

// UpdateSignCount updates the sign count and last used timestamp
func (r *PasskeyRepository) UpdateSignCount(ctx context.Context, credentialID string, signCount uint32) error {
	query := `
		UPDATE passkey
		SET sign_count = $sign_count, last_used_on = time::now()
		WHERE credential_id = $credential_id
	`
	vars := map[string]interface{}{
		"credential_id": credentialID,
		"sign_count":    signCount,
	}

	return r.db.Execute(ctx, query, vars)
}

// Delete deletes a passkey
func (r *PasskeyRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE type::record($id)`
	vars := map[string]interface{}{"id": id}

	return r.db.Execute(ctx, query, vars)
}

// DeleteByCredentialID deletes a passkey by credential ID
func (r *PasskeyRepository) DeleteByCredentialID(ctx context.Context, credentialID string) error {
	query := `DELETE passkey WHERE credential_id = $credential_id`
	vars := map[string]interface{}{"credential_id": credentialID}

	return r.db.Execute(ctx, query, vars)
}

// CountByUserID counts passkeys for a user
func (r *PasskeyRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT count() AS count FROM passkey WHERE user = type::record($user_id) GROUP ALL`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		if count, ok := data["count"].(float64); ok {
			return int(count), nil
		}
	}

	return 0, nil
}

func parsePasskeyResult(result interface{}) (*model.Passkey, error) {
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
		data["id"] = convertTokenID(id)
	}
	if userID, ok := data["user"]; ok {
		data["user_id"] = convertTokenID(userID)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var passkey model.Passkey
	if err := json.Unmarshal(jsonBytes, &passkey); err != nil {
		return nil, err
	}

	// Handle byte array for public key
	if pk, ok := data["public_key"].(string); ok {
		passkey.PublicKey = []byte(pk)
	}

	return &passkey, nil
}

func parsePasskeysResult(results []interface{}) ([]*model.Passkey, error) {
	passkeys := make([]*model.Passkey, 0)

	for _, result := range results {
		if resp, ok := result.(map[string]interface{}); ok {
			if status, ok := resp["status"].(string); ok && status == "OK" {
				if resultData, ok := resp["result"].([]interface{}); ok {
					for _, item := range resultData {
						passkey, err := parsePasskeyResult(item)
						if err == nil {
							passkeys = append(passkeys, passkey)
						}
					}
				}
			}
		}
	}

	return passkeys, nil
}

// WebAuthnCredential represents a passkey credential for WebAuthn operations
type WebAuthnCredential struct {
	ID              []byte
	PublicKey       []byte
	SignCount       uint32
	AttestationType string
	Transport       []string
}

// PasskeyToWebAuthnCredential converts a Passkey to WebAuthn credential format
func PasskeyToWebAuthnCredential(p *model.Passkey) *WebAuthnCredential {
	return &WebAuthnCredential{
		ID:        []byte(p.CredentialID),
		PublicKey: p.PublicKey,
		SignCount: p.SignCount,
	}
}

// WebAuthnUser interface for go-webauthn library
type WebAuthnUser struct {
	UserID      string
	UserEmail   string
	UserName    string
	Credentials []*model.Passkey
}

// WebAuthnID returns the user ID as bytes
func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.UserID)
}

// WebAuthnName returns the username
func (u *WebAuthnUser) WebAuthnName() string {
	return u.UserEmail
}

// WebAuthnDisplayName returns the display name
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	if u.UserName != "" {
		return u.UserName
	}
	return u.UserEmail
}

// WebAuthnCredentials returns the user's credentials
func (u *WebAuthnUser) WebAuthnCredentials() [][]byte {
	creds := make([][]byte, len(u.Credentials))
	for i, c := range u.Credentials {
		creds[i] = []byte(c.CredentialID)
	}
	return creds
}

// LastUsedAt returns the most recent last_used_on timestamp or zero time
func LastUsedAt(passkeys []*model.Passkey) time.Time {
	var latest time.Time
	for _, p := range passkeys {
		if p.LastUsedOn != nil && p.LastUsedOn.After(latest) {
			latest = *p.LastUsedOn
		}
	}
	return latest
}
