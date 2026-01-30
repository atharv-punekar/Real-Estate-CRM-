package models

import (
	"time"

	"gorm.io/datatypes"
)

type Campaign struct {
	ID             string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OrganizationID string `gorm:"type:uuid"`

	Name       string
	TemplateID string `gorm:"type:uuid"`

	AudienceIDs datatypes.JSONSlice[string] `gorm:"type:jsonb;default:'[]'"`
	ContactID   *string                     `gorm:"type:uuid"`

	ScheduleType string // once | recurring
	ScheduledAt  time.Time

	Recurrence           *string
	RecurrenceDayOfWeek  *int
	RecurrenceDayOfMonth *int
	RecurrenceTime       *time.Time
	LastRunAt            *time.Time `gorm:"column:last_run_at"`

	Status    string
	CreatedBy string `gorm:"type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Campaign) TableName() string {
	return "campaign"
}
