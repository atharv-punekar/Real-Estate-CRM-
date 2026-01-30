package routes

import (
	"github.com/atharvpunekar/real_estate_crm_backend/internal/handlers"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// RegisterAgentRoutes registers routes for both org_admin and org_user
func RegisterAgentRoutes(app *fiber.App) {
	// Agent routes - accessible by both org_admin and org_user
	agent := app.Group("/agent", middleware.JWTProtected, middleware.OrgUserOrAdmin)

	// Contact routes
	contacts := agent.Group("/contacts")
	contacts.Post("/", handlers.CreateContact)
	contacts.Get("/", handlers.GetContacts)
	contacts.Get("/:id", handlers.GetContactByID)
	contacts.Put("/:id", handlers.UpdateContact)
	contacts.Delete("/:id", handlers.DeleteContact)
	contacts.Post("/import", handlers.ImportContactsCSV)

	// Audience routes
	audiences := agent.Group("/audiences")
	audiences.Post("/", handlers.CreateAudience)
	audiences.Get("/", handlers.GetAudiences)
	audiences.Get("/:id", handlers.GetAudienceByID)
	audiences.Put("/:id", handlers.UpdateAudience)
	audiences.Delete("/:id", handlers.DeleteAudience)
	audiences.Post("/:id/contacts", handlers.AddContactsToAudience)
	audiences.Delete("/:id/contacts", handlers.RemoveContactsFromAudience)
	audiences.Get("/:id/contacts", handlers.GetAudienceContacts)

	// Email template routes
	templates := agent.Group("/email-templates")
	templates.Post("/", handlers.CreateEmailTemplate)
	templates.Get("/", handlers.GetEmailTemplates)
	templates.Get("/:id", handlers.GetEmailTemplateByID)
	templates.Put("/:id", handlers.UpdateEmailTemplate)
	templates.Delete("/:id", handlers.DeleteEmailTemplate)
	templates.Post("/:id/test-send", handlers.TestSendEmail)

	// Campaign routes
	campaigns := agent.Group("/campaigns")
	campaigns.Post("/", handlers.CreateCampaign)
	campaigns.Get("/", handlers.GetCampaigns)
	campaigns.Get("/:id", handlers.GetCampaignByID)
	campaigns.Put("/:id", handlers.UpdateCampaign)
	campaigns.Delete("/:id", handlers.DeleteCampaign)
	campaigns.Post("/:id/pause", handlers.PauseCampaign)
	campaigns.Post("/:id/resume", handlers.ResumeCampaign)
	campaigns.Get("/:id/logs", handlers.GetCampaignLogs)

	// Notification routes
	notifications := agent.Group("/notifications")
	notifications.Get("/", handlers.GetNotifications)
	notifications.Put("/:id/read", handlers.MarkNotificationRead)
	notifications.Get("/unread-count", handlers.GetUnreadCount)
}
