package models

import "time"

type Audience struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`

	Name        string
	Description string

	CreatedBy string `gorm:"type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Audience) TableName() string {
	return "audience"
}
