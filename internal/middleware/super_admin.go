package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func SuperAdminOnly(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role != "super_admin" {
		return c.Status(403).JSON(fiber.Map{
			"error": "Access denied (Super Admin only)",
		})
	}
	return c.Next()
}
