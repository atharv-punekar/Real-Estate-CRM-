package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	repository "github.com/atharvpunekar/real_estate_crm_backend/internal/repositories"
	"gorm.io/gorm"
)

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

	// Validate email (required and valid format with regex)
	if strings.TrimSpace(contact.Email) == "" {
		return errors.New("email is required")
	}
	// Email regex validation: must have @ and domain
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, contact.Email)
	if !matched {
		return errors.New("email must be a valid format (e.g., user@example.com)")
	}
	if strings.Count(contact.Email, "@") != 1 {
		return errors.New("email must contain exactly one @ symbol")
	}
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

// ParseCSV parses a CSV file and returns contacts
func (s *ContactService) ParseCSV(file io.Reader, orgID, createdBy string) ([]models.Contact, error) {
	// Read all content first to detect delimiter
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.New("failed to read file content")
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
		return nil, errors.New("failed to read CSV headers")
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
			return nil, fmt.Errorf("CSV must contain '%s' column", required)
		}
	}

	var contacts []models.Contact
	var parseErrors []string
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("Row %d: error reading CSV line", lineNum))
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
		var rowErrors []string

		// First Name
		if idx, ok := headerMap["first_name"]; ok && idx < len(record) {
			contact.FirstName = strings.TrimSpace(record[idx])
			if contact.FirstName == "" {
				rowErrors = append(rowErrors, "first_name is required")
			} else if len(contact.FirstName) < 2 {
				rowErrors = append(rowErrors, "first_name must be at least 2 characters long")
			} else {
				// Check alphabets only
				for _, char := range contact.FirstName {
					if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
						rowErrors = append(rowErrors, "first_name must contain only alphabetic characters")
						break
					}
				}
			}
		} else {
			rowErrors = append(rowErrors, "first_name is required")
		}

		// Last Name
		if idx, ok := headerMap["last_name"]; ok && idx < len(record) {
			contact.LastName = strings.TrimSpace(record[idx])
			if contact.LastName == "" {
				rowErrors = append(rowErrors, "last_name is required")
			} else if len(contact.LastName) < 2 {
				rowErrors = append(rowErrors, "last_name must be at least 2 characters long")
			} else {
				// Check alphabets only
				for _, char := range contact.LastName {
					if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
						rowErrors = append(rowErrors, "last_name must contain only alphabetic characters")
						break
					}
				}
			}
		} else {
			rowErrors = append(rowErrors, "last_name is required")
		}

		// Email
		if idx, ok := headerMap["email"]; ok && idx < len(record) {
			contact.Email = strings.TrimSpace(record[idx])
			if contact.Email == "" {
				rowErrors = append(rowErrors, "email is required")
			} else {
				// Email regex validation
				emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
				matched, _ := regexp.MatchString(emailRegex, contact.Email)
				if !matched {
					rowErrors = append(rowErrors, "email must be valid format")
				} else if strings.Count(contact.Email, "@") != 1 {
					rowErrors = append(rowErrors, "email must contain exactly one @ symbol")
				}
			}
		} else {
			rowErrors = append(rowErrors, "email is required")
		}

		// Phone
		if idx, ok := headerMap["phone"]; ok && idx < len(record) {
			contact.Phone = strings.TrimSpace(record[idx])
			if contact.Phone == "" {
				rowErrors = append(rowErrors, "phone is required")
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
					rowErrors = append(rowErrors, "phone must contain only digits")
				} else if len(cleaned) != 10 {
					rowErrors = append(rowErrors, "phone must be exactly 10 digits")
				}
			}
		} else {
			rowErrors = append(rowErrors, "phone is required")
		}

		// Property Type
		if idx, ok := headerMap["property_type"]; ok && idx < len(record) {
			contact.PropertyType = strings.TrimSpace(record[idx])
			if contact.PropertyType == "" {
				rowErrors = append(rowErrors, "property_type is required")
			}
		} else {
			rowErrors = append(rowErrors, "property_type is required")
		}

		// Preferred Location
		if idx, ok := headerMap["preferred_location"]; ok && idx < len(record) {
			contact.PreferredLocation = strings.TrimSpace(record[idx])
			if contact.PreferredLocation == "" {
				rowErrors = append(rowErrors, "preferred_location is required")
			}
		} else {
			rowErrors = append(rowErrors, "preferred_location is required")
		}

		// Budget Min
		if idx, ok := headerMap["budget_min"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				rowErrors = append(rowErrors, "budget_min is required")
			} else {
				if budgetMin, err := strconv.ParseFloat(val, 64); err == nil {
					if budgetMin < 0 {
						rowErrors = append(rowErrors, "budget_min cannot be negative")
					} else {
						contact.BudgetMin = budgetMin
					}
				} else {
					rowErrors = append(rowErrors, "budget_min must be a valid number")
				}
			}
		} else {
			rowErrors = append(rowErrors, "budget_min is required")
		}

		// Budget Max
		if idx, ok := headerMap["budget_max"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				rowErrors = append(rowErrors, "budget_max is required")
			} else {
				if budgetMax, err := strconv.ParseFloat(val, 64); err == nil {
					if budgetMax < 0 {
						rowErrors = append(rowErrors, "budget_max cannot be negative")
					} else {
						contact.BudgetMax = budgetMax
					}
				} else {
					rowErrors = append(rowErrors, "budget_max must be a valid number")
				}
			}
		} else {
			rowErrors = append(rowErrors, "budget_max is required")
		}

		// Bedrooms
		if idx, ok := headerMap["bedrooms"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				rowErrors = append(rowErrors, "bedrooms is required")
			} else {
				if bedrooms, err := strconv.Atoi(val); err == nil {
					if bedrooms < 0 {
						rowErrors = append(rowErrors, "bedrooms cannot be negative")
					} else if bedrooms > 5 {
						rowErrors = append(rowErrors, "bedrooms cannot exceed 5")
					} else {
						contact.Bedrooms = bedrooms
					}
				} else {
					rowErrors = append(rowErrors, "bedrooms must be a valid number")
				}
			}
		} else {
			rowErrors = append(rowErrors, "bedrooms is required")
		}

		// Bathrooms
		if idx, ok := headerMap["bathrooms"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				rowErrors = append(rowErrors, "bathrooms is required")
			} else {
				if bathrooms, err := strconv.Atoi(val); err == nil {
					if bathrooms < 0 {
						rowErrors = append(rowErrors, "bathrooms cannot be negative")
					} else if bathrooms > 5 {
						rowErrors = append(rowErrors, "bathrooms cannot exceed 5")
					} else {
						contact.Bathrooms = bathrooms
					}
				} else {
					rowErrors = append(rowErrors, "bathrooms must be a valid number")
				}
			}
		} else {
			rowErrors = append(rowErrors, "bathrooms is required")
		}

		// Square Feet
		if idx, ok := headerMap["square_feet"]; ok && idx < len(record) {
			val := strings.TrimSpace(record[idx])
			if val == "" {
				rowErrors = append(rowErrors, "square_feet is required")
			} else {
				if sqft, err := strconv.Atoi(val); err == nil {
					if sqft < 0 {
						rowErrors = append(rowErrors, "square_feet cannot be negative")
					} else {
						contact.SquareFeet = sqft
					}
				} else {
					rowErrors = append(rowErrors, "square_feet must be a valid number")
				}
			}
		} else {
			rowErrors = append(rowErrors, "square_feet is required")
		}

		// Validate budget range
		if contact.BudgetMin > 0 && contact.BudgetMax > 0 && contact.BudgetMin > contact.BudgetMax {
			rowErrors = append(rowErrors, "budget_min cannot be greater than budget_max")
		}

		// If there are any row errors, skip this row
		if len(rowErrors) > 0 {
			parseErrors = append(parseErrors, fmt.Sprintf("Row %d: %s", lineNum, strings.Join(rowErrors, "; ")))
			continue
		}

		contacts = append(contacts, contact)
	}

	// If there were any parse errors, return them
	if len(parseErrors) > 0 {
		// Return first 10 errors to avoid overwhelming the response
		maxErrors := 10
		if len(parseErrors) > maxErrors {
			return nil, fmt.Errorf("CSV contains validation errors in %d rows. First %d errors:\n%s",
				len(parseErrors), maxErrors, strings.Join(parseErrors[:maxErrors], "\n"))
		}
		return nil, fmt.Errorf("CSV contains validation errors:\n%s", strings.Join(parseErrors, "\n"))
	}

	if len(contacts) == 0 {
		return nil, errors.New("no valid contacts found in CSV")
	}

	return contacts, nil
}

// BulkCreateContacts creates multiple contacts, skipping duplicates
func (s *ContactService) BulkCreateContacts(contacts []models.Contact) (int, int, error) {
	successCount := 0
	skipCount := 0

	for i, contact := range contacts {
		// Check for duplicate email
		if contact.Email != "" {
			existing, err := s.contactRepo.FindByEmailOrPhone(contact.Email, "", contact.OrganizationID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				fmt.Printf("Row %d: Error checking duplicate email: %v\n", i+2, err)
				skipCount++
				continue
			}
			if existing != nil {
				fmt.Printf("Row %d: Skipped - duplicate email: %s\n", i+2, contact.Email)
				skipCount++
				continue
			}
		}

		// Create contact
		if err := s.contactRepo.Create(&contact); err != nil {
			fmt.Printf("Row %d: Failed to create contact: %v\n", i+2, err)
			skipCount++
			continue
		}
		successCount++
		if successCount%10 == 0 {
			fmt.Printf("Created %d contacts so far...\n", successCount)
		}
	}

	fmt.Printf("Bulk create completed: %d success, %d skipped\n", successCount, skipCount)

	return successCount, skipCount, nil
}
