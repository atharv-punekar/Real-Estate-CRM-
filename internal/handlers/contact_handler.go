package handlers

import (
	"io"
	"strconv"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/services"
	"github.com/gofiber/fiber/v2"
)

var (
	contactRepo    = &repository.ContactRepository{}
	contactService = services.NewContactService()
	bgJobRepo      = &repository.BackgroundJobRepository{}
	bgJobService   = services.NewBackgroundJobService()
	notifService   = services.NewNotificationService()
)

// CreateContact creates a new contact
func CreateContact(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		FirstName         string  `json:"first_name"`
		LastName          string  `json:"last_name"`
		Email             string  `json:"email"`
		Phone             string  `json:"phone"`
		BudgetMin         float64 `json:"budget_min"`
		BudgetMax         float64 `json:"budget_max"`
		PropertyType      string  `json:"property_type"`
		Bedrooms          int     `json:"bedrooms"`
		Bathrooms         int     `json:"bathrooms"`
		SquareFeet        int     `json:"square_feet"`
		PreferredLocation string  `json:"preferred_location"`
		Notes             string  `json:"notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	contact := models.Contact{
		OrganizationID:    orgID,
		CreatedBy:         userID,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Email:             req.Email,
		Phone:             req.Phone,
		BudgetMin:         req.BudgetMin,
		BudgetMax:         req.BudgetMax,
		PropertyType:      req.PropertyType,
		Bedrooms:          req.Bedrooms,
		Bathrooms:         req.Bathrooms,
		SquareFeet:        req.SquareFeet,
		PreferredLocation: req.PreferredLocation,
		Notes:             req.Notes,
	}

	if err := contactService.CreateContact(&contact); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Contact created successfully",
		"contact": contact,
	})
}

// GetContacts returns paginated contacts with optional search
func GetContacts(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	contacts, total, err := contactRepo.FindAllByOrg(orgID, page, limit, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch contacts"})
	}

	return c.JSON(fiber.Map{
		"contacts": contacts,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetContactByID returns a single contact
func GetContactByID(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	contactID := c.Params("id")

	contact, err := contactRepo.FindByID(contactID, orgID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Contact not found"})
	}

	return c.JSON(contact)
}

// UpdateContact updates a contact
func UpdateContact(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	contactID := c.Params("id")

	contact, err := contactRepo.FindByID(contactID, orgID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Contact not found"})
	}

	var req struct {
		FirstName         *string  `json:"first_name"`
		LastName          *string  `json:"last_name"`
		Email             *string  `json:"email"`
		Phone             *string  `json:"phone"`
		BudgetMin         *float64 `json:"budget_min"`
		BudgetMax         *float64 `json:"budget_max"`
		PropertyType      *string  `json:"property_type"`
		Bedrooms          *int     `json:"bedrooms"`
		Bathrooms         *int     `json:"bathrooms"`
		SquareFeet        *int     `json:"square_feet"`
		PreferredLocation *string  `json:"preferred_location"`
		Notes             *string  `json:"notes"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update fields if provided
	if req.FirstName != nil {
		contact.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		contact.LastName = *req.LastName
	}
	if req.Email != nil {
		contact.Email = *req.Email
	}
	if req.Phone != nil {
		contact.Phone = *req.Phone
	}
	if req.BudgetMin != nil {
		contact.BudgetMin = *req.BudgetMin
	}
	if req.BudgetMax != nil {
		contact.BudgetMax = *req.BudgetMax
	}
	if req.PropertyType != nil {
		contact.PropertyType = *req.PropertyType
	}
	if req.Bedrooms != nil {
		contact.Bedrooms = *req.Bedrooms
	}
	if req.Bathrooms != nil {
		contact.Bathrooms = *req.Bathrooms
	}
	if req.SquareFeet != nil {
		contact.SquareFeet = *req.SquareFeet
	}
	if req.PreferredLocation != nil {
		contact.PreferredLocation = *req.PreferredLocation
	}
	if req.Notes != nil {
		contact.Notes = *req.Notes
	}

	if err := contactRepo.Update(contact); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update contact"})
	}

	return c.JSON(fiber.Map{
		"message": "Contact updated successfully",
		"contact": contact,
	})
}

// DeleteContact soft deletes a contact
func DeleteContact(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	contactID := c.Params("id")

	if err := contactRepo.Delete(contactID, orgID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete contact"})
	}

	return c.JSON(fiber.Map{"message": "Contact deleted successfully"})
}

// ImportContactsCSV imports contacts from a CSV file
func ImportContactsCSV(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	// Check file type
	if file.Header.Get("Content-Type") != "text/csv" && file.Header.Get("Content-Type") != "application/vnd.ms-excel" {
		// Also accept if filename ends with .csv
		if len(file.Filename) < 4 || file.Filename[len(file.Filename)-4:] != ".csv" {
			return c.Status(400).JSON(fiber.Map{"error": "File must be a CSV"})
		}
	}

	// Open file
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read file"})
	}
	defer fileContent.Close()

	// Read file content
	csvData, err := io.ReadAll(fileContent)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read file content"})
	}

	// Create background job
	job := models.BackgroundJobLog{
		JobType:        "csv_import",
		OrganizationID: orgID,
		Status:         "queued",
	}
	if err := bgJobRepo.Create(&job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create import job"})
	}

	// Process CSV in background
	go bgJobService.ProcessCSVImport(job.ID, orgID, userID, csvData)

	return c.Status(202).JSON(fiber.Map{
		"message": "CSV import started",
		"job_id":  job.ID,
	})
}

// AddContactPreference adds a preference to a contact
