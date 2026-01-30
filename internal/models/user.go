package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	Name           string     `json:"name"`
	Email          string     `gorm:"unique;not null" json:"email"`
	PasswordHash   string     `json:"-"`
	Role           string     `json:"role"` // org_admin, org_user
	IsActive       bool       `json:"is_active"`
	InviteToken    *string    `gorm:"type:varchar(255);index" json:"-"`
	TokenExpiresAt *time.Time `json:"-"`
	IsPasswordSet  bool       `gorm:"default:false" json:"is_password_set"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}
