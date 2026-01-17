package model

import "time"

// UserRole represents the role of a user in the system
type UserRole string

const (
	UserRoleUser      UserRole = "user"      // Default role
	UserRoleModerator UserRole = "moderator" // Can review reports, issue warnings
	UserRoleAdmin     UserRole = "admin"     // Full access including bans, system settings
)

// User represents a user account
type User struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Username      *string    `json:"username,omitempty"`
	Hash          *string    `json:"-"` // Never expose password hash
	Firstname     *string    `json:"firstname,omitempty"`
	Lastname      *string    `json:"lastname,omitempty"`
	Role          UserRole   `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	CreatedOn     time.Time  `json:"created_on"`
	UpdatedOn     time.Time  `json:"updated_on"`
	LoginOn       *time.Time `json:"login_on,omitempty"`
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsModerator returns true if the user has moderator or admin role
func (u *User) IsModerator() bool {
	return u.Role == UserRoleModerator || u.Role == UserRoleAdmin
}

// CanModerate returns true if the user can perform moderation actions
func (u *User) CanModerate() bool {
	return u.IsModerator()
}

// Identity represents a linked OAuth provider
type Identity struct {
	ID                      string    `json:"id"`
	UserID                  string    `json:"user_id"`
	Provider                string    `json:"provider"` // "google", "apple"
	ProviderUserID          string    `json:"provider_user_id"`
	ProviderEmail           *string   `json:"provider_email,omitempty"`
	EmailVerifiedByProvider bool      `json:"email_verified_by_provider"`
	CreatedOn               time.Time `json:"created_on"`
	UpdatedOn               time.Time `json:"updated_on"`
}

// Passkey represents a WebAuthn credential
type Passkey struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	CredentialID string     `json:"credential_id"`
	PublicKey    []byte     `json:"-"` // Don't expose in API
	SignCount    uint32     `json:"sign_count"`
	Name         string     `json:"name"` // "iPhone 15", "MacBook Pro"
	CreatedOn    time.Time  `json:"created_on"`
	LastUsedOn   *time.Time `json:"last_used_on,omitempty"`
}

// UserWithIdentities includes user with their linked identities and passkeys
type UserWithIdentities struct {
	User       *User       `json:"user"`
	Identities []*Identity `json:"identities"`
	Passkeys   []*Passkey  `json:"passkeys"`
}

// TokenClaims represents extracted JWT claims
type TokenClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
}

// PasskeyPublic represents a passkey for API responses (without sensitive data)
type PasskeyPublic struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CreatedOn  time.Time  `json:"created_on"`
	LastUsedOn *time.Time `json:"last_used_on,omitempty"`
}

// ToPublic converts a Passkey to its public representation
func (p *Passkey) ToPublic() *PasskeyPublic {
	return &PasskeyPublic{
		ID:         p.ID,
		Name:       p.Name,
		CreatedOn:  p.CreatedOn,
		LastUsedOn: p.LastUsedOn,
	}
}
