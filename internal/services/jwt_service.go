package services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ybds/pkg/jwt"
)

// JWTServiceWrapper wraps the pkg/jwt.JWTService to provide additional functionality
type JWTServiceWrapper struct {
	jwtService *jwt.JWTService
}

// NewJWTServiceWrapper creates a new JWTServiceWrapper
func NewJWTServiceWrapper(jwtService *jwt.JWTService) *JWTServiceWrapper {
	return &JWTServiceWrapper{
		jwtService: jwtService,
	}
}

// GenerateToken generates a new JWT token
func (s *JWTServiceWrapper) GenerateToken(userID string, roles []string) (string, error) {
	return s.jwtService.GenerateToken(userID, roles)
}

// ValidateToken validates a JWT token
func (s *JWTServiceWrapper) ValidateToken(tokenString string) (*jwt.CustomClaims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// AuthMiddleware is a middleware that validates JWT tokens
func (s *JWTServiceWrapper) AuthMiddleware(c *fiber.Ctx) error {
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
	claims, err := s.ValidateToken(tokenString)
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
