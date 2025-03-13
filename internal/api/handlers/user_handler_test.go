package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ybds/internal/models/account"
	"github.com/ybds/internal/services"
	"github.com/ybds/internal/testutil"
)

// MockUserNotificationService is a mock implementation of services.NotificationService for user tests
type MockUserNotificationService struct {
	*services.NotificationService
}

// NewMockUserNotificationService creates a new mock notification service for user tests
func NewMockUserNotificationService() *MockUserNotificationService {
	return &MockUserNotificationService{
		NotificationService: &services.NotificationService{},
	}
}

// mockJWTMiddleware creates a simple JWT middleware for testing
func mockJWTMiddleware(c *fiber.Ctx) error {
	// Set user ID in locals for testing
	c.Locals("user_id", "test-user-id")
	c.Locals("roles", []string{"admin"})
	return c.Next()
}

func TestUserHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a mock user service
	mockUserService := new(testutil.MockUserService)

	// Register routes with mock middleware
	api := app.Group("/api")
	users := api.Group("/users")
	users.Use(mockJWTMiddleware)

	// Register the GetUsers endpoint
	users.Get("/", func(c *fiber.Ctx) error {
		// Parse pagination parameters
		page := 1
		pageSize := 10

		// Get users from service
		users, total, err := mockUserService.GetAllUsers(page, pageSize)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to retrieve users",
				"error":   err.Error(),
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
	})

	// Register the GetUserByID endpoint
	users.Get("/:id", func(c *fiber.Ctx) error {
		// Parse user ID from path
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID format",
				"error":   err.Error(),
			})
		}

		// Get user from service
		user, err := mockUserService.GetUserByID(id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "User retrieved successfully",
			"data":    user,
		})
	})

	// Test case 1: Get all users
	t.Run("GetUsers", func(t *testing.T) {
		// Setup mock expectations
		users := []account.User{
			{
				Username: "user1",
				Email:    "user1@example.com",
			},
			{
				Username: "user2",
				Email:    "user2@example.com",
			},
		}
		users[0].ID = uuid.New()
		users[1].ID = uuid.New()

		mockUserService.On("GetAllUsers", 1, 10).Return(users, int64(2), nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
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
		assert.Equal(t, "Users retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["users"])
		assert.Equal(t, float64(2), data["total"])
		assert.Equal(t, float64(1), data["page"])
		assert.Equal(t, float64(10), data["page_size"])

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	// Test case 2: Get user by ID
	t.Run("GetUserByID", func(t *testing.T) {
		// Setup mock expectations
		userID := uuid.New()
		user := &account.User{
			Username: "testuser",
			Email:    "test@example.com",
		}
		user.ID = userID

		mockUserService.On("GetUserByID", userID).Return(user, nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/users/"+userID.String(), nil)
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
		assert.Equal(t, "User retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "testuser", data["username"])
		assert.Equal(t, "test@example.com", data["email"])

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	// Test case 3: Get user by ID - not found
	t.Run("GetUserByID_NotFound", func(t *testing.T) {
		// Setup mock expectations
		userID := uuid.New()
		mockUserService.On("GetUserByID", userID).Return(nil, errors.New("user not found")).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/users/"+userID.String(), nil)
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify response
		assert.Equal(t, false, response["success"])
		assert.Equal(t, "User not found", response["message"])
		assert.Equal(t, "user not found", response["error"])

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	// Test case 4: Get user by ID - invalid ID
	t.Run("GetUserByID_InvalidID", func(t *testing.T) {
		// Create request with invalid UUID
		req := httptest.NewRequest(http.MethodGet, "/api/users/invalid-uuid", nil)
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
		assert.Equal(t, "Invalid user ID format", response["message"])
	})
}
