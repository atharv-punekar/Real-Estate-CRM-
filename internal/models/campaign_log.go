package models

import "time"

type CampaignLog struct {
	ID         string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CampaignID string `gorm:"type:uuid"`
	ContactID  string `gorm:"type:uuid"`

	RecipientEmail string
	Subject        string

	Status       string // queued, sent, failed
	ErrorMessage string

	SentAt    *time.Time
	CreatedAt time.Time
}

func (CampaignLog) TableName() string {
	return "campaign_log"
}
