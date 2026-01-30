package services

import (
	"fmt"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/google/uuid"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	userRepo         *repository.UserRepository
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		notificationRepo: &repository.NotificationRepository{},
		userRepo:         &repository.UserRepository{},
	}
}

// NotifyAgentAdded creates notifications for org admins when an agent is added
func (s *NotificationService) NotifyAgentAdded(orgID, addedUserID, agentName string) error {
	// Get all org admins
	admins, err := s.userRepo.FindAdminsByOrg(parseUUID(orgID))
	if err != nil {
		return err
	}

	var notifications []models.Notification
	for _, admin := range admins {
		notification := models.Notification{
			OrganizationID:   orgID,
			UserID:           admin.ID.String(),
			NotificationType: "agent_added",
			Title:            "New Agent Added",
			Message:          fmt.Sprintf("Agent %s has been added to your organization", agentName),
			RelatedUserID:    &addedUserID,
			IsRead:           false,
		}
		notifications = append(notifications, notification)
	}

	if len(notifications) > 0 {
		return s.notificationRepo.CreateBulk(notifications)
	}
	return nil
}

// NotifyAgentRemoved creates notifications when an agent is removed
func (s *NotificationService) NotifyAgentRemoved(orgID, removedUserID, agentName string) error {
	admins, err := s.userRepo.FindAdminsByOrg(parseUUID(orgID))
	if err != nil {
		return err
	}

	var notifications []models.Notification
	for _, admin := range admins {
		notification := models.Notification{
			OrganizationID:   orgID,
			UserID:           admin.ID.String(),
			NotificationType: "agent_removed",
			Title:            "Agent Removed",
			Message:          fmt.Sprintf("Agent %s has been removed from your organization", agentName),
			RelatedUserID:    &removedUserID,
			IsRead:           false,
		}
		notifications = append(notifications, notification)
	}

	if len(notifications) > 0 {
		return s.notificationRepo.CreateBulk(notifications)
	}
	return nil
}

// NotifyCampaignSent creates a notification when a campaign is sent
func (s *NotificationService) NotifyCampaignSent(orgID, userID, campaignID string, recipientCount int) error {
	notification := models.Notification{
		OrganizationID:    orgID,
		UserID:            userID,
		NotificationType:  "campaign_sent",
		Title:             "Campaign Sent Successfully",
		Message:           fmt.Sprintf("Your campaign has been sent to %d recipients", recipientCount),
		RelatedCampaignID: &campaignID,
		IsRead:            false,
	}
	return s.notificationRepo.Create(&notification)
}

// NotifyCSVImportCompleted creates a notification when CSV import completes
func (s *NotificationService) NotifyCSVImportCompleted(orgID, userID string, importedCount int) error {
	notification := models.Notification{
		OrganizationID:   orgID,
		UserID:           userID,
		NotificationType: "csv_import_completed",
		Title:            "CSV Import Completed",
		Message:          fmt.Sprintf("Successfully imported %d contacts from CSV", importedCount),
		IsRead:           false,
	}
	return s.notificationRepo.Create(&notification)
}

// NotifyCSVImportFailed creates a notification when CSV import fails
func (s *NotificationService) NotifyCSVImportFailed(orgID, userID, errorMessage string) error {
	notification := models.Notification{
		OrganizationID:   orgID,
		UserID:           userID,
		NotificationType: "csv_import_failed",
		Title:            "CSV Import Failed",
		Message:          fmt.Sprintf("CSV import failed: %s", errorMessage),
		IsRead:           false,
	}
	return s.notificationRepo.Create(&notification)
}

// Helper function to parse UUID string
func parseUUID(uuidStr string) uuid.UUID {
	parsed, _ := uuid.Parse(uuidStr)
	return parsed
}
