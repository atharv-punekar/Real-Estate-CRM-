package repository

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
)

type EmailTemplateRepository struct{}

// Create creates a new email template
func (r *EmailTemplateRepository) Create(template *models.EmailTemplate) error {
	return database.DB.Create(template).Error
}

// FindByID finds an email template by ID within an organization, filtered by creator
func (r *EmailTemplateRepository) FindByID(id, orgID, createdBy string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	if err := database.DB.Where("id = ? AND organization_id = ? AND created_by = ?", id, orgID, createdBy).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

// FindAllByOrg returns all email templates for an organization filtered by creator
func (r *EmailTemplateRepository) FindAllByOrg(orgID, createdBy string) ([]models.EmailTemplate, error) {
	var templates []models.EmailTemplate
	if err := database.DB.Where("organization_id = ? AND created_by = ?", orgID, createdBy).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// Update updates an email template
func (r *EmailTemplateRepository) Update(template *models.EmailTemplate) error {
	return database.DB.Save(template).Error
}

// Delete deletes an email template
func (r *EmailTemplateRepository) Delete(id, orgID string) error {
	return database.DB.Where("id = ? AND organization_id = ?", id, orgID).Delete(&models.EmailTemplate{}).Error
}

// FindByNameAndOrg finds a template by name within an organization (case-insensitive)
func (r *EmailTemplateRepository) FindByNameAndOrg(name, orgID string) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	err := database.DB.Where("LOWER(name) = LOWER(?) AND organization_id = ?", name, orgID).First(&template).Error
	return &template, err
}
