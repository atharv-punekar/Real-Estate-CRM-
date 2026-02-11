package utils

import (
	"errors"
	"fmt"
	"strings"
)

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

// NormalizeEmail converts email to lowercase and validates format
func NormalizeEmail(email string) (string, error) {
	if email == "" {
		return "", nil // Empty email is handled by other validators
	}

	normalized := strings.ToLower(strings.TrimSpace(email))

	if !IsValidEmail(normalized) {
		return "", errors.New("invalid email format")
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

// ValidateAgentEmail validates that email contains only lowercase letters and numbers with exactly 1 @ and 1 domain
func ValidateAgentEmail(email string) error {
	trimmed := strings.TrimSpace(email)

	if trimmed == "" {
		return errors.New("email cannot be empty")
	}

	// Check for exactly one @ symbol
	atCount := strings.Count(trimmed, "@")
	if atCount != 1 {
		return errors.New("email must contain exactly one @ symbol")
	}

	// Split by @ to get local and domain parts
	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("invalid email format")
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Validate local part: only lowercase letters and numbers
	for _, char := range localPart {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return errors.New("email local part must contain only lowercase letters and numbers")
		}
	}

	// Validate domain part: must have exactly one dot and follow pattern text.text
	dotCount := strings.Count(domainPart, ".")
	if dotCount != 1 {
		return errors.New("email domain must contain exactly one dot (e.g., example.com)")
	}

	domainParts := strings.Split(domainPart, ".")
	if len(domainParts) != 2 || domainParts[0] == "" || domainParts[1] == "" {
		return errors.New("invalid email domain format")
	}

	// Validate each domain part: only lowercase letters and numbers
	for _, domPart := range domainParts {
		for _, char := range domPart {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				return errors.New("email domain must contain only lowercase letters and numbers")
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

	// Check for at least one lowercase letter
	hasLower := false
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one numeric digit
	hasDigit := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return errors.New("password must contain at least one numeric digit")
	}

	// Check for at least one special character
	hasSpecial := false
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?/~`"
	for _, char := range password {
		for _, special := range specialChars {
			if char == special {
				hasSpecial = true
				break
			}
		}
		if hasSpecial {
			break
		}
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?/~`)")
	}

	return nil
}
