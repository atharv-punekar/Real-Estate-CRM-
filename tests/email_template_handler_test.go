package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateEmailTemplate_Success(t *testing.T) {
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

	// Create template request
	reqBody := map[string]interface{}{
		"name":            "Test Template",
		"subject":         "Test Subject",
		"preheader":       "Test Preheader",
		"from_name":       "Test Sender",
		"reply_to":        "reply@test.com",
		"html_body":       "<html><body><p>Test email</p></body></html>",
		"plain_text_body": "Test email",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/templates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestCreateEmailTemplate_MissingFields(t *testing.T) {
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

	// Missing required fields
	reqBody := map[string]interface{}{
		"name": "Test Template",
		// Missing subject and html_body
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/templates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetEmailTemplates_Success(t *testing.T) {
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

	// Create templates
	for i := 1; i <= 2; i++ {
		template := models.EmailTemplate{
			OrganizationID: org.ID.String(),
			Name:           "Template " + string(rune(i)),
			Subject:        "Subject",
			HtmlBody:       "<p>Body</p>",
			CreatedBy:      user.ID.String(),
		}
		db.Create(&template)
	}

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/templates", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var templates []interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &templates)

	assert.GreaterOrEqual(t, len(templates), 2)
}

func TestGetEmailTemplateByID_Success(t *testing.T) {
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

	// Create template
	template := models.EmailTemplate{
		OrganizationID: org.ID.String(),
		Name:           "Test Template",
		Subject:        "Test Subject",
		HtmlBody:       "<p>Test</p>",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&template)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/templates/"+template.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUpdateEmailTemplate_Success(t *testing.T) {
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

	// Create template
	template := models.EmailTemplate{
		OrganizationID: org.ID.String(),
		Name:           "Test Template",
		Subject:        "Test Subject",
		HtmlBody:       "<p>Test</p>",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&template)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	// Update template
	updateBody := map[string]interface{}{
		"name":      "Updated Template",
		"subject":   "Updated Subject",
		"html_body": "<p>Updated body</p>",
	}
	body, _ := json.Marshal(updateBody)

	req := httptest.NewRequest("PUT", "/api/templates/"+template.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDeleteEmailTemplate_Success(t *testing.T) {
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

	// Create template
	template := models.EmailTemplate{
		OrganizationID: org.ID.String(),
		Name:           "Test Template",
		Subject:        "Test Subject",
		HtmlBody:       "<p>Test</p>",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&template)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("DELETE", "/api/templates/"+template.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
