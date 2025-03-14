package services

import (
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

// GetJWTService returns the underlying JWT service
func (s *JWTServiceWrapper) GetJWTService() *jwt.JWTService {
	return s.jwtService
}

// Note: The AuthMiddleware functionality has been moved to the middleware package
// Use middleware.JWTAuth instead
