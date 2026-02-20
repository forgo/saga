package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

// AdminUserRepository defines the user repo interface needed by AdminUsersService
type AdminUserRepository interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
	SetRole(ctx context.Context, userID string, role model.UserRole) error
	Delete(ctx context.Context, id string) error
}

// AdminProfileRepository defines the profile repo interface needed by AdminUsersService
type AdminProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*model.UserProfile, error)
	Delete(ctx context.Context, userID string) error
}

// AdminUsersService handles admin user management operations
type AdminUsersService struct {
	db            database.Database
	userRepo      AdminUserRepository
	profileRepo   AdminProfileRepository
	moderationSvc *ModerationService
}

// NewAdminUsersService creates a new admin users service
func NewAdminUsersService(
	db database.Database,
	userRepo AdminUserRepository,
	profileRepo AdminProfileRepository,
	moderationSvc *ModerationService,
) *AdminUsersService {
	return &AdminUsersService{
		db:            db,
		userRepo:      userRepo,
		profileRepo:   profileRepo,
		moderationSvc: moderationSvc,
	}
}

// ListUsersRequest defines the request for listing users
type ListUsersRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Search   string `json:"search,omitempty"`
	Role     string `json:"role,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`
	SortDir  string `json:"sort_dir,omitempty"`
}

// AdminUserItem represents a user in the admin list
type AdminUserItem struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      *string `json:"username,omitempty"`
	Firstname     *string `json:"firstname,omitempty"`
	Lastname      *string `json:"lastname,omitempty"`
	Role          string  `json:"role"`
	EmailVerified bool    `json:"email_verified"`
	CreatedOn     string  `json:"created_on"`
	UpdatedOn     string  `json:"updated_on"`
	LoginOn       *string `json:"login_on,omitempty"`
	Status        string  `json:"status"` // active, suspended, banned
}

