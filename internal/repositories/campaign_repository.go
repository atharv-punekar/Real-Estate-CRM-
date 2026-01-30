package repository

import (
	"fmt"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
)

type CampaignRepository struct{}

// Create creates a new campaign
func (r *CampaignRepository) Create(campaign *models.Campaign) error {
	// return database.DB.Create(campaign).Error
	err := database.DB.Create(campaign).Error
	if err != nil {
		fmt.Println("CAMPAIGN DB ERROR:", err)
	}
	return err

}

// FindByID finds a campaign by ID within an organization
func (r *CampaignRepository) FindByID(id, orgID string) (*models.Campaign, error) {
	var campaign models.Campaign
	if err := database.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&campaign).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

// FindByIDOnly finds a campaign by ID without organization constraint (for scheduler use)
func (r *CampaignRepository) FindByIDOnly(id string) (*models.Campaign, error) {
	var campaign models.Campaign
	if err := database.DB.Where("id = ?", id).First(&campaign).Error; err != nil {
		return nil, err
	}
	return &campaign, nil
}

// FindAllByOrg returns paginated campaigns for an organization with optional status filter
func (r *CampaignRepository) FindAllByOrg(orgID string, status string, page, limit int) ([]models.Campaign, int64, error) {
	var campaigns []models.Campaign
	var total int64

	query := database.DB.Where("organization_id = ?", orgID)

	// Add status filter if provided
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Model(&models.Campaign{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// Update updates a campaign
func (r *CampaignRepository) Update(campaign *models.Campaign) error {
	return database.DB.Save(campaign).Error
}

// Delete deletes a campaign
func (r *CampaignRepository) Delete(id, orgID string) error {
	return database.DB.Where("id = ? AND organization_id = ?", id, orgID).Delete(&models.Campaign{}).Error
}

// FindScheduledCampaigns finds campaigns that are scheduled and ready to run
func (r *CampaignRepository) FindScheduledCampaigns(currentTime time.Time) ([]models.Campaign, error) {
	var campaigns []models.Campaign
	// Only find campaigns that:
	// 1. Have status = 'scheduled' (not completed, not failed, not running)
	// 2. Are due (scheduled_at <= now)
	// 3. Have valid organization_id (not NULL)
	err := database.DB.
		Where("status = ?", "scheduled").
		Where("scheduled_at <= ?", currentTime).
		Where("last_run_at IS NULL"). // <-- IMPORTANT FIX
		Not("organization_id IS NULL").
		Find(&campaigns).Error
	return campaigns, err
}

// FindRecurringCampaigns finds active recurring campaigns
func (r *CampaignRepository) FindRecurringCampaigns() ([]models.Campaign, error) {
	var campaigns []models.Campaign
	// Only find recurring campaigns that:
	// 1. Have schedule_type = 'recurring'
	// 2. Are in scheduled or running state
	// 3. Have valid organization_id
	err := database.DB.
		Where("schedule_type = ?", "recurring").
		Where("status IN ?", []string{"scheduled", "running"}).
		Not("organization_id IS NULL").
		Find(&campaigns).Error
	return campaigns, err
}

// UpdateStatus updates only the status of a campaign
func (r *CampaignRepository) UpdateStatus(id, status string) error {
	return database.DB.Model(&models.Campaign{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateLastRunAt updates the last_run_at timestamp
func (r *CampaignRepository) UpdateLastRunAt(id string, lastRunAt time.Time) error {
	return database.DB.Model(&models.Campaign{}).Where("id = ?", id).Update("last_run_at", lastRunAt).Error
}

// GetRecipientContacts returns all contact IDs for a campaign (from audiences or single contact)
func (r *CampaignRepository) GetRecipientContacts(campaign *models.Campaign) ([]string, error) {
	var contactIDs []string

	// Single contact
	if campaign.ContactID != nil {
		return []string{*campaign.ContactID}, nil
	}

	// Multiple audiences
	if len(campaign.AudienceIDs) > 0 {
		var ids []string
		err := database.DB.Table("audience_contact").
			Select("DISTINCT contact_id").
			Where("audience_id IN ?", []string(campaign.AudienceIDs)).
			Pluck("contact_id", &ids).Error

		if err != nil {
			return nil, err
		}

		return ids, nil
	}

	return contactIDs, nil
}
