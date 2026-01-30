package repository

import (
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"gorm.io/gorm"
)

type ContactRepository struct{}

// Create creates a new contact
func (r *ContactRepository) Create(contact *models.Contact) error {
	return database.DB.Create(contact).Error
}

// FindByID finds a contact by ID within an organization
func (r *ContactRepository) FindByID(id, orgID string) (*models.Contact, error) {
	var contact models.Contact
	if err := database.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&contact).Error; err != nil {
		return nil, err
	}
	return &contact, nil
}

// FindAllByOrg returns paginated contacts for an organization with optional search
func (r *ContactRepository) FindAllByOrg(orgID string, page, limit int, search string) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	query := database.DB.Where("organization_id = ?", orgID)

	// Add search filter if provided
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Count total
	if err := query.Model(&models.Contact{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

// Update updates a contact
func (r *ContactRepository) Update(contact *models.Contact) error {
	return database.DB.Save(contact).Error
}

// Delete soft deletes a contact (sets is_active to false)
func (r *ContactRepository) Delete(id, orgID string) error {
	return database.DB.Model(&models.Contact{}).
		Where("id = ? AND organization_id = ?", id, orgID).
		Update("is_active", false).Error
}

// FindByEmailOrPhone checks if a contact with the given email or phone exists in the organization
func (r *ContactRepository) FindByEmailOrPhone(email, phone, orgID string) (*models.Contact, error) {
	var contact models.Contact
	query := database.DB.Where("organization_id = ?", orgID)

	if email != "" && phone != "" {
		query = query.Where("email = ? OR phone = ?", email, phone)
	} else if email != "" {
		query = query.Where("email = ?", email)
	} else if phone != "" {
		query = query.Where("phone = ?", phone)
	} else {
		return nil, gorm.ErrRecordNotFound
	}

	if err := query.First(&contact).Error; err != nil {
		return nil, err
	}
	return &contact, nil
}

// BulkCreate creates multiple contacts in a transaction
func (r *ContactRepository) BulkCreate(contacts []models.Contact) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, contact := range contacts {
			if err := tx.Create(&contact).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByIDs finds multiple contacts by their IDs within an organization
func (r *ContactRepository) FindByIDs(ids []string, orgID string) ([]models.Contact, error) {
	var contacts []models.Contact
	if err := database.DB.Where("id IN ? AND organization_id = ?", ids, orgID).Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

// FindWithFilter finds contacts based on dynamic filter criteria
func (r *ContactRepository) FindWithFilter(f models.ContactFilter) ([]models.Contact, error) {
	query := database.DB.Where("organization_id = ?", f.OrganizationID)

	// Property Type (case-insensitive)
	if len(f.PropertyType) > 0 {
		query = query.Where("LOWER(property_type) IN ?", toLower(f.PropertyType))
	}

	// Bedrooms
	if len(f.Bedrooms) > 0 {
		query = query.Where("bedrooms IN ?", f.Bedrooms)
	}

	// Bathrooms
	if len(f.Bathrooms) > 0 {
		query = query.Where("bathrooms IN ?", f.Bathrooms)
	}

	// Preferred Location (case-insensitive)
	if len(f.Locations) > 0 {
		query = query.Where("LOWER(preferred_location) IN ?", toLower(f.Locations))
	}

	// Budget logic (correct overlap)
	if f.MinBudget > 0 {
		query = query.Where("budget_max >= ?", f.MinBudget)
	}
	if f.MaxBudget > 0 {
		query = query.Where("budget_min <= ?", f.MaxBudget)
	}

	var contacts []models.Contact
	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}

	return contacts, nil
}

func toLower(list []string) []string {
	out := make([]string, len(list))
	for i, v := range list {
		out[i] = strings.ToLower(v)
	}
	return out
}
