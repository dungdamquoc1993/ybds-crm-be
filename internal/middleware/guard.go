package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ybds/internal/utils"
)

// RoleGuard creates a middleware that checks if the user has any of the required roles
func RoleGuard(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user roles from context
		userRoles, ok := c.Locals("roles").([]string)
		if !ok {
			return utils.UnauthorizedResponse(c)
		}

		// Check if the user has any of the required roles
		hasRequiredRole := false
		for _, required := range requiredRoles {
			for _, role := range userRoles {
				if role == required {
					hasRequiredRole = true
					// Once we find a matching role, we can skip checking other roles
					break
				}
			}
			// If we found a matching role, no need to check other required roles
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			return utils.ForbiddenResponse(c)
		}

		return c.Next()
	}
}

// AdminGuard creates a middleware that checks if the user is an admin
func AdminGuard() fiber.Handler {
	return RoleGuard("admin")
}

// AgentGuard creates a middleware that checks if the user is an AI agent
func AgentGuard() fiber.Handler {
	return RoleGuard("agent")
}

// AdminOrAgentGuard creates a middleware that checks if the user is an admin or an AI agent
func AdminOrAgentGuard() fiber.Handler {
	return RoleGuard("admin", "agent")
}
