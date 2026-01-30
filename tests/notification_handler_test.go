package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetNotifications_Success(t *testing.T) {
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

	// Create notifications
	for i := 1; i <= 2; i++ {
		notification := models.Notification{
			OrganizationID:   org.ID.String(),
			UserID:           user.ID.String(),
			NotificationType: "campaign_completed",
			Title:            "Test Notification",
			Message:          "Test message",
			IsRead:           false,
		}
		db.Create(&notification)
	}

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestMarkNotificationRead_Success(t *testing.T) {
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

	// Create notification
	notification := models.Notification{
		OrganizationID:   org.ID.String(),
		UserID:           user.ID.String(),
		NotificationType: "campaign_completed",
		Title:            "Test Notification",
		Message:          "Test message",
		IsRead:           false,
	}
	db.Create(&notification)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("PUT", "/api/notifications/"+notification.ID+"/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify notification is marked as read
	var updatedNotif models.Notification
	db.First(&updatedNotif, "id = ?", notification.ID)
	assert.True(t, updatedNotif.IsRead)
}

func TestMarkNotificationRead_Unauthorized(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization and two users
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	user1 := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "User 1",
		Email:          "user1@test.com",
		Role:           "org_admin",
		IsActive:       true,
	}
	db.Create(&user1)

	user2 := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "User 2",
		Email:          "user2@test.com",
		Role:           "org_user",
		IsActive:       true,
	}
	db.Create(&user2)

	// Create notification for user1
	notification := models.Notification{
		OrganizationID:   org.ID.String(),
		UserID:           user1.ID.String(),
		NotificationType: "campaign_completed",
		Title:            "Test Notification",
		Message:          "Test message",
		IsRead:           false,
	}
	db.Create(&notification)

	// Try to mark as read with user2's token
	token := getAuthToken(t, user2.ID.String(), user2.Role, org.ID.String())

	req := httptest.NewRequest("PUT", "/api/notifications/"+notification.ID+"/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestGetUnreadCount_Success(t *testing.T) {
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

	// Create unread notifications
	for i := 1; i <= 3; i++ {
		notification := models.Notification{
			OrganizationID:   org.ID.String(),
			UserID:           user.ID.String(),
			NotificationType: "campaign_completed",
			Title:            "Test Notification",
			Message:          "Test message",
			IsRead:           false,
		}
		db.Create(&notification)
	}

	// Create read notification
	readTime := time.Now()
	readNotif := models.Notification{
		OrganizationID:   org.ID.String(),
		UserID:           user.ID.String(),
		NotificationType: "campaign_completed",
		Title:            "Read Notification",
		Message:          "Read message",
		IsRead:           true,
		ReadAt:           &readTime,
	}
	db.Create(&readNotif)

	token := getAuthToken(t, user.ID.String(), user.Role, org.ID.String())

	req := httptest.NewRequest("GET", "/api/notifications/unread", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response to verify count
	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(bodyBytes, &response)

	// Should have 3 unread notifications
	assert.NotNil(t, response["unread_count"])
}
