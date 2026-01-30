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
	// At least one of email or phone must be provided
	if contact.Email == "" && contact.Phone == "" {
		return errors.New("at least one of email or phone is required")
	}

	// Validate budget range
	if contact.BudgetMin > 0 && contact.BudgetMax > 0 && contact.BudgetMin > contact.BudgetMax {
		return errors.New("budget_min cannot be greater than budget_max")
	}

	return nil
}

// CreateContact creates a new contact with uniqueness check
func (s *ContactService) CreateContact(contact *models.Contact) error {
	// Validate contact
	if err := s.ValidateContact(contact); err != nil {
		return err
	}

	// Check for duplicates within organization
	existing, err := s.contactRepo.FindByEmailOrPhone(contact.Email, contact.Phone, contact.OrganizationID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		if existing.Email == contact.Email && contact.Email != "" {
			return errors.New("contact with this email already exists in your organization")
		}
		if existing.Phone == contact.Phone && contact.Phone != "" {
			return errors.New("contact with this phone already exists in your organization")
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
		fmt.Println("DEBUG: Detected TAB delimiter")
	} else {
		fmt.Println("DEBUG: Detected COMMA delimiter")
	}

	// Create reader with detected delimiter
	reader := csv.NewReader(strings.NewReader(string(content)))
	reader.Comma = delimiter

	// Read header
	headers, err := reader.Read()
	fmt.Println("DEBUG HEADERS:", headers)
	fmt.Printf("DEBUG: Number of headers: %d\n", len(headers))
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

	// Required headers
	requiredHeaders := []string{"email", "phone"}
	hasRequired := false
	for _, req := range requiredHeaders {
		if _, ok := headerMap[req]; ok {
			hasRequired = true
			break
		}
	}
	if !hasRequired {
		return nil, errors.New("CSV must contain at least one of: email, phone")
	}

	var contacts []models.Contact
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV line %d: %v", lineNum, err)
		}
		lineNum++

		contact := models.Contact{
			OrganizationID: orgID,
			CreatedBy:      createdBy,
			IsActive:       true,
		}

		// Parse fields
		if idx, ok := headerMap["first_name"]; ok && idx < len(record) {
			contact.FirstName = strings.TrimSpace(record[idx])
		}
		if idx, ok := headerMap["last_name"]; ok && idx < len(record) {
			contact.LastName = strings.TrimSpace(record[idx])
		}
		if idx, ok := headerMap["email"]; ok && idx < len(record) {
			contact.Email = strings.TrimSpace(record[idx])
		}
		if idx, ok := headerMap["phone"]; ok && idx < len(record) {
			contact.Phone = strings.TrimSpace(record[idx])
		}
		if idx, ok := headerMap["budget_min"]; ok && idx < len(record) {
			if val := strings.TrimSpace(record[idx]); val != "" {
				if budgetMin, err := strconv.ParseFloat(val, 64); err == nil {
					contact.BudgetMin = budgetMin
				}
			}
		}
		if idx, ok := headerMap["budget_max"]; ok && idx < len(record) {
			if val := strings.TrimSpace(record[idx]); val != "" {
				if budgetMax, err := strconv.ParseFloat(val, 64); err == nil {
					contact.BudgetMax = budgetMax
				}
			}
		}
		if idx, ok := headerMap["property_type"]; ok && idx < len(record) {
			contact.PropertyType = strings.TrimSpace(record[idx])
		}
		if idx, ok := headerMap["bedrooms"]; ok && idx < len(record) {
			if val := strings.TrimSpace(record[idx]); val != "" {
				if bedrooms, err := strconv.Atoi(val); err == nil {
					contact.Bedrooms = bedrooms
				}
			}
		}
		if idx, ok := headerMap["bathrooms"]; ok && idx < len(record) {
			if val := strings.TrimSpace(record[idx]); val != "" {
				if bathrooms, err := strconv.Atoi(val); err == nil {
					contact.Bathrooms = bathrooms
				}
			}
		}
		if idx, ok := headerMap["square_feet"]; ok && idx < len(record) {
			if val := strings.TrimSpace(record[idx]); val != "" {
				if sqft, err := strconv.Atoi(val); err == nil {
					contact.SquareFeet = sqft
				}
			}
		}
		if idx, ok := headerMap["preferred_location"]; ok && idx < len(record) {
			contact.PreferredLocation = strings.TrimSpace(record[idx])
		}
		if idx, ok := headerMap["notes"]; ok && idx < len(record) {
			contact.Notes = strings.TrimSpace(record[idx])
		}

		// Skip if both email and phone are empty
		if contact.Email == "" && contact.Phone == "" {
			continue
		}

		contacts = append(contacts, contact)
	}

	if len(contacts) == 0 {
		return nil, errors.New("no valid contacts found in CSV")
	}

	fmt.Printf("DEBUG: Parsed %d contacts from CSV\n", len(contacts))
	if len(contacts) > 0 {
		fmt.Printf("DEBUG: Sample contact: Email=%s, Name=%s %s\n",
			contacts[0].Email, contacts[0].FirstName, contacts[0].LastName)
	}

	return contacts, nil
}

// BulkCreateContacts creates multiple contacts, skipping duplicates
func (s *ContactService) BulkCreateContacts(contacts []models.Contact) (int, int, error) {
	successCount := 0
	skipCount := 0

	fmt.Printf("DEBUG: Starting bulk create for %d contacts\n", len(contacts))

	for i, contact := range contacts {
		// Check for duplicates
		existing, err := s.contactRepo.FindByEmailOrPhone(contact.Email, contact.Phone, contact.OrganizationID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("DEBUG: Error checking duplicate for contact %d: %v\n", i+1, err)
			continue // Skip on error
		}
		if existing != nil {
			fmt.Printf("DEBUG: Skipping duplicate contact %d: Email=%s, Phone=%s\n", i+1, contact.Email, contact.Phone)
			skipCount++
			continue // Skip duplicate
		}

		// Create contact
		if err := s.contactRepo.Create(&contact); err != nil {
			fmt.Printf("DEBUG: Failed to create contact %d: %v\n", i+1, err)
			skipCount++
			continue
		}
		successCount++
		if successCount%10 == 0 {
			fmt.Printf("DEBUG: Created %d contacts so far...\n", successCount)
		}
	}

	fmt.Printf("DEBUG: Bulk create completed: %d success, %d skipped\n", successCount, skipCount)

	return successCount, skipCount, nil
}
