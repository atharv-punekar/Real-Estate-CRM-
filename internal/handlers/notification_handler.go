package handlers

import (
	"strconv"

	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/gofiber/fiber/v2"
)

var notificationRepo = &repository.NotificationRepository{}

// GetNotifications returns paginated notifications for the current user
func GetNotifications(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	notifications, total, err := notificationRepo.FindByUser(userID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch notifications"})
	}

	return c.JSON(fiber.Map{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

// MarkNotificationRead marks a notification as read
func MarkNotificationRead(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	notificationID := c.Params("id")

	// Verify notification exists and belongs to user
	notification, err := notificationRepo.FindByID(notificationID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Notification not found"})
	}

	if notification.UserID != userID {
		return c.Status(403).JSON(fiber.Map{"error": "You don't have permission to access this notification"})
	}

	if err := notificationRepo.MarkAsRead(notificationID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to mark notification as read"})
	}

	return c.JSON(fiber.Map{"message": "Notification marked as read"})
}

// GetUnreadCount returns the count of unread notifications
func GetUnreadCount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	count, err := notificationRepo.FindUnreadCount(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch unread count"})
	}

	return c.JSON(fiber.Map{"unread_count": count})
}
