package service

import (
	"context"
	"errors"

	"github.com/forgo/saga/api/internal/model"
)

var (
	ErrRoleNotFound           = errors.New("role not found")
	ErrAssignmentNotFound     = errors.New("assignment not found")
	ErrRoleFull               = errors.New("role is full")
	ErrAlreadyAssignedToRole  = errors.New("already assigned to this role")
	ErrCannotDeleteDefault    = errors.New("cannot delete default role")
	ErrMaxRolesReached        = errors.New("maximum roles reached")
	ErrCannotAssignOthers     = errors.New("cannot assign roles to others")
	ErrMaxRolesPerUserReached = errors.New("maximum roles per user reached")
	// Note: ErrNotEventHost is defined in event.go
)

// MaxRolesPerUser limits how many roles a single user can take at one event
const MaxRolesPerUser = 10

// EventRoleRepositoryInterface defines the repository interface
type EventRoleRepositoryInterface interface {
	CreateRole(ctx context.Context, role *model.EventRole) error
	GetRole(ctx context.Context, roleID string) (*model.EventRole, error)
	GetRolesByEvent(ctx context.Context, eventID string) ([]*model.EventRole, error)
	UpdateRole(ctx context.Context, roleID string, updates map[string]interface{}) (*model.EventRole, error)
	DeleteRole(ctx context.Context, roleID string) error
	CreateAssignment(ctx context.Context, assignment *model.EventRoleAssignment) error
	GetAssignment(ctx context.Context, assignmentID string) (*model.EventRoleAssignment, error)
	GetUserAssignmentForRole(ctx context.Context, roleID, userID string) (*model.EventRoleAssignment, error)
	GetUserAssignmentsForEvent(ctx context.Context, eventID, userID string) ([]*model.EventRoleAssignment, error)
	GetAssignmentsByRole(ctx context.Context, roleID string) ([]*model.EventRoleAssignment, error)
	GetAssignmentsByEvent(ctx context.Context, eventID string) ([]*model.EventRoleAssignment, error)
	UpdateAssignment(ctx context.Context, assignmentID string, updates map[string]interface{}) (*model.EventRoleAssignment, error)
	DeleteAssignment(ctx context.Context, assignmentID string) error
	CountAssignmentsByRole(ctx context.Context, roleID string) (int, error)
	GetRolesWithAssignments(ctx context.Context, eventID string) ([]model.EventRoleWithAssignments, error)
}

// InterestServiceInterface for suggesting roles based on interests
type InterestServiceForRoles interface {
	GetUserInterests(ctx context.Context, userID string) ([]*model.UserInterest, error)
}

// EventRoleService handles event role business logic
type EventRoleService struct {
	repo            EventRoleRepositoryInterface
	interestService InterestServiceForRoles
}

// NewEventRoleService creates a new event role service
func NewEventRoleService(repo EventRoleRepositoryInterface, interestService InterestServiceForRoles) *EventRoleService {
	return &EventRoleService{
		repo:            repo,
		interestService: interestService,
	}
}

