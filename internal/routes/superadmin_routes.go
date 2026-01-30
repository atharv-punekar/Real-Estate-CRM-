package routes

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/handlers"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterSuperAdminRoutes(app *fiber.App) {
	superAdmin := app.Group("/superadmin", middleware.JWTProtected, middleware.SuperAdminOnly)

	// Organization CRUD
	org := superAdmin.Group("/orgs")
	org.Post("/", handlers.CreateOrganization)
	org.Get("/", handlers.GetOrganizations)
	org.Put("/:id", handlers.UpdateOrganization)

	// Agent Management (org_admin and org_user)
	agents := superAdmin.Group("/orgs/:org_id/agents")
	agents.Post("/", handlers.SuperAdminCreateAgent)                           // Create agent
	agents.Get("/", handlers.SuperAdminGetAgents)                              // List all agents for org
	agents.Put("/:id", handlers.SuperAdminUpdateAgent)                         // Update agent
	agents.Delete("/:id", handlers.SuperAdminDeactivateAgent)                  // Deactivate agent
	agents.Post("/:id/regenerate-invite", handlers.SuperAdminRegenerateInvite) // Regenerate invite
}
