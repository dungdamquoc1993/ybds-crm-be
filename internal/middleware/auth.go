package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ybds/pkg/jwt"
)

// Protected creates a middleware that verifies JWT tokens
func Protected(jwtService *jwt.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication required",
				"error":   "Missing authorization header",
			})
		}

		// Check if authorization header has Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication failed",
				"error":   "Invalid authorization format",
			})
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication failed",
				"error":   err.Error(),
			})
		}

		// Store claims in context for later use
		c.Locals("user", claims)
		return c.Next()
	}
}

// JWTAuth creates a middleware that validates JWT tokens and sets userID and roles in context
func JWTAuth(jwtService *jwt.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization header is required")
		}

		// Check if the header starts with "Bearer "
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid authorization header format")
		}

		// Extract the token
		tokenString := authHeader[7:]

		// Validate the token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
		}

		// Convert user ID string to UUID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid user ID format in token")
		}

		// Set the user ID and roles in the context
		c.Locals("userID", userID)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}

// RoleRequired creates a middleware that checks if the user has the required roles
func RoleRequired(jwtService *jwt.JWTService, roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user claims from context
		claims, ok := c.Locals("user").(*jwt.CustomClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication required",
				"error":   "Invalid user context",
			})
		}

		// Check if user has any of the required roles
		if !jwtService.HasRole(claims, roles...) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Access denied",
				"error":   "Insufficient permissions",
			})
		}

		return c.Next()
	}
}
