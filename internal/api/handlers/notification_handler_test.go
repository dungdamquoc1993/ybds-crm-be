package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/testutil"
)

// notificationMockJWTMiddleware creates a simple JWT middleware for testing
func notificationMockJWTMiddleware(c *fiber.Ctx) error {
	// Set user ID in locals for testing
	userID := uuid.New().String()
	c.Locals("user_id", userID)
	c.Locals("roles", []string{"admin"})
	return c.Next()
}

func TestNotificationHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a mock notification service
	mockNotificationService := new(testutil.MockNotificationService)

	// Register routes with mock middleware
	api := app.Group("/api")
	notifications := api.Group("/notifications")
	notifications.Use(notificationMockJWTMiddleware)

	// Register the GetNotifications endpoint
	notifications.Get("/", func(c *fiber.Ctx) error {
		// Parse pagination parameters
		page := 1
		pageSize := 10
		unreadOnly := false

		// Get user ID from context
		userID := c.Locals("user_id").(string)
		recipientID, err := uuid.Parse(userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID",
				"error":   err.Error(),
			})
		}

		// Get notifications from service
		var notifs []notification.Notification
		if unreadOnly {
			notifs, err = mockNotificationService.GetUnreadNotificationsByRecipient(recipientID, notification.RecipientUser)
		} else {
			notifs, err = mockNotificationService.GetNotificationsByRecipient(recipientID, notification.RecipientUser)
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to retrieve notifications",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Notifications retrieved successfully",
			"data": fiber.Map{
				"notifications": notifs,
				"page":          page,
				"page_size":     pageSize,
				"total":         len(notifs),
			},
		})
	})

	// Register the MarkAsRead endpoint
	notifications.Patch("/:id/read", func(c *fiber.Ctx) error {
		// Parse notification ID from path
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid notification ID format",
				"error":   err.Error(),
			})
		}

		// Mark notification as read
		err = mockNotificationService.MarkNotificationAsRead(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to mark notification as read",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Notification marked as read successfully",
		})
	})

	// Register the MarkAllAsRead endpoint
	notifications.Patch("/read-all", func(c *fiber.Ctx) error {
		// Get user ID from context
		userID := c.Locals("user_id").(string)
		recipientID, err := uuid.Parse(userID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID",
				"error":   err.Error(),
			})
		}

		// Mark all notifications as read
		err = mockNotificationService.MarkAllNotificationsAsRead(recipientID, notification.RecipientUser)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to mark all notifications as read",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "All notifications marked as read successfully",
		})
	})

	// Register the CreateNotification endpoint
	notifications.Post("/", func(c *fiber.Ctx) error {
		// Parse request body
		var request struct {
			RecipientID   *string                    `json:"recipient_id"`
			RecipientType string                     `json:"recipient_type"`
			Title         string                     `json:"title"`
			Message       string                     `json:"message"`
			Metadata      notification.Metadata      `json:"metadata"`
			Channels      []notification.ChannelType `json:"channels"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		}

		// Validate request
		if request.Title == "" || request.Message == "" || request.RecipientType == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Title, message, and recipient type are required",
			})
		}

		// Parse recipient ID if provided
		var recipientID *uuid.UUID
		if request.RecipientID != nil {
			parsed, err := uuid.Parse(*request.RecipientID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "Invalid recipient ID format",
					"error":   err.Error(),
				})
			}
			recipientID = &parsed
		}

		// Parse recipient type
		recipientType := notification.RecipientType(request.RecipientType)

		// Create notification
		result, err := mockNotificationService.CreateNotification(
			recipientID,
			recipientType,
			request.Title,
			request.Message,
			request.Metadata,
			request.Channels,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create notification",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Notification created successfully",
			"data":    result,
		})
	})

	t.Run("MarkAsRead", func(t *testing.T) {
		// Setup mock expectations
		notificationID := uuid.New()
		mockNotificationService.On("MarkNotificationAsRead", notificationID).Return(nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodPatch, "/api/notifications/"+notificationID.String()+"/read", nil)
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify response
		assert.Equal(t, true, response["success"])
		assert.Equal(t, "Notification marked as read successfully", response["message"])

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("MarkAsRead_InvalidID", func(t *testing.T) {
		// Create request with invalid UUID
		req := httptest.NewRequest(http.MethodPatch, "/api/notifications/invalid-uuid/read", nil)
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify response
		assert.Equal(t, false, response["success"])
		assert.Equal(t, "Invalid notification ID format", response["message"])
	})

	t.Run("CreateNotification", func(t *testing.T) {
		// Setup mock expectations
		recipientID := uuid.New()
		recipientIDStr := recipientID.String()

		result := &testutil.NotificationResult{
			Success:        true,
			Message:        "Notification created successfully",
			NotificationID: uuid.New(),
		}

		metadata := notification.Metadata{
			"order_id": "12345",
			"status":   "shipped",
		}

		channels := []notification.ChannelType{notification.ChannelWebsocket, notification.ChannelEmail}

		mockNotificationService.On(
			"CreateNotification",
			&recipientID,
			notification.RecipientUser,
			"New Order Status",
			"Your order has been shipped",
			metadata,
			channels,
		).Return(result, nil).Once()

		// Create request body
		requestBody := map[string]interface{}{
			"recipient_id":   recipientIDStr,
			"recipient_type": "user",
			"title":          "New Order Status",
			"message":        "Your order has been shipped",
			"metadata":       metadata,
			"channels":       []string{"websocket", "email"},
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/notifications", bytes.NewBuffer(requestJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify response
		assert.Equal(t, true, response["success"])
		assert.Equal(t, "Notification created successfully", response["message"])

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("CreateNotification_ValidationError", func(t *testing.T) {
		// Create request body with missing required fields
		requestBody := map[string]interface{}{
			"recipient_type": "user",
			"metadata":       map[string]interface{}{},
			"channels":       []string{"in_app"},
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/notifications", bytes.NewBuffer(requestJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify response
		assert.Equal(t, false, response["success"])
		assert.Equal(t, "Title, message, and recipient type are required", response["message"])
	})

	t.Run("CreateNotification_ServerError", func(t *testing.T) {
		// Setup mock expectations
		recipientID := uuid.New()
		recipientIDStr := recipientID.String()

		metadata := notification.Metadata{
			"order_id": "12345",
			"status":   "shipped",
		}

		channels := []notification.ChannelType{notification.ChannelWebsocket, notification.ChannelEmail}

		mockNotificationService.On(
			"CreateNotification",
			&recipientID,
			notification.RecipientUser,
			"Error Notification",
			"This will cause an error",
			metadata,
			channels,
		).Return(nil, errors.New("database error")).Once()

		// Create request body
		requestBody := map[string]interface{}{
			"recipient_id":   recipientIDStr,
			"recipient_type": "user",
			"title":          "Error Notification",
			"message":        "This will cause an error",
			"metadata":       metadata,
			"channels":       []string{"websocket", "email"},
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/notifications", bytes.NewBuffer(requestJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify response
		assert.Equal(t, false, response["success"])
		assert.Equal(t, "Failed to create notification", response["message"])
		assert.Equal(t, "database error", response["error"])

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})
}
