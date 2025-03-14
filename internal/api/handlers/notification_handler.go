package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/api/responses"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/services"
	"github.com/ybds/pkg/websocket"
	"gorm.io/gorm"
)

// NotificationHandler handles HTTP requests related to notifications
type NotificationHandler struct {
	notificationService *services.NotificationService
	db                  *gorm.DB
}

// NewNotificationHandler creates a new instance of NotificationHandler
func NewNotificationHandler(db *gorm.DB, notificationService *services.NotificationService, hub *websocket.Hub) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		db:                  db,
	}
}

// RegisterRoutes registers all routes related to notifications
func (h *NotificationHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	notifications := router.Group("/notifications")
	notifications.Use(authMiddleware)

	notifications.Get("/", h.GetNotifications)
	notifications.Get("/unread", h.GetUnreadNotifications)
	notifications.Put("/:id/read", h.MarkAsRead)
	notifications.Put("/read-all", h.MarkAllAsRead)
}

// GetNotifications godoc
// @Summary Get all notifications for the current user
// @Description Get a list of all notifications for the current user with pagination
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param unread_only query bool false "Get only unread notifications"
// @Success 200 {object} responses.NotificationsResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/notifications [get]
// @Security ApiKeyAuth
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	// Parse request parameters
	req := requests.GetNotificationsRequest{
		Page:       1,
		PageSize:   10,
		UnreadOnly: false,
	}

	// Parse pagination parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}

	// Ensure page and pageSize are valid
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	if unreadOnlyStr := c.Query("unread_only"); unreadOnlyStr == "true" {
		req.UnreadOnly = true
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
	}

	var notifications []notification.Notification
	var err error

	// Get notifications based on unread_only flag
	if req.UnreadOnly {
		notifications, err = h.notificationService.GetUnreadNotificationsByRecipient(userID, notification.RecipientUser)
	} else {
		notifications, err = h.notificationService.GetNotificationsByRecipient(userID, notification.RecipientUser)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve notifications",
			Error:   err.Error(),
		})
	}

	// Manual pagination since the service doesn't support it
	total := int64(len(notifications))

	// Calculate total pages
	totalPages := (total + int64(req.PageSize) - 1) / int64(req.PageSize)

	// Adjust page if it exceeds total pages
	if totalPages > 0 && int64(req.Page) > totalPages {
		req.Page = int(totalPages)
	}

	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start >= len(notifications) {
		start = 0
		end = 0
		notifications = []notification.Notification{}
	} else if end > len(notifications) {
		end = len(notifications)
	}

	paginatedNotifications := notifications
	if len(notifications) > 0 {
		paginatedNotifications = notifications[start:end]
	}

	// Convert to response format
	notificationResponses := make([]responses.NotificationResponse, len(paginatedNotifications))
	for i, n := range paginatedNotifications {
		var userID uuid.UUID
		if n.RecipientID != nil {
			userID = *n.RecipientID
		}

		notificationResponses[i] = responses.NotificationResponse{
			ID:          n.ID,
			UserID:      userID,
			Type:        string(n.RecipientType),
			Title:       n.Title,
			Message:     n.Message,
			IsRead:      n.IsRead,
			RedirectURL: "", // Add logic to extract from metadata if needed
			CreatedAt:   n.CreatedAt,
			UpdatedAt:   n.UpdatedAt,
		}
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.NotificationsResponse{
		Success:    true,
		Message:    "Notifications retrieved successfully",
		Data:       notificationResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int(totalPages),
	})
}

// GetUnreadNotifications godoc
// @Summary Get unread notifications for the current user
// @Description Get a list of unread notifications for the current user
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} responses.NotificationsResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/notifications/unread [get]
// @Security ApiKeyAuth
func (h *NotificationHandler) GetUnreadNotifications(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	// Parse request parameters
	req := requests.GetNotificationsRequest{
		Page:       1,
		PageSize:   10,
		UnreadOnly: true,
	}

	// Parse pagination parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request parameters",
			Error:   err.Error(),
		})
	}

	// Get unread notifications
	notifications, err := h.notificationService.GetUnreadNotificationsByRecipient(userID, notification.RecipientUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve unread notifications",
			Error:   err.Error(),
		})
	}

	// Manual pagination
	total := int64(len(notifications))
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start >= len(notifications) {
		start = 0
		end = 0
		notifications = []notification.Notification{}
	} else if end > len(notifications) {
		end = len(notifications)
	}

	paginatedNotifications := notifications
	if len(notifications) > 0 {
		paginatedNotifications = notifications[start:end]
	}

	// Convert to response format
	notificationResponses := make([]responses.NotificationResponse, len(paginatedNotifications))
	for i, n := range paginatedNotifications {
		var userID uuid.UUID
		if n.RecipientID != nil {
			userID = *n.RecipientID
		}

		notificationResponses[i] = responses.NotificationResponse{
			ID:          n.ID,
			UserID:      userID,
			Type:        string(n.RecipientType),
			Title:       n.Title,
			Message:     n.Message,
			IsRead:      n.IsRead,
			RedirectURL: "", // Add logic to extract from metadata if needed
			CreatedAt:   n.CreatedAt,
			UpdatedAt:   n.UpdatedAt,
		}
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.NotificationsResponse{
		Success:    true,
		Message:    "Unread notifications retrieved successfully",
		Data:       notificationResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
	})
}

// MarkAsRead godoc
// @Summary Mark a notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/notifications/{id}/read [put]
// @Security ApiKeyAuth
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	_, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	// Parse notification ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid notification ID format",
			Error:   err.Error(),
		})
	}

	// Create and validate request
	req := requests.MarkNotificationAsReadRequest{
		NotificationID: id,
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Mark notification as read
	err = h.notificationService.MarkNotificationAsRead(req.NotificationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to mark notification as read",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: "Notification marked as read successfully",
	})
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications for the current user as read
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} responses.SuccessResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/notifications/read-all [put]
// @Security ApiKeyAuth
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	// Mark all notifications as read
	err := h.notificationService.MarkAllNotificationsAsRead(userID, notification.RecipientUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to mark all notifications as read",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: "All notifications marked as read successfully",
	})
}
