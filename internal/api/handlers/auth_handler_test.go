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
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/api/responses"
	"github.com/ybds/internal/testutil"
)

// TestLoginAPI tests the login API with static responses
func TestLoginAPI(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a mock auth service
	mockAuthService := new(testutil.MockAuthService)

	// Register the login endpoint
	app.Post("/api/auth/login", func(c *fiber.Ctx) error {
		// Parse request body
		var request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		}

		// Validate request
		if request.Email == "" || request.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Email and password are required",
			})
		}

		// Attempt login
		result, err := mockAuthService.Login(request.Email, request.Password)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid credentials",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Login successful",
			"data": fiber.Map{
				"token": result.Token,
				"roles": result.Roles,
			},
		})
	})

	t.Run("Successful login", func(t *testing.T) {
		// Setup mock expectations
		result := &testutil.AuthResult{
			Success:  true,
			Message:  "Login successful",
			UserID:   uuid.New(),
			Username: "testuser",
			Email:    "test@example.com",
			Token:    "jwt-token",
			Roles:    []string{"user"},
		}

		mockAuthService.On("Login", "test@example.com", "password123").Return(result, nil).Once()

		// Create request body
		requestBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Perform request
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "Login successful", response["message"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "jwt-token", data["token"])
		assert.Equal(t, []interface{}{"user"}, data["roles"])

		// Verify mock expectations
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Login with invalid data", func(t *testing.T) {
		// Create request body with missing required fields
		requestBody := map[string]interface{}{
			"email": "test@example.com",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "Email and password are required", response["message"])
	})

	t.Run("Login with incorrect credentials", func(t *testing.T) {
		// Setup mock expectations
		mockAuthService.On("Login", "wrong@example.com", "wrongpassword").Return(nil, errors.New("invalid credentials")).Once()

		// Create request body
		requestBody := map[string]interface{}{
			"email":    "wrong@example.com",
			"password": "wrongpassword",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(requestJSON))
		req.Header.Set("Content-Type", "application/json")

		// Perform request
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		// Parse response
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		// Verify response
		assert.Equal(t, false, response["success"])
		assert.Equal(t, "Invalid credentials", response["message"])
		assert.Equal(t, "invalid credentials", response["error"])

		// Verify mock expectations
		mockAuthService.AssertExpectations(t)
	})
}

func TestRegisterAPI(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a mock auth service
	mockAuthService := new(testutil.MockAuthService)

	// Register the handler directly as a route
	app.Post("/auth/register", func(c *fiber.Ctx) error {
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
		result, err := mockAuthService.Register(registerRequest.Email, registerRequest.Phone, registerRequest.Password)
		if err != nil {
			// Default to internal server error
			return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
				Success: false,
				Message: "Registration failed",
				Error:   "Internal server error",
			})
		}

		// Check for specific error messages in the result
		if result != nil && !result.Success {
			if result.Error == "Email or phone number already registered" {
				return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
					Success: false,
					Message: result.Message,
					Error:   result.Error,
				})
			}
		}

		// Return success response
		return c.Status(fiber.StatusOK).JSON(responses.RegisterResponse{
			Success:  true,
			Message:  result.Message,
			UserID:   result.UserID,
			Username: result.Username,
			Email:    result.Email,
		})
	})

	// Test case 1: Successful registration
	t.Run("Successful registration", func(t *testing.T) {
		// Setup mock expectations
		userID := uuid.New()
		mockAuthService.On("Register", "test@example.com", "", "password123").Return(&testutil.AuthResult{
			Success:  true,
			Message:  "User registered successfully",
			UserID:   userID,
			Username: "test",
			Email:    "test@example.com",
		}, nil).Once()

		// Create request body
		registerRequest := requests.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(registerRequest)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
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
		assert.Equal(t, "User registered successfully", response["message"])
		assert.NotNil(t, response["user_id"])
		assert.Equal(t, "test", response["username"])
		assert.Equal(t, "test@example.com", response["email"])

		// Verify mock expectations
		mockAuthService.AssertExpectations(t)
	})

	// Test case 2: Registration with invalid data
	t.Run("Registration with invalid data", func(t *testing.T) {
		// Create request body with missing required fields
		registerRequest := requests.RegisterRequest{
			// Missing both email and phone
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(registerRequest)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
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
		assert.Equal(t, "Validation failed", response["message"])
		assert.Contains(t, response["error"], "email or phone")
	})

	// Test case 3: Registration with existing email
	t.Run("Registration with existing email", func(t *testing.T) {
		// Setup mock expectations
		mockAuthService.On("Register", "existing@example.com", "", "password123").Return(&testutil.AuthResult{
			Success: false,
			Message: "Registration failed",
			Error:   "Email or phone number already registered",
		}, nil).Once()

		// Create request body
		registerRequest := requests.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(registerRequest)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
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
		assert.Equal(t, "Registration failed", response["message"])
		assert.Equal(t, "Email or phone number already registered", response["error"])

		// Verify mock expectations
		mockAuthService.AssertExpectations(t)
	})

	// Test case 4: Registration with server error
	t.Run("Registration with server error", func(t *testing.T) {
		// Setup mock expectations
		mockAuthService.On("Register", "error@example.com", "", "password123").Return(nil,
			errors.New("Database error")).Once()

		// Create request body
		registerRequest := requests.RegisterRequest{
			Email:    "error@example.com",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(registerRequest)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(jsonBody))
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
		assert.Equal(t, "Registration failed", response["message"])
		assert.Equal(t, "Internal server error", response["error"])

		// Verify mock expectations
		mockAuthService.AssertExpectations(t)
	})
}
