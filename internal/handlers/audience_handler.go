package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var audienceRepo = &repository.AudienceRepository{}

// CreateAudience creates a new audience
func CreateAudience(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`

		PropertyType      []string `json:"property_type"`
		Bedrooms          []int    `json:"bedrooms"`
		Bathrooms         []int    `json:"bathrooms"`
		PreferredLocation []string `json:"preferred_location"`

		MinBudget float64 `json:"min_budget"`
		MaxBudget float64 `json:"max_budget"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate name (required, alphabets only, unique)
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}

	// Validate name format (alphabets and spaces only)
	trimmedName := strings.TrimSpace(req.Name)
	if len(trimmedName) < 2 {
		return c.Status(400).JSON(fiber.Map{"error": "name must be at least 2 characters long"})
	}

	// Check for alphabets and spaces only
	for _, char := range trimmedName {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == ' ') {
			return c.Status(400).JSON(fiber.Map{"error": "name must contain only alphabetic characters and spaces"})
		}
	}

	// Check for multiple consecutive spaces
	if strings.Contains(trimmedName, "  ") {
		return c.Status(400).JSON(fiber.Map{"error": "name cannot contain multiple consecutive spaces"})
	}

	// Check for duplicate name in organization
	existing, err := audienceRepo.FindByNameAndOrg(trimmedName, orgID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to check for duplicate audience"})
	}
	if existing != nil && err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "an audience with this name already exists in your organization"})
	}

	// Create audience
	audience := models.Audience{
		OrganizationID: orgID,
		Name:           trimmedName,
		Description:    req.Description,
		CreatedBy:      userID,
	}

	if err := audienceRepo.Create(&audience); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create audience"})
	}

	// Build query
	filter := models.ContactFilter{
		OrganizationID: orgID,
		PropertyType:   req.PropertyType,
		Bedrooms:       req.Bedrooms,
		Bathrooms:      req.Bathrooms,
		Locations:      req.PreferredLocation,
		MinBudget:      req.MinBudget,
		MaxBudget:      req.MaxBudget,
	}

	contacts, err := contactRepo.FindWithFilter(filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to filter contacts"})
	}

	// Add contacts to audience
	var ids []string
	for _, ct := range contacts {
		ids = append(ids, ct.ID)
	}

	if len(ids) > 0 {
		audienceRepo.AddContacts(audience.ID, ids)
	}

	return c.Status(201).JSON(fiber.Map{
		"message":          "Audience created successfully",
		"audience":         audience,
		"matched_contacts": len(ids),
	})
}

// GetAudiences returns all audiences for an organization
func GetAudiences(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	audiences, err := audienceRepo.FindAllByOrg(orgID, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch audiences"})
	}

	// Add contact counts
	type AudienceWithCount struct {
		models.Audience
		ContactCount int64 `json:"contact_count"`
	}

	var result []AudienceWithCount
	for _, audience := range audiences {
		count, _ := audienceRepo.CountContactsByAudience(audience.ID)
		result = append(result, AudienceWithCount{
			Audience:     audience,
			ContactCount: count,
		})
	}

	return c.JSON(result)
}

// GetAudienceByID returns a single audience
func GetAudienceByID(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	audienceID := c.Params("id")

	audience, err := audienceRepo.FindByID(audienceID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Audience not found"})
	}

	count, _ := audienceRepo.CountContactsByAudience(audienceID)

	return c.JSON(fiber.Map{
		"audience":      audience,
		"contact_count": count,
	})
}

// UpdateAudience updates an audience
func UpdateAudience(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	audienceID := c.Params("id")

	audience, err := audienceRepo.FindByID(audienceID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Audience not found"})
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name != nil {
		// Validate name format (alphabets and spaces only)
		trimmedName := strings.TrimSpace(*req.Name)
		if trimmedName == "" {
			return c.Status(400).JSON(fiber.Map{"error": "name cannot be empty"})
		}

		if len(trimmedName) < 2 {
			return c.Status(400).JSON(fiber.Map{"error": "name must be at least 2 characters long"})
		}

		// Check for alphabets and spaces only
		for _, char := range trimmedName {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == ' ') {
				return c.Status(400).JSON(fiber.Map{"error": "name must contain only alphabetic characters and spaces"})
			}
		}

		// Check for multiple consecutive spaces
		if strings.Contains(trimmedName, "  ") {
			return c.Status(400).JSON(fiber.Map{"error": "name cannot contain multiple consecutive spaces"})
		}

		// Check for duplicate name in organization (excluding current audience)
		existing, err := audienceRepo.FindByNameAndOrg(trimmedName, orgID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to check for duplicate audience"})
		}
		if existing != nil && err == nil && existing.ID != audienceID {
			return c.Status(400).JSON(fiber.Map{"error": "an audience with this name already exists in your organization"})
		}

		audience.Name = trimmedName
	}
	if req.Description != nil {
		audience.Description = *req.Description
	}

	if err := audienceRepo.Update(audience); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update audience"})
	}

	return c.JSON(fiber.Map{
		"message":  "Audience updated successfully",
		"audience": audience,
	})
}

// DeleteAudience deletes an audience
func DeleteAudience(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	audienceID := c.Params("id")

	// Verify ownership before deletion
	_, err := audienceRepo.FindByID(audienceID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Audience not found"})
	}

	if err := audienceRepo.Delete(audienceID, orgID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete audience"})
	}

	return c.JSON(fiber.Map{"message": "Audience deleted successfully"})
}

// AddContactsToAudience adds contacts to an audience
func AddContactsToAudience(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	audienceID := c.Params("id")

	// Verify audience exists and belongs to the agent
	_, err := audienceRepo.FindByID(audienceID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Audience not found"})
	}

	var req struct {
		ContactIDs []string `json:"contact_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.ContactIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No contact IDs provided"})
	}

	// Verify all contacts exist and belong to org
	contacts, err := contactRepo.FindByIDs(req.ContactIDs, orgID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify contacts"})
	}

	if len(contacts) != len(req.ContactIDs) {
		return c.Status(400).JSON(fiber.Map{"error": "Some contacts not found or don't belong to your organization"})
	}

	if err := audienceRepo.AddContacts(audienceID, req.ContactIDs); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to add contacts to audience"})
	}

	return c.JSON(fiber.Map{
		"message": "Contacts added to audience successfully",
		"count":   len(req.ContactIDs),
	})
}

// RemoveContactsFromAudience removes contacts from an audience
func RemoveContactsFromAudience(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	audienceID := c.Params("id")

	// Verify audience exists and belongs to the agent
	_, err := audienceRepo.FindByID(audienceID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Audience not found"})
	}

	var req struct {
		ContactIDs []string `json:"contact_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.ContactIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No contact IDs provided"})
	}

	if err := audienceRepo.RemoveContacts(audienceID, req.ContactIDs); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to remove contacts from audience"})
	}

	return c.JSON(fiber.Map{
		"message": "Contacts removed from audience successfully",
		"count":   len(req.ContactIDs),
	})
}

// GetAudienceContacts returns paginated contacts for an audience
func GetAudienceContacts(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	audienceID := c.Params("id")

	// Verify audience exists and belongs to the agent
	_, err := audienceRepo.FindByID(audienceID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Audience not found"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	contacts, total, err := audienceRepo.FindContactsByAudience(audienceID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch audience contacts"})
	}

	return c.JSON(fiber.Map{
		"contacts": contacts,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}
