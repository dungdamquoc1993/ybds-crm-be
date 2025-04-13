package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/api/responses"
	"github.com/ybds/internal/models/account"
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
	users.Patch("/:id/telegram", h.UpdateTelegramID)
}

// convertUserToResponse converts a user model to a user response
func convertUserToResponse(user *account.User) responses.UserDetailResponse {
	// Extract role names
	roles := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, string(role.Name))
	}

	return responses.UserDetailResponse{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		Phone:      user.Phone,
		IsActive:   user.IsActive,
		TelegramID: user.TelegramID,
		Roles:      roles,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

// GetUsers godoc
// @Summary Get all users with their relationships
// @Description Get a list of all users with their roles
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
// @Router /api/admin/users [get]
// @Security ApiKeyAuth
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	// Ensure page and pageSize are valid
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// First, get the total count to calculate total pages
	_, total, err := h.userService.GetAllUsers(1, 1)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve users count",
			Error:   err.Error(),
		})
	}

	// Calculate total pages
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	// Adjust page if it exceeds total pages
	if totalPages > 0 && int64(page) > totalPages {
		page = int(totalPages)
	}

	// Get users from service
	users, _, err := h.userService.GetAllUsers(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve users",
			Error:   err.Error(),
		})
	}

	// Convert users to response format
	userResponses := make([]responses.UserDetailResponse, len(users))
	for i, user := range users {
		userResponses[i] = convertUserToResponse(&user)
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.UsersResponse{
		Success:    true,
		Message:    "Users retrieved successfully",
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(totalPages),
	})
}

// GetUserByID godoc
// @Summary Get a user by ID with all relationships
// @Description Get a specific user with their roles
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} responses.SingleUserResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/admin/users/{id} [get]
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

	// Convert to response format
	userResponse := convertUserToResponse(user)

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SingleUserResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    userResponse,
	})
}

// UpdateTelegramID godoc
// @Summary Update a user's Telegram ID
// @Description Update the Telegram ID for a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param telegramRequest body requests.UpdateTelegramIDRequest true "Telegram ID info"
// @Success 200 {object} responses.SingleUserResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/admin/users/{id}/telegram [patch]
// @Security ApiKeyAuth
func (h *UserHandler) UpdateTelegramID(c *fiber.Ctx) error {
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

	// Parse request body
	var request requests.UpdateTelegramIDRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Update user's Telegram ID
	result, err := h.userService.UpdateTelegramID(id, request.TelegramID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if result != nil && result.Error == "User not found" {
			statusCode = fiber.StatusNotFound
		}
		return c.Status(statusCode).JSON(responses.ErrorResponse{
			Success: false,
			Message: result.Message,
			Error:   result.Error,
		})
	}

	// Get updated user
	user, err := h.userService.GetUserByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve updated user",
			Error:   err.Error(),
		})
	}

	// Convert to response format
	userResponse := convertUserToResponse(user)

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SingleUserResponse{
		Success: true,
		Message: "Telegram ID updated successfully",
		Data:    userResponse,
	})
}
