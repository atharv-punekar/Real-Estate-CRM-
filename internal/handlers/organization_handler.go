package handlers

import (
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/services"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Shared repository and service instances across handlers package
var (
	orgRepo      = repository.OrganizationRepository{}
	userRepo     = repository.UserRepository{}
	emailService = services.NewEmailService()
)

func CreateOrganization(c *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Trim and validate organization name
	orgName := strings.TrimSpace(req.Name)
	if err := utils.ValidateNotEmpty(orgName, "Organization name"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Check for duplicate organization name
	existing, err := orgRepo.FindByName(orgName)
	if err != nil && err != gorm.ErrRecordNotFound {
		// Database error
		return c.Status(500).JSON(fiber.Map{"error": "Failed to check for duplicate organization"})
	}
	if existing != nil {
		return c.Status(400).JSON(fiber.Map{"error": "An organization with this name already exists"})
	}

	// Extract super admin ID from JWT
	createdBy := c.Locals("user_id")
	if createdBy == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	superAdminID, err := uuid.Parse(createdBy.(string))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid super admin ID"})
	}

	org := models.Organization{
		Name:      orgName,
		CreatedBy: superAdminID,
	}

	if err := orgRepo.Create(&org); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create organization"})
	}

	return c.JSON(org)
}

func GetOrganizations(c *fiber.Ctx) error {
	orgs, err := orgRepo.FindAll()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch organizations"})
	}

	return c.JSON(orgs)
}

func UpdateOrganization(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	org, err := orgRepo.FindByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Organization not found"})
	}

	var req struct {
		Name     string `json:"name"`
		IsActive *bool  `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Name != "" {
		// Trim and validate new name
		orgName := strings.TrimSpace(req.Name)
		if err := utils.ValidateNotEmpty(orgName, "Organization name"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Check for duplicate organization name (excluding current org)
		existing, err := orgRepo.FindByNameExcluding(orgName, id)
		if err != nil && err != gorm.ErrRecordNotFound {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to check for duplicate organization"})
		}
		if existing != nil {
			return c.Status(400).JSON(fiber.Map{"error": "An organization with this name already exists"})
		}

		org.Name = orgName
	}

	// Track if organization is being deactivated
	var orgDeactivated bool
	if req.IsActive != nil {
		// Check if organization is being deactivated
		if org.IsActive && !*req.IsActive {
			orgDeactivated = true
		}
		org.IsActive = *req.IsActive
	}

	if err := orgRepo.Update(org); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update organization"})
	}

	// If organization was deactivated, deactivate all agents in that organization
	if orgDeactivated {
		if err := userRepo.DeactivateAllByOrg(id); err != nil {
			// Log error but don't fail the request since org is already updated
			return c.JSON(fiber.Map{
				"organization": org,
				"warning":      "Organization updated but failed to deactivate agents",
			})
		}
		return c.JSON(fiber.Map{
			"organization":       org,
			"message":            "Organization deactivated successfully",
			"agents_deactivated": true,
		})
	}

	return c.JSON(org)
}
