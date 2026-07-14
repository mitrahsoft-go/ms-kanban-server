package models

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleSuperAdmin     Role = "super_admin"
	RoleOrgAdmin       Role = "org_admin"
	RoleProjectManager Role = "project_manager"
	RoleDeveloper      Role = "developer"
	RoleViewer         Role = "viewer"
)

type Organization struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:50;not null;unique"`
	Domain    *string        `json:"domain" validate:"required" gorm:"size:150;not null;unique"`
	LogoURL   *string        `json:"logo_url" validate:"required" gorm:"size:150;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type User struct {
	ID             uuid.UUID
	OrganizationID *uuid.UUID     `json:"organization_id,omitempty" gorm:"index:idx_users_organization_id"`
	Organization   Organization   `json:"organization"`
	FullName       string         `json:"name" gorm:"size:100;not null;unique"`
	Email          string         `json:"email" validate:"required,email" gorm:"size:100;not null;unique;index:idx_users_email"`
	PasswordHash   string         `json:"password_hash" validate:"required"`
	Role           string         `json:"role" gorm:"size:30;index:idx_users_role"`
	AvatarURL      *string        `json:"avatar_url" gorm:"size:255"`
	Timezone       string         `json:"timezone" gorm:"size:50;default:'UTC'"`
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index:idx_users_deleted_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return
}
