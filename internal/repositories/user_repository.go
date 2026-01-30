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
