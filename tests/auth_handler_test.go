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
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSuperAdminLogin_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create super admin
	hashedPassword, _ := utils.HashPassword("admin123")
	admin := models.SuperAdmin{
		ID:           uuid.New().String(),
		Name:         "Super Admin",
		Email:        "admin@test.com",
		PasswordHash: hashedPassword,
	}
	db.Create(&admin)

	// Login request
	reqBody := map[string]string{
		"email":    "admin@test.com",
		"password": "admin123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/superadmin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.NotEmpty(t, response["token"])
	assert.NotNil(t, response["user"])
}

func TestSuperAdminLogin_InvalidCredentials(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create super admin
	hashedPassword, _ := utils.HashPassword("admin123")
	admin := models.SuperAdmin{
		ID:           uuid.New().String(),
		Name:         "Super Admin",
		Email:        "admin@test.com",
		PasswordHash: hashedPassword,
	}
	db.Create(&admin)

	// Login with wrong password
	reqBody := map[string]string{
		"email":    "admin@test.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/superadmin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestSuperAdminLogin_InvalidEmail(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// No admin created

	reqBody := map[string]string{
		"email":    "nonexistent@test.com",
		"password": "admin123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/superadmin/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func TestOrgAdminLogin_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create org admin user
	hashedPassword, _ := utils.HashPassword("password123")
	user := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "Org Admin",
		Email:          "orgadmin@test.com",
		PasswordHash:   hashedPassword,
		Role:           "org_admin",
		IsActive:       true,
		IsPasswordSet:  true,
	}
	db.Create(&user)

	// Login request
	reqBody := map[string]string{
		"email":    "orgadmin@test.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.NotEmpty(t, response["token"])
	assert.NotNil(t, response["user"])
}

func TestOrgAdminLogin_Deactivated(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create deactivated user
	hashedPassword, _ := utils.HashPassword("password123")
	user := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "Deactivated User",
		Email:          "deactivated@test.com",
		PasswordHash:   hashedPassword,
		Role:           "org_admin",
		IsActive:       false, // Deactivated
		IsPasswordSet:  true,
	}
	db.Create(&user)

	// Login request
	reqBody := map[string]string{
		"email":    "deactivated@test.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestOrgAdminLogin_PasswordNotSet(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create user without password set
	user := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "New User",
		Email:          "newuser@test.com",
		Role:           "org_admin",
		IsActive:       true,
		IsPasswordSet:  false, // Password not set
	}
	db.Create(&user)

	// Login request
	reqBody := map[string]string{
		"email":    "newuser@test.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestActivatePassword_Success(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create user with invite token
	inviteToken := "valid-invite-token-123"
	tokenExpiry := time.Now().Add(7 * 24 * time.Hour)
	user := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "New User",
		Email:          "newuser@test.com",
		Role:           "org_admin",
		IsActive:       true,
		IsPasswordSet:  false,
		InviteToken:    &inviteToken,
		TokenExpiresAt: &tokenExpiry,
	}
	db.Create(&user)

	// Activate password request
	reqBody := map[string]string{
		"token":    inviteToken,
		"password": "newpassword123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify user password is set
	var updatedUser models.User
	db.Where("email = ?", "newuser@test.com").First(&updatedUser)
	assert.True(t, updatedUser.IsPasswordSet)
	assert.Nil(t, updatedUser.InviteToken)
}

func TestActivatePassword_ExpiredToken(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Create organization
	org := models.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
	}
	db.Create(&org)

	// Create user with expired invite token
	inviteToken := "expired-invite-token-123"
	tokenExpiry := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	user := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "New User",
		Email:          "newuser@test.com",
		Role:           "org_admin",
		IsActive:       true,
		IsPasswordSet:  false,
		InviteToken:    &inviteToken,
		TokenExpiresAt: &tokenExpiry,
	}
	db.Create(&user)

	// Activate password request
	reqBody := map[string]string{
		"token":    inviteToken,
		"password": "newpassword123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestActivatePassword_MissingFields(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Missing password
	reqBody := map[string]string{
		"token": "some-token",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestActivatePassword_ShortPassword(t *testing.T) {
	db := SetupTestDB()
	database.DB = db
	defer CleanupTestDB(db)

	app := SetupTestApp()

	// Password too short
	reqBody := map[string]string{
		"token":    "some-token",
		"password": "abc", // Less than 8 characters
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/auth/activate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}
