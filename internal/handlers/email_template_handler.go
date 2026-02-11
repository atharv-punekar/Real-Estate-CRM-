package handlers

import (
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

var (
	templateRepo = &repository.EmailTemplateRepository{}
)

// CreateEmailTemplate creates a new email template
func CreateEmailTemplate(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	userID := c.Locals("user_id").(string)

	var req struct {
		Name          string `json:"name"`
		Subject       string `json:"subject"`
		Preheader     string `json:"preheader"`
		FromName      string `json:"from_name"`
		ReplyTo       string `json:"reply_to"`
		HtmlBody      string `json:"html_body"`
		PlainTextBody string `json:"plain_text_body"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate and trim name
	templateName := strings.TrimSpace(req.Name)
	if err := utils.ValidateNotEmpty(templateName, "Template name"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate subject
	templateSubject := strings.TrimSpace(req.Subject)
	if err := utils.ValidateNotEmpty(templateSubject, "Subject"); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate template body (at least one required)
	if err := utils.ValidatePlainTextTemplate(req.HtmlBody, req.PlainTextBody); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate no placeholder content
	if err := utils.ValidateTemplateContent(req.FromName, req.Subject, req.HtmlBody, req.PlainTextBody); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	template := models.EmailTemplate{
		OrganizationID: orgID,
		Name:           templateName,
		Subject:        templateSubject,
		Preheader:      strings.TrimSpace(req.Preheader),
		FromName:       strings.TrimSpace(req.FromName),
		ReplyTo:        strings.TrimSpace(req.ReplyTo),
		HtmlBody:       req.HtmlBody,
		PlainTextBody:  req.PlainTextBody,
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
		if err := utils.ValidateNotEmpty(templateName, "Template name"); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
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
		template.FromName = strings.TrimSpace(*req.FromName)
	}
	if req.ReplyTo != nil {
		template.ReplyTo = strings.TrimSpace(*req.ReplyTo)
	}
	if req.HtmlBody != nil {
		template.HtmlBody = *req.HtmlBody
	}
	if req.PlainTextBody != nil {
		template.PlainTextBody = *req.PlainTextBody
	}

	// Validate template content after updates
	if err := utils.ValidatePlainTextTemplate(template.HtmlBody, template.PlainTextBody); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate no placeholder content
	if err := utils.ValidateTemplateContent(template.FromName, template.Subject, template.HtmlBody, template.PlainTextBody); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
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
