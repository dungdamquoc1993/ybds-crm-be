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
	"github.com/stretchr/testify/mock"
	"github.com/ybds/internal/models/product"
	"github.com/ybds/internal/testutil"
)

// mockJWTMiddleware creates a simple JWT middleware for testing
func productMockJWTMiddleware(c *fiber.Ctx) error {
	// Set user ID in locals for testing
	c.Locals("user_id", "test-user-id")
	c.Locals("roles", []string{"admin"})
	return c.Next()
}

func TestProductHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a mock product service
	mockProductService := new(testutil.MockProductService)

	// Register routes with mock middleware
	api := app.Group("/api")
	products := api.Group("/products")
	products.Use(productMockJWTMiddleware)

	// Register the GetProducts endpoint
	products.Get("/", func(c *fiber.Ctx) error {
		// Parse pagination parameters
		page := 1
		pageSize := 10

		// Get products from service
		products, total, err := mockProductService.GetAllProducts(page, pageSize, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to retrieve products",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Products retrieved successfully",
			"data": fiber.Map{
				"products":    products,
				"total":       total,
				"page":        page,
				"page_size":   pageSize,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	})

	// Register the GetProductByID endpoint
	products.Get("/:id", func(c *fiber.Ctx) error {
		// Parse product ID from path
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid product ID format",
				"error":   err.Error(),
			})
		}

		// Get product from service
		product, err := mockProductService.GetProductByID(id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Product not found",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Product retrieved successfully",
			"data":    product,
		})
	})

	// Register the CreateProduct endpoint
	products.Post("/", func(c *fiber.Ctx) error {
		// Parse request body
		var request struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			SKU         string `json:"sku"`
			Category    string `json:"category"`
			ImageURL    string `json:"image_url"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		}

		// Validate request
		if request.Name == "" || request.SKU == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Name and SKU are required",
			})
		}

		// Create product
		result, err := mockProductService.CreateProduct(
			request.Name,
			request.Description,
			request.SKU,
			request.Category,
			request.ImageURL,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create product",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Product created successfully",
			"data":    result,
		})
	})

	// Test case 1: Get all products
	t.Run("GetProducts", func(t *testing.T) {
		// Setup mock expectations
		products := []product.Product{
			{
				Name:        "Product 1",
				Description: "Description 1",
				SKU:         "SKU-001",
				Category:    "Category 1",
			},
			{
				Name:        "Product 2",
				Description: "Description 2",
				SKU:         "SKU-002",
				Category:    "Category 2",
			},
		}
		products[0].ID = uuid.New()
		products[1].ID = uuid.New()

		mockProductService.On("GetAllProducts", 1, 10, mock.Anything).Return(products, int64(2), nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
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
		assert.Equal(t, "Products retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["products"])
		assert.Equal(t, float64(2), data["total"])
		assert.Equal(t, float64(1), data["page"])
		assert.Equal(t, float64(10), data["page_size"])

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	// Test case 2: Get product by ID
	t.Run("GetProductByID", func(t *testing.T) {
		// Setup mock expectations
		productID := uuid.New()
		product := &product.Product{
			Name:        "Test Product",
			Description: "Test Description",
			SKU:         "TEST-SKU",
			Category:    "Test Category",
		}
		product.ID = productID

		mockProductService.On("GetProductByID", productID).Return(product, nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/products/"+productID.String(), nil)
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
		assert.Equal(t, "Product retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Test Product", data["name"])
		assert.Equal(t, "Test Description", data["description"])
		assert.Equal(t, "TEST-SKU", data["sku"])
		assert.Equal(t, "Test Category", data["category"])

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	// Test case 3: Get product by ID - not found
	t.Run("GetProductByID_NotFound", func(t *testing.T) {
		// Setup mock expectations
		productID := uuid.New()
		mockProductService.On("GetProductByID", productID).Return(nil, errors.New("product not found")).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/products/"+productID.String(), nil)
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
		assert.Equal(t, "Product not found", response["message"])
		assert.Equal(t, "product not found", response["error"])

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	// Test case 4: Get product by ID - invalid ID
	t.Run("GetProductByID_InvalidID", func(t *testing.T) {
		// Create request with invalid UUID
		req := httptest.NewRequest(http.MethodGet, "/api/products/invalid-uuid", nil)
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
		assert.Equal(t, "Invalid product ID format", response["message"])
	})

	// Test case 5: Create product
	t.Run("CreateProduct", func(t *testing.T) {
		// Setup mock expectations
		productID := uuid.New()
		result := &testutil.ProductResult{
			Success:   true,
			Message:   "Product created successfully",
			ProductID: productID,
			Name:      "New Product",
			SKU:       "NEW-SKU",
		}

		mockProductService.On(
			"CreateProduct",
			"New Product",
			"New Description",
			"NEW-SKU",
			"New Category",
			"https://example.com/image.jpg",
		).Return(result, nil).Once()

		// Create request body
		requestBody := map[string]interface{}{
			"name":        "New Product",
			"description": "New Description",
			"sku":         "NEW-SKU",
			"category":    "New Category",
			"image_url":   "https://example.com/image.jpg",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "Product created successfully", response["message"])

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	// Test case 6: Create product - validation error
	t.Run("CreateProduct_ValidationError", func(t *testing.T) {
		// Create request body with missing required fields
		requestBody := map[string]interface{}{
			"description": "New Description",
			"category":    "New Category",
			"image_url":   "https://example.com/image.jpg",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "Name and SKU are required", response["message"])
	})

	// Test case 7: Create product - server error
	t.Run("CreateProduct_ServerError", func(t *testing.T) {
		// Setup mock expectations
		mockProductService.On(
			"CreateProduct",
			"Error Product",
			"Error Description",
			"ERROR-SKU",
			"Error Category",
			"https://example.com/error.jpg",
		).Return(nil, errors.New("database error")).Once()

		// Create request body
		requestBody := map[string]interface{}{
			"name":        "Error Product",
			"description": "Error Description",
			"sku":         "ERROR-SKU",
			"category":    "Error Category",
			"image_url":   "https://example.com/error.jpg",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "Failed to create product", response["message"])
		assert.Equal(t, "database error", response["error"])

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})
}
