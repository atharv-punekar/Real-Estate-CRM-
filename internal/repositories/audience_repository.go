package repository

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"gorm.io/gorm"
)

type AudienceRepository struct{}

// Create creates a new audience
func (r *AudienceRepository) Create(audience *models.Audience) error {
	return database.DB.Create(audience).Error
}

// FindByID finds an audience by ID within an organization
func (r *AudienceRepository) FindByID(id, orgID string) (*models.Audience, error) {
	var audience models.Audience
	if err := database.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&audience).Error; err != nil {
		return nil, err
	}
	return &audience, nil
}

// FindAllByOrg returns all audiences for an organization
func (r *AudienceRepository) FindAllByOrg(orgID string) ([]models.Audience, error) {
	var audiences []models.Audience
	if err := database.DB.Where("organization_id = ?", orgID).Order("created_at DESC").Find(&audiences).Error; err != nil {
		return nil, err
	}
	return audiences, nil
}

// Update updates an audience
func (r *AudienceRepository) Update(audience *models.Audience) error {
	return database.DB.Save(audience).Error
}

// Delete deletes an audience
func (r *AudienceRepository) Delete(id, orgID string) error {
	return database.DB.Where("id = ? AND organization_id = ?", id, orgID).Delete(&models.Audience{}).Error
}

// AddContacts adds multiple contacts to an audience
func (r *AudienceRepository) AddContacts(audienceID string, contactIDs []string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, contactID := range contactIDs {
			audienceContact := models.AudienceContact{
				AudienceID: audienceID,
				ContactID:  contactID,
			}
			// Use FirstOrCreate to avoid duplicates
			if err := tx.Where("audience_id = ? AND contact_id = ?", audienceID, contactID).
				FirstOrCreate(&audienceContact).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveContacts removes multiple contacts from an audience
func (r *AudienceRepository) RemoveContacts(audienceID string, contactIDs []string) error {
	return database.DB.Where("audience_id = ? AND contact_id IN ?", audienceID, contactIDs).
		Delete(&models.AudienceContact{}).Error
}

// FindContactsByAudience returns paginated contacts for an audience
func (r *AudienceRepository) FindContactsByAudience(audienceID string, page, limit int) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	// Count total
	if err := database.DB.Model(&models.Contact{}).
		Joins("JOIN audience_contact ON audience_contact.contact_id = contact.id").
		Where("audience_contact.audience_id = ?", audienceID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := database.DB.
		Joins("JOIN audience_contact ON audience_contact.contact_id = contact.id").
		Where("audience_contact.audience_id = ?", audienceID).
		Offset(offset).Limit(limit).
		Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

// CountContactsByAudience returns the count of contacts in an audience
func (r *AudienceRepository) CountContactsByAudience(audienceID string) (int64, error) {
	var count int64
	err := database.DB.Model(&models.AudienceContact{}).
		Where("audience_id = ?", audienceID).
		Count(&count).Error
	return count, err
}

func (r *AudienceRepository) DB() *gorm.DB {
	return database.DB
}
