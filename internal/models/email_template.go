package models

import "time"

type EmailTemplate struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`

	Name          string
	Subject       string
	Preheader     string
	FromName      string
	ReplyTo       string
	HtmlBody      string
	PlainTextBody string

	CreatedBy string `gorm:"type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (EmailTemplate) TableName() string {
	return "email_template"
}
