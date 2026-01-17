package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// TrustRepository handles trust and IRL verification data access
type TrustRepository struct {
	db database.Database
}

// NewTrustRepository creates a new trust repository
func NewTrustRepository(db database.Database) *TrustRepository {
	return &TrustRepository{db: db}
}

// CreateTrustRelation creates a trust relation (one-way)
func (r *TrustRepository) CreateTrustRelation(ctx context.Context, trust *model.TrustRelation) error {
	query := `
		CREATE trust_relation CONTENT {
			user_a_id: $user_a_id,
			user_b_id: $user_b_id,
			status: $status,
			created_on: time::now(),
			updated_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"user_a_id": trust.UserAID,
		"user_b_id": trust.UserBID,
		"status":    trust.Status,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	trust.ID = created.ID
	trust.CreatedOn = created.CreatedOn
	trust.UpdatedOn = created.UpdatedOn
	return nil
}

// GetTrustRelation retrieves trust from user A to user B
func (r *TrustRepository) GetTrustRelation(ctx context.Context, userAID, userBID string) (*model.TrustRelation, error) {
	query := `
		SELECT * FROM trust_relation
		WHERE user_a_id = $user_a_id AND user_b_id = $user_b_id
		LIMIT 1
	`
	vars := map[string]interface{}{
		"user_a_id": userAID,
		"user_b_id": userBID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseTrustRelationResult(result)
}

// UpdateTrustRelation updates a trust relation status
func (r *TrustRepository) UpdateTrustRelation(ctx context.Context, id, status string) error {
	query := `
		UPDATE trust_relation SET status = $status, updated_on = time::now()
		WHERE id = type::record($id)
	`
	vars := map[string]interface{}{
		"id":     id,
		"status": status,
	}

	return r.db.Execute(ctx, query, vars)
}

// DeleteTrustRelation deletes a trust relation
func (r *TrustRepository) DeleteTrustRelation(ctx context.Context, id string) error {
	query := `DELETE trust_relation WHERE id = type::record($id)`
	vars := map[string]interface{}{"id": id}

	return r.db.Execute(ctx, query, vars)
}

// GetUserTrustRelations gets all trust relations for a user (both directions)
func (r *TrustRepository) GetUserTrustRelations(ctx context.Context, userID string) ([]*model.TrustRelation, error) {
	query := `
		SELECT * FROM trust_relation
		WHERE (user_a_id = $user_id OR user_b_id = $user_id) AND status = "active"
	`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseTrustRelationsResult(result)
}

// CheckMutualTrust checks if two users have mutual trust
func (r *TrustRepository) CheckMutualTrust(ctx context.Context, userAID, userBID string) (bool, error) {
	query := `
		SELECT count() as cnt FROM trust_relation
		WHERE status = "active" AND (
			(user_a_id = $user_a_id AND user_b_id = $user_b_id) OR
			(user_a_id = $user_b_id AND user_b_id = $user_a_id)
		)
		GROUP ALL
	`
	vars := map[string]interface{}{
		"user_a_id": userAID,
		"user_b_id": userBID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	if data, ok := result.(map[string]interface{}); ok {
		cnt := getInt(data, "cnt")
		return cnt >= 2, nil // Both directions needed for mutual
	}
	return false, nil
}

// CreateIRLVerification creates an IRL verification record
func (r *TrustRepository) CreateIRLVerification(ctx context.Context, irl *model.IRLVerification) error {
	query := `
		CREATE irl_verification CONTENT {
			user_a_id: $user_a_id,
			user_b_id: $user_b_id,
			context: $context,
			reference_id: $reference_id,
			user_a_confirmed: $user_a_confirmed,
			user_b_confirmed: $user_b_confirmed,
			user_a_confirmed_on: $user_a_confirmed_on,
			user_b_confirmed_on: $user_b_confirmed_on,
			verified_on: time::now()
		}
	`

	vars := map[string]interface{}{
		"user_a_id":           irl.UserAID,
		"user_b_id":           irl.UserBID,
		"context":             irl.Context,
		"reference_id":        irl.ReferenceID,
		"user_a_confirmed":    irl.UserAConfirmed,
		"user_b_confirmed":    irl.UserBConfirmed,
		"user_a_confirmed_on": irl.UserAConfirmedOn,
		"user_b_confirmed_on": irl.UserBConfirmedOn,
	}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return err
	}

	created, err := extractCreatedRecord(result)
	if err != nil {
		return err
	}

	irl.ID = created.ID
	irl.VerifiedOn = created.CreatedOn
	return nil
}

// GetIRLVerification retrieves IRL verification between two users
func (r *TrustRepository) GetIRLVerification(ctx context.Context, userAID, userBID string) (*model.IRLVerification, error) {
	query := `
		SELECT * FROM irl_verification
		WHERE (user_a_id = $user_a_id AND user_b_id = $user_b_id)
		   OR (user_a_id = $user_b_id AND user_b_id = $user_a_id)
		ORDER BY verified_on DESC
		LIMIT 1
	`
	vars := map[string]interface{}{
		"user_a_id": userAID,
		"user_b_id": userBID,
	}

	result, err := r.db.QueryOne(ctx, query, vars)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.parseIRLVerificationResult(result)
}

// UpdateIRLVerification updates an IRL verification record
func (r *TrustRepository) UpdateIRLVerification(ctx context.Context, id string, updates map[string]interface{}) error {
	query := `UPDATE irl_verification SET `

	vars := map[string]interface{}{
		"id": id,
	}

	first := true
	if userAConfirmed, ok := updates["user_a_confirmed"]; ok {
		if !first {
			query += ", "
		}
		query += "user_a_confirmed = $user_a_confirmed"
		vars["user_a_confirmed"] = userAConfirmed
		first = false
	}
	if userAConfirmedOn, ok := updates["user_a_confirmed_on"]; ok {
		if !first {
			query += ", "
		}
		query += "user_a_confirmed_on = $user_a_confirmed_on"
		vars["user_a_confirmed_on"] = userAConfirmedOn
		first = false
	}
	if userBConfirmed, ok := updates["user_b_confirmed"]; ok {
		if !first {
			query += ", "
		}
		query += "user_b_confirmed = $user_b_confirmed"
		vars["user_b_confirmed"] = userBConfirmed
		first = false
	}
	if userBConfirmedOn, ok := updates["user_b_confirmed_on"]; ok {
		if !first {
			query += ", "
		}
		query += "user_b_confirmed_on = $user_b_confirmed_on"
		vars["user_b_confirmed_on"] = userBConfirmedOn
	}

	query += ` WHERE id = type::record($id)`

	return r.db.Execute(ctx, query, vars)
}

// CheckIRLConfirmed checks if IRL is confirmed by both users
func (r *TrustRepository) CheckIRLConfirmed(ctx context.Context, userAID, userBID string) (bool, error) {
	irl, err := r.GetIRLVerification(ctx, userAID, userBID)
	if err != nil {
		return false, err
	}
	if irl == nil {
		return false, nil
	}

	return irl.UserAConfirmed && irl.UserBConfirmed, nil
}

// GetUserIRLConnections gets all confirmed IRL connections for a user
func (r *TrustRepository) GetUserIRLConnections(ctx context.Context, userID string) ([]*model.IRLVerification, error) {
	query := `
		SELECT * FROM irl_verification
		WHERE (user_a_id = $user_id OR user_b_id = $user_id)
		  AND user_a_confirmed = true AND user_b_confirmed = true
		ORDER BY verified_on DESC
	`
	vars := map[string]interface{}{"user_id": userID}

	result, err := r.db.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	return r.parseIRLVerificationsResult(result)
}

// GetTrustProfile gets trust statistics for a user
func (r *TrustRepository) GetTrustProfile(ctx context.Context, userID string) (*model.UserTrustProfile, error) {
	// Count IRL connections
	irlCount := 0
	irls, err := r.GetUserIRLConnections(ctx, userID)
	if err == nil {
		irlCount = len(irls)
	}

	// Count trust relationships
	trustedByQuery := `
		SELECT count() as cnt FROM trust_relation
		WHERE user_b_id = $user_id AND status = "active"
		GROUP ALL
	`
	vars := map[string]interface{}{"user_id": userID}

	trustedByCount := 0
	result, err := r.db.QueryOne(ctx, trustedByQuery, vars)
	if err == nil && result != nil {
		if data, ok := result.(map[string]interface{}); ok {
			trustedByCount = getInt(data, "cnt")
		}
	}

	// Count users this user trusts
	trustsQuery := `
		SELECT count() as cnt FROM trust_relation
		WHERE user_a_id = $user_id AND status = "active"
		GROUP ALL
	`
	trustsCount := 0
	result, err = r.db.QueryOne(ctx, trustsQuery, vars)
	if err == nil && result != nil {
		if data, ok := result.(map[string]interface{}); ok {
			trustsCount = getInt(data, "cnt")
		}
	}

	// Count mutual trust
	mutualQuery := `
		SELECT count() as cnt FROM trust_relation tr1
		WHERE tr1.user_a_id = $user_id AND tr1.status = "active"
		AND (SELECT count() FROM trust_relation WHERE user_a_id = tr1.user_b_id AND user_b_id = $user_id AND status = "active") > 0
		GROUP ALL
	`
	mutualCount := 0
	result, err = r.db.QueryOne(ctx, mutualQuery, vars)
	if err == nil && result != nil {
		if data, ok := result.(map[string]interface{}); ok {
			mutualCount = getInt(data, "cnt")
		}
	}

	profile := &model.UserTrustProfile{
		UserID:             userID,
		IRLConnectionCount: irlCount,
		TrustedByCount:     trustedByCount,
		TrustsCount:        trustsCount,
		MutualTrustCount:   mutualCount,
		CanOfferCommute:    irlCount >= model.MinIRLForCommute && mutualCount >= model.MinTrustForCommute,
	}

	return profile, nil
}

// Helper functions

func (r *TrustRepository) parseTrustRelationResult(result interface{}) (*model.TrustRelation, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var trust model.TrustRelation
	if err := json.Unmarshal(jsonBytes, &trust); err != nil {
		return nil, err
	}

	if t := getTime(data, "created_on"); t != nil {
		trust.CreatedOn = *t
	}
	if t := getTime(data, "updated_on"); t != nil {
		trust.UpdatedOn = *t
	}

	return &trust, nil
}

func (r *TrustRepository) parseTrustRelationsResult(result []interface{}) ([]*model.TrustRelation, error) {
	relations := make([]*model.TrustRelation, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					trust, err := r.parseTrustRelationResult(item)
					if err != nil {
						continue
					}
					relations = append(relations, trust)
				}
				continue
			}
		}

		trust, err := r.parseTrustRelationResult(res)
		if err != nil {
			continue
		}
		relations = append(relations, trust)
	}

	return relations, nil
}

func (r *TrustRepository) parseIRLVerificationResult(result interface{}) (*model.IRLVerification, error) {
	if result == nil {
		return nil, database.ErrNotFound
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected result format")
	}

	if id, ok := data["id"]; ok {
		data["id"] = convertSurrealID(id)
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var irl model.IRLVerification
	if err := json.Unmarshal(jsonBytes, &irl); err != nil {
		return nil, err
	}

	if t := getTime(data, "verified_on"); t != nil {
		irl.VerifiedOn = *t
	}
	irl.UserAConfirmedOn = getTime(data, "user_a_confirmed_on")
	irl.UserBConfirmedOn = getTime(data, "user_b_confirmed_on")

	return &irl, nil
}

func (r *TrustRepository) parseIRLVerificationsResult(result []interface{}) ([]*model.IRLVerification, error) {
	irls := make([]*model.IRLVerification, 0)

	for _, res := range result {
		if resp, ok := res.(map[string]interface{}); ok {
			if resultData, ok := resp["result"].([]interface{}); ok {
				for _, item := range resultData {
					irl, err := r.parseIRLVerificationResult(item)
					if err != nil {
						continue
					}
					irls = append(irls, irl)
				}
				continue
			}
		}

		irl, err := r.parseIRLVerificationResult(res)
		if err != nil {
			continue
		}
		irls = append(irls, irl)
	}

	return irls, nil
}

// Unused - silence linter
var _ = time.Now