// CreateRole creates a new role for an event (host only)
// MaxSlots defaults to 1 if not specified (one person per role by default)
func (s *EventRoleService) CreateRole(ctx context.Context, eventID, hostUserID string, req *model.CreateEventRoleRequest) (*model.EventRole, error) {
	// Check max roles
	existing, err := s.repo.GetRolesByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if len(existing) >= model.MaxRolesPerEvent {
		return nil, ErrMaxRolesReached
	}

	// Default to 1 slot per role if not specified
	maxSlots := req.MaxSlots
	if maxSlots == 0 {
		maxSlots = model.DefaultMaxSlotsPerRole
	}

	// Create the role
	role := &model.EventRole{
		EventID:            eventID,
		Name:               req.Name,
		Description:        req.Description,
		MaxSlots:           maxSlots,
		IsDefault:          false,
		SortOrder:          len(existing) + 1,
		CreatedBy:          hostUserID,
		SuggestedInterests: req.SuggestedInterests,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

// CreateDefaultRole creates the default "Guest" role for an event
func (s *EventRoleService) CreateDefaultRole(ctx context.Context, eventID, hostUserID string, maxSlots int) (*model.EventRole, error) {
	description := model.DefaultRoleDescription
	role := &model.EventRole{
		EventID:     eventID,
		Name:        model.DefaultRoleName,
		Description: &description,
		MaxSlots:    maxSlots,
		IsDefault:   true,
		SortOrder:   0,
		CreatedBy:   hostUserID,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

// GetRole retrieves a role by ID
func (s *EventRoleService) GetRole(ctx context.Context, roleID string) (*model.EventRole, error) {
	role, err := s.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

// GetEventRoles retrieves all roles for an event
func (s *EventRoleService) GetEventRoles(ctx context.Context, eventID string) ([]*model.EventRole, error) {
	return s.repo.GetRolesByEvent(ctx, eventID)
}

// GetEventRolesOverview retrieves roles with assignment counts
func (s *EventRoleService) GetEventRolesOverview(ctx context.Context, eventID string) (*model.EventRolesOverview, error) {
	rolesWithAssignments, err := s.repo.GetRolesWithAssignments(ctx, eventID)
	if err != nil {
		return nil, err
	}

	totalAttendees := 0
	totalSlots := 0
	for _, rwa := range rolesWithAssignments {
		totalAttendees += len(rwa.Assignments)
		// Count total slots (0 = unlimited, so don't count those)
		if rwa.Role.MaxSlots > 0 {
			totalSlots += rwa.Role.MaxSlots
		}
	}

	overview := &model.EventRolesOverview{
		EventID:        eventID,
		TotalAttendees: totalAttendees,
		Roles:          rolesWithAssignments,
	}

	return overview, nil
}

// GetTotalSlotsForEvent returns the total number of role slots defined for an event
// This is used to validate against max_attendees (roles define the capacity)
func (s *EventRoleService) GetTotalSlotsForEvent(ctx context.Context, eventID string) (int, error) {
	roles, err := s.repo.GetRolesByEvent(ctx, eventID)
	if err != nil {
		return 0, err
	}

	totalSlots := 0
	hasUnlimited := false
	for _, role := range roles {
		if role.MaxSlots == 0 {
			hasUnlimited = true // Unlimited slots (typically the default Guest role)
		} else {
			totalSlots += role.MaxSlots
		}
	}

	// If there's an unlimited role, return -1 to indicate unlimited capacity
	if hasUnlimited {
		return -1, nil
	}

	return totalSlots, nil
}

// GetFilledSlotsForEvent returns the total number of filled slots for an event
func (s *EventRoleService) GetFilledSlotsForEvent(ctx context.Context, eventID string) (int, error) {
	assignments, err := s.repo.GetAssignmentsByEvent(ctx, eventID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, a := range assignments {
		if a.Status == model.RoleAssignmentStatusConfirmed {
			count++
		}
	}

	return count, nil
}

// UpdateRole updates a role (host only)
func (s *EventRoleService) UpdateRole(ctx context.Context, roleID string, req *model.UpdateEventRoleRequest) (*model.EventRole, error) {
	role, err := s.repo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.MaxSlots != nil {
		updates["max_slots"] = *req.MaxSlots
	}
	if req.SuggestedInterests != nil {
		updates["suggested_interests"] = req.SuggestedInterests
	}

	if len(updates) == 0 {
		return role, nil
	}

	return s.repo.UpdateRole(ctx, roleID, updates)
}

// DeleteRole deletes a role (host only)
func (s *EventRoleService) DeleteRole(ctx context.Context, roleID string) error {
	role, err := s.repo.GetRole(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}
	if role.IsDefault {
		return ErrCannotDeleteDefault
	}

	return s.repo.DeleteRole(ctx, roleID)
}

// AssignRole assigns the current user to a role (self-assignment only)
// Users can have multiple roles at an event (e.g., DJ + bring lasagna + wash dishes)
func (s *EventRoleService) AssignRole(ctx context.Context, userID string, req *model.AssignRoleRequest) (*model.EventRoleAssignment, error) {
	// Get the role to verify it exists and check capacity
	role, err := s.repo.GetRole(ctx, req.RoleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}

	// Check if user is already assigned to THIS specific role
	existingForRole, err := s.repo.GetUserAssignmentForRole(ctx, req.RoleID, userID)
	if err != nil {
		return nil, err
	}
	if existingForRole != nil && existingForRole.Status != model.RoleAssignmentStatusCancelled {
		return nil, ErrAlreadyAssignedToRole
	}

	// Check if user has reached max roles per user limit
	existingAssignments, err := s.repo.GetUserAssignmentsForEvent(ctx, role.EventID, userID)
	if err != nil {
		return nil, err
	}
	activeCount := 0
	for _, a := range existingAssignments {
		if a.Status != model.RoleAssignmentStatusCancelled {
			activeCount++
		}
	}
	if activeCount >= MaxRolesPerUser {
		return nil, ErrMaxRolesPerUserReached
	}

	// Check if role is full
	if role.MaxSlots > 0 {
		count, err := s.repo.CountAssignmentsByRole(ctx, req.RoleID)
		if err != nil {
			return nil, err
		}
		if count >= role.MaxSlots {
			return nil, ErrRoleFull
		}
	}

	// Create the assignment
	assignment := &model.EventRoleAssignment{
		EventID: role.EventID,
		RoleID:  req.RoleID,
		UserID:  userID,
		Note:    req.Note,
		Status:  model.RoleAssignmentStatusConfirmed,
	}

	if err := s.repo.CreateAssignment(ctx, assignment); err != nil {
		return nil, err
	}

	assignment.RoleName = &role.Name
	return assignment, nil
}

// UpdateAssignment updates an assignment (user can update their own note)
func (s *EventRoleService) UpdateAssignment(ctx context.Context, userID, assignmentID string, req *model.UpdateAssignmentRequest) (*model.EventRoleAssignment, error) {
	assignment, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	if assignment == nil {
		return nil, ErrAssignmentNotFound
	}
	if assignment.UserID != userID {
		return nil, ErrCannotAssignOthers
	}

	updates := make(map[string]interface{})
	if req.Note != nil {
		updates["note"] = *req.Note
	}

	if len(updates) == 0 {
		return assignment, nil
	}

	return s.repo.UpdateAssignment(ctx, assignmentID, updates)
}

// CancelAssignment cancels a user's role assignment
func (s *EventRoleService) CancelAssignment(ctx context.Context, userID, assignmentID string) error {
	assignment, err := s.repo.GetAssignment(ctx, assignmentID)
	if err != nil {
		return err
	}
	if assignment == nil {
		return ErrAssignmentNotFound
	}
	if assignment.UserID != userID {
		return ErrCannotAssignOthers
	}

	_, err = s.repo.UpdateAssignment(ctx, assignmentID, map[string]interface{}{
		"status": model.RoleAssignmentStatusCancelled,
	})
	return err
}

// GetUserRoles retrieves all of a user's role assignments for an event
func (s *EventRoleService) GetUserRoles(ctx context.Context, eventID, userID string) (*model.UserEventRoles, error) {
	assignments, err := s.repo.GetUserAssignmentsForEvent(ctx, eventID, userID)
	if err != nil {
		return nil, err
	}

	// Filter out cancelled assignments and convert to non-pointer
	activeAssignments := make([]model.EventRoleAssignment, 0)
	for _, a := range assignments {
		if a.Status != model.RoleAssignmentStatusCancelled {
			activeAssignments = append(activeAssignments, *a)
		}
	}

	return &model.UserEventRoles{
		UserID:      userID,
		EventID:     eventID,
		Assignments: activeAssignments,
	}, nil
}

// GetRoleSuggestions suggests roles for a user based on their interests
func (s *EventRoleService) GetRoleSuggestions(ctx context.Context, eventID, userID string) ([]model.RoleSuggestion, error) {
	// Get user's interests
	interests, err := s.interestService.GetUserInterests(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build a map of user's interest IDs
	userInterests := make(map[string]*model.UserInterest)
	for _, interest := range interests {
		userInterests[interest.InterestID] = interest
	}

	// Get event roles
	roles, err := s.repo.GetRolesByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Match roles to user interests
	suggestions := make([]model.RoleSuggestion, 0)
	for _, role := range roles {
		if role.IsDefault {
			continue // Don't suggest the default role
		}

		for _, suggestedInterestID := range role.SuggestedInterests {
			if interest, ok := userInterests[suggestedInterestID]; ok {
				suggestions = append(suggestions, model.RoleSuggestion{
					Role:            *role,
					MatchedInterest: suggestedInterestID,
					Reason:          "You're interested in " + interest.Name,
				})
				break // Only one suggestion per role
			}
		}
	}

	return suggestions, nil
}

// AddRole is an alias for AssignRole - users can take on multiple roles at an event
func (s *EventRoleService) AddRole(ctx context.Context, userID string, req *model.AssignRoleRequest) (*model.EventRoleAssignment, error) {
	return s.AssignRole(ctx, userID, req)
}

// RemoveRole removes a specific role assignment (cancels it)
func (s *EventRoleService) RemoveRole(ctx context.Context, userID, assignmentID string) error {
	return s.CancelAssignment(ctx, userID, assignmentID)
}
