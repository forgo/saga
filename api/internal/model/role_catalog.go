package model

import "time"

// RoleCatalogScopeType determines if the catalog is guild-owned or user-owned
type RoleCatalogScopeType string

const (
	RoleCatalogScopeGuild RoleCatalogScopeType = "guild"
	RoleCatalogScopeUser  RoleCatalogScopeType = "user"
)

// RoleCatalogRoleType determines if roles are for events or rideshares
type RoleCatalogRoleType string

const (
	RoleCatalogRoleEvent     RoleCatalogRoleType = "event"
	RoleCatalogRoleRideshare RoleCatalogRoleType = "rideshare"
)

// RoleCatalog represents a reusable role template
type RoleCatalog struct {
	ID          string               `json:"id"`
	ScopeType   RoleCatalogScopeType `json:"scope_type"`   // guild or user
	ScopeID     string               `json:"scope_id"`     // guild:<id> or user:<id>
	RoleType    RoleCatalogRoleType  `json:"role_type"`    // event or rideshare
	Name        string               `json:"name"`         // e.g., "DJ", "Dessert Bringer"
	Description *string              `json:"description,omitempty"`
	Icon        *string              `json:"icon,omitempty"` // Emoji or icon name
	IsActive    bool                 `json:"is_active"`
	CreatedBy   string               `json:"created_by"` // User ID
	CreatedOn   time.Time            `json:"created_on"`
	UpdatedOn   time.Time            `json:"updated_on"`
}

// RideshareRole represents a role for a specific rideshare (mirrors EventRole)
type RideshareRole struct {
	ID            string    `json:"id"`
	RideshareID   string    `json:"rideshare_id"`
	CatalogRoleID *string   `json:"catalog_role_id,omitempty"` // If created from catalog
	Name          string    `json:"name"`
	Description   *string   `json:"description,omitempty"`
	MaxSlots      int       `json:"max_slots"`    // Default 1, 0 = unlimited
	FilledSlots   int       `json:"filled_slots"` // Computed from assignments
	SortOrder     int       `json:"sort_order"`
	CreatedBy     string    `json:"created_by"` // User ID
	CreatedOn     time.Time `json:"created_on"`
	UpdatedOn     time.Time `json:"updated_on"`
}

// RideshareRoleAssignment represents a user's assignment to a rideshare role
type RideshareRoleAssignment struct {
	ID          string    `json:"id"`
	RideshareID string    `json:"rideshare_id"`
	RoleID      string    `json:"role_id"`
	UserID      string    `json:"user_id"`
	Note        *string   `json:"note,omitempty"` // User's note about their contribution
	Status      string    `json:"status"`         // pending, confirmed, cancelled
	AssignedOn  time.Time `json:"assigned_on"`
	UpdatedOn   time.Time `json:"updated_on"`
	// Populated by joins
	RoleName *string `json:"role_name,omitempty"`
}

// RideshareRoleWithAssignments includes the role and its current assignments
type RideshareRoleWithAssignments struct {
	Role        RideshareRole             `json:"role"`
	Assignments []RideshareRoleAssignment `json:"assignments"`
	IsFull      bool                      `json:"is_full"`
	SpotsLeft   int                       `json:"spots_left"` // -1 if unlimited
}

// RideshareRolesOverview provides a summary of all roles for a rideshare
type RideshareRolesOverview struct {
	RideshareID string                         `json:"rideshare_id"`
	TotalSlots  int                            `json:"total_slots"`  // Total role slots available (-1 if unlimited)
	FilledSlots int                            `json:"filled_slots"` // Currently filled slots
	IsFull      bool                           `json:"is_full"`
	Roles       []RideshareRoleWithAssignments `json:"roles"`
}

// Constraints
const (
	MaxRoleCatalogsPerScope     = 50
	MaxRoleCatalogNameLength    = 50
	MaxRoleCatalogDescLength    = 500
	MaxRideshareRolesPerRideshare = 20
	MaxRideshareRoleNoteLength  = 200
)

