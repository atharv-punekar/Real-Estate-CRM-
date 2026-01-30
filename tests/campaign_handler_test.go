package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestCreateCampaign_Success(t *testing.T) {
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

	// Create email template
	template := models.EmailTemplate{
		OrganizationID: org.ID.String(),
		Name:           "Test Template",
		Subject:        "Test Subject",
		HtmlBody:       "<p>Test Body</p>",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&template)

	// Create contact
	contact := models.Contact{
		OrganizationID: org.ID.String(),
		CreatedBy:      user.ID.String(),
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@test.com",
	}
	db.Create(&contact)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	// Create campaign request
	scheduledAt := time.Now().Add(1 * time.Hour)
	reqBody := map[string]interface{}{
		"name":          "Test Campaign",
		"template_id":   template.ID,
		"contact_id":    contact.ID,
		"schedule_type": "once",
		"scheduled_at":  scheduledAt.Format(time.RFC3339),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/campaigns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestCreateCampaign_MissingFields(t *testing.T) {
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
		"name": "Test Campaign",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/campaigns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	// Should return error for missing fields
	assert.True(t, resp.StatusCode >= 400)
}

func TestGetCampaigns_Success(t *testing.T) {
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
		Subject:        "Test",
		HtmlBody:       "<p>Test</p>",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&template)

	// Create campaigns
	for i := 1; i <= 2; i++ {
		campaign := models.Campaign{
			OrganizationID: org.ID.String(),
			Name:           "Campaign " + string(rune(i)),
			TemplateID:     template.ID,
			ScheduleType:   "once",
			ScheduledAt:    time.Now().Add(1 * time.Hour),
			Status:         "scheduled",
			CreatedBy:      user.ID.String(),
			AudienceIDs:    datatypes.JSONSlice[string]{},
		}
		db.Create(&campaign)
	}

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/campaigns", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetCampaignByID_Success(t *testing.T) {
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
		Subject:        "Test",
		HtmlBody:       "<p>Test</p>",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&template)

	// Create campaign
	campaign := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Test Campaign",
		TemplateID:     template.ID,
		ScheduleType:   "once",
		ScheduledAt:    time.Now().Add(1 * time.Hour),
		Status:         "scheduled",
		CreatedBy:      user.ID.String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/campaigns/"+campaign.ID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetCampaignLogs_Success(t *testing.T) {
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
		Subject:        "Test",
		HtmlBody:       "<p>Test</p>",
		CreatedBy:      user.ID.String(),
	}
	db.Create(&template)

	// Create campaign
	campaign := models.Campaign{
		OrganizationID: org.ID.String(),
		Name:           "Test Campaign",
		TemplateID:     template.ID,
		ScheduleType:   "once",
		ScheduledAt:    time.Now().Add(1 * time.Hour),
		Status:         "scheduled",
		CreatedBy:      user.ID.String(),
		AudienceIDs:    datatypes.JSONSlice[string]{},
	}
	db.Create(&campaign)

	// Create campaign logs
	log1 := models.CampaignLog{
		CampaignID: campaign.ID,
		ContactID:  uuid.New().String(),
		Status:     "sent",
	}
	db.Create(&log1)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/campaigns/"+campaign.ID+"/logs", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	logs := response["logs"].([]interface{})
	assert.GreaterOrEqual(t, len(logs), 1)
}
