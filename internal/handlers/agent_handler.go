package handlers

import (
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/services"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Variables are shared across handlers package

func CreateAgent(c *fiber.Ctx) error {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"` // org_admin or org_user
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate name with strict rules: only alphabets and exactly one space
	agentName := strings.TrimSpace(req.Name)
	if err := utils.ValidateAgentName(agentName); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate email with strict rules: lowercase letters/numbers, exactly 1 @, simple domain
	if err := utils.ValidateAgentEmail(req.Email); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// Validate role
	if req.Role == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Role is required"})
	}
	// Validate role - org_admin can create both org_admin and org_user
	if req.Role != "org_admin" && req.Role != "org_user" {
		return c.Status(400).JSON(fiber.Map{"error": "Role must be 'org_admin' or 'org_user'"})
	}

	callerRole := c.Locals("role").(string)
	orgID := c.Locals("org_id")

	if callerRole != "org_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Only Org Admin allowed to create agents"})
	}

	orgUUID, err := uuid.Parse(orgID.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid organization ID"})
	}

	// Get organization and check if active
	org, err := orgRepo.FindByID(orgUUID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Organization not found"})
	}
	if !org.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Cannot add agents. Organization is inactive"})
	}

	// Check global email uniqueness
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

	newUser := models.User{
		OrganizationID: orgUUID,
		Name:           agentName,
		Email:          normalizedEmail,
		Role:           req.Role,
		IsActive:       true,
		InviteToken:    &inviteToken,
		TokenExpiresAt: &expiresAt,
		IsPasswordSet:  false,
	}

	if err := userRepo.Create(&newUser); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create agent"})
	}

	// Send invite email using only token
	go emailService.SendInviteEmail(newUser.Email, newUser.Name, org.Name, inviteToken)

	// Build correct frontend invite link to return in API response
	inviteLink := services.BuildFrontendInviteLink(inviteToken)

	// Notify org admins about new agent (async)
	go notifService.NotifyAgentAdded(orgID.(string), newUser.ID.String(), newUser.Name)

	return c.JSON(fiber.Map{
		"message":     "Agent created successfully. Invite email sent.",
		"user":        newUser,
		"invite_link": inviteLink,
	})
}

func GetAgents(c *fiber.Ctx) error {
	orgID := c.Locals("org_id")
	role := c.Locals("role")

	if orgID == nil || role == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Only Org Admin can view agents
	if role.(string) != "org_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Only Org Admin can view agents"})
	}

	orgUUID, err := uuid.Parse(orgID.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid organization ID"})
	}

	users, err := userRepo.FindAllByOrg(orgUUID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch agents"})
	}

	return c.JSON(users)
}

func UpdateAgent(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid agent ID"})
	}

	orgID := c.Locals("org_id")
	role := c.Locals("role")

	if orgID == nil || role == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Only Org Admin can update other agents
	if role.(string) != "org_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Only Org Admin can update agents"})
	}

	// Check if organization is active
	orgUUID, err := uuid.Parse(orgID.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid organization ID"})
	}
	org, err := orgRepo.FindByID(orgUUID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Organization not found"})
	}
	if !org.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Cannot update agents. Organization is inactive"})
	}

	agent, err := userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Agent not found"})
	}

	// Ensure this agent belongs to same org
	if agent.OrganizationID.String() != orgID.(string) {
		return c.Status(403).JSON(fiber.Map{"error": "You cannot modify agents of another organization"})
	}

	var req struct {
		Name     string `json:"name"`
		IsActive *bool  `json:"is_active"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update name, if given (validate with strict rules)
	if req.Name != "" {
		agentName := strings.TrimSpace(req.Name)
		if err := utils.ValidateAgentName(agentName); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		agent.Name = agentName
	}

	// Update active/inactive
	if req.IsActive != nil {
		agent.IsActive = *req.IsActive
	}

	// Validate + update role - org_admin can change roles to both org_admin and org_user
	if req.Role != "" {
		if req.Role != "org_admin" && req.Role != "org_user" {
			return c.Status(400).JSON(fiber.Map{"error": "Role must be 'org_admin' or 'org_user'"})
		}
		agent.Role = req.Role
	}

	// Save
	if err := userRepo.Update(agent); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update agent"})
	}

	return c.JSON(agent)
}

// DeactivateAgent deactivates an agent (soft delete)
func DeactivateAgent(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid agent ID"})
	}

	orgID := c.Locals("org_id")
	role := c.Locals("role")

	if orgID == nil || role == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Only Org Admin can deactivate agents
	if role.(string) != "org_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Only Org Admin can deactivate agents"})
	}

	// Check if organization is active
	orgUUID, err := uuid.Parse(orgID.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid organization ID"})
	}
	org, err := orgRepo.FindByID(orgUUID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Organization not found"})
	}
	if !org.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Cannot deactivate agents. Organization is inactive"})
	}

	agent, err := userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Agent not found"})
	}

	// Ensure this agent belongs to same org
	if agent.OrganizationID.String() != orgID.(string) {
		return c.Status(403).JSON(fiber.Map{"error": "You cannot modify agents of another organization"})
	}

	agent.IsActive = false

	if err := userRepo.Update(agent); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to deactivate agent"})
	}

	return c.JSON(fiber.Map{
		"message": "Agent deactivated successfully",
	})
}

// RegenerateInvite regenerates invite token for an agent
func RegenerateInvite(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid agent ID"})
	}

	orgID := c.Locals("org_id")
	role := c.Locals("role")

	if orgID == nil || role == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Only Org Admin can regenerate invites
	if role.(string) != "org_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Only Org Admin can regenerate invites"})
	}

	// Check if organization is active
	orgUUID, err := uuid.Parse(orgID.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid organization ID"})
	}
	org, err := orgRepo.FindByID(orgUUID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Organization not found"})
	}
	if !org.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Cannot resend invitation. Organization is inactive"})
	}

	agent, err := userRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Agent not found"})
	}

	// Ensure this agent belongs to same org
	if agent.OrganizationID.String() != orgID.(string) {
		return c.Status(403).JSON(fiber.Map{"error": "You cannot modify agents of another organization"})
	}

	// Generate new invite token
	inviteToken, err := utils.GenerateInviteToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate invite token"})
	}
	expiresAt := utils.GetTokenExpiry()

	agent.InviteToken = &inviteToken
	agent.TokenExpiresAt = &expiresAt

	if err := userRepo.Update(agent); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to regenerate invite token"})
	}

	// Get organization for email
	orgName := "your organization"
	if org != nil {
		orgName = org.Name
	}

	// Generate invite link and send email
	go emailService.SendInviteEmail(agent.Email, agent.Name, orgName, inviteToken)
	inviteLink := services.BuildFrontendInviteLink(inviteToken)

	return c.JSON(fiber.Map{
		"message":     "Invite token regenerated and email sent",
		"invite_link": inviteLink,
	})
}
