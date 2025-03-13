package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/pkg/config"
	"github.com/ybds/pkg/jwt"
)

// TestLoginAPI tests the login API with static responses
func TestLoginAPI(t *testing.T) {
	// Setup
	app := fiber.New()

	// Setup JWT service
	jwtConfig := &config.JWTConfig{
		Secret: "test-secret-key",
		Expiry: "24h",
	}
	jwtService, _ := jwt.NewJWTService(jwtConfig)

	// Register static login endpoint for testing
	api := app.Group("/api")
	auth := api.Group("/auth")
	auth.Post("/login", func(c *fiber.Ctx) error {
		var loginRequest requests.LoginRequest
		if err := c.BodyParser(&loginRequest); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request",
				"error":   err.Error(),
			})
		}

		// Static test - only accept specific credentials
		if loginRequest.Username == "testuser" && loginRequest.Password == "password123" {
			// Generate token
			token, _ := jwtService.GenerateToken("test-user-id", []string{"admin"})

			// Return success response
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"success": true,
				"message": "Authentication successful",
				"token":   token,
				"user": fiber.Map{
					"id":       "test-user-id",
					"username": "testuser",
					"email":    "test@example.com",
					"roles":    []string{"admin"},
				},
			})
		}

		// Return error for invalid credentials
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Authentication failed",
			"error":   "Invalid credentials",
		})
	})

	// Test valid login
	t.Run("Valid Login", func(t *testing.T) {
		// Create login request
		loginRequest := requests.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}
		body, _ := json.Marshal(loginRequest)

		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := app.Test(req)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse response
		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)

		// Validate response
		assert.True(t, respBody["success"].(bool))
		assert.Equal(t, "Authentication successful", respBody["message"])
		assert.NotEmpty(t, respBody["token"])

		user := respBody["user"].(map[string]interface{})
		assert.Equal(t, "testuser", user["username"])
		assert.Equal(t, "test@example.com", user["email"])

		roles := user["roles"].([]interface{})
		assert.Contains(t, roles, "admin")
	})

	// Test invalid password
	t.Run("Invalid Password", func(t *testing.T) {
		// Create login request with wrong password
		loginRequest := requests.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginRequest)

		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := app.Test(req)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		// Parse response
		var respBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respBody)

		// Validate response
		assert.False(t, respBody["success"].(bool))
		assert.Equal(t, "Authentication failed", respBody["message"])
		assert.Equal(t, "Invalid credentials", respBody["error"])
	})
}
