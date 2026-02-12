package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestImportContactsCSV_PartialSuccess(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create request specifics
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	user := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "Test User",
		Email:          "user@test.com",
		Role:           "org_admin",
		IsActive:       true,
	}
	db.Create(&user)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	// CSV Content
	csvContent := "first_name,last_name,email,phone,budget_min,budget_max,property_type,bedrooms,bathrooms,square_feet,preferred_location\n" +
		"Valid,User,valid@example.com,1234567891,100000,200000,Apartment,2,1,1000,Downtown\n" +
		"Invalid,User,,1234567892,100000,200000,Apartment,2,1,1000,Downtown\n" +
		"Duplicate,User,duplicate@example.com,1234567893,100000,200000,Apartment,2,1,1000,Downtown\n" +
		"Negative,Budget,negative@example.com,1234567894,-1000,200000,Apartment,2,1,1000,Downtown\n"

	// Create duplicate contact beforehand
	duplicateContact := models.Contact{
		OrganizationID: org.ID.String(),
		FirstName:      "Existing",
		LastName:       "User",
		Email:          "duplicate@example.com",
		Phone:          "1234567890",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&duplicateContact)

	// Create multipart request
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "contacts.csv")
	assert.NoError(t, err)
	part.Write([]byte(csvContent))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/contacts/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var result map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &result)

	// Assertions
	// Total rows: 4 data rows
	assert.Equal(t, float64(4), result["total_rows"])

	// Imported: 1 (Valid User)
	assert.Equal(t, float64(1), result["imported_rows"])

	// Duplicate: 1 (Duplicate User)
	assert.Equal(t, float64(1), result["duplicate_rows"])

	// Invalid: 2 (Missing email, Negative budget)
	assert.Equal(t, float64(2), result["invalid_rows"])
}
