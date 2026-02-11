package handlers

import (
	"errors"
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var (
	templateRepo = &repository.EmailTemplateRepository{}
)

// CreateEmailTemplate creates a new email template
func CreateEmailTemplate(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		Name          string  `json:"name"`
		Subject       string  `json:"subject"`
		Preheader     string  `json:"preheader"`
		FromName      *string `json:"from_name"` // Optional
		ReplyTo       *string `json:"reply_to"`  // Optional
		HtmlBody      string  `json:"html_body"`
		PlainTextBody string  `json:"plain_text_body"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate name (required, alphabets and spaces, unique)
	templateName := strings.TrimSpace(req.Name)
	if templateName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if len(templateName) < 2 {
		return c.Status(400).JSON(fiber.Map{"error": "name must be at least 2 characters long"})
	}
	// Check for alphabets and spaces only
	for _, char := range templateName {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == ' ') {
			return c.Status(400).JSON(fiber.Map{"error": "name must contain only alphabetic characters and spaces"})
		}
	}
	// Check for multiple consecutive spaces
	if strings.Contains(templateName, "  ") {
		return c.Status(400).JSON(fiber.Map{"error": "name cannot contain multiple consecutive spaces"})
	}

	// Check for duplicate name in organization
	existing, err := templateRepo.FindByNameAndOrg(templateName, orgID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to check for duplicate template"})
	}
	if existing != nil && err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "a template with this name already exists in your organization"})
	}

	// Validate subject
	templateSubject := strings.TrimSpace(req.Subject)
	if err := utils.ValidateNotEmpty(templateSubject, "Subject"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate template body (at least one required, not both empty)
	htmlBody := strings.TrimSpace(req.HtmlBody)
	plainTextBody := strings.TrimSpace(req.PlainTextBody)
	if htmlBody == "" && plainTextBody == "" {
		return c.Status(400).JSON(fiber.Map{"error": "either html_body or plain_text_body must be provided"})
	}

	template := models.EmailTemplate{
		OrganizationID: orgID,
		Name:           templateName,
		Subject:        templateSubject,
		Preheader:      strings.TrimSpace(req.Preheader),
		FromName:       req.FromName, // Optional, can be nil
		ReplyTo:        req.ReplyTo,  // Optional, can be nil
		HtmlBody:       htmlBody,
		PlainTextBody:  plainTextBody,
		CreatedBy:      userID,
	}

	if err := templateRepo.Create(&template); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create email template"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":  "Email template created successfully",
		"template": template,
	})
}

// GetEmailTemplates returns all email templates for an organization
func GetEmailTemplates(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	templates, err := templateRepo.FindAllByOrg(orgID, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch email templates"})
	}

	return c.JSON(templates)
}

// GetEmailTemplateByID returns a single email template
func GetEmailTemplateByID(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	templateID := c.Params("id")

	template, err := templateRepo.FindByID(templateID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Email template not found"})
	}

	return c.JSON(template)
}

// UpdateEmailTemplate updates an email template
func UpdateEmailTemplate(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	templateID := c.Params("id")

	template, err := templateRepo.FindByID(templateID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Email template not found"})
	}

	var req struct {
		Name          *string `json:"name"`
		Subject       *string `json:"subject"`
		Preheader     *string `json:"preheader"`
		FromName      *string `json:"from_name"`
		ReplyTo       *string `json:"reply_to"`
		HtmlBody      *string `json:"html_body"`
		PlainTextBody *string `json:"plain_text_body"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name != nil {
		templateName := strings.TrimSpace(*req.Name)
		if templateName == "" {
			return c.Status(400).JSON(fiber.Map{"error": "name cannot be empty"})
		}
		if len(templateName) < 2 {
			return c.Status(400).JSON(fiber.Map{"error": "name must be at least 2 characters long"})
		}
		// Check for alphabets and spaces only
		for _, char := range templateName {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == ' ') {
				return c.Status(400).JSON(fiber.Map{"error": "name must contain only alphabetic characters and spaces"})
			}
		}
		// Check for multiple consecutive spaces
		if strings.Contains(templateName, "  ") {
			return c.Status(400).JSON(fiber.Map{"error": "name cannot contain multiple consecutive spaces"})
		}
		// Check for duplicate name (excluding current template)
		existing, err := templateRepo.FindByNameAndOrg(templateName, orgID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to check for duplicate template"})
		}
		if existing != nil && err == nil && existing.ID != templateID {
			return c.Status(400).JSON(fiber.Map{"error": "a template with this name already exists in your organization"})
		}
		template.Name = templateName
	}
	if req.Subject != nil {
		templateSubject := strings.TrimSpace(*req.Subject)
		if err := utils.ValidateNotEmpty(templateSubject, "Subject"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		template.Subject = templateSubject
	}
	if req.Preheader != nil {
		template.Preheader = strings.TrimSpace(*req.Preheader)
	}
	if req.FromName != nil {
		template.FromName = req.FromName // Already a pointer
	}
	if req.ReplyTo != nil {
		template.ReplyTo = req.ReplyTo // Already a pointer
	}
	if req.HtmlBody != nil {
		template.HtmlBody = *req.HtmlBody
	}
	if req.PlainTextBody != nil {
		template.PlainTextBody = *req.PlainTextBody
	}

	// Validate template body after updates (at least one required)
	if strings.TrimSpace(template.HtmlBody) == "" && strings.TrimSpace(template.PlainTextBody) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "either html_body or plain_text_body must be provided"})
	}

	if err := templateRepo.Update(template); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update email template"})
	}

	return c.JSON(fiber.Map{
		"message":  "Email template updated successfully",
		"template": template,
	})
}

// DeleteEmailTemplate deletes an email template
func DeleteEmailTemplate(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	templateID := c.Params("id")

	if err := templateRepo.Delete(templateID, orgID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete email template"})
	}

	return c.JSON(fiber.Map{"message": "Email template deleted successfully"})
}

// TestSendEmail sends a test email using a template
func TestSendEmail(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)
	templateID := c.Params("id")

	template, err := templateRepo.FindByID(templateID, orgID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Email template not found"})
	}

	var req struct {
		TestEmail string `json:"test_email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.TestEmail == "" {
		return c.Status(400).JSON(fiber.Map{"error": "test_email is required"})
	}

	// Send test email
	err = emailService.SendCampaignEmail(
		req.TestEmail,
		template.Subject,
		template.HtmlBody,
		template.PlainTextBody,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send test email"})
	}

	return c.JSON(fiber.Map{
		"message": "Test email sent successfully",
		"to":      req.TestEmail,
	})
}
