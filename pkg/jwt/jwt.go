package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ybds/pkg/config"
)

// CustomClaims represents the claims in the JWT
type CustomClaims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTService provides methods to generate, parse and validate JWT tokens
type JWTService struct {
	secretKey []byte
	expiry    time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.JWTConfig) (*JWTService, error) {
	if cfg.Secret == "" {
		return nil, errors.New("JWT secret key is required")
	}

	expiry, err := time.ParseDuration(cfg.Expiry)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT expiry duration: %w", err)
	}

	return &JWTService{
		secretKey: []byte(cfg.Secret),
		expiry:    expiry,
	}, nil
}

// GenerateToken generates a new JWT token
func (s *JWTService) GenerateToken(userID string, roles []string) (string, error) {
	claims := &CustomClaims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates the JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// HasRole checks if the user has any of the required roles
func (s *JWTService) HasRole(claims *CustomClaims, requiredRoles ...string) bool {
	if len(requiredRoles) == 0 {
		return true // No roles required
	}

	userRolesMap := make(map[string]bool)
	for _, role := range claims.Roles {
		userRolesMap[role] = true
	}

	for _, requiredRole := range requiredRoles {
		if userRolesMap[requiredRole] {
			return true
		}
	}

	return false
}
