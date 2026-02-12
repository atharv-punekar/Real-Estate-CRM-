package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"gorm.io/gorm"
)

// RowError represents an error in a CSV row
type RowError struct {
	Row    int      `json:"row"`
	Errors []string `json:"errors"`
}

type ContactService struct {
	contactRepo *repository.ContactRepository
}

func NewContactService() *ContactService {
	return &ContactService{
		contactRepo: &repository.ContactRepository{},
	}
}

// ValidateContact validates contact data
func (s *ContactService) ValidateContact(contact *models.Contact) error {
	// Validate first name (alphabets only, min 2 chars)
	trimmedFirstName := strings.TrimSpace(contact.FirstName)
	if trimmedFirstName == "" {
		return errors.New("first_name is required")
	}
	if len(trimmedFirstName) < 2 {
		return errors.New("first_name must be at least 2 characters long")
	}
	for _, char := range trimmedFirstName {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			return errors.New("first_name must contain only alphabetic characters")
		}
	}

	// Validate last name (alphabets only, min 2 chars)
	trimmedLastName := strings.TrimSpace(contact.LastName)
	if trimmedLastName == "" {
		return errors.New("last_name is required")
	}
	if len(trimmedLastName) < 2 {
		return errors.New("last_name must be at least 2 characters long")
	}
	for _, char := range trimmedLastName {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			return errors.New("last_name must contain only alphabetic characters")
		}
	}

	// Validate and normalize email (required and valid format)
	normalizedEmail, err := utils.NormalizeEmail(contact.Email)
	if err != nil {
		return err
	}
	contact.Email = normalizedEmail
	if strings.TrimSpace(contact.Phone) == "" {
		return errors.New("phone is required")
	}
	if strings.TrimSpace(contact.PropertyType) == "" {
		return errors.New("property_type is required")
	}
	if strings.TrimSpace(contact.PreferredLocation) == "" {
		return errors.New("preferred_location is required")
	}

	// Phone number validation (exactly 10 digits)
	// Clean phone number
	cleaned := strings.ReplaceAll(contact.Phone, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Check if only digits
	for _, char := range cleaned {
		if char < '0' || char > '9' {
			return errors.New("phone number must contain only digits")
		}
	}

	// Check length
	if len(cleaned) != 10 {
		return errors.New("phone number must be exactly 10 digits")
	}

	// Validate budget (non-negative and max >= min)
	if contact.BudgetMin < 0 {
		return errors.New("minimum budget cannot be negative")
	}
	if contact.BudgetMax < 0 {
		return errors.New("maximum budget cannot be negative")
	}
	if contact.BudgetMin > 0 && contact.BudgetMax > 0 && contact.BudgetMin > contact.BudgetMax {
		return errors.New("maximum budget must be greater than or equal to minimum budget")
	}

	// Validate bedrooms (0-5)
	if contact.Bedrooms < 0 {
		return errors.New("bedrooms cannot be negative")
	}
	if contact.Bedrooms > 5 {
		return errors.New("bedrooms cannot exceed 5")
	}

	// Validate bathrooms (0-5)
	if contact.Bathrooms < 0 {
		return errors.New("bathrooms cannot be negative")
	}
	if contact.Bathrooms > 5 {
		return errors.New("bathrooms cannot exceed 5")
	}

	// Validate square feet (non-negative)
	if contact.SquareFeet < 0 {
		return errors.New("square feet cannot be negative")
	}

	return nil
}

