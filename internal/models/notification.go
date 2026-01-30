package models

import "time"

type Notification struct {
	ID string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	OrganizationID string `gorm:"type:uuid"`
	UserID         string `gorm:"type:uuid"`

	NotificationType string
	Title            string
	Message          string

	RelatedUserID     *string `gorm:"type:uuid"`
	RelatedCampaignID *string `gorm:"type:uuid"`

	IsRead    bool
	ReadAt    *time.Time
	CreatedAt time.Time
}

func (Notification) TableName() string {
	return "notification"
}
