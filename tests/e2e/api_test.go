package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig holds configuration for the E2E tests
type TestConfig struct {
	BaseURL  string
	Username string
	Password string
}

// loadConfig loads the test configuration from environment variables
func loadConfig() TestConfig {
	return TestConfig{
		BaseURL:  getEnvOrDefault("TEST_API_URL", "http://localhost:8080"),
		Username: getEnvOrDefault("TEST_USERNAME", "admin@example.com"),
		Password: getEnvOrDefault("TEST_PASSWORD", "password123"),
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// TestClient is a helper for making API requests
type TestClient struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthToken  string
}

// NewTestClient creates a new test client
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Login authenticates with the API
func (c *TestClient) Login(username, password string) error {
	loginData := map[string]string{
		"email":    username,
		"password": password,
	}

	resp, err := c.SendRequest("POST", "/api/auth/login", loginData, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var loginResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Token   string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	if !loginResp.Success {
		return fmt.Errorf("login failed: %s", loginResp.Message)
	}

	c.AuthToken = loginResp.Token
	return nil
}

// SendRequest sends an HTTP request to the API
func (c *TestClient) SendRequest(method, path string, body interface{}, queryParams map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.BaseURL + path
	if len(queryParams) > 0 {
		url += "?"
		for key, value := range queryParams {
			url += fmt.Sprintf("%s=%s&", key, value)
		}
		// Remove trailing &
		url = url[:len(url)-1]
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	return c.HTTPClient.Do(req)
}

// TestE2EUserFlow tests the complete user flow
func TestE2EUserFlow(t *testing.T) {
	// Skip if running in CI environment without proper setup
	if os.Getenv("CI") != "" && os.Getenv("RUN_E2E_TESTS") != "true" {
		t.Skip("Skipping E2E tests in CI environment")
	}

	config := loadConfig()
	client := NewTestClient(config.BaseURL)

	// Test login
	t.Run("Login", func(t *testing.T) {
		err := client.Login(config.Username, config.Password)
		require.NoError(t, err, "Login should succeed")
		assert.NotEmpty(t, client.AuthToken, "Auth token should be set after login")
	})

	var userID string
	var productID string

	// Test user creation
	t.Run("CreateUser", func(t *testing.T) {
		userData := map[string]interface{}{
			"email":     fmt.Sprintf("test-user-%d@example.com", time.Now().Unix()),
			"password":  "Test@123",
			"firstName": "Test",
			"lastName":  "User",
			"role":      "user",
		}

		resp, err := client.SendRequest("POST", "/api/users", userData, nil)
		require.NoError(t, err, "Create user request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Create user should return 201")

		var createResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			User    struct {
				ID string `json:"id"`
			} `json:"user"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.User.ID, "User ID should be returned")

		userID = createResp.User.ID
	})

	// Test get user
	t.Run("GetUser", func(t *testing.T) {
		if userID == "" {
			t.Skip("Skipping test because user creation failed")
		}

		resp, err := client.SendRequest("GET", "/api/users/"+userID, nil, nil)
		require.NoError(t, err, "Get user request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Get user should return 200")

		var getResp struct {
			Success bool `json:"success"`
			User    struct {
				ID string `json:"id"`
			} `json:"user"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, getResp.Success, "Response should indicate success")
		assert.Equal(t, userID, getResp.User.ID, "User ID should match")
	})

	// Test product creation
	t.Run("CreateProduct", func(t *testing.T) {
		productData := map[string]interface{}{
			"name":        fmt.Sprintf("Test Product %d", time.Now().Unix()),
			"description": "This is a test product for E2E testing",
			"sku":         fmt.Sprintf("TP-%d", time.Now().Unix()),
			"category":    "Test",
			"imageUrl":    "https://example.com/test-image.jpg",
		}

		resp, err := client.SendRequest("POST", "/api/products", productData, nil)
		require.NoError(t, err, "Create product request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Create product should return 201")

		var createResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Product struct {
				ID string `json:"id"`
			} `json:"product"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.Product.ID, "Product ID should be returned")

		productID = createResp.Product.ID
	})

	// Test get product
	t.Run("GetProduct", func(t *testing.T) {
		if productID == "" {
			t.Skip("Skipping test because product creation failed")
		}

		resp, err := client.SendRequest("GET", "/api/products/"+productID, nil, nil)
		require.NoError(t, err, "Get product request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Get product should return 200")

		var getResp struct {
			Success bool `json:"success"`
			Product struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				SKU         string `json:"sku"`
				Category    string `json:"category"`
				ImageURL    string `json:"image_url"`
			} `json:"product"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, getResp.Success, "Response should indicate success")
		assert.Equal(t, productID, getResp.Product.ID, "Product ID should match")
	})

	// Test create inventory
	t.Run("CreateInventory", func(t *testing.T) {
		if productID == "" {
			t.Skip("Skipping test because product creation failed")
		}

		inventoryData := map[string]interface{}{
			"size":     "M",
			"color":    "Blue",
			"quantity": 100,
			"location": "Warehouse A",
		}

		resp, err := client.SendRequest("POST", "/api/products/"+productID+"/inventories", inventoryData, nil)
		require.NoError(t, err, "Create inventory request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Create inventory should return 201")

		var createResp struct {
			Success   bool   `json:"success"`
			Message   string `json:"message"`
			Inventory struct {
				ID string `json:"id"`
			} `json:"inventory"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.Inventory.ID, "Inventory ID should be returned")
	})

	// Test create price
	t.Run("CreatePrice", func(t *testing.T) {
		if productID == "" {
			t.Skip("Skipping test because product creation failed")
		}

		now := time.Now()
		priceData := map[string]interface{}{
			"price":     99.99,
			"currency":  "USD",
			"startDate": now.Format(time.RFC3339),
			"endDate":   now.AddDate(0, 1, 0).Format(time.RFC3339),
		}

		resp, err := client.SendRequest("POST", "/api/products/"+productID+"/prices", priceData, nil)
		require.NoError(t, err, "Create price request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Create price should return 201")

		var createResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Price   struct {
				ID string `json:"id"`
			} `json:"price"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.Price.ID, "Price ID should be returned")
	})

	// Test get all products
	t.Run("GetAllProducts", func(t *testing.T) {
		resp, err := client.SendRequest("GET", "/api/products", nil, map[string]string{
			"page":      "1",
			"page_size": "10",
		})
		require.NoError(t, err, "Get products request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Get products should return 200")

		var getResp struct {
			Success  bool `json:"success"`
			Products []struct {
				ID string `json:"id"`
			} `json:"products"`
			Total int `json:"total"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, getResp.Success, "Response should indicate success")
		assert.Greater(t, getResp.Total, 0, "Should have at least one product")
	})

	// Test create order
	t.Run("CreateOrder", func(t *testing.T) {
		if productID == "" {
			t.Skip("Skipping test because product creation failed")
		}

		orderData := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"product_id": productID,
					"quantity":   1,
					"price":      99.99,
				},
			},
			"shipping_address": map[string]interface{}{
				"street":      "123 Test St",
				"city":        "Test City",
				"state":       "TS",
				"postal_code": "12345",
				"country":     "Test Country",
			},
			"payment_method": "credit_card",
			"total_amount":   99.99,
		}

		resp, err := client.SendRequest("POST", "/api/orders", orderData, nil)
		require.NoError(t, err, "Create order request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Create order should return 201")

		var createResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Order   struct {
				ID string `json:"id"`
			} `json:"order"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.Order.ID, "Order ID should be returned")
	})

	// Test notifications
	t.Run("GetNotifications", func(t *testing.T) {
		resp, err := client.SendRequest("GET", "/api/notifications", nil, map[string]string{
			"page":      "1",
			"page_size": "10",
		})
		require.NoError(t, err, "Get notifications request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Get notifications should return 200")

		var getResp struct {
			Success       bool `json:"success"`
			Notifications []struct {
				ID string `json:"id"`
			} `json:"notifications"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, getResp.Success, "Response should indicate success")
	})
}
