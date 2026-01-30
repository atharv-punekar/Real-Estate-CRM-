package routes

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	// Login endpoints
	auth.Post("/superadmin/login", handlers.SuperAdminLogin)
	auth.Post("/orgadmin/login", handlers.OrgAdminLogin)

	// Password activation (public endpoint)
	auth.Post("/activate", handlers.ActivatePassword)
}
