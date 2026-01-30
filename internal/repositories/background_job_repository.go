package repository

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
)

type BackgroundJobRepository struct{}

// Create creates a new background job log
func (r *BackgroundJobRepository) Create(job *models.BackgroundJobLog) error {
	return database.DB.Create(job).Error
}

// FindByID finds a background job by ID
func (r *BackgroundJobRepository) FindByID(id string) (*models.BackgroundJobLog, error) {
	var job models.BackgroundJobLog
	if err := database.DB.Where("id = ?", id).First(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// Update updates a background job
func (r *BackgroundJobRepository) Update(job *models.BackgroundJobLog) error {
	return database.DB.Save(job).Error
}

// FindByType finds jobs by type and optional status
func (r *BackgroundJobRepository) FindByType(jobType, status string) ([]models.BackgroundJobLog, error) {
	var jobs []models.BackgroundJobLog
	query := database.DB.Where("job_type = ?", jobType)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

// UpdateStatus updates the status of a job
func (r *BackgroundJobRepository) UpdateStatus(id, status, errorMessage string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": "NOW()",
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	return database.DB.Model(&models.BackgroundJobLog{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateProgress updates the processed records count
func (r *BackgroundJobRepository) UpdateProgress(id string, processedRecords int) error {
	return database.DB.Model(&models.BackgroundJobLog{}).
		Where("id = ?", id).
		Update("processed_records", processedRecords).Error
}

// FindByOrg finds jobs by organization
func (r *BackgroundJobRepository) FindByOrg(orgID string, page, limit int) ([]models.BackgroundJobLog, int64, error) {
	var jobs []models.BackgroundJobLog
	var total int64

	query := database.DB.Where("organization_id = ?", orgID)

	// Count total
	if err := query.Model(&models.BackgroundJobLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}
