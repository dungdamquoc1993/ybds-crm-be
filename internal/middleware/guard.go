package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ybds/internal/utils"
)

// RoleGuard creates a middleware that checks if the user has the required role
func RoleGuard(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user role from context
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return utils.UnauthorizedResponse(c)
		}

		// Check if the user has one of the required roles
		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return utils.ForbiddenResponse(c)
	}
}

// AdminGuard creates a middleware that checks if the user is an admin
func AdminGuard() fiber.Handler {
	return RoleGuard("admin")
}

// UserGuard creates a middleware that checks if the user is a regular user
func UserGuard() fiber.Handler {
	return RoleGuard("user", "admin")
}
