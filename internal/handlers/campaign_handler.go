package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

var (
	campaignRepo    = &repository.CampaignRepository{}
	campaignLogRepo = &repository.CampaignLogRepository{}
)

// CreateCampaign creates a new campaign
func CreateCampaign(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		Name                 string                      `json:"name"`
		TemplateID           string                      `json:"template_id"`
		AudienceIDs          datatypes.JSONSlice[string] `json:"audience_ids"`
		ContactID            *string                     `json:"contact_id"`
		ScheduleType         string                      `json:"schedule_type"` // once | recurring
		ScheduledAt          time.Time                   `json:"scheduled_at"`
		Recurrence           *string                     `json:"recurrence"`              // daily | weekly | monthly
		RecurrenceDayOfWeek  *int                        `json:"recurrence_day_of_week"`  // 0-6 (Sunday-Saturday)
		RecurrenceDayOfMonth *int                        `json:"recurrence_day_of_month"` // 1-31
		RecurrenceTime       *string                     `json:"recurrence_time"`         // HH:MM format
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate and normalize campaign name
	campaignName := strings.TrimSpace(req.Name)
	if err := utils.ValidateNotEmpty(campaignName, "Campaign name"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate required fields
	if req.TemplateID == "" || req.ScheduleType == "" {
		return c.Status(400).JSON(fiber.Map{"error": "template_id and schedule_type are required"})
	}

	// Validate schedule type
	if req.ScheduleType != "once" && req.ScheduleType != "recurring" {
		return c.Status(400).JSON(fiber.Map{"error": "schedule_type must be 'once' or 'recurring'"})
	}

	// Reject campaign creation for past dates
	if req.ScheduledAt.Before(time.Now()) {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot schedule campaign in the past. Please select a future date and time"})
	}

	// Validate recipients (must have either audience_ids or contact_id, not both)
	if (len(req.AudienceIDs) == 0 && req.ContactID == nil) || (len(req.AudienceIDs) > 0 && req.ContactID != nil) {
		return c.Status(400).JSON(fiber.Map{"error": "Must provide either audience_ids or contact_id, not both"})
	}

	// Verify template exists and belongs to the agent
	_, err := templateRepo.FindByID(req.TemplateID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Email template not found"})
	}

	// Verify audiences exist if provided
	if len(req.AudienceIDs) > 0 {
		for _, audienceID := range req.AudienceIDs {
			_, err := audienceRepo.FindByID(audienceID, orgID, userID)
			if err != nil {
				return c.Status(404).JSON(fiber.Map{"error": "Audience not found: " + audienceID})
			}
		}
	}

	// Verify contact exists and belongs to the agent if provided
	if req.ContactID != nil {
		_, err := contactRepo.FindByID(*req.ContactID, orgID, userID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Contact not found"})
		}
	}

	// Validate recurring settings
	if req.ScheduleType == "recurring" {
		if req.Recurrence == nil {
			return c.Status(400).JSON(fiber.Map{"error": "recurrence is required for recurring campaigns"})
		}
		if *req.Recurrence != "daily" && *req.Recurrence != "weekly" && *req.Recurrence != "monthly" {
			return c.Status(400).JSON(fiber.Map{"error": "recurrence must be 'daily', 'weekly', or 'monthly'"})
		}
	}

	// Parse recurrence time if provided
	var recurrenceTime *time.Time
	if req.RecurrenceTime != nil {
		parsedTime, err := time.Parse("15:04", *req.RecurrenceTime)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid recurrence_time format. Use HH:MM"})
		}
		recurrenceTime = &parsedTime
	}

	campaign := models.Campaign{
		OrganizationID:       orgID,
		Name:                 campaignName,
		TemplateID:           req.TemplateID,
		AudienceIDs:          req.AudienceIDs,
		ContactID:            req.ContactID,
		ScheduleType:         req.ScheduleType,
		ScheduledAt:          req.ScheduledAt,
		Recurrence:           req.Recurrence,
		RecurrenceDayOfWeek:  req.RecurrenceDayOfWeek,
		RecurrenceDayOfMonth: req.RecurrenceDayOfMonth,
		RecurrenceTime:       recurrenceTime,
		Status:               "scheduled",
		CreatedBy:            userID,
	}

	fmt.Printf("DEBUG: Creating campaign with AudienceIDs: %+v\n", campaign.AudienceIDs)
	if err := campaignRepo.Create(&campaign); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create campaign",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":  "Campaign created successfully",
		"campaign": campaign,
	})
}

