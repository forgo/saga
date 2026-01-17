package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/forgo/saga/api/internal/database"
	"github.com/forgo/saga/api/internal/model"
)

// GuildRepository defines the interface for guild storage
type GuildRepository interface {
	Create(ctx context.Context, guild *model.Guild) error
	GetByID(ctx context.Context, id string) (*model.Guild, error)
	Update(ctx context.Context, guild *model.Guild) error
	Delete(ctx context.Context, id string) error
	GetGuildsForUser(ctx context.Context, userID string) ([]*model.Guild, error)
	CountGuildsForUser(ctx context.Context, userID string) (int, error)
	AddMember(ctx context.Context, memberID, guildID string, pendingApproval bool) error
	AddMemberWithRole(ctx context.Context, memberID, guildID string, role model.GuildRole, pendingApproval bool) error
	RemoveMember(ctx context.Context, memberID, guildID string) error
	IsMember(ctx context.Context, userID, guildID string) (bool, error)
	CountMembers(ctx context.Context, guildID string) (int, error)
	GetMembers(ctx context.Context, guildID string) ([]*model.Member, error)
	GetMemberRole(ctx context.Context, userID, guildID string) (model.GuildRole, error)
	IsGuildAdmin(ctx context.Context, userID, guildID string) (bool, error)
	IsGuildModerator(ctx context.Context, userID, guildID string) (bool, error)
	UpdateMemberRole(ctx context.Context, userID, guildID string, role model.GuildRole) error
}

// MemberRepository defines the interface for member storage
type MemberRepository interface {
	Create(ctx context.Context, member *model.Member) error
	GetByID(ctx context.Context, id string) (*model.Member, error)
	GetByUserID(ctx context.Context, userID string) (*model.Member, error)
	GetOrCreate(ctx context.Context, userID, name, email string) (*model.Member, error)
	Update(ctx context.Context, member *model.Member) error
	Delete(ctx context.Context, id string) error
}

// Error definitions moved to errors.go

// GuildService handles guild business logic
type GuildService struct {
	guildRepo  GuildRepository
	memberRepo MemberRepository
	userRepo   UserRepository
}

// GuildServiceConfig holds dependencies for GuildService
type GuildServiceConfig struct {
	GuildRepo  GuildRepository
	MemberRepo MemberRepository
	UserRepo   UserRepository
}

// NewGuildService creates a new guild service
func NewGuildService(cfg GuildServiceConfig) *GuildService {
	return &GuildService{
		guildRepo:  cfg.GuildRepo,
		memberRepo: cfg.MemberRepo,
		userRepo:   cfg.UserRepo,
	}
}

// CreateGuildRequest represents a request to create a guild
type CreateGuildRequest struct {
	Name        string
	Description string
	Icon        string
	Color       string
	Visibility  string
}

// CreateGuild creates a new guild with the given user as the initial admin member
func (s *GuildService) CreateGuild(ctx context.Context, userID string, req CreateGuildRequest) (*model.Guild, error) {
	// Validate name
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrGuildNameRequired
	}
	if len(name) > model.MaxGuildNameLength {
		return nil, ErrGuildNameTooLong
	}

	// Validate description
	if len(req.Description) > model.MaxGuildDescLength {
		return nil, ErrGuildDescTooLong
	}

	// Check user hasn't exceeded max guilds
	count, err := s.guildRepo.CountGuildsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("checking guild count: %w", err)
	}
	if count >= model.MaxGuildsPerUser {
		return nil, ErrMaxGuildsReached
	}

	// Get user for member creation
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Set default visibility
	visibility := req.Visibility
	if visibility == "" {
		visibility = model.GuildVisibilityPrivate
	}

	// Create guild
	guild := &model.Guild{
		Name:        name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		Visibility:  visibility,
	}

	if err := s.guildRepo.Create(ctx, guild); err != nil {
		if errors.Is(err, database.ErrDuplicate) {
			return nil, ErrGuildNameExists
		}
		return nil, fmt.Errorf("creating guild: %w", err)
	}

	// Get or create member for user
	member, err := s.memberRepo.GetOrCreate(ctx, userID, user.Email, user.Email)
	if err != nil {
		return nil, fmt.Errorf("getting/creating member: %w", err)
	}

	// Add member to guild as admin (not pending approval since they're the creator)
	if err := s.guildRepo.AddMemberWithRole(ctx, member.ID, guild.ID, model.GuildRoleAdmin, false); err != nil {
		return nil, fmt.Errorf("adding member to guild: %w", err)
	}

	return guild, nil
}

