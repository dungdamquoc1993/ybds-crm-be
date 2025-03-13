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
		BaseURL:  getEnvOrDefault("TEST_API_URL", "http://localhost:3000"),
		Username: getEnvOrDefault("TEST_USERNAME", "admin@example.com"),
		Password: getEnvOrDefault("TEST_PASSWORD", "admin123"),
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
		"username": username,
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
		User    struct {
			ID       string   `json:"id"`
			Username string   `json:"username"`
			Email    string   `json:"email"`
			Roles    []string `json:"roles"`
		} `json:"user"`
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
		fmt.Printf("DEBUG: Setting Authorization header: Bearer %s\n", c.AuthToken)
	} else {
		fmt.Println("DEBUG: No AuthToken set")
	}

	return c.HTTPClient.Do(req)
}

// TestE2EUserFlow tests the complete user flow from login to creating and managing resources
func TestE2EUserFlow(t *testing.T) {
	// Only skip if explicitly requested
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E tests as requested by environment variable")
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
			"email":    fmt.Sprintf("test-user-%d@example.com", time.Now().Unix()),
			"phone":    "",
			"password": "Test@123",
		}

		resp, err := client.SendRequest("POST", "/api/auth/register", userData, nil)
		require.NoError(t, err, "Create user request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Create user should return 200")

		var createResp struct {
			Success  bool   `json:"success"`
			Message  string `json:"message"`
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.UserID, "User ID should be returned")

		userID = createResp.UserID
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
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				ID string `json:"id"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, getResp.Success, "Response should indicate success")
		assert.Equal(t, userID, getResp.Data.ID, "User ID should match")
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
			Data    struct {
				ProductID string `json:"ProductID"`
				Name      string `json:"Name"`
				SKU       string `json:"SKU"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.Data.ProductID, "Product ID should be returned")

		productID = createResp.Data.ProductID
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
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				SKU         string `json:"sku"`
				Category    string `json:"category"`
				ImageURL    string `json:"image_url"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, getResp.Success, "Response should indicate success")
		assert.Equal(t, productID, getResp.Data.ID, "Product ID should match")
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
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				InventoryID string `json:"InventoryID"`
				ProductID   string `json:"ProductID"`
				Quantity    int    `json:"Quantity"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.Data.InventoryID, "Inventory ID should be returned")
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
			Data    struct {
				PriceID   string  `json:"PriceID"`
				ProductID string  `json:"ProductID"`
				Price     float64 `json:"Price"`
				Currency  string  `json:"Currency"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.Data.PriceID, "Price ID should be returned")
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
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				Products []struct {
					ID string `json:"id"`
				} `json:"products"`
				Total      int64 `json:"total"`
				Page       int   `json:"page"`
				PageSize   int   `json:"page_size"`
				TotalPages int   `json:"total_pages"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, getResp.Success, "Response should indicate success")
		assert.Greater(t, getResp.Data.Total, int64(0), "Should have at least one product")
	})

	// Test create order
	t.Run("CreateOrder", func(t *testing.T) {
		if productID == "" {
			t.Skip("Skipping test because product creation failed")
		}

		// First, we need to get the inventory ID for the product
		resp, err := client.SendRequest("GET", "/api/products/"+productID, nil, nil)
		require.NoError(t, err, "Get product request should succeed")
		defer resp.Body.Close()

		var getProductResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				ID          string `json:"id"`
				Inventories []struct {
					ID string `json:"id"`
				} `json:"inventories"`
				Prices []struct {
					ID string `json:"id"`
				} `json:"prices"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getProductResp)
		require.NoError(t, err, "Should decode response JSON")

		// Skip if no inventories or prices
		if len(getProductResp.Data.Inventories) == 0 || len(getProductResp.Data.Prices) == 0 {
			t.Skip("Skipping test because product has no inventories or prices")
		}

		inventoryID := getProductResp.Data.Inventories[0].ID
		priceID := getProductResp.Data.Prices[0].ID

		orderData := map[string]interface{}{
			"customer_id":         userID, // Use the user ID from the CreateUser test
			"shipping_address_id": userID, // Use the user ID as a placeholder
			"payment_method":      "credit_card",
			"status":              "pending",
			"notes":               "Test order",
			"items": []map[string]interface{}{
				{
					"product_id":   productID,
					"inventory_id": inventoryID,
					"price_id":     priceID,
					"quantity":     1,
					"notes":        "Test item",
				},
			},
		}

		resp, err = client.SendRequest("POST", "/api/orders", orderData, nil)
		require.NoError(t, err, "Create order request should succeed")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Create order should return 201")

		var createResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				OrderID   string  `json:"OrderID"`
				Status    string  `json:"Status"`
				Total     float64 `json:"Total"`
				CreatedBy string  `json:"CreatedBy"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&createResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, createResp.Success, "Response should indicate success")
		assert.NotEmpty(t, createResp.Data.OrderID, "Order ID should be returned")
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
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    []struct {
				ID string `json:"id"`
			} `json:"data"`
			Total      int64 `json:"total"`
			Page       int   `json:"page"`
			PageSize   int   `json:"page_size"`
			TotalPages int   `json:"total_pages"`
		}

		err = json.NewDecoder(resp.Body).Decode(&getResp)
		require.NoError(t, err, "Should decode response JSON")
		assert.True(t, getResp.Success, "Response should indicate success")
	})
}
