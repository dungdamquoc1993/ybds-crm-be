package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ybds/internal/api/responses"
	"github.com/ybds/internal/services"
	"gorm.io/gorm"
)

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(db *gorm.DB, notificationService *services.NotificationService) *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(db, notificationService),
	}
}

// RegisterRoutes registers all routes related to users
func (h *UserHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	users := router.Group("/users")
	users.Use(authMiddleware)

	users.Get("/", h.GetUsers)
	users.Get("/:id", h.GetUserByID)
}

// GetUsers godoc
// @Summary Get all users with their relationships
// @Description Get a list of all users with their addresses and roles
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search term"
// @Param role query string false "Filter by role"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} responses.UsersResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/users [get]
// @Security ApiKeyAuth
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	// Get users from service
	users, total, err := h.userService.GetAllUsers(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve users",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Users retrieved successfully",
		"data": fiber.Map{
			"users":       users,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetUserByID godoc
// @Summary Get a user by ID with all relationships
// @Description Get a specific user with their addresses and roles
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} responses.UserDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/users/{id} [get]
// @Security ApiKeyAuth
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	// Parse user ID from path
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid user ID format",
			Error:   err.Error(),
		})
	}

	// Get user from service
	user, err := h.userService.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "User not found",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User retrieved successfully",
		"data":    user,
	})
}
