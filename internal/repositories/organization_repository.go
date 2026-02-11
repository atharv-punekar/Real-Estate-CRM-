package repository

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"

	"github.com/google/uuid"
)

type OrganizationRepository struct{}

func (OrganizationRepository) Create(org *models.Organization) error {
	return database.DB.Create(org).Error
}

func (OrganizationRepository) FindAll() ([]models.Organization, error) {
	var orgs []models.Organization
	err := database.DB.Find(&orgs).Error
	return orgs, err
}

func (OrganizationRepository) Update(org *models.Organization) error {
	return database.DB.Save(org).Error
}

func (OrganizationRepository) FindByID(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := database.DB.First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByName finds an organization by name (case-insensitive)
func (OrganizationRepository) FindByName(name string) (*models.Organization, error) {
	var org models.Organization
	err := database.DB.Where("LOWER(name) = LOWER(?)", name).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByNameExcluding finds an organization by name, excluding a specific ID (for update validation)
func (OrganizationRepository) FindByNameExcluding(name string, excludeID uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := database.DB.Where("LOWER(name) = LOWER(?) AND id != ?", name, excludeID).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}
