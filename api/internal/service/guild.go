package service

import (
	"context"

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
	RemoveMember(ctx context.Context, memberID, guildID string) error
	IsMember(ctx context.Context, userID, guildID string) (bool, error)
	CountMembers(ctx context.Context, guildID string) (int, error)
	GetMembers(ctx context.Context, guildID string) ([]*model.Member, error)
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