// ListUsersResponse contains the paginated user list
type ListUsersResponse struct {
	Users    []AdminUserItem `json:"users"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// AdminUserDetail represents detailed user info for the admin panel
type AdminUserDetail struct {
	// User fields
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      *string `json:"username,omitempty"`
	Firstname     *string `json:"firstname,omitempty"`
	Lastname      *string `json:"lastname,omitempty"`
	Role          string  `json:"role"`
	EmailVerified bool    `json:"email_verified"`
	CreatedOn     string  `json:"created_on"`
	UpdatedOn     string  `json:"updated_on"`
	LoginOn       *string `json:"login_on,omitempty"`

	// Profile
	Profile *AdminUserProfile `json:"profile,omitempty"`

	// Moderation
	Moderation *model.UserModerationStatus `json:"moderation,omitempty"`

	// Stats
	Stats *AdminUserStats `json:"stats,omitempty"`
}

// AdminUserProfile is a subset of profile data for the admin panel
type AdminUserProfile struct {
	Bio        *string `json:"bio,omitempty"`
	Tagline    *string `json:"tagline,omitempty"`
	City       string  `json:"city,omitempty"`
	Country    string  `json:"country,omitempty"`
	Visibility string  `json:"visibility"`
	LastActive *string `json:"last_active,omitempty"`
}

// AdminUserStats contains user statistics
type AdminUserStats struct {
	GuildCount int `json:"guild_count"`
	EventCount int `json:"event_count"`
}

// UpdateRoleRequest defines the request for updating a user's role
type UpdateRoleRequest struct {
	Role string `json:"role"`
}

// ListUsers returns a paginated list of users with search/filter/sort
func (s *AdminUsersService) ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	// Defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Build WHERE clause
	var conditions []string
	vars := map[string]interface{}{
		"limit":  req.PageSize,
		"offset": (req.Page - 1) * req.PageSize,
	}

	if req.Search != "" {
		conditions = append(conditions, "(string::lowercase(email) CONTAINS string::lowercase($search) OR string::lowercase(username ?? '') CONTAINS string::lowercase($search) OR string::lowercase(firstname ?? '') CONTAINS string::lowercase($search) OR string::lowercase(lastname ?? '') CONTAINS string::lowercase($search))")
		vars["search"] = req.Search
	}

	if req.Role != "" {
		conditions = append(conditions, "role = $role")
		vars["role"] = req.Role
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build ORDER BY
	sortBy := "created_on"
	sortDir := "DESC"
	validSorts := map[string]bool{"email": true, "username": true, "role": true, "created_on": true, "updated_on": true}
	if req.SortBy != "" && validSorts[req.SortBy] {
		sortBy = req.SortBy
	}
	if req.SortDir == "asc" || req.SortDir == "ASC" {
		sortDir = "ASC"
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT count() AS total FROM user %s GROUP ALL", whereClause)
	countResults, err := s.db.Query(ctx, countQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	total := extractCountValue(countResults)

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT
			id,
			email,
			username,
			firstname,
			lastname,
			role,
			email_verified,
			created_on,
			updated_on,
			login_on
		FROM user
		%s
		ORDER BY %s %s
		LIMIT $limit
		START $offset
	`, whereClause, sortBy, sortDir)

	results, err := s.db.Query(ctx, dataQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	rows := extractResultArray(results)

	// Convert to typed response and enrich with moderation status
	users := make([]AdminUserItem, 0, len(rows))
	for _, row := range rows {
		item := AdminUserItem{
			ID:            getStringField(row, "id"),
			Email:         getStringField(row, "email"),
			Username:      getOptStringField(row, "username"),
			Firstname:     getOptStringField(row, "firstname"),
			Lastname:      getOptStringField(row, "lastname"),
			Role:          getStringField(row, "role"),
			EmailVerified: getBoolField(row, "email_verified"),
			CreatedOn:     getTimeStringField(row, "created_on"),
			UpdatedOn:     getTimeStringField(row, "updated_on"),
			LoginOn:       getOptTimeStringField(row, "login_on"),
			Status:        "active",
		}

		// Get moderation status
		if item.ID != "" {
			modStatus, err := s.moderationSvc.GetUserModerationStatus(ctx, item.ID)
			if err == nil && modStatus != nil {
				if modStatus.IsBanned {
					item.Status = "banned"
				} else if modStatus.IsSuspended {
					item.Status = "suspended"
				}
			}
		}

		users = append(users, item)
	}

	return &ListUsersResponse{
		Users:    users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetUserDetail returns detailed information about a single user
func (s *AdminUsersService) GetUserDetail(ctx context.Context, userID string) (*AdminUserDetail, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	detail := &AdminUserDetail{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		Firstname:     user.Firstname,
		Lastname:      user.Lastname,
		Role:          string(user.Role),
		EmailVerified: user.EmailVerified,
		CreatedOn:     user.CreatedOn.Format(time.RFC3339),
		UpdatedOn:     user.UpdatedOn.Format(time.RFC3339),
	}
	if user.LoginOn != nil {
		loginStr := user.LoginOn.Format(time.RFC3339)
		detail.LoginOn = &loginStr
	}

	// Get profile
	profile, err := s.profileRepo.GetByUserID(ctx, userID)
	if err == nil && profile != nil {
		detail.Profile = &AdminUserProfile{
			Bio:        profile.Bio,
			Tagline:    profile.Tagline,
			Visibility: profile.Visibility,
		}
		if profile.Location != nil {
			detail.Profile.City = profile.Location.City
			detail.Profile.Country = profile.Location.Country
		}
		if profile.LastActive != nil {
			la := profile.LastActive.Format(time.RFC3339)
			detail.Profile.LastActive = &la
		}
	}

	// Get moderation status
	modStatus, err := s.moderationSvc.GetUserModerationStatus(ctx, userID)
	if err == nil {
		detail.Moderation = modStatus
	}

	// Get stats (guild and event counts)
	detail.Stats = s.getUserStats(ctx, userID)

	return detail, nil
}

// getUserStats retrieves user statistics from the database
func (s *AdminUsersService) getUserStats(ctx context.Context, userID string) *AdminUserStats {
	stats := &AdminUserStats{}

	// Count guilds the user is a member of
	guildQuery := `SELECT count() AS total FROM responsible_for WHERE in = type::record($user_id) GROUP ALL`
	guildResults, err := s.db.Query(ctx, guildQuery, map[string]interface{}{"user_id": userID})
	if err == nil {
		stats.GuildCount = extractCountValue(guildResults)
	}

	// Count events the user created
	eventQuery := `SELECT count() AS total FROM event WHERE created_by = type::record($user_id) GROUP ALL`
	eventResults, err := s.db.Query(ctx, eventQuery, map[string]interface{}{"user_id": userID})
	if err == nil {
		stats.EventCount = extractCountValue(eventResults)
	}

	return stats
}

// UpdateUserRole updates a user's role with self-demotion protection
func (s *AdminUsersService) UpdateUserRole(ctx context.Context, adminUserID, targetUserID string, role model.UserRole) error {
	// Validate role
	switch role {
	case model.UserRoleUser, model.UserRoleModerator, model.UserRoleAdmin:
		// valid
	default:
		return fmt.Errorf("invalid role: %s", role)
	}

	// Self-demotion protection
	if adminUserID == targetUserID && role != model.UserRoleAdmin {
		return fmt.Errorf("cannot demote yourself")
	}

	// Verify target user exists
	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.SetRole(ctx, targetUserID, role)
}

// DeleteUser deletes a user — soft delete (ban) by default, hard delete if requested
func (s *AdminUsersService) DeleteUser(ctx context.Context, adminUserID, targetUserID string, hard bool) error {
	// Verify target user exists
	user, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Self-deletion protection
	if adminUserID == targetUserID {
		return fmt.Errorf("cannot delete yourself")
	}

	if !hard {
		// Soft delete = ban via moderation
		_, err := s.moderationSvc.TakeAction(ctx, adminUserID, &model.CreateModerationActionRequest{
			UserID: targetUserID,
			Level:  string(model.ModerationLevelBan),
			Reason: "Deleted by admin",
		})
		return err
	}

	// Hard delete — cascade delete user data
	slog.Info("hard deleting user", slog.String("user_id", targetUserID), slog.String("admin_id", adminUserID))

	// Delete profile
	if err := s.profileRepo.Delete(ctx, targetUserID); err != nil {
		slog.Warn("failed to delete profile during hard delete", slog.String("error", err.Error()))
	}

	// Delete identities via direct query
	identityQuery := `DELETE identity WHERE user = type::record($user_id)`
	if err := s.db.Execute(ctx, identityQuery, map[string]interface{}{"user_id": targetUserID}); err != nil {
		slog.Warn("failed to delete identities during hard delete", slog.String("error", err.Error()))
	}

	// Delete user record
	return s.userRepo.Delete(ctx, targetUserID)
}

// Helper: extract count value from SurrealDB count() query result
func extractCountValue(results []interface{}) int {
	if len(results) == 0 {
		return 0
	}

	resp, ok := results[0].(map[string]interface{})
	if !ok {
		return 0
	}

	result, ok := resp["result"]
	if !ok {
		return 0
	}

	arr, ok := result.([]interface{})
	if !ok || len(arr) == 0 {
		return 0
	}

	data, ok := arr[0].(map[string]interface{})
	if !ok {
		return 0
	}

	if total, ok := data["total"].(float64); ok {
		return int(total)
	}
	if total, ok := data["total"].(int); ok {
		return total
	}

	return 0
}

// Helper: get string field from result map
func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		// Handle SurrealDB record IDs
		return formatID(v)
	}
	return ""
}

// Helper: get optional string field from result map
func getOptStringField(m map[string]interface{}, key string) *string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok && s != "" {
			return &s
		}
	}
	return nil
}

// Helper: get bool field from result map
func getBoolField(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// Helper: get time field as RFC3339 string
func getTimeStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case string:
			return val
		case time.Time:
			return val.Format(time.RFC3339)
		case models.CustomDateTime:
			return val.Format(time.RFC3339)
		case *models.CustomDateTime:
			if val != nil {
				return val.Format(time.RFC3339)
			}
		}
	}
	return ""
}

// Helper: get optional time field as RFC3339 string
func getOptTimeStringField(m map[string]interface{}, key string) *string {
	if v, ok := m[key]; ok && v != nil {
		var s string
		switch val := v.(type) {
		case string:
			s = val
		case time.Time:
			s = val.Format(time.RFC3339)
		case models.CustomDateTime:
			s = val.Format(time.RFC3339)
		case *models.CustomDateTime:
			if val != nil {
				s = val.Format(time.RFC3339)
			}
		default:
			return nil
		}
		if s != "" {
			return &s
		}
	}
	return nil
}
