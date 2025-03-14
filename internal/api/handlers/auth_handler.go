package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/api/responses"
	"github.com/ybds/internal/services"
	"github.com/ybds/pkg/jwt"
	"gorm.io/gorm"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(db *gorm.DB, jwtService *jwt.JWTService, userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(db, jwtService, userService),
	}
}

// RegisterRoutes registers all routes related to authentication
func (h *AuthHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/login", h.Login)
	router.Post("/register", h.Register)
}

// Login godoc
// @Summary Login to the application
// @Description Login for admin and AI agent users
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body requests.LoginRequest true "Login credentials"
// @Success 200 {object} responses.LoginResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Parse request
	var loginRequest requests.LoginRequest
	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Validate request
	if err := loginRequest.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Call service to handle login
	result, err := h.authService.Login(loginRequest.Username, loginRequest.Password)
	if err != nil {
		// Check the result for specific error messages
		if result != nil {
			if !result.Success {
				if result.Error == "Invalid credentials" || result.Error == "Account is inactive" {
					return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
						Success: false,
						Message: result.Message,
						Error:   result.Error,
					})
				}
			}
		}

		// Default to internal server error
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Authentication failed",
			Error:   "Internal server error",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.LoginResponse{
		Success: true,
		Message: result.Message,
		Token:   result.Token,
		User: responses.UserResponse{
			ID:       result.UserID,
			Username: result.Username,
			Email:    result.Email,
			Roles:    result.Roles,
		},
	})
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email or phone and password
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body requests.RegisterRequest true "Registration details"
// @Success 200 {object} responses.RegisterResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// Parse request
	var registerRequest requests.RegisterRequest
	if err := c.BodyParser(&registerRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Validate request
	if err := registerRequest.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Call service to handle registration
	result, err := h.authService.Register(registerRequest.Email, registerRequest.Phone, registerRequest.Password)
	if err != nil {
		// Check the result for specific error messages
		if result != nil && !result.Success {
			if result.Error == "Email or phone number already registered" {
				return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
					Success: false,
					Message: result.Message,
					Error:   result.Error,
				})
			}
		}

		// Default to internal server error
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Registration failed",
			Error:   "Internal server error",
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.RegisterResponse{
		Success:  true,
		Message:  result.Message,
		UserID:   result.UserID,
		Username: result.Username,
		Email:    result.Email,
	})
}
