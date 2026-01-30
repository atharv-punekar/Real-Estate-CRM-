package handlers

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
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

	if req.Name == "" || req.Subject == "" || req.HtmlBody == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Name, subject, and html_body are required"})
	}

	template := models.EmailTemplate{
		OrganizationID: orgID,
		Name:           req.Name,
		Subject:        req.Subject,
		Preheader:      req.Preheader,
		FromName:       req.FromName,
		ReplyTo:        req.ReplyTo,
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

	templates, err := templateRepo.FindAllByOrg(orgID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch email templates"})
	}

	return c.JSON(templates)
}

// GetEmailTemplateByID returns a single email template
func GetEmailTemplateByID(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	templateID := c.Params("id")

	template, err := templateRepo.FindByID(templateID, orgID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Email template not found"})
	}

	return c.JSON(template)
}

// UpdateEmailTemplate updates an email template
func UpdateEmailTemplate(c *fiber.Ctx) error {
	orgID := c.Locals("org_id").(string)
	templateID := c.Params("id")

	template, err := templateRepo.FindByID(templateID, orgID)
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
		template.Name = *req.Name
	}
	if req.Subject != nil {
		template.Subject = *req.Subject
	}
	if req.Preheader != nil {
		template.Preheader = *req.Preheader
	}
	if req.FromName != nil {
		template.FromName = *req.FromName
	}
	if req.ReplyTo != nil {
		template.ReplyTo = *req.ReplyTo
	}
	if req.HtmlBody != nil {
		template.HtmlBody = *req.HtmlBody
	}
	if req.PlainTextBody != nil {
		template.PlainTextBody = *req.PlainTextBody
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
	templateID := c.Params("id")

	template, err := templateRepo.FindByID(templateID, orgID)
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