// CreateContact creates a new contact with uniqueness check
func (s *ContactService) CreateContact(contact *models.Contact) error {
	// Validate contact
	if err := s.ValidateContact(contact); err != nil {
		return err
	}

	// Check for duplicate email in organization
	if contact.Email != "" {
		existing, err := s.contactRepo.FindByEmailOrPhone(contact.Email, "", contact.OrganizationID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if existing != nil {
			return errors.New("contact with this email already exists in your organization")
		}
	}

	// Set default active status
	contact.IsActive = true

	return s.contactRepo.Create(contact)
}

// ParseCSV parses a CSV file and returns valid contacts and row errors
func (s *ContactService) ParseCSV(file io.Reader, orgID, createdBy string) ([]models.Contact, []RowError, int, error) {
	// Read all content first to detect delimiter
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, 0, errors.New("failed to read file content")
	}

	// Auto-detect delimiter: check first line for comma or tab
	firstLine := string(content)
	if idx := strings.Index(firstLine, "\n"); idx != -1 {
		firstLine = firstLine[:idx]
	}

	delimiter := ','
	if strings.Count(firstLine, "\t") > strings.Count(firstLine, ",") {
		delimiter = '\t'
	}

	// Create reader with detected delimiter
	reader := csv.NewReader(strings.NewReader(string(content)))
	reader.Comma = delimiter

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, 0, errors.New("failed to read CSV headers")
	}

	// Map headers to indices — clean BOM & normalize
	headerMap := make(map[string]int)
	for i, header := range headers {
		clean := strings.ToLower(
			strings.TrimSpace(
				strings.ReplaceAll(header, "\uFEFF", ""), // remove UTF-8 BOM
			),
		)
		headerMap[clean] = i
	}

	// Required headers validation
	requiredHeaders := []string{"first_name", "last_name", "email", "phone", "budget_min", "budget_max", "property_type", "bedrooms", "bathrooms", "square_feet", "preferred_location"}
	for _, required := range requiredHeaders {
		if _, exists := headerMap[required]; !exists {
			return nil, nil, 0, fmt.Errorf("CSV must contain '%s' column", required)
		}
	}

	var contacts []models.Contact
	var rowErrors []RowError
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			rowErrors = append(rowErrors, RowError{
				Row:    lineNum,
				Errors: []string{fmt.Sprintf("error reading CSV line: %v", err)},
			})
			lineNum++
			continue
		}
		lineNum++

		contact := models.Contact{
			OrganizationID: orgID,
			CreatedBy:      createdBy,
			IsActive:       true,
		}

		// Parse and validate required fields
		var currentErrors []string

		// First Name
		if idx, ok := headerMap["first_name"]; ok && idx < len(record) {
			contact.FirstName = strings.TrimSpace(record[idx])
			if contact.FirstName == "" {
				currentErrors = append(currentErrors, "first_name is required")
			} else if len(contact.FirstName) < 2 {
				currentErrors = append(currentErrors, "first_name must be at least 2 characters long")
			} else {
				// Check alphabets only
				for _, char := range contact.FirstName {
					if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
						currentErrors = append(currentErrors, "first_name must contain only alphabetic characters")
						break
					}
				}
			}
		} else {
			currentErrors = append(currentErrors, "first_name is required")
		}

		// Last Name
		if idx, ok := headerMap["last_name"]; ok && idx < len(record) {
			contact.LastName = strings.TrimSpace(record[idx])
			if contact.LastName == "" {
				currentErrors = append(currentErrors, "last_name is required")
			} else if len(contact.LastName) < 2 {
				currentErrors = append(currentErrors, "last_name must be at least 2 characters long")
			} else {
				// Check alphabets only
				for _, char := range contact.LastName {
					if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
						currentErrors = append(currentErrors, "last_name must contain only alphabetic characters")
						break
					}
				}
			}
		} else {
			currentErrors = append(currentErrors, "last_name is required")
		}

		// Email
		if idx, ok := headerMap["email"]; ok && idx < len(record) {
			email := strings.TrimSpace(record[idx])
			if email == "" {
				currentErrors = append(currentErrors, "email is required")
			} else {
				// Validate and normalize email
				normalizedEmail, err := utils.NormalizeEmail(email)
				if err != nil {
					currentErrors = append(currentErrors, "email: "+err.Error())
				} else {
					contact.Email = normalizedEmail
				}
			}
		} else {
			currentErrors = append(currentErrors, "email is required")
		}

		// Phone
		if idx, ok := headerMap["phone"]; ok && idx < len(record) {
			contact.Phone = strings.TrimSpace(record[idx])
			if contact.Phone == "" {
				currentErrors = append(currentErrors, "phone is required")
			} else {
				// Validate phone number (exactly 10 digits)
				cleaned := strings.ReplaceAll(contact.Phone, "-", "")
				cleaned = strings.ReplaceAll(cleaned, " ", "")
				cleaned = strings.ReplaceAll(cleaned, "(", "")
				cleaned = strings.ReplaceAll(cleaned, ")", "")
				cleaned = strings.ReplaceAll(cleaned, "+", "")

				isDigit := true
				for _, char := range cleaned {
					if char < '0' || char > '9' {
						isDigit = false
						break
					}
				}

				if !isDigit {
					currentErrors = append(currentErrors, "phone must contain only digits")
				} else if len(cleaned) != 10 {
					currentErrors = append(currentErrors, "phone must be exactly 10 digits")
				}
			}
		} else {
			currentErrors = append(currentErrors, "phone is required")
		}

		// Property Type
		if idx, ok := headerMap["property_type"]; ok && idx < len(record) {
			contact.PropertyType = strings.TrimSpace(record[idx])
			if contact.PropertyType == "" {
				currentErrors = append(currentErrors, "property_type is required")
			}
		} else {
			currentErrors = append(currentErrors, "property_type is required")
		}

		// Preferred Location
		if idx, ok := headerMap["preferred_location"]; ok && idx < len(record) {
			contact.PreferredLocation = strings.TrimSpace(record[idx])
			if contact.PreferredLocation == "" {
				currentErrors = append(currentErrors, "preferred_location is required")
			}
		} else {
			currentErrors = append(currentErrors, "preferred_location is required")
		}

		// Budget Min
		if idx, ok := headerMap["budget_min"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				currentErrors = append(currentErrors, "budget_min is required")
			} else {
				if budgetMin, err := strconv.ParseFloat(val, 64); err == nil {
					if budgetMin < 0 {
						currentErrors = append(currentErrors, "budget_min cannot be negative")
					} else {
						contact.BudgetMin = budgetMin
					}
				} else {
					currentErrors = append(currentErrors, "budget_min must be a valid number")
				}
			}
		} else {
			currentErrors = append(currentErrors, "budget_min is required")
		}

		// Budget Max
		if idx, ok := headerMap["budget_max"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				currentErrors = append(currentErrors, "budget_max is required")
			} else {
				if budgetMax, err := strconv.ParseFloat(val, 64); err == nil {
					if budgetMax < 0 {
						currentErrors = append(currentErrors, "budget_max cannot be negative")
					} else {
						contact.BudgetMax = budgetMax
					}
				} else {
					currentErrors = append(currentErrors, "budget_max must be a valid number")
				}
			}
		} else {
			currentErrors = append(currentErrors, "budget_max is required")
		}

		// Bedrooms
		if idx, ok := headerMap["bedrooms"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				currentErrors = append(currentErrors, "bedrooms is required")
			} else {
				if bedrooms, err := strconv.Atoi(val); err == nil {
					if bedrooms < 0 {
						currentErrors = append(currentErrors, "bedrooms cannot be negative")
					} else if bedrooms > 5 {
						currentErrors = append(currentErrors, "bedrooms cannot exceed 5")
					} else {
						contact.Bedrooms = bedrooms
					}
				} else {
					currentErrors = append(currentErrors, "bedrooms must be a valid number")
				}
			}
		} else {
			currentErrors = append(currentErrors, "bedrooms is required")
		}

		// Bathrooms
		if idx, ok := headerMap["bathrooms"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				currentErrors = append(currentErrors, "bathrooms is required")
			} else {
				if bathrooms, err := strconv.Atoi(val); err == nil {
					if bathrooms < 0 {
						currentErrors = append(currentErrors, "bathrooms cannot be negative")
					} else if bathrooms > 5 {
						currentErrors = append(currentErrors, "bathrooms cannot exceed 5")
					} else {
						contact.Bathrooms = bathrooms
					}
				} else {
					currentErrors = append(currentErrors, "bathrooms must be a valid number")
				}
			}
		} else {
			currentErrors = append(currentErrors, "bathrooms is required")
		}

		// Square Feet
		if idx, ok := headerMap["square_feet"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				currentErrors = append(currentErrors, "square_feet is required")
			} else {
				if sqft, err := strconv.Atoi(val); err == nil {
					if sqft < 0 {
						currentErrors = append(currentErrors, "square_feet cannot be negative")
					} else {
						contact.SquareFeet = sqft
					}
				} else {
					currentErrors = append(currentErrors, "square_feet must be a valid number")
				}
			}
		} else {
			currentErrors = append(currentErrors, "square_feet is required")
		}

		// Validate budget range
		if contact.BudgetMin > 0 && contact.BudgetMax > 0 && contact.BudgetMin > contact.BudgetMax {
			currentErrors = append(currentErrors, "budget_min cannot be greater than budget_max")
		}

		// If there are any row errors, skip this row
		if len(currentErrors) > 0 {
			rowErrors = append(rowErrors, RowError{
				Row:    lineNum,
				Errors: currentErrors,
			})
			continue
		}

		contacts = append(contacts, contact)
	}

	return contacts, rowErrors, lineNum - 1, nil
}

// BulkCreateContacts creates multiple contacts, skipping duplicates and returning detailed stats
// Returns: importedCount, duplicateCount, failedCount, failures, systemError
func (s *ContactService) BulkCreateContacts(contacts []models.Contact) (int, int, int, []string, error) {
	successCount := 0
	duplicateCount := 0
	failedCount := 0
	var failureMessages []string

	for _, contact := range contacts {
		// Check for duplicate email
		if contact.Email != "" {
			existing, err := s.contactRepo.FindByEmailOrPhone(contact.Email, "", contact.OrganizationID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				msg := fmt.Sprintf("Error checking duplicate email %s: %v", contact.Email, err)
				failureMessages = append(failureMessages, msg)
				failedCount++
				continue
			}
			if existing != nil {
				duplicateCount++
				continue
			}
		}

		// Create contact
		if err := s.contactRepo.Create(&contact); err != nil {
			msg := fmt.Sprintf("Failed to create contact %s: %v", contact.Email, err)
			failureMessages = append(failureMessages, msg)
			failedCount++
			continue
		}
		successCount++
	}

	// Log summary (removed fmt.Printf for cleaner logs)
	return successCount, duplicateCount, failedCount, failureMessages, nil
}
