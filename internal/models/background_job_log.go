package models

import (
	"time"

	"gorm.io/datatypes"
)

type BackgroundJobLog struct {
	ID string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	JobType        string  // csv_import | campaign_run | campaign_scheduler
	OrganizationID string  `gorm:"type:uuid"`
	ReferenceID    *string `gorm:"type:uuid"`

	Status string // queued, running, success, failed

	TotalRecords     *int
	ProcessedRecords *int
	Details          datatypes.JSON `gorm:"type:jsonb"`
	ErrorMessage     string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (BackgroundJobLog) TableName() string {
	return "background_job_log"
}
