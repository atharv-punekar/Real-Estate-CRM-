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

func TestCreateAudience_Success(t *testing.T) {
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

	// Create audience request
	reqBody := map[string]interface{}{
		"name":               "Test Audience",
		"description":        "Test Description",
		"property_type":      []string{"apartment"},
		"bedrooms":           []int{3},
		"bathrooms":          []int{2},
		"preferred_location": []string{"Downtown"},
		"min_budget":         200000,
		"max_budget":         500000,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/audiences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestCreateAudience_MissingName(t *testing.T) {
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

	// Missing name
	reqBody := map[string]interface{}{
		"description": "Test Description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/audiences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetAudiences_Success(t *testing.T) {
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

	// Create audiences
	for i := 1; i <= 2; i++ {
		audience := models.Audience{
			OrganizationID: org.ID.String(),
			Name:           "Audience " + string(rune(i)),
			Description:    "Test",
			CreatedBy:      user.ID.String(),
		}
		db.Create(&audience)
	}

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/audiences", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var audiences []interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &audiences)

	assert.GreaterOrEqual(t, len(audiences), 2)
}

func TestGetAudienceByID_Success(t *testing.T) {
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

	// Create audience
	audience := models.Audience{
		OrganizationID: org.ID.String(),
		Name:           "Test Audience",
		Description:    "Test",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&audience)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/audiences/"+audience.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUpdateAudience_Success(t *testing.T) {
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

	// Create audience
	audience := models.Audience{
		OrganizationID: org.ID.String(),
		Name:           "Test Audience",
		Description:    "Test",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&audience)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	// Update audience
	updateBody := map[string]interface{}{
		"name":        "Updated Audience",
		"description": "Updated Description",
	}
	body, _ := json.Marshal(updateBody)

	req := httptest.NewRequest("PUT", "/api/audiences/"+audience.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDeleteAudience_Success(t *testing.T) {
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

	// Create audience
	audience := models.Audience{
		OrganizationID: org.ID.String(),
		Name:           "Test Audience",
		Description:    "Test",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&audience)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("DELETE", "/api/audiences/"+audience.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
