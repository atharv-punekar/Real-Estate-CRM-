package main

import (
	"log"
	"time"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/config"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/routes"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config load failed:", err)
	}

	database.Connect(cfg.DSN())
	database.EnsureSchema()

	if err := database.DB.AutoMigrate(
		// &models.SuperAdmin{},
		&models.Organization{},
		&models.User{},
		&models.Contact{},
		&models.Audience{},
		&models.AudienceContact{},
		&models.EmailTemplate{},
		&models.Campaign{},
		&models.CampaignLog{},
		&models.Notification{},
		&models.BackgroundJobLog{},
	); err != nil {
		log.Fatal("Database migration failed:", err)
	}

	app := fiber.New()

	// Configure CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, PATCH, OPTIONS",
		AllowCredentials: false,
	}))

	routes.AuthRoutes(app)
	routes.RegisterSuperAdminRoutes(app)
	routes.RegisterOrgAdminRoutes(app)
	routes.RegisterAgentRoutes(app)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Start background job scheduler in a goroutine
	go startBackgroundScheduler()

	log.Println("🚀 Server running on port:", cfg.Server.Port)
	app.Listen(":" + cfg.Server.Port)
}

// startBackgroundScheduler runs the campaign scheduler periodically
func startBackgroundScheduler() {
	bgJobService := services.NewBackgroundJobService()
	ticker := time.NewTicker(1 * time.Minute) // Run every  minutes
	defer ticker.Stop()

	log.Println("📅 Background job scheduler started")

	// Run immediately on startup
	bgJobService.ProcessCampaignScheduler()

	for range ticker.C {
		bgJobService.ProcessCampaignScheduler()
	}
}
