package repository

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
)

type CampaignLogRepository struct{}

// Create creates a new campaign log entry
func (r *CampaignLogRepository) Create(log *models.CampaignLog) error {
	return database.DB.Create(log).Error
}

// FindByCampaign returns paginated logs for a campaign
func (r *CampaignLogRepository) FindByCampaign(campaignID string, page, limit int) ([]models.CampaignLog, int64, error) {
	var logs []models.CampaignLog
	var total int64

	query := database.DB.Where("campaign_id = ?", campaignID)

	// Count total
	if err := query.Model(&models.CampaignLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// UpdateStatus updates the status of a log entry
func (r *CampaignLogRepository) UpdateStatus(logID, status, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	if status == "sent" {
		updates["sent_at"] = "NOW()"
	}
	return database.DB.Model(&models.CampaignLog{}).Where("id = ?", logID).Updates(updates).Error
}

// FindByID finds a campaign log by ID
func (r *CampaignLogRepository) FindByID(id string) (*models.CampaignLog, error) {
	var log models.CampaignLog
	if err := database.DB.Where("id = ?", id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetStatsByCampaign returns statistics for a campaign
func (r *CampaignLogRepository) GetStatsByCampaign(campaignID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Count by status
	var results []struct {
		Status string
		Count  int64
	}

	err := database.DB.Model(&models.CampaignLog{}).
		Select("status, COUNT(*) as count").
		Where("campaign_id = ?", campaignID).
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for _, result := range results {
		stats[result.Status] = result.Count
	}

	return stats, nil
}
