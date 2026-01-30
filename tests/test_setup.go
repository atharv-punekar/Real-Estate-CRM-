package tests

import (
	"log"
	"os"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/handlers"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/middleware"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// init sets up environment variables before tests run
func init() {
	os.Setenv("JWT_SECRET", "this_is_a_test_secret_key_which_is_very_long_123")
}

var TestDB *gorm.DB

// SetupTestDB initializes an in-memory SQLite database for testing
func SetupTestDB() *gorm.DB {
	// Set test env variables
	os.Setenv("JWT_SECRET", "this_is_a_test_secret_key_which_is_very_long_123")

	dsn := "host=localhost user=postgres password=password dbname=crm_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to test database:", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.SuperAdmin{},
		&models.Contact{},
		&models.Audience{},
		&models.AudienceContact{},
		&models.EmailTemplate{},
		&models.Campaign{},
		&models.CampaignLog{},
		&models.Notification{},
		&models.BackgroundJobLog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// --- Seed Test Data ---

	org := models.Organization{
		ID:        uuid.New(),
		Name:      "Test Org",
		IsActive:  true,
		CreatedBy: uuid.UUID{},
	}
	db.Create(&org)

	hashed, _ := utils.HashPassword("test1234")

	user := models.User{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		Name:           "Test User",
		Email:          "test@example.com",
		PasswordHash:   hashed,
		Role:           "org_admin",
		IsActive:       true,
	}
	db.Create(&user)

	TestDB = db
	return db
}

// SetupTestApp creates a test Fiber app instance with routes
func SetupTestApp() *fiber.App {
	app := fiber.New()

	// Auth routes (no middleware)
	app.Post("/auth/superadmin/login", handlers.SuperAdminLogin)
	app.Post("/auth/login", handlers.OrgAdminLogin)
	app.Post("/auth/activate", handlers.ActivatePassword)

	// Protected routes with middleware
	protected := app.Group("/api", middleware.JWTProtected)

	// Contact routes
	protected.Post("/contacts", handlers.CreateContact)
	protected.Get("/contacts", handlers.GetContacts)
	protected.Get("/contacts/:id", handlers.GetContactByID)
	protected.Put("/contacts/:id", handlers.UpdateContact)
	protected.Delete("/contacts/:id", handlers.DeleteContact)
	protected.Post("/contacts/import", handlers.ImportContactsCSV)

	// Audience routes
	protected.Post("/audiences", handlers.CreateAudience)
	protected.Get("/audiences", handlers.GetAudiences)
	protected.Get("/audiences/:id", handlers.GetAudienceByID)
	protected.Put("/audiences/:id", handlers.UpdateAudience)
	protected.Delete("/audiences/:id", handlers.DeleteAudience)
	protected.Post("/audiences/:id/contacts", handlers.AddContactsToAudience)
	protected.Delete("/audiences/:id/contacts", handlers.RemoveContactsFromAudience)
	protected.Get("/audiences/:id/contacts", handlers.GetAudienceContacts)

	// Email Template routes
	protected.Post("/templates", handlers.CreateEmailTemplate)
	protected.Get("/templates", handlers.GetEmailTemplates)
	protected.Get("/templates/:id", handlers.GetEmailTemplateByID)
	protected.Put("/templates/:id", handlers.UpdateEmailTemplate)
	protected.Delete("/templates/:id", handlers.DeleteEmailTemplate)
	protected.Post("/templates/:id/test", handlers.TestSendEmail)

	// Campaign routes
	protected.Post("/campaigns", handlers.CreateCampaign)
	protected.Get("/campaigns", handlers.GetCampaigns)
	protected.Get("/campaigns/:id", handlers.GetCampaignByID)
	protected.Put("/campaigns/:id", handlers.UpdateCampaign)
	protected.Delete("/campaigns/:id", handlers.DeleteCampaign)
	protected.Post("/campaigns/:id/pause", handlers.PauseCampaign)
	protected.Post("/campaigns/:id/resume", handlers.ResumeCampaign)
	protected.Get("/campaigns/:id/logs", handlers.GetCampaignLogs)

	// Notification routes
	protected.Get("/notifications", handlers.GetNotifications)
	protected.Put("/notifications/:id/read", handlers.MarkNotificationRead)
	protected.Get("/notifications/unread", handlers.GetUnreadCount)

	return app
}

// CleanupTestDB cleans up all tables
func CleanupTestDB(db *gorm.DB) {
	db.Exec("DELETE FROM notification")
	db.Exec("DELETE FROM campaign_log")
	db.Exec("DELETE FROM campaign")
	db.Exec("DELETE FROM email_template")
	db.Exec("DELETE FROM audience_contact")
	db.Exec("DELETE FROM audience")
	db.Exec("DELETE FROM contact")
	db.Exec("DELETE FROM background_job_log")
	db.Exec("DELETE FROM \"user\"")
	db.Exec("DELETE FROM organization")
	db.Exec("DELETE FROM super_admin")
}
