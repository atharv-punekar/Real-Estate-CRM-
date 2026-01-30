package handlers

import (
	"os"

	"github.com/atharvpunekar/real_estate_crm_backend/internal/database"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/models"
	"github.com/atharvpunekar/real_estate_crm_backend/internal/utils"
	"github.com/gofiber/fiber/v2"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ActivatePasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// Variables are shared across handlers package

func SuperAdminLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var admin models.SuperAdmin
	if err := database.DB.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	if !utils.CheckPassword(admin.PasswordHash, req.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	token, err := utils.GenerateJWT(jwtSecret, admin.ID, "super_admin", "")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    admin.ID,
			"name":  admin.Name,
			"email": admin.Email,
			"role":  "super_admin",
		},
	})
}

// OrgAdminLogin handles login for org_admin and org_user roles
func OrgAdminLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	user, err := userRepo.FindByEmail(req.Email)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Check if user is active
	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Account is deactivated"})
	}

	// Check if password is set
	if !user.IsPasswordSet {
		return c.Status(403).JSON(fiber.Map{"error": "Password not set. Please activate your account using the invite link."})
	}

	// Validate role
	if user.Role != "org_admin" && user.Role != "org_user" {
		return c.Status(403).JSON(fiber.Map{"error": "Invalid user role"})
	}

	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	token, err := utils.GenerateJWT(jwtSecret, user.ID.String(), user.Role, user.OrganizationID.String())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":              user.ID,
			"name":            user.Name,
			"email":           user.Email,
			"role":            user.Role,
			"organization_id": user.OrganizationID,
		},
	})
}

// ActivatePassword handles password activation via invite token
func ActivatePassword(c *fiber.Ctx) error {
	var req ActivatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Token == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Token and password are required"})
	}

	if len(req.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}

	// Find user by invite token
	user, err := userRepo.FindByInviteToken(req.Token)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invalid or expired invite token"})
	}

	// Check if token is expired
	if utils.IsTokenExpired(user.TokenExpiresAt) {
		return c.Status(400).JSON(fiber.Map{"error": "Invite token has expired"})
	}

	// Check if password already set
	if user.IsPasswordSet {
		return c.Status(400).JSON(fiber.Map{"error": "Password already set. Please login instead."})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Update user
	user.PasswordHash = hashedPassword
	user.IsPasswordSet = true
	user.InviteToken = nil
	user.TokenExpiresAt = nil

	if err := userRepo.Update(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to activate password"})
	}

	return c.JSON(fiber.Map{
		"message": "Password activated successfully. You can now login.",
	})
}
