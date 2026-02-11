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

// FindByID finds a contact by ID within an organization, filtered by creator
func (r *ContactRepository) FindByID(id, orgID, createdBy string) (*models.Contact, error) {
	var contact models.Contact
	if err := database.DB.Where("id = ? AND organization_id = ? AND created_by = ?", id, orgID, createdBy).First(&contact).Error; err != nil {
		return nil, err
	}
	return &contact, nil
}

// FindAllByOrg returns paginated contacts for an organization filtered by creator
func (r *ContactRepository) FindAllByOrg(orgID, createdBy string, page, limit int, search, sortBy, sortOrder string) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	// Filter by organization, creator, and active status
	query := database.DB.Where("organization_id = ? AND created_by = ? AND is_active = ?", orgID, createdBy, true)

	// Add search filter if provided (trim whitespace)
	if search != "" {
		search = strings.TrimSpace(search)
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

	// Build sort query
	allowedFields := []string{"created_at", "updated_at", "first_name", "last_name", "email", "budget_min", "budget_max"}
	var orderClause string

	if sortBy != "" && sortOrder != "" {
		sortOrder = strings.ToUpper(sortOrder)
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
	if err := query.Offset(offset).Limit(limit).Order(orderClause).Find(&contacts).Error; err != nil {
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
