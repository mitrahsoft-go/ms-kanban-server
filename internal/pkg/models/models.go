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

type RefreshToken struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey"`
	UserID    uuid.UUID      `json:"user_id" gorm:"index:idx_refresh_tokens_user_id;not null;unique"`
	TokenHash string         `json:"token_hash" gorm:"size:255;not null;unique"`
	UserAgent *string        `json:"user_agent,omitempty" gorm:"type:text"`
	IPAddress *string        `json:"ip_address,omitempty" gorm:"size:45"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	RevokedAt *time.Time     `json:"revoked_at,omitempty"`
	CreatedAt time.Time      `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index:idx_refresh_tokens_deleted_at"`
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

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return
}
