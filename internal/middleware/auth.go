package middleware

import (
	"os"
	"strings"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

func JWTProtected(c *fiber.Ctx) error {
	authorization := c.Get("Authorization")
	if authorization == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Missing token"})
	}

	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token format"})
	}

	tokenString := parts[1]

	jwtSecret := os.Getenv("JWT_SECRET")
	claims, err := utils.ParseJWT(tokenString, jwtSecret)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	c.Locals("user_id", claims.UserID)
	c.Locals("role", claims.Role)
	c.Locals("org_id", claims.OrgID)

	return c.Next()
}

func OrgAdminOnly(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role != "org_admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Only Org Admin allowed"})
	}
	return c.Next()
}

// OrgUserOrAdmin allows both org_admin and org_user roles
func OrgUserOrAdmin(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role != "org_admin" && role != "org_user" {
		return c.Status(403).JSON(fiber.Map{"error": "Only Org Admin or Org User allowed"})
	}
	return c.Next()
}