// GetGuild retrieves a guild by ID, checking membership for private guilds
func (s *GuildService) GetGuild(ctx context.Context, userID, guildID string) (*model.Guild, error) {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("getting guild: %w", err)
	}
	if guild == nil {
		return nil, ErrGuildNotFound
	}

	// For private guilds, verify membership
	if guild.Visibility == model.GuildVisibilityPrivate {
		isMember, err := s.guildRepo.IsMember(ctx, userID, guildID)
		if err != nil {
			return nil, fmt.Errorf("checking membership: %w", err)
		}
		if !isMember {
			return nil, ErrNotGuildMember
		}
	}

	return guild, nil
}

// GetGuildWithMembers retrieves a guild with its members
func (s *GuildService) GetGuildWithMembers(ctx context.Context, userID, guildID string) (*model.GuildData, error) {
	guild, err := s.GetGuild(ctx, userID, guildID)
	if err != nil {
		return nil, err
	}

	members, err := s.guildRepo.GetMembers(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("getting members: %w", err)
	}

	memberSlice := make([]model.Member, len(members))
	for i, m := range members {
		memberSlice[i] = *m
	}

	return &model.GuildData{
		Guild:   *guild,
		Members: memberSlice,
	}, nil
}

// ListUserGuilds lists all guilds a user is a member of
func (s *GuildService) ListUserGuilds(ctx context.Context, userID string) ([]*model.Guild, error) {
	guilds, err := s.guildRepo.GetGuildsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing guilds: %w", err)
	}
	return guilds, nil
}

// UpdateGuildRequest represents a request to update a guild
type UpdateGuildRequest struct {
	Name        *string
	Description *string
	Icon        *string
	Color       *string
	Visibility  *string
}

// UpdateGuild updates a guild (requires membership)
func (s *GuildService) UpdateGuild(ctx context.Context, userID, guildID string, req UpdateGuildRequest) (*model.Guild, error) {
	// Verify membership
	isMember, err := s.guildRepo.IsMember(ctx, userID, guildID)
	if err != nil {
		return nil, fmt.Errorf("checking membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotGuildMember
	}

	// Get current guild
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("getting guild: %w", err)
	}
	if guild == nil {
		return nil, ErrGuildNotFound
	}

	// Apply updates
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrGuildNameRequired
		}
		if len(name) > model.MaxGuildNameLength {
			return nil, ErrGuildNameTooLong
		}
		guild.Name = name
	}

	if req.Description != nil {
		if len(*req.Description) > model.MaxGuildDescLength {
			return nil, ErrGuildDescTooLong
		}
		guild.Description = *req.Description
	}

	if req.Icon != nil {
		guild.Icon = *req.Icon
	}

	if req.Color != nil {
		guild.Color = *req.Color
	}

	if req.Visibility != nil {
		guild.Visibility = *req.Visibility
	}

	if err := s.guildRepo.Update(ctx, guild); err != nil {
		return nil, fmt.Errorf("updating guild: %w", err)
	}

	return guild, nil
}

// JoinGuild allows a user to join a public guild
func (s *GuildService) JoinGuild(ctx context.Context, userID, guildID string) error {
	// Get guild
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("getting guild: %w", err)
	}
	if guild == nil {
		return ErrGuildNotFound
	}

	// Check if already a member
	isMember, err := s.guildRepo.IsMember(ctx, userID, guildID)
	if err != nil {
		return fmt.Errorf("checking membership: %w", err)
	}
	if isMember {
		return ErrAlreadyGuildMember
	}

	// Check member limit
	memberCount, err := s.guildRepo.CountMembers(ctx, guildID)
	if err != nil {
		return fmt.Errorf("counting members: %w", err)
	}
	if memberCount >= model.MaxMembersPerGuild {
		return ErrMaxMembersReached
	}

	// Check user's guild limit
	guildCount, err := s.guildRepo.CountGuildsForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("counting user guilds: %w", err)
	}
	if guildCount >= model.MaxGuildsPerUser {
		return ErrMaxGuildsReached
	}

	// Get user for member creation
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Get or create member
	member, err := s.memberRepo.GetOrCreate(ctx, userID, user.Email, user.Email)
	if err != nil {
		return fmt.Errorf("getting/creating member: %w", err)
	}

	// For private guilds, require pending approval
	pendingApproval := guild.Visibility == model.GuildVisibilityPrivate

	// Add member to guild
	if err := s.guildRepo.AddMember(ctx, member.ID, guildID, pendingApproval); err != nil {
		return fmt.Errorf("adding member: %w", err)
	}

	return nil
}

