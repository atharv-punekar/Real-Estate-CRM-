package handlers

import (
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

	if req.Name == "" || req.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name & Email required"})
	}

	if !utils.IsValidEmail(req.Email) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid email format"})
	}

	// Validate role
	if req.Role != "org_admin" && req.Role != "org_user" {
		return c.Status(400).JSON(fiber.Map{"error": "Role must be 'org_admin' or 'org_user'"})
	}

	// Unique email check
	existing, _ := userRepo.FindByEmail(req.Email)
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
		Email:          req.Email,
		Role:           req.Role,
		IsActive:       true,
		InviteToken:    &inviteToken,
		TokenExpiresAt: &expiresAt,
		IsPasswordSet:  false,
	}

	if err := userRepo.Create(&user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create agent"})
	}

	// Generate invite link and send email
	inviteLink := services.BuildFrontendInviteLink(inviteToken)

	// Send invite email (async in production)
	go emailService.SendInviteEmail(user.Email, user.Name, org.Name, inviteToken)

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

	users, err := userRepo.FindAllByOrg(orgUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch agents"})
	}

	return c.JSON(users)
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

	if req.Name != "" {
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
	inviteLink := services.GenerateInviteLink(cfg.Server.BaseURL, inviteToken)

	// Send invite email
	go emailService.SendInviteEmail(user.Email, user.Name, orgName, inviteToken)

	return c.JSON(fiber.Map{
		"message":     "Invite token regenerated and email sent",
		"invite_link": inviteLink,
	})
}
