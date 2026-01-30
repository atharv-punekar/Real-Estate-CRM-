package models

import "time"

type AudienceContact struct {
	ID         string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AudienceID string `gorm:"type:uuid;index"`
	ContactID  string `gorm:"type:uuid;index"`
	AddedAt    time.Time
}

func (AudienceContact) TableName() string {
	return "audience_contact"
}
