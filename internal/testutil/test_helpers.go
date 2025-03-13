package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB is a mock implementation of *gorm.DB
type MockDB struct {
	mock.Mock
}

// MockWebsocketHub is a mock implementation of websocket hub
type MockWebsocketHub struct {
	mock.Mock
}

// BroadcastToUser mocks the BroadcastToUser method
func (m *MockWebsocketHub) BroadcastToUser(userID string, message []byte) {
	m.Called(userID, message)
}

// BroadcastToAll mocks the BroadcastToAll method
func (m *MockWebsocketHub) BroadcastToAll(message []byte) {
	m.Called(message)
}

// TestRequest represents a test request configuration
type TestRequest struct {
	Method      string
	URL         string
	Body        interface{}
	Headers     map[string]string
	QueryParams map[string]string
}

// TestResponse represents a test response
type TestResponse struct {
	StatusCode int
	Body       []byte
}

// SetupTestApp creates a new Fiber app for testing
func SetupTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": "Error occurred",
				"error":   err.Error(),
			})
		},
	})
	return app
}

// CreateAuthMiddleware creates a simple auth middleware for testing
func CreateAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set user ID in context
		userID := uuid.New()
		c.Locals("userID", userID)
		return c.Next()
	}
}

// CreateAuthMiddlewareWithUserID creates an auth middleware with a specific user ID
func CreateAuthMiddlewareWithUserID(userID uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}
}

// ExecuteRequest executes a test request against the Fiber app
func ExecuteRequest(t *testing.T, app *fiber.App, req TestRequest) *TestResponse {
	// Create HTTP request
	var reqBody []byte
	var err error
	if req.Body != nil {
		reqBody, err = json.Marshal(req.Body)
		assert.NoError(t, err)
	}

	httpReq := httptest.NewRequest(req.Method, req.URL, bytes.NewReader(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Add headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add query params
	if len(req.QueryParams) > 0 {
		q := httpReq.URL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Execute request
	resp, err := app.Test(httpReq)
	assert.NoError(t, err)

	// Read response body
	var respBody []byte
	if resp.Body != nil {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		assert.NoError(t, err)
		respBody = buf.Bytes()
	}

	return &TestResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
	}
}

// AssertJSONResponse asserts that the response matches the expected JSON
func AssertJSONResponse(t *testing.T, resp *TestResponse, expectedStatus int, expectedJSON interface{}) {
	assert.Equal(t, expectedStatus, resp.StatusCode)

	// If we expect a specific JSON response
	if expectedJSON != nil {
		expected, err := json.Marshal(expectedJSON)
		assert.NoError(t, err)

		// Compare JSON objects (ignoring whitespace differences)
		var expectedObj, actualObj interface{}
		err = json.Unmarshal(expected, &expectedObj)
		assert.NoError(t, err)

		err = json.Unmarshal(resp.Body, &actualObj)
		assert.NoError(t, err)

		assert.Equal(t, expectedObj, actualObj)
	}
}

// AssertJSONContains asserts that the response contains the expected fields
func AssertJSONContains(t *testing.T, resp *TestResponse, expectedStatus int, expectedFields map[string]interface{}) {
	assert.Equal(t, expectedStatus, resp.StatusCode)

	var actualObj map[string]interface{}
	err := json.Unmarshal(resp.Body, &actualObj)
	assert.NoError(t, err)

	for key, expectedValue := range expectedFields {
		actualValue, exists := actualObj[key]
		assert.True(t, exists, fmt.Sprintf("Expected field '%s' not found in response", key))
		assert.Equal(t, expectedValue, actualValue, fmt.Sprintf("Field '%s' value mismatch", key))
	}
}

// MockTransaction is a mock implementation of a database transaction
type MockTransaction struct {
	mock.Mock
}

// Begin mocks the Begin method
func (m *MockTransaction) Begin() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// Commit mocks the Commit method
func (m *MockTransaction) Commit() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// Rollback mocks the Rollback method
func (m *MockTransaction) Rollback() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// Error mocks the Error method
func (m *MockTransaction) Error() error {
	args := m.Called()
	return args.Error(0)
}
