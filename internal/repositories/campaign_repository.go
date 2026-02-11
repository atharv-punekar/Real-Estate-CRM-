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

// FindByID finds a campaign by ID within an organization, filtered by creator
func (r *CampaignRepository) FindByID(id, orgID, createdBy string) (*models.Campaign, error) {
	var campaign models.Campaign
	if err := database.DB.Where("id = ? AND organization_id = ? AND created_by = ?", id, orgID, createdBy).First(&campaign).Error; err != nil {
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

// FindAllByOrg returns paginated campaigns for an organization filtered by creator with optional status filter and sorting
func (r *CampaignRepository) FindAllByOrg(orgID, createdBy, status string, page, limit int, sortBy, sortOrder string) ([]models.Campaign, int64, error) {
	var campaigns []models.Campaign
	var total int64

	// Filter by organization and creator
	query := database.DB.Where("organization_id = ? AND created_by = ?", orgID, createdBy)

	// Add status filter if provided
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total
	if err := query.Model(&models.Campaign{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build sort query
	allowedFields := []string{"created_at", "updated_at", "scheduled_at", "name", "status"}
	var orderClause string

	if sortBy != "" && sortOrder != "" {
		sortOrder = fmt.Sprintf("%s", sortOrder) // Use fmt to avoid import issues
		if sortOrder != "ASC" && sortOrder != "DESC" {
			sortOrder = "DESC"
		}

		// Validate sortBy is in allowed fields
		isAllowed := false
		for _, field := range allowedFields {
			if sortBy == field {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			orderClause = sortBy + " " + sortOrder
		} else {
			orderClause = "created_at DESC"
		}
	} else {
		orderClause = "created_at DESC"
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order(orderClause).Find(&campaigns).Error; err != nil {
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

// FindByNameAndOrg finds a campaign by name within an organization (case-insensitive)
func (r *CampaignRepository) FindByNameAndOrg(name, orgID string) (*models.Campaign, error) {
	var campaign models.Campaign
	err := database.DB.Where("LOWER(name) = LOWER(?) AND organization_id = ?", name, orgID).First(&campaign).Error
	return &campaign, err
}
