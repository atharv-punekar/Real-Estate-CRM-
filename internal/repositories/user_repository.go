package repository

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/google/uuid"
)

type UserRepository struct{}

func (r *UserRepository) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAllByOrg(orgID uuid.UUID) ([]models.User, error) {
	var users []models.User
	if err := database.DB.Where("organization_id = ?", orgID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindAllByOrgPaginated returns paginated users for an organization
func (r *UserRepository) FindAllByOrgPaginated(orgID uuid.UUID, page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Count total
	if err := database.DB.Model(&models.User{}).Where("organization_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := database.DB.Where("organization_id = ?", orgID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
}

func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return database.DB.Save(user).Error
}

func (r *UserRepository) FindAdminsByOrg(orgID uuid.UUID) ([]models.User, error) {
	var admins []models.User
	err := database.DB.Where("organization_id = ? AND role = ?", orgID, "org_admin").Find(&admins).Error
	return admins, err
}

func (r *UserRepository) FindByInviteToken(token string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("invite_token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByOrgAndRole(orgID uuid.UUID, role string) ([]models.User, error) {
	var users []models.User
	err := database.DB.Where("organization_id = ? AND role = ?", orgID, role).Find(&users).Error
	return users, err
}

// DeactivateAllByOrg sets is_active to false for all users in an organization
func (r *UserRepository) DeactivateAllByOrg(orgID uuid.UUID) error {
	return database.DB.Model(&models.User{}).
		Where("organization_id = ?", orgID).
		Update("is_active", false).Error
}

// ReactivateAllByOrg sets is_active to true for all users in an organization
func (r *UserRepository) ReactivateAllByOrg(orgID uuid.UUID) error {
	return database.DB.Model(&models.User{}).
		Where("organization_id = ?", orgID).
		Update("is_active", true).Error
}

// FindByNameAndOrg finds a user by name within a specific organization (case-insensitive)
func (r *UserRepository) FindByNameAndOrg(name string, orgID uuid.UUID) (*models.User, error) {
	var user models.User
	err := database.DB.Where("LOWER(name) = LOWER(?) AND organization_id = ?", name, orgID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByNameAndOrgExcluding finds a user by name within an organization, excluding a specific user ID
func (r *UserRepository) FindByNameAndOrgExcluding(name string, orgID uuid.UUID, excludeID uuid.UUID) (*models.User, error) {
	var user models.User
	err := database.DB.Where("LOWER(name) = LOWER(?) AND organization_id = ? AND id != ?", name, orgID, excludeID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
