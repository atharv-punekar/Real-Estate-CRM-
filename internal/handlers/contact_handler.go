package handlers

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/services"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
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
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate negative values
	if err := utils.ValidateNonNegative(req.BudgetMin, "Minimum budget"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := utils.ValidateNonNegative(req.BudgetMax, "Maximum budget"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := utils.ValidateNonNegativeInt(req.Bedrooms, "Bedrooms"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := utils.ValidateNonNegativeInt(req.Bathrooms, "Bathrooms"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := utils.ValidateNonNegativeInt(req.SquareFeet, "Square feet"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate budget range
	if err := utils.ValidateBudgetRange(req.BudgetMin, req.BudgetMax); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Normalize email
	normalizedEmail, err := utils.NormalizeEmail(req.Email)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	contact := models.Contact{
		OrganizationID:    orgID,
		CreatedBy:         userID,
		FirstName:         strings.TrimSpace(req.FirstName),
		LastName:          strings.TrimSpace(req.LastName),
		Email:             normalizedEmail,
		Phone:             strings.TrimSpace(req.Phone),
		BudgetMin:         req.BudgetMin,
		BudgetMax:         req.BudgetMax,
		PropertyType:      strings.TrimSpace(req.PropertyType),
		Bedrooms:          req.Bedrooms,
		Bathrooms:         req.Bathrooms,
		SquareFeet:        req.SquareFeet,
		PreferredLocation: strings.TrimSpace(req.PreferredLocation),
	}

	if err := contactService.CreateContact(&contact); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Contact created successfully",
		"contact": contact,
	})
}

// GetContacts returns paginated contacts with optional search and sorting
func GetContacts(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	search := c.Query("search", "")
	sortBy := c.Query("sort_by", "")
	sortOrder := c.Query("sort_order", "")

	// Validate pagination params
	page, limit = utils.ValidatePaginationParams(page, limit)

	// Filter by creator (per-agent ownership)
	contacts, total, err := contactRepo.FindAllByOrg(orgID, userID, page, limit, search, sortBy, sortOrder)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch contacts"})
	}

	// Calculate pagination metadata
	pagination := utils.CalculatePagination(page, limit, total)

	return c.JSON(fiber.Map{
		"contacts":     contacts,
		"total_count":  pagination.Total,
		"page":         pagination.Page,
		"limit":        pagination.Limit,
		"total_pages":  pagination.TotalPages,
		"offset_start": pagination.OffsetStart,
		"offset_end":   pagination.OffsetEnd,
	})
}

// GetContactByID returns a single contact
func GetContactByID(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	contactID := c.Params("id")

	// Verify ownership
	contact, err := contactRepo.FindByID(contactID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Contact not found"})
	}

	return c.JSON(contact)
}

func UpdateContact(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	contactID := c.Params("id")

	// Verify ownership
	contact, err := contactRepo.FindByID(contactID, orgID, userID)
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
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// -------- STRING FIELDS --------

	if req.FirstName != nil {
		contact.FirstName = strings.TrimSpace(*req.FirstName)
	}

	if req.LastName != nil {
		contact.LastName = strings.TrimSpace(*req.LastName)
	}

	if req.Phone != nil {
		contact.Phone = strings.TrimSpace(*req.Phone)
	}

	if req.PropertyType != nil {
		contact.PropertyType = strings.TrimSpace(*req.PropertyType)
	}

	if req.PreferredLocation != nil {
		contact.PreferredLocation = strings.TrimSpace(*req.PreferredLocation)
	}

	// -------- EMAIL VALIDATION --------

	if req.Email != nil {
		normalizedEmail, err := utils.NormalizeEmail(*req.Email)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		contact.Email = normalizedEmail
	}

	// -------- NUMERIC VALIDATION --------

	if req.BudgetMin != nil {
		if err := utils.ValidateNonNegative(*req.BudgetMin, "Minimum budget"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		contact.BudgetMin = *req.BudgetMin
	}

	if req.BudgetMax != nil {
		if err := utils.ValidateNonNegative(*req.BudgetMax, "Maximum budget"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		contact.BudgetMax = *req.BudgetMax
	}

	if req.Bedrooms != nil {
		if err := utils.ValidateNonNegativeInt(*req.Bedrooms, "Bedrooms"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		contact.Bedrooms = *req.Bedrooms
	}

	if req.Bathrooms != nil {
		if err := utils.ValidateNonNegativeInt(*req.Bathrooms, "Bathrooms"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		contact.Bathrooms = *req.Bathrooms
	}

	if req.SquareFeet != nil {
		if err := utils.ValidateNonNegativeInt(*req.SquareFeet, "Square feet"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		contact.SquareFeet = *req.SquareFeet
	}

	// -------- BUDGET RANGE VALIDATION --------
	// Validate after updating min/max values

	if err := utils.ValidateBudgetRange(contact.BudgetMin, contact.BudgetMax); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// -------- SAVE --------

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
	userID := c.Locals("user_id").(string)
	contactID := c.Params("id")

	// Verify ownership before deletion
	_, err := contactRepo.FindByID(contactID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Contact not found"})
	}

	// Delete contact and all its relationships (audiences, campaigns)
	if err := contactRepo.DeleteWithRelationships(contactID, orgID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete contact"})
	}

	return c.JSON(fiber.Map{"message": "Contact deleted successfully"})
}

// ImportContactsCSV imports contacts from a CSV file
func ImportContactsCSV(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	// Verify organization exists to avoid foreign key errors in background_job_log
	var org models.Organization
	if err := database.DB.Where("id = ?", orgID).First(&org).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid organization session. Please log in again."})
	}

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

	// Create job record for audit
	job := models.BackgroundJobLog{
		JobType:        "csv_import",
		OrganizationID: orgID,
		Status:         "running",
	}
	if err := bgJobRepo.Create(&job); err != nil {
		fmt.Printf("Error creating audit job: %v\n", err)
	}

	// Parse CSV
	reader := bytes.NewReader(csvData)
	contacts, parseErrors, totalRows, err := contactService.ParseCSV(reader, orgID, userID)
	if err != nil {
		if job.ID != "" {
			bgJobService.FailJob(job.ID, err.Error())
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if job.ID != "" {
		job.TotalRecords = &totalRows
		database.DB.Save(&job)
	}

	// Bulk create contacts
	successCount, duplicateCount, failedCount, createErrors, err := contactService.BulkCreateContacts(contacts)
	if err != nil {
		// Systemic error
		if job.ID != "" {
			bgJobService.FailJob(job.ID, err.Error())
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Combine errors for API response
	type APIError struct {
		Row    int      `json:"row"`
		Errors []string `json:"errors"`
	}
	var apiErrors []APIError

	// Add parse errors
	for _, pe := range parseErrors {
		apiErrors = append(apiErrors, APIError{
			Row:    pe.Row,
			Errors: pe.Errors,
		})
	}

	// Add create errors (these are strings, need to parse row number if possible, or just add as general errors)
	// The CreateErrors from BulkCreate are simple strings.
	// We'll append them to a separate list or try to parse.
	// For API simplicity, let's return them as a list of strings if they don't fit the structure,
	// or better, standardise the BulkCreate error format later. For now, we'll return them as `invalid_rows` if we can't map them.
	// But wait, the previous code returned `errors`.
	// The prompt wants:
	// "invalid_details": [ { "row": 4, "errors": ["email invalid"] } ]

	// Current `createErrors` are strings like "Failed to create contact ..." or "Error checking duplicates ...".
	// They don't have row numbers cleanly attached in a struct, just in the string "Row X: ...".
	// Regex parsing? Or just return the strings in a separate field?
	// The prompt asked for specific JSON structure.
	// Let's iterate `createErrors` and try to extract row number if present.
	// Or just return `createErrors` as a list of strings in `general_errors`.

	// Let's stick to the prompt's `invalid_details`.
	// Since `BulkCreateContacts` returns `[]string`, and `s.contactRepo.Create` might not fail often for individual rows if validation passes...
	// Duplicate checks are done explicitly.
	// So `createErrors` are rare database errors.

	// Let's just return `parseErrors` as `invalid_details` and `createErrors` as just `errors`.

	// Update job progress and finish
	if job.ID != "" {
		job.ProcessedRecords = &successCount
		var allErrors []string
		for _, pe := range parseErrors {
			allErrors = append(allErrors, fmt.Sprintf("Row %d: %s", pe.Row, strings.Join(pe.Errors, ", ")))
		}
		allErrors = append(allErrors, createErrors...)

		if len(allErrors) > 0 {
			job.ErrorMessage = strings.Join(allErrors, "; ")
			// limit length
			if len(job.ErrorMessage) > 1000 {
				job.ErrorMessage = job.ErrorMessage[:997] + "..."
			}
		}
		database.DB.Save(&job)
		bgJobService.FinishJob(job.ID)
	}

	// Notify user via system notification as well
	notifService.NotifyCSVImportCompleted(orgID, userID, successCount, duplicateCount, len(parseErrors)+failedCount)

	return c.JSON(fiber.Map{
		"message":         "CSV import completed",
		"total_rows":      totalRows,
		"imported_rows":   successCount,
		"duplicate_rows":  duplicateCount,
		"invalid_rows":    len(parseErrors) + failedCount,
		"invalid_details": apiErrors,
		"general_errors":  createErrors,
	})
}

// AddContactPreference adds a preference to a contact