// LeaveGuild removes a user from a guild
func (s *GuildService) LeaveGuild(ctx context.Context, userID, guildID string) error {
	// Check membership
	isMember, err := s.guildRepo.IsMember(ctx, userID, guildID)
	if err != nil {
		return fmt.Errorf("checking membership: %w", err)
	}
	if !isMember {
		return ErrNotGuildMember
	}

	// Check if sole member
	memberCount, err := s.guildRepo.CountMembers(ctx, guildID)
	if err != nil {
		return fmt.Errorf("counting members: %w", err)
	}
	if memberCount <= 1 {
		return ErrCannotLeaveSoleMember
	}

	// Get member record
	member, err := s.memberRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting member: %w", err)
	}
	if member == nil {
		return ErrNotGuildMember
	}

	// Remove from guild
	if err := s.guildRepo.RemoveMember(ctx, member.ID, guildID); err != nil {
		return fmt.Errorf("removing member: %w", err)
	}

	return nil
}

// DeleteGuild deletes a guild (only allowed if sole member)
func (s *GuildService) DeleteGuild(ctx context.Context, userID, guildID string) error {
	// Check membership
	isMember, err := s.guildRepo.IsMember(ctx, userID, guildID)
	if err != nil {
		return fmt.Errorf("checking membership: %w", err)
	}
	if !isMember {
		return ErrNotGuildMember
	}

	// Delete guild (will cascade delete memberships)
	if err := s.guildRepo.Delete(ctx, guildID); err != nil {
		return fmt.Errorf("deleting guild: %w", err)
	}

	return nil
}

// GetMemberCount returns the number of members in a guild
func (s *GuildService) GetMemberCount(ctx context.Context, guildID string) (int, error) {
	count, err := s.guildRepo.CountMembers(ctx, guildID)
	if err != nil {
		return 0, fmt.Errorf("counting members: %w", err)
	}
	return count, nil
}

// IsMember checks if a user is a member of a guild
func (s *GuildService) IsMember(ctx context.Context, userID, guildID string) (bool, error) {
	return s.guildRepo.IsMember(ctx, userID, guildID)
}

// GetMemberRole returns a user's role in a guild
func (s *GuildService) GetMemberRole(ctx context.Context, userID, guildID string) (model.GuildRole, error) {
	return s.guildRepo.GetMemberRole(ctx, userID, guildID)
}

// IsGuildAdmin checks if a user has admin privileges in a guild
func (s *GuildService) IsGuildAdmin(ctx context.Context, userID, guildID string) (bool, error) {
	return s.guildRepo.IsGuildAdmin(ctx, userID, guildID)
}

// IsGuildModerator checks if a user has moderator privileges in a guild
func (s *GuildService) IsGuildModerator(ctx context.Context, userID, guildID string) (bool, error) {
	return s.guildRepo.IsGuildModerator(ctx, userID, guildID)
}

// UpdateMemberRole updates a user's role in a guild (requires admin)
func (s *GuildService) UpdateMemberRole(ctx context.Context, adminUserID, targetUserID, guildID string, newRole model.GuildRole) error {
	// Validate role
	if !newRole.IsValid() {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	// Check that the requesting user is an admin
	isAdmin, err := s.guildRepo.IsGuildAdmin(ctx, adminUserID, guildID)
	if err != nil {
		return fmt.Errorf("checking admin status: %w", err)
	}
	if !isAdmin {
		return ErrNotGuildAdmin
	}

	// Check that target is a member
	isMember, err := s.guildRepo.IsMember(ctx, targetUserID, guildID)
	if err != nil {
		return fmt.Errorf("checking membership: %w", err)
	}
	if !isMember {
		return ErrNotGuildMember
	}

	// Prevent demoting the last admin
	if newRole != model.GuildRoleAdmin {
		targetRole, err := s.guildRepo.GetMemberRole(ctx, targetUserID, guildID)
		if err != nil {
			return fmt.Errorf("getting target role: %w", err)
		}
		if targetRole.IsAdmin() {
			// Check if this is the last admin
			members, err := s.guildRepo.GetMembers(ctx, guildID)
			if err != nil {
				return fmt.Errorf("getting members: %w", err)
			}
			adminCount := 0
			for _, m := range members {
				role, err := s.guildRepo.GetMemberRole(ctx, m.UserID, guildID)
				if err == nil && role.IsAdmin() {
					adminCount++
				}
			}
			if adminCount <= 1 {
				return fmt.Errorf("cannot demote the last admin")
			}
		}
	}

	// Update the role
	if err := s.guildRepo.UpdateMemberRole(ctx, targetUserID, guildID, newRole); err != nil {
		return fmt.Errorf("updating role: %w", err)
	}

	return nil
}

// RequireGuildAdmin checks if a user is a guild admin and returns an error if not
func (s *GuildService) RequireGuildAdmin(ctx context.Context, userID, guildID string) error {
	isAdmin, err := s.guildRepo.IsGuildAdmin(ctx, userID, guildID)
	if err != nil {
		return fmt.Errorf("checking admin status: %w", err)
	}
	if !isAdmin {
		return ErrNotGuildAdmin
	}
	return nil
}
