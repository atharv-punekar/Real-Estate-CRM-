package routes

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/handlers"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterOrgAdminRoutes(app *fiber.App) {
	orgAdmin := app.Group("/orgadmin", middleware.JWTProtected, middleware.OrgAdminOnly)

	// Agent Management (org_admin can create both org_admin and org_user)
	agentRoutes := orgAdmin.Group("/agents")
	agentRoutes.Post("/", handlers.CreateAgent)                           // Create agent
	agentRoutes.Get("/", handlers.GetAgents)                              // List all agents
	agentRoutes.Put("/:id", handlers.UpdateAgent)                         // Update agent
	agentRoutes.Delete("/:id", handlers.DeactivateAgent)                  // Deactivate agent
	agentRoutes.Post("/:id/regenerate-invite", handlers.RegenerateInvite) // Regenerate invite
}
