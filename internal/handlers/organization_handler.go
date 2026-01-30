package handlers

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
		Name:      req.Name,
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
		org.Name = req.Name
	}
	if req.IsActive != nil {
		org.IsActive = *req.IsActive
	}

	if err := orgRepo.Update(org); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update organization"})
	}

	return c.JSON(org)
}
