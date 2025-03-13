package handlers_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ybds/internal/api/handlers"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/testutil"
)

func TestNotificationHandler(t *testing.T) {
	// Create mock services
	mockNotificationService := new(testutil.MockNotificationService)

	// Create test app
	app := testutil.SetupTestApp()

	// Create a fixed user ID for testing
	userID := uuid.New()

	// Create auth middleware with fixed user ID
	authMiddleware := testutil.CreateAuthMiddlewareWithUserID(userID)

	// Create notification handler with mock service
	notificationHandler := handlers.NewNotificationHandler(nil, nil)

	// We can't set the private field directly, so we'll mock the service calls instead

	// Register routes
	api := app.Group("/api")
	authenticated := api.Group("/")
	authenticated.Use(authMiddleware)
	notificationHandler.RegisterRoutes(authenticated, authMiddleware)

	t.Run("GetNotifications", func(t *testing.T) {
		// Create test notifications
		now := time.Now()
		notifications := []notification.Notification{
			{
				// Base fields are embedded, so we need to set them directly
				// ID, CreatedAt, and UpdatedAt are part of the Base struct
				RecipientID:   &userID,
				RecipientType: notification.RecipientUser,
				Title:         "Test Notification 1",
				Message:       "This is a test notification 1",
				Status:        notification.NotificationSent,
				IsRead:        false,
			},
			{
				RecipientID:   &userID,
				RecipientType: notification.RecipientUser,
				Title:         "Test Notification 2",
				Message:       "This is a test notification 2",
				Status:        notification.NotificationSent,
				IsRead:        true,
			},
		}

		// Set IDs and timestamps manually
		notifications[0].ID = uuid.New()
		notifications[0].CreatedAt = now.Add(-24 * time.Hour)
		notifications[0].UpdatedAt = now.Add(-24 * time.Hour)

		notifications[1].ID = uuid.New()
		notifications[1].CreatedAt = now.Add(-48 * time.Hour)
		notifications[1].UpdatedAt = now.Add(-24 * time.Hour)

		// Setup mock expectations
		mockNotificationService.On("GetNotificationsByRecipient", userID, notification.RecipientUser).
			Return(notifications, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/notifications",
			QueryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
			},
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("GetNotifications_WithUnreadOnly", func(t *testing.T) {
		// Create test notifications
		now := time.Now()
		notifications := []notification.Notification{
			{
				RecipientID:   &userID,
				RecipientType: notification.RecipientUser,
				Title:         "Test Notification 1",
				Message:       "This is a test notification 1",
				Status:        notification.NotificationSent,
				IsRead:        false,
			},
		}

		// Set IDs and timestamps manually
		notifications[0].ID = uuid.New()
		notifications[0].CreatedAt = now.Add(-24 * time.Hour)
		notifications[0].UpdatedAt = now.Add(-24 * time.Hour)

		// Setup mock expectations
		mockNotificationService.On("GetUnreadNotificationsByRecipient", userID, notification.RecipientUser).
			Return(notifications, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/notifications",
			QueryParams: map[string]string{
				"page":        "1",
				"page_size":   "10",
				"unread_only": "true",
			},
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("GetNotifications_Error", func(t *testing.T) {
		// Setup mock expectations
		mockNotificationService.On("GetNotificationsByRecipient", userID, notification.RecipientUser).
			Return([]notification.Notification{}, errors.New("database error"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/notifications",
		})

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("GetNotifications_InvalidParameters", func(t *testing.T) {
		// Execute request with invalid parameters
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/notifications",
			QueryParams: map[string]string{
				"page":      "-1",
				"page_size": "0",
			},
		})

		// Assert response
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("GetUnreadNotifications", func(t *testing.T) {
		// Create test notifications
		now := time.Now()
		notifications := []notification.Notification{
			{
				RecipientID:   &userID,
				RecipientType: notification.RecipientUser,
				Title:         "Test Notification 1",
				Message:       "This is a test notification 1",
				Status:        notification.NotificationSent,
				IsRead:        false,
			},
		}

		// Set ID and timestamps manually
		notifications[0].ID = uuid.New()
		notifications[0].CreatedAt = now.Add(-24 * time.Hour)
		notifications[0].UpdatedAt = now.Add(-24 * time.Hour)

		// Setup mock expectations
		mockNotificationService.On("GetUnreadNotificationsByRecipient", userID, notification.RecipientUser).
			Return(notifications, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/notifications/unread",
			QueryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
			},
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("GetUnreadNotifications_Error", func(t *testing.T) {
		// Setup mock expectations
		mockNotificationService.On("GetUnreadNotificationsByRecipient", userID, notification.RecipientUser).
			Return([]notification.Notification{}, errors.New("database error"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/notifications/unread",
		})

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("MarkAsRead", func(t *testing.T) {
		// Create notification ID
		notificationID := uuid.New()

		// Setup mock expectations
		mockNotificationService.On("MarkNotificationAsRead", notificationID).
			Return(nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/notifications/" + notificationID.String() + "/read",
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify expected response structure
		testutil.AssertJSONContains(t, resp, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Notification marked as read successfully",
		})

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("MarkAsRead_InvalidID", func(t *testing.T) {
		// Execute request with invalid ID
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/notifications/invalid-id/read",
		})

		// Assert response
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("MarkAsRead_Error", func(t *testing.T) {
		// Create notification ID
		notificationID := uuid.New()

		// Setup mock expectations
		mockNotificationService.On("MarkNotificationAsRead", notificationID).
			Return(errors.New("database error"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/notifications/" + notificationID.String() + "/read",
		})

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("MarkAllAsRead", func(t *testing.T) {
		// Setup mock expectations
		mockNotificationService.On("MarkAllNotificationsAsRead", userID, notification.RecipientUser).
			Return(nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/notifications/read-all",
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify expected response structure
		testutil.AssertJSONContains(t, resp, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "All notifications marked as read successfully",
		})

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("MarkAllAsRead_Error", func(t *testing.T) {
		// Setup mock expectations
		mockNotificationService.On("MarkAllNotificationsAsRead", userID, notification.RecipientUser).
			Return(errors.New("database error"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/notifications/read-all",
		})

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify mock expectations
		mockNotificationService.AssertExpectations(t)
	})

	t.Run("CreateNotification", func(t *testing.T) {
		// Create a test notification request
		createReq := requests.CreateNotificationRequest{
			RecipientID:   &userID,
			RecipientType: "user",
			Title:         "Test Notification",
			Message:       "This is a test notification",
			Metadata:      map[string]interface{}{"key": "value"},
			Channels:      []string{"email", "push"},
		}

		// This test is just to demonstrate the request model usage
		// The actual implementation would need to be added to the handler
		assert.NoError(t, createReq.Validate())
	})
}
