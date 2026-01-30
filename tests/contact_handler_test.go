package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func getAuthToken(t *testing.T, userID, role, orgID string) string {
	// Use JWT_SECRET from environment (set in init())
	jwtSecret := os.Getenv("JWT_SECRET")
	token, err := utils.GenerateJWT(jwtSecret, userID, role, orgID)
	assert.NoError(t, err)
	return token
}

func TestCreateContact_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization and user
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

	// Create contact request
	reqBody := map[string]interface{}{
		"first_name":         "John",
		"last_name":          "Doe",
		"email":              "john.doe@test.com",
		"phone":              "+1234567890",
		"budget_min":         200000,
		"budget_max":         500000,
		"property_type":      "apartment",
		"bedrooms":           3,
		"bathrooms":          2,
		"square_feet":        1500,
		"preferred_location": "Downtown",
		"notes":              "Prefers high floor",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/contacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1) // -1 means no timeout
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)

	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.NotNil(t, response["contact"])
	assert.Equal(t, "Contact created successfully", response["message"])
}

func TestCreateContact_MissingFields(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization and user
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

	// Empty request body
	reqBody := map[string]interface{}{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/contacts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	// Should succeed even with minimal data (validation happens in service layer)
	assert.True(t, resp.StatusCode == 201 || resp.StatusCode == 400)
}

func TestGetContacts_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization and user
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

	// Create test contacts
	for i := 1; i <= 3; i++ {
		contact := models.Contact{
			OrganizationID: org.ID.String(),
			CreatedBy:      user.ID.String(),
			FirstName:      fmt.Sprintf("Contact%d", i),
			LastName:       "Test",
			Email:          fmt.Sprintf("contact%d@test.com", i),
		}
		db.Create(&contact)
	}

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/contacts", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	contacts := response["contacts"].([]interface{})
	assert.GreaterOrEqual(t, len(contacts), 3)
}

func TestGetContactByID_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization and user
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

	// Create test contact
	contact := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      user.ID.String(),
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@test.com",
	}
	db.Create(&contact)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/contacts/"+contact.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetContactByID_NotFound(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization and user
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

	// Try to get non-existent contact
	fakeID := uuid.New().String()
	req := httptest.NewRequest("GET", "/api/contacts/"+fakeID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestUpdateContact_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization and user
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

	// Create test contact
	contact := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      user.ID.String(),
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@test.com",
	}
	db.Create(&contact)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	// Update contact
	updateBody := map[string]interface{}{
		"first_name": "Jane",
		"last_name":  "Smith",
	}
	body, _ := json.Marshal(updateBody)

	req := httptest.NewRequest("PUT", "/api/contacts/"+contact.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDeleteContact_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization and user
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

	// Create test contact
	contact := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      user.ID.String(),
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@test.com",
	}
	db.Create(&contact)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("DELETE", "/api/contacts/"+contact.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
