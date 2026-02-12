package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Pre-compiled email regex for performance (compiled once at package init)
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)

// ValidateNotEmpty validates that a string is not empty or whitespace-only
func ValidateNotEmpty(value, fieldName string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

// ValidateNonNegative validates that a float value is not negative
func ValidateNonNegative(value float64, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
	return nil
}

// ValidateNonNegativeInt validates that an int value is not negative
func ValidateNonNegativeInt(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s cannot be negative", fieldName)
	}
	return nil
}

// ValidateBudgetRange validates that min_budget <= max_budget
func ValidateBudgetRange(min, max float64) error {
	if min > 0 && max > 0 && min > max {
		return errors.New("minimum budget cannot be greater than maximum budget")
	}
	return nil
}

// ValidatePhoneNumber validates that phone number is exactly 10 digits
func ValidatePhoneNumber(phone string) error {
	trimmed := strings.TrimSpace(phone)

	if trimmed == "" {
		return errors.New("phone number cannot be empty")
	}

	// Remove common non-digit characters
	cleaned := strings.ReplaceAll(trimmed, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Check if only digits remain
	for _, char := range cleaned {
		if char < '0' || char > '9' {
			return errors.New("phone number must contain only digits")
		}
	}

	// Check length
	if len(cleaned) != 10 {
		return errors.New("phone number must be exactly 10 digits")
	}

	return nil
}

// ValidateContactName validates that a name contains only alphabetic characters
func ValidateContactName(name, fieldName string) error {
	trimmed := strings.TrimSpace(name)

	if trimmed == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	// Check minimum length (at least 2 characters)
	if len(trimmed) < 2 {
		return fmt.Errorf("%s must be at least 2 characters long", fieldName)
	}

	// Check that name contains only alphabets
	for _, char := range trimmed {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			return fmt.Errorf("%s must contain only alphabetic characters", fieldName)
		}
	}

	return nil
}

// NormalizeEmail validates and normalizes an email address.
// It trims whitespace, converts to lowercase, and validates the format.
// Returns normalized email or error with descriptive message.
func NormalizeEmail(email string) (string, error) {
	trimmed := strings.TrimSpace(email)

	if trimmed == "" {
		return "", errors.New("email cannot be empty")
	}

	// Normalize to lowercase
	normalized := strings.ToLower(trimmed)

	// Validate format using pre-compiled regex
	if !emailRegex.MatchString(normalized) {
		return "", errors.New("email must be a valid format (e.g., user@example.com)")
	}

	// Ensure exactly one @ symbol
	if strings.Count(normalized, "@") != 1 {
		return "", errors.New("email must contain exactly one @ symbol")
	}

	return normalized, nil
}

// ValidatePlainTextTemplate validates template body requirements
func ValidatePlainTextTemplate(htmlBody, plainTextBody string) error {
	// At least one body type must be provided
	if strings.TrimSpace(htmlBody) == "" && strings.TrimSpace(plainTextBody) == "" {
		return errors.New("at least one of html_body or plain_text_body is required")
	}
	return nil
}

// TrimAndValidate trims whitespace and validates that the result is not empty
func TrimAndValidate(value, fieldName string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s cannot be empty", fieldName)
	}
	return trimmed, nil
}

// ContainsPlaceholder checks if a string contains placeholder text that shouldn't be saved
func ContainsPlaceholder(value string, placeholders []string) bool {
	valueLower := strings.ToLower(strings.TrimSpace(value))
	for _, placeholder := range placeholders {
		if strings.Contains(valueLower, strings.ToLower(placeholder)) {
			return true
		}
	}
	return false
}

// ValidateTemplateContent validates that template content doesn't contain placeholder values
func ValidateTemplateContent(fromName, subject, htmlBody, plainTextBody string) error {
	placeholders := []string{"Your Organization", "Organization Name", "[Organization]", "{{org}}", "Example Org"}

	fields := map[string]string{
		"from_name":       fromName,
		"subject":         subject,
		"html_body":       htmlBody,
		"plain_text_body": plainTextBody,
	}

	for fieldName, fieldValue := range fields {
		if fieldValue != "" && ContainsPlaceholder(fieldValue, placeholders) {
			return fmt.Errorf("%s contains placeholder text that should be replaced with actual values", fieldName)
		}
	}

	return nil
}

// ValidateOrganizationName validates that organization name contains only alphabetic characters and spaces
func ValidateOrganizationName(name string) error {
	trimmed := strings.TrimSpace(name)

	if trimmed == "" {
		return errors.New("organization name cannot be empty")
	}

	// Check that name contains only alphabets and spaces
	for _, char := range trimmed {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == ' ') {
			return errors.New("organization name can only contain alphabetic characters and spaces")
		}
	}

	// Ensure no multiple consecutive spaces
	if strings.Contains(trimmed, "  ") {
		return errors.New("organization name cannot contain multiple consecutive spaces")
	}

	return nil
}

// ValidateAgentName validates that name contains only alphabets and exactly one whitespace between first and last name
func ValidateAgentName(name string) error {
	trimmed := strings.TrimSpace(name)

	if trimmed == "" {
		return errors.New("name cannot be empty")
	}

	// Count spaces
	spaceCount := strings.Count(trimmed, " ")
	if spaceCount != 1 {
		return errors.New("name must contain exactly one space between first name and surname")
	}

	// Split by space to check each part
	parts := strings.Split(trimmed, " ")
	if len(parts) != 2 {
		return errors.New("name must contain exactly one space between first name and surname")
	}

	// Check that each part contains only alphabets and has minimum length
	for i, part := range parts {
		if part == "" {
			return errors.New("name must contain exactly one space between first name and surname")
		}

		// Minimum length check (at least 2 characters)
		if len(part) < 2 {
			if i == 0 {
				return errors.New("first name must be at least 2 characters long")
			}
			return errors.New("last name must be at least 2 characters long")
		}

		// Alphabets only check
		for _, char := range part {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
				return errors.New("name must contain only alphabetic characters")
			}
		}
	}

	return nil
}

// ValidatePassword validates password strength with strict requirements
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	// Check length constraints
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	if len(password) > 20 {
		return errors.New("password must not exceed 20 characters")
	}

	hasLower := false
	hasDigit := false
	hasSpecial := false

	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?/~`"

	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
		for _, sp := range specialChars {
			if char == sp {
				hasSpecial = true
				break
			}
		}
	}

	// Build a single combined error message
	var missing []string

	if !hasLower {
		missing = append(missing, "one lowercase letter")
	}
	if !hasDigit {
		missing = append(missing, "one numeric digit")
	}
	if !hasSpecial {
		missing = append(missing, "one special character")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain at least: %s", strings.Join(missing, ", "))
	}

	return nil
}
