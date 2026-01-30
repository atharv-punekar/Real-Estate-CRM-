package repository

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
)

type NotificationRepository struct{}

// Create creates a new notification
func (r *NotificationRepository) Create(notification *models.Notification) error {
	return database.DB.Create(notification).Error
}

// FindByUser returns paginated notifications for a user
func (r *NotificationRepository) FindByUser(userID string, page, limit int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := database.DB.Where("user_id = ?", userID)

	// Count total
	if err := query.Model(&models.Notification{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(notificationID string) error {
	return database.DB.Model(&models.Notification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": "NOW()",
		}).Error
}

// FindUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepository) FindUnreadCount(userID string) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// FindByID finds a notification by ID
func (r *NotificationRepository) FindByID(id string) (*models.Notification, error) {
	var notification models.Notification
	if err := database.DB.Where("id = ?", id).First(&notification).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

// CreateBulk creates multiple notifications at once
func (r *NotificationRepository) CreateBulk(notifications []models.Notification) error {
	return database.DB.Create(&notifications).Error
}

// FindByOrg returns all notifications for an organization
func (r *NotificationRepository) FindByOrg(orgID string, page, limit int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := database.DB.Where("organization_id = ?", orgID)

	// Count total
	if err := query.Model(&models.Notification{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}
