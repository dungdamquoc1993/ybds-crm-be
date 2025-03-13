package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/ybds/internal/utils"
	"github.com/ybds/pkg/jwt"
)

// JWTMiddleware creates a middleware that validates JWT tokens
func JWTMiddleware(jwtService *jwt.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.UnauthorizedResponse(c)
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.UnauthorizedResponse(c)
		}

		// Validate the token
		tokenString := parts[1]
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			return utils.UnauthorizedResponse(c)
		}

		// Set user information in the context
		c.Locals("user", claims)
		c.Locals("userID", claims.UserID)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}
