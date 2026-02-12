package handlers

import (
	"strconv"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/config"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/services"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Variables are shared from organization_handler.go

// -------------------------
// 1) CREATE AGENT (org_admin or org_user)
// -------------------------
func SuperAdminCreateAgent(c *fiber.Ctx) error {
	orgID := c.Params("org_id")
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid organization ID"})
	}

	// Verify organization exists
	org, err := orgRepo.FindByID(orgUUID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Organization not found"})
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"` // org_admin or org_user
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate name with strict rules: only alphabets and exactly one space
	if err := utils.ValidateAgentName(req.Name); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate and normalize email
	normalizedEmail, err := utils.NormalizeEmail(req.Email)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Check if organization is active
	if !org.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Cannot add agents. Organization is inactive"})
	}

	// Check name uniqueness within organization
	existingName, _ := userRepo.FindByNameAndOrg(req.Name, orgUUID)
	if existingName != nil {
		return c.Status(400).JSON(fiber.Map{"error": "An agent with this name already exists in this organization"})
	}

	// Validate role
	if req.Role != "org_admin" && req.Role != "org_user" {
		return c.Status(400).JSON(fiber.Map{"error": "Role must be 'org_admin' or 'org_user'"})
	}

	// Unique email check
	existing, _ := userRepo.FindByEmail(normalizedEmail)
	if existing != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Email already exists"})
	}

	// Generate invite token
	inviteToken, err := utils.GenerateInviteToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate invite token"})
	}
	expiresAt := utils.GetTokenExpiry()

	user := models.User{
		OrganizationID: orgUUID,
		Name:           req.Name,
		Email:          normalizedEmail,
		Role:           req.Role,
		IsActive:       true,
		InviteToken:    &inviteToken,
		TokenExpiresAt: &expiresAt,
		IsPasswordSet:  false,
	}

	if err := userRepo.Create(&user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create agent"})
	}

	// Load config for frontend URL
	cfg, _ := config.Load()

	// Generate invite link and send email
	inviteLink := services.GenerateInviteLink(cfg.Server.FrontendURL, inviteToken)

	// Send invite email (async in production)
	go emailService.SendInviteEmail(cfg.Server.FrontendURL, user.Email, user.Name, org.Name, inviteToken)

	return c.JSON(fiber.Map{
		"message":     "Agent created successfully. Invite email sent.",
		"user":        user,
		"invite_link": inviteLink,
	})
}

// -------------------------
// 2) GET ALL AGENTS FOR AN ORGANIZATION
// -------------------------
func SuperAdminGetAgents(c *fiber.Ctx) error {
	orgID := c.Params("org_id")
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid organization ID"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	// Validate pagination params
	page, limit = utils.ValidatePaginationParams(page, limit)

	users, total, err := userRepo.FindAllByOrgPaginated(orgUUID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch agents"})
	}

	// Calculate pagination metadata
	pagination := utils.CalculatePagination(page, limit, total)

	return c.JSON(fiber.Map{
		"agents":       users,
		"total_count":  pagination.Total,
		"page":         pagination.Page,
		"limit":        pagination.Limit,
		"total_pages":  pagination.TotalPages,
		"offset_start": pagination.OffsetStart,
		"offset_end":   pagination.OffsetEnd,
	})
}

// -------------------------
// 3) UPDATE AGENT
// -------------------------
func SuperAdminUpdateAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	user, err := userRepo.FindByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	var req struct {
		Name     string `json:"name"`
		IsActive *bool  `json:"is_active"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update name with strict validation
	if req.Name != "" {
		if err := utils.ValidateAgentName(req.Name); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Check name uniqueness within organization (excluding current user)
		existingName, _ := userRepo.FindByNameAndOrgExcluding(req.Name, user.OrganizationID, userID)
		if existingName != nil {
			return c.Status(400).JSON(fiber.Map{"error": "An agent with this name already exists in this organization"})
		}

		user.Name = req.Name
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.Role != "" {
		if req.Role != "org_admin" && req.Role != "org_user" {
			return c.Status(400).JSON(fiber.Map{"error": "Role must be 'org_admin' or 'org_user'"})
		}
		user.Role = req.Role
	}

	if err := userRepo.Update(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update agent"})
	}

	return c.JSON(fiber.Map{
		"message": "Agent updated successfully",
		"user":    user,
	})
}

// -------------------------
// 4) DEACTIVATE AGENT (Soft Delete)
// -------------------------
func SuperAdminDeactivateAgent(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	user, err := userRepo.FindByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Agent not found"})
	}

	user.IsActive = false

	if err := userRepo.Update(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to deactivate agent"})
	}

	return c.JSON(fiber.Map{
		"message": "Agent deactivated successfully",
	})
}

// -------------------------
// 5) REGENERATE INVITE TOKEN
// -------------------------
func SuperAdminRegenerateInvite(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	user, err := userRepo.FindByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Agent not found"})
	}

	// Generate new invite token
	inviteToken, err := utils.GenerateInviteToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate invite token"})
	}
	expiresAt := utils.GetTokenExpiry()

	user.InviteToken = &inviteToken
	user.TokenExpiresAt = &expiresAt

	if err := userRepo.Update(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to regenerate invite token"})
	}

	// Get organization for email
	org, _ := orgRepo.FindByID(user.OrganizationID)
	orgName := "your organization"
	if org != nil {
		orgName = org.Name
	}

	// Generate invite link and send email
	cfg, _ := config.Load()
	inviteLink := services.GenerateInviteLink(cfg.Server.FrontendURL, inviteToken)

	// Send invite email
	go emailService.SendInviteEmail(cfg.Server.FrontendURL, user.Email, user.Name, orgName, inviteToken)

	return c.JSON(fiber.Map{
		"message":     "Invite token regenerated and email sent",
		"invite_link": inviteLink,
	})
}