// CreateRoleCatalogRequest represents a request to create a role template
type CreateRoleCatalogRequest struct {
	RoleType    string  `json:"role_type"`              // event or rideshare
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

// Validate checks if the create request is valid
func (r *CreateRoleCatalogRequest) Validate() []FieldError {
	var errors []FieldError

	if r.RoleType == "" {
		errors = append(errors, FieldError{Field: "role_type", Message: "role_type is required"})
	} else if r.RoleType != string(RoleCatalogRoleEvent) && r.RoleType != string(RoleCatalogRoleRideshare) {
		errors = append(errors, FieldError{Field: "role_type", Message: "role_type must be 'event' or 'rideshare'"})
	}
	if r.Name == "" {
		errors = append(errors, FieldError{Field: "name", Message: "name is required"})
	} else if len(r.Name) > MaxRoleCatalogNameLength {
		errors = append(errors, FieldError{Field: "name", Message: "name must be 50 characters or less"})
	}
	if r.Description != nil && len(*r.Description) > MaxRoleCatalogDescLength {
		errors = append(errors, FieldError{Field: "description", Message: "description must be 500 characters or less"})
	}

	return errors
}

// UpdateRoleCatalogRequest represents a request to update a role template
type UpdateRoleCatalogRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// Validate checks if the update request is valid
func (r *UpdateRoleCatalogRequest) Validate() []FieldError {
	var errors []FieldError

	if r.Name != nil {
		if *r.Name == "" {
			errors = append(errors, FieldError{Field: "name", Message: "name cannot be empty"})
		} else if len(*r.Name) > MaxRoleCatalogNameLength {
			errors = append(errors, FieldError{Field: "name", Message: "name must be 50 characters or less"})
		}
	}
	if r.Description != nil && len(*r.Description) > MaxRoleCatalogDescLength {
		errors = append(errors, FieldError{Field: "description", Message: "description must be 500 characters or less"})
	}

	return errors
}

// CreateRoleFromCatalogRequest represents a request to create a role from a catalog template
type CreateRoleFromCatalogRequest struct {
	CatalogRoleID string  `json:"catalog_role_id"`
	MaxSlots      *int    `json:"max_slots,omitempty"` // Override catalog default
	SortOrder     *int    `json:"sort_order,omitempty"`
}

// Validate checks if the request is valid
func (r *CreateRoleFromCatalogRequest) Validate() []FieldError {
	var errors []FieldError

	if r.CatalogRoleID == "" {
		errors = append(errors, FieldError{Field: "catalog_role_id", Message: "catalog_role_id is required"})
	}
	if r.MaxSlots != nil && *r.MaxSlots < 0 {
		errors = append(errors, FieldError{Field: "max_slots", Message: "max_slots must be 0 or greater"})
	}

	return errors
}

// CreateRideshareRoleRequest represents a request to create a rideshare role
type CreateRideshareRoleRequest struct {
	CatalogRoleID *string `json:"catalog_role_id,omitempty"` // Optional: create from catalog
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	MaxSlots      int     `json:"max_slots,omitempty"` // 0 = unlimited
}

// Validate checks if the create request is valid
func (r *CreateRideshareRoleRequest) Validate() []FieldError {
	var errors []FieldError

	if r.Name == "" {
		errors = append(errors, FieldError{Field: "name", Message: "name is required"})
	} else if len(r.Name) > MaxRoleCatalogNameLength {
		errors = append(errors, FieldError{Field: "name", Message: "name must be 50 characters or less"})
	}
	if r.Description != nil && len(*r.Description) > MaxRoleCatalogDescLength {
		errors = append(errors, FieldError{Field: "description", Message: "description must be 500 characters or less"})
	}
	if r.MaxSlots < 0 {
		errors = append(errors, FieldError{Field: "max_slots", Message: "max_slots must be 0 or greater"})
	}

	return errors
}

// UpdateRideshareRoleRequest represents a request to update a rideshare role
type UpdateRideshareRoleRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	MaxSlots    *int    `json:"max_slots,omitempty"`
}

// Validate checks if the update request is valid
func (r *UpdateRideshareRoleRequest) Validate() []FieldError {
	var errors []FieldError

	if r.Name != nil {
		if *r.Name == "" {
			errors = append(errors, FieldError{Field: "name", Message: "name cannot be empty"})
		} else if len(*r.Name) > MaxRoleCatalogNameLength {
			errors = append(errors, FieldError{Field: "name", Message: "name must be 50 characters or less"})
		}
	}
	if r.Description != nil && len(*r.Description) > MaxRoleCatalogDescLength {
		errors = append(errors, FieldError{Field: "description", Message: "description must be 500 characters or less"})
	}
	if r.MaxSlots != nil && *r.MaxSlots < 0 {
		errors = append(errors, FieldError{Field: "max_slots", Message: "max_slots must be 0 or greater"})
	}

	return errors
}

// AssignRideshareRoleRequest represents a request to assign oneself to a rideshare role
type AssignRideshareRoleRequest struct {
	RoleID string  `json:"role_id"`
	Note   *string `json:"note,omitempty"`
}

// Validate checks if the request is valid
func (r *AssignRideshareRoleRequest) Validate() []FieldError {
	var errors []FieldError

	if r.RoleID == "" {
		errors = append(errors, FieldError{Field: "role_id", Message: "role_id is required"})
	}
	if r.Note != nil && len(*r.Note) > MaxRideshareRoleNoteLength {
		errors = append(errors, FieldError{Field: "note", Message: "note must be 200 characters or less"})
	}

	return errors
}
