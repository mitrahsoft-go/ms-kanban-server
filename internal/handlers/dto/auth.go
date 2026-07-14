package dto

import "fmt"

type Role string

const (
	RoleSuperAdmin     Role = "super_admin"
	RoleOrgAdmin       Role = "org_admin"
	RoleProjectManager Role = "project_manager"
	RoleDeveloper      Role = "developer"
	RoleViewer         Role = "viewer"
)

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type SignUpRequest struct {
	Email          string  `json:"email" validate:"required,email"`
	Password       string  `json:"password" validate:"required"`
	FullName       string  `json:"full_name" validate:"required"`
	OrganizationID string  `json:"organization_id,omitempty"`
	Role           Role    `json:"role" gorm:"size:30"`
	AvatarURL      *string `json:"avatar_url" gorm:"size:255"`
	Timezone       string  `json:"timezone" gorm:"size:50;default:'UTC'"`
}

func (r Role) Validate() error {
	switch r {
	case RoleSuperAdmin,
		RoleOrgAdmin,
		RoleProjectManager,
		RoleDeveloper,
		RoleViewer:
		return nil
	default:
		return fmt.Errorf("invalid role: %s", r)
	}
}
