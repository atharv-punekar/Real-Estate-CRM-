package models

import "time"

type SuperAdmin struct {
	ID           string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name         string
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	CreatedAt    time.Time
}

func (SuperAdmin) TableName() string {
	return "super_admin"
}
