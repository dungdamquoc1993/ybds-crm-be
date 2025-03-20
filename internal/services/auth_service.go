package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ybds/pkg/jwt"
	passwordpkg "github.com/ybds/pkg/password"
	"gorm.io/gorm"
)

// AuthService handles authentication-related business logic
type AuthService struct {
	db          *gorm.DB
	jwtService  *jwt.JWTService
	userService *UserService
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(db *gorm.DB, jwtService *jwt.JWTService, userService *UserService) *AuthService {
	return &AuthService{
		db:          db,
		jwtService:  jwtService,
		userService: userService,
	}
}

// LoginResult represents the result of a login attempt
type LoginResult struct {
	Success  bool
	Message  string
	Error    string
	Token    string
	UserID   uuid.UUID
	Username string
	Email    string
	Roles    []string
}

// Login authenticates a user and returns a JWT token if successful
func (s *AuthService) Login(username, plainPassword string) (*LoginResult, error) {
	// Find user by username, email, or phone using UserService
	user, err := s.userService.GetUserByCredentials(username)
	if err != nil {
		return &LoginResult{
			Success: false,
			Message: "Authentication failed",
			Error:   "Invalid credentials",
		}, err
	}

	// Check if user is active
	if !user.IsActive {
		return &LoginResult{
			Success: false,
			Message: "Authentication failed",
			Error:   "Account is inactive",
		}, fmt.Errorf("account is inactive")
	}

	// Verify password
	if !passwordpkg.Verify(plainPassword, user.PasswordHash, user.Salt) {
		return &LoginResult{
			Success: false,
			Message: "Authentication failed",
			Error:   "Invalid credentials",
		}, fmt.Errorf("invalid password")
	}

	// Extract roles
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, string(role.Name))
	}

	// Generate JWT token
	token, err := s.jwtService.GenerateToken(user.ID.String(), roles)
	if err != nil {
		return &LoginResult{
			Success: false,
			Message: "Authentication failed",
			Error:   "Failed to generate token",
		}, err
	}

	// Return successful result
	return &LoginResult{
		Success:  true,
		Message:  "Authentication successful",
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    roles,
	}, nil
}

// RegistrationResult represents the result of a registration attempt
type RegistrationResult struct {
	Success  bool
	Message  string
	Error    string
	UserID   uuid.UUID
	Username string
	Email    string
}

// Register creates a new user account
func (s *AuthService) Register(email, phone, password string) (*RegistrationResult, error) {
	// Use UserService to create the user
	userResult, err := s.userService.CreateUser(email, phone, password)
	if err != nil {
		return &RegistrationResult{
			Success: false,
			Message: "Registration failed",
			Error:   err.Error(),
		}, err
	}

	// Convert UserResult to RegistrationResult
	return &RegistrationResult{
		Success:  userResult.Success,
		Message:  "Registration successful",
		UserID:   userResult.UserID,
		Username: userResult.Username,
		Email:    userResult.Email,
	}, nil
}
