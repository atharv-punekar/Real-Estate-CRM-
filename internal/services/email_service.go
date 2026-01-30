package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	textTemplate "text/template"

	"github.com/go-gomail/gomail"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

// LoadSMTPConfig loads SMTP configuration from environment variables
func LoadSMTPConfig() (*SMTPConfig, error) {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if host == "" || portStr == "" || user == "" || pass == "" || from == "" {
		return nil, fmt.Errorf("missing SMTP configuration in environment variables")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	return &SMTPConfig{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		From: from,
	}, nil
}

// EmailService handles sending emails
type EmailService struct {
	smtpConfig *SMTPConfig
}

func NewEmailService() *EmailService {
	config, err := LoadSMTPConfig()
	if err != nil {
		log.Printf("Warning: Failed to load SMTP config: %v. Email sending will be disabled.", err)
		return &EmailService{smtpConfig: nil}
	}
	return &EmailService{smtpConfig: config}
}

func BuildFrontendInviteLink(token string) string {
	frontend := os.Getenv("FRONTEND_URL")
	return fmt.Sprintf("%s/auth/activate?token=%s", frontend, token)
}

// SendInviteEmail sends an invite email to a new agent
func (s *EmailService) SendInviteEmail(email, name, orgName, token string) error {
	inviteLink := BuildFrontendInviteLink(token)
	// If SMTP is not configured, just log
	if s.smtpConfig == nil {
		log.Printf("📧 INVITE EMAIL (SMTP not configured - logging only)")
		log.Printf("   To: %s", email)
		log.Printf("   Name: %s", name)
		log.Printf("   Organization: %s", orgName)
		log.Printf("   Invite Link: %s", inviteLink)
		return nil
	}

	// Load templates
	htmlBody, err := s.renderInviteEmailHTML(name, orgName, inviteLink)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	textBody, err := s.renderInviteEmailText(name, orgName, inviteLink)
	if err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}

	// Send email
	subject := fmt.Sprintf("Welcome to %s - Set Your Password", orgName)
	return s.sendEmail(email, subject, htmlBody, textBody)
}

// SendCampaignEmail sends a campaign email to a recipient
func (s *EmailService) SendCampaignEmail(recipientEmail, subject, htmlBody, plainBody string) error {
	// If SMTP is not configured, just log
	if s.smtpConfig == nil {
		log.Printf("📧 CAMPAIGN EMAIL (SMTP not configured - logging only)")
		log.Printf("   To: %s", recipientEmail)
		log.Printf("   Subject: %s", subject)
		return nil
	}

	// Send email
	return s.sendEmail(recipientEmail, subject, htmlBody, plainBody)
}

// sendEmail is the internal method that actually sends via SMTP
func (s *EmailService) sendEmail(to, subject, htmlBody, textBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.smtpConfig.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	// Set both plain text and HTML bodies
	m.SetBody("text/plain", textBody)
	m.AddAlternative("text/html", htmlBody)

	// Create dialer and send
	d := gomail.NewDialer(s.smtpConfig.Host, s.smtpConfig.Port, s.smtpConfig.User, s.smtpConfig.Pass)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("❌ Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ Email sent successfully to %s", to)
	return nil
}

// renderInviteEmailHTML renders the invite email HTML template
func (s *EmailService) renderInviteEmailHTML(name, orgName, inviteLink string) (string, error) {
	templatePath := filepath.Join("templates", "invite_email.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		// Fallback to inline template if file not found
		return s.fallbackInviteHTML(name, orgName, inviteLink), nil
	}

	data := map[string]string{
		"Name":             name,
		"OrganizationName": orgName,
		"InviteLink":       inviteLink,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderInviteEmailText renders the invite email text template
func (s *EmailService) renderInviteEmailText(name, orgName, inviteLink string) (string, error) {
	templatePath := filepath.Join("templates", "invite_email.txt")
	tmpl, err := textTemplate.ParseFiles(templatePath)
	if err != nil {
		// Fallback to inline template if file not found
		return s.fallbackInviteText(name, orgName, inviteLink), nil
	}

	data := map[string]string{
		"Name":             name,
		"OrganizationName": orgName,
		"InviteLink":       inviteLink,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// fallbackInviteHTML returns a simple HTML template if file is not found
func (s *EmailService) fallbackInviteHTML(name, orgName, inviteLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body>
	<h1>Welcome to %s!</h1>
	<p>Hi %s,</p>
	<p>You have been invited to join %s as a team member.</p>
	<p><a href="%s">Activate Your Account</a></p>
	<p>Or copy this link: %s</p>
</body>
</html>
`, orgName, name, orgName, inviteLink, inviteLink)
}

// fallbackInviteText returns a simple text template if file is not found
func (s *EmailService) fallbackInviteText(name, orgName, inviteLink string) string {
	return fmt.Sprintf(`Welcome to %s!

Hi %s,

You have been invited to join %s as a team member.

Activate your account: %s

---
This is an automated email.
`, orgName, name, orgName, inviteLink)
}

// GenerateInviteLink generates a full invite link URL
func GenerateInviteLink(baseURL, token string) string {
	return fmt.Sprintf("%s/auth/activate?token=%s", baseURL, token)
}

// SubstituteTemplateVariables replaces template variables with actual values
func SubstituteTemplateVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
