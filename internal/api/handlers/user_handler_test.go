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

func TestUserHandler(t *testing.T) {
	// Create mock services
	mockUserService := new(testutil.MockUserService)
	mockNotificationService := NewMockUserNotificationService()

	// Create test app
	app := testutil.SetupTestApp()

	// Create a fixed user ID for testing
	userID := uuid.New()

	// Create auth middleware with fixed user ID
	authMiddleware := testutil.CreateAuthMiddlewareWithUserID(userID)

	// Create user handler
	userHandler := handlers.NewUserHandler(nil, mockNotificationService.NotificationService)

	// Register routes
	api := app.Group("/api")
	authenticated := api.Group("/")
	authenticated.Use(authMiddleware)
	userHandler.RegisterRoutes(authenticated, authMiddleware)

	t.Run("GetUsers", func(t *testing.T) {
		// Create test users
		users := []account.User{
			{
				Username: "testuser1",
				Email:    "test1@example.com",
				IsActive: true,
			},
			{
				Username: "testuser2",
				Email:    "test2@example.com",
				IsActive: true,
			},
		}

		// Set IDs and timestamps manually
		now := time.Now()
		users[0].ID = uuid.New()
		users[0].CreatedAt = now.Add(-24 * time.Hour)
		users[0].UpdatedAt = now.Add(-24 * time.Hour)

		users[1].ID = uuid.New()
		users[1].CreatedAt = now.Add(-48 * time.Hour)
		users[1].UpdatedAt = now.Add(-24 * time.Hour)

		// Setup mock expectations
		mockUserService.On("GetAllUsers", 1, 10).
			Return(users, int64(2), nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/users",
			QueryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
			},
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	t.Run("GetUsers_Error", func(t *testing.T) {
		// Setup mock expectations
		mockUserService.On("GetAllUsers", 1, 10).
			Return([]account.User{}, int64(0), errors.New("database error"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/users",
			QueryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
			},
		})

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	t.Run("GetUser", func(t *testing.T) {
		// Create test user
		testUserID := uuid.New()
		testUser := &account.User{
			Username: "testuser",
			Email:    "test@example.com",
			IsActive: true,
		}
		testUser.ID = testUserID
		testUser.CreatedAt = time.Now().Add(-24 * time.Hour)
		testUser.UpdatedAt = time.Now().Add(-24 * time.Hour)

		// Setup mock expectations
		mockUserService.On("GetUserByID", testUserID).
			Return(testUser, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/users/" + testUserID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	t.Run("GetUser_NotFound", func(t *testing.T) {
		// Create user ID
		testUserID := uuid.New()

		// Setup mock expectations
		mockUserService.On("GetUserByID", testUserID).
			Return(nil, errors.New("user not found"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/users/" + testUserID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	t.Run("CreateUser", func(t *testing.T) {
		// Create user request
		createReq := requests.CreateUserRequest{
			Username:  "newuser",
			Email:     "newuser@example.com",
			Password:  "Password123!",
			FirstName: "New",
			LastName:  "User",
			Phone:     "1234567890",
			Roles:     []string{"user"},
		}

		// Create user result
		userResult := &testutil.UserResult{
			Success:  true,
			Message:  "User created successfully",
			UserID:   uuid.New(),
			Username: createReq.Username,
			Email:    createReq.Email,
			Roles:    createReq.Roles,
		}

		// Setup mock expectations
		mockUserService.On("CreateUser",
			createReq.Username,
			createReq.Email,
			createReq.Password,
			createReq.Phone,
			createReq.FirstName,
			createReq.LastName,
			createReq.Roles).
			Return(userResult, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPost,
			URL:    "/api/users",
			Body:   createReq,
		})

		// Assert response
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	t.Run("CreateUser_ValidationError", func(t *testing.T) {
		// Create invalid user request (missing required fields)
		createReq := requests.CreateUserRequest{
			// Username is required but missing
			Email: "newuser@example.com",
			// Password is required but missing
			FirstName: "New",
			LastName:  "User",
		}

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPost,
			URL:    "/api/users",
			Body:   createReq,
		})

		// Assert response
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UpdateUser", func(t *testing.T) {
		// Create user ID
		testUserID := uuid.New()

		// Create update request
		email := "updated@example.com"
		phone := "9876543210"
		firstName := "Updated"
		lastName := "User"
		isActive := true

		updateReq := requests.UpdateUserRequest{
			Email:     &email,
			Phone:     &phone,
			FirstName: &firstName,
			LastName:  &lastName,
			IsActive:  &isActive,
			Roles:     []string{"user", "admin"},
		}

		// Create user result
		userResult := &testutil.UserResult{
			Success: true,
			Message: "User updated successfully",
			UserID:  testUserID,
			Email:   email,
			Roles:   updateReq.Roles,
		}

		// Setup mock expectations
		mockUserService.On("UpdateUser",
			testUserID,
			updateReq.Email,
			updateReq.Phone,
			updateReq.FirstName,
			updateReq.LastName,
			updateReq.IsActive,
			updateReq.Roles).
			Return(userResult, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/users/" + testUserID.String(),
			Body:   updateReq,
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	t.Run("UpdateUser_NotFound", func(t *testing.T) {
		// Create user ID
		testUserID := uuid.New()

		// Create update request
		email := "updated@example.com"
		updateReq := requests.UpdateUserRequest{
			Email: &email,
		}

		// Setup mock expectations
		mockUserService.On("UpdateUser",
			testUserID,
			updateReq.Email,
			updateReq.Phone,
			updateReq.FirstName,
			updateReq.LastName,
			updateReq.IsActive,
			updateReq.Roles).
			Return(nil, errors.New("user not found"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/users/" + testUserID.String(),
			Body:   updateReq,
		})

		// Assert response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		// Create user ID
		testUserID := uuid.New()

		// Create user result
		userResult := &testutil.UserResult{
			Success: true,
			Message: "User deleted successfully",
			UserID:  testUserID,
		}

		// Setup mock expectations
		mockUserService.On("DeleteUser", testUserID).
			Return(userResult, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodDelete,
			URL:    "/api/users/" + testUserID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify expected response structure
		testutil.AssertJSONContains(t, resp, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "User deleted successfully",
		})

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})

	t.Run("DeleteUser_NotFound", func(t *testing.T) {
		// Create user ID
		testUserID := uuid.New()

		// Setup mock expectations
		mockUserService.On("DeleteUser", testUserID).
			Return(nil, errors.New("user not found"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodDelete,
			URL:    "/api/users/" + testUserID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify mock expectations
		mockUserService.AssertExpectations(t)
	})
}