// GetCampaigns returns paginated campaigns with optional status filter and sorting
func GetCampaigns(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	status := c.Query("status", "")
	sortBy := c.Query("sort_by", "")
	sortOrder := c.Query("sort_order", "")

	// Validate pagination params
	page, limit = utils.ValidatePaginationParams(page, limit)

	campaigns, total, err := campaignRepo.FindAllByOrg(orgID, userID, status, page, limit, sortBy, sortOrder)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch campaigns"})
	}

	// Calculate pagination metadata
	pagination := utils.CalculatePagination(page, limit, total)

	return c.JSON(fiber.Map{
		"campaigns":    campaigns,
		"total_count":  pagination.Total,
		"page":         pagination.Page,
		"limit":        pagination.Limit,
		"total_pages":  pagination.TotalPages,
		"offset_start": pagination.OffsetStart,
		"offset_end":   pagination.OffsetEnd,
	})
}

// GetCampaignByID returns a single campaign
func GetCampaignByID(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	campaignID := c.Params("id")

	campaign, err := campaignRepo.FindByID(campaignID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Campaign not found"})
	}

	return c.JSON(campaign)
}

// UpdateCampaign updates a campaign
func UpdateCampaign(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	campaignID := c.Params("id")

	campaign, err := campaignRepo.FindByID(campaignID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Campaign not found"})
	}

	// Only allow updates to draft or scheduled campaigns
	if campaign.Status != "draft" && campaign.Status != "scheduled" {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot update a campaign that is running, paused, or completed"})
	}

	var req struct {
		Name        *string    `json:"name"`
		ScheduledAt *time.Time `json:"scheduled_at"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name != nil {
		campaign.Name = *req.Name
	}
	if req.ScheduledAt != nil {
		campaign.ScheduledAt = *req.ScheduledAt
	}

	if err := campaignRepo.Update(campaign); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update campaign"})
	}

	return c.JSON(fiber.Map{
		"message":  "Campaign updated successfully",
		"campaign": campaign,
	})
}

// DeleteCampaign deletes a campaign
func DeleteCampaign(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	campaignID := c.Params("id")

	campaign, err := campaignRepo.FindByID(campaignID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Campaign not found"})
	}

	// Only allow deletion of draft or completed campaigns
	if campaign.Status == "running" {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot delete a running campaign. Pause it first."})
	}

	if err := campaignRepo.Delete(campaignID, orgID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete campaign"})
	}

	return c.JSON(fiber.Map{"message": "Campaign deleted successfully"})
}

// PauseCampaign pauses a campaign
func PauseCampaign(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	campaignID := c.Params("id")

	campaign, err := campaignRepo.FindByID(campaignID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Campaign not found"})
	}

	if campaign.Status != "scheduled" && campaign.Status != "running" {
		return c.Status(400).JSON(fiber.Map{"error": "Can only pause scheduled or running campaigns"})
	}

	if err := campaignRepo.UpdateStatus(campaignID, "paused"); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to pause campaign"})
	}

	return c.JSON(fiber.Map{"message": "Campaign paused successfully"})
}

// ResumeCampaign resumes a paused campaign
func ResumeCampaign(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	campaignID := c.Params("id")

	campaign, err := campaignRepo.FindByID(campaignID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Campaign not found"})
	}

	if campaign.Status != "paused" {
		return c.Status(400).JSON(fiber.Map{"error": "Can only resume paused campaigns"})
	}

	if err := campaignRepo.UpdateStatus(campaignID, "scheduled"); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to resume campaign"})
	}

	return c.JSON(fiber.Map{"message": "Campaign resumed successfully"})
}

// GetCampaignLogs returns paginated logs for a campaign
func GetCampaignLogs(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	campaignID := c.Params("id")

	// Verify campaign exists and belongs to the agent
	_, err := campaignRepo.FindByID(campaignID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Campaign not found"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	logs, total, err := campaignLogRepo.FindByCampaign(campaignID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch campaign logs"})
	}

	// Get statistics
	stats, _ := campaignLogRepo.GetStatsByCampaign(campaignID)

	return c.JSON(fiber.Map{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
		"stats": stats,
	})
}
