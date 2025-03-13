package handlers_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ybds/internal/api/handlers"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/models/product"
	"github.com/ybds/internal/services"
	"github.com/ybds/internal/testutil"
)

// MockProductNotificationService is a mock implementation of services.NotificationService for product tests
type MockProductNotificationService struct {
	*services.NotificationService
}

// NewMockProductNotificationService creates a new mock notification service for product tests
func NewMockProductNotificationService() *MockProductNotificationService {
	return &MockProductNotificationService{
		NotificationService: &services.NotificationService{},
	}
}

func TestProductHandler(t *testing.T) {
	// Create mock services
	mockProductService := new(testutil.MockProductService)
	mockNotificationService := NewMockProductNotificationService()

	// Create test app
	app := testutil.SetupTestApp()

	// Create a fixed user ID for testing
	userID := uuid.New()

	// Create auth middleware with fixed user ID
	authMiddleware := testutil.CreateAuthMiddlewareWithUserID(userID)

	// Create product handler with correct parameters
	productHandler := handlers.NewProductHandler(nil, mockNotificationService.NotificationService)

	// Register routes
	api := app.Group("/api")
	authenticated := api.Group("/")
	authenticated.Use(authMiddleware)
	productHandler.RegisterRoutes(authenticated, authMiddleware)

	t.Run("GetProducts", func(t *testing.T) {
		// Create test products
		products := []product.Product{
			{
				Name:        "Test Product 1",
				Description: "This is a test product 1",
				SKU:         "TP001",
				Category:    "Test Category",
				ImageURL:    "https://example.com/image1.jpg",
			},
			{
				Name:        "Test Product 2",
				Description: "This is a test product 2",
				SKU:         "TP002",
				Category:    "Test Category",
				ImageURL:    "https://example.com/image2.jpg",
			},
		}

		// Set IDs and timestamps manually
		now := time.Now()
		products[0].ID = uuid.New()
		products[0].CreatedAt = now.Add(-24 * time.Hour)
		products[0].UpdatedAt = now.Add(-24 * time.Hour)

		products[1].ID = uuid.New()
		products[1].CreatedAt = now.Add(-48 * time.Hour)
		products[1].UpdatedAt = now.Add(-24 * time.Hour)

		// Setup mock expectations
		mockProductService.On("GetAllProducts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(products, 2, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/products",
			QueryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
			},
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	t.Run("GetProducts_Error", func(t *testing.T) {
		// Setup mock expectations
		mockProductService.On("GetAllProducts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return([]product.Product{}, 0, errors.New("database error"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/products",
		})

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	t.Run("GetProduct", func(t *testing.T) {
		// Create test product
		productID := uuid.New()
		testProduct := product.Product{
			Name:        "Test Product",
			Description: "This is a test product",
			SKU:         "TP001",
			Category:    "Test Category",
			ImageURL:    "https://example.com/image.jpg",
		}
		testProduct.ID = productID
		testProduct.CreatedAt = time.Now().Add(-24 * time.Hour)
		testProduct.UpdatedAt = time.Now().Add(-24 * time.Hour)

		// Setup mock expectations
		mockProductService.On("GetProductByID", productID).
			Return(testProduct, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/products/" + productID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	t.Run("GetProduct_NotFound", func(t *testing.T) {
		// Create product ID
		productID := uuid.New()

		// Setup mock expectations
		mockProductService.On("GetProductByID", productID).
			Return(product.Product{}, errors.New("product not found"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/products/" + productID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	t.Run("CreateProduct", func(t *testing.T) {
		// Create product request
		createReq := requests.CreateProductRequest{
			Name:        "New Product",
			Description: "This is a new product",
			SKU:         "NP001",
			Category:    "New Category",
			ImageURL:    "https://example.com/new-image.jpg",
		}

		// Create product result
		productID := uuid.New()
		createdProduct := product.Product{
			Name:        createReq.Name,
			Description: createReq.Description,
			SKU:         createReq.SKU,
			Category:    createReq.Category,
			ImageURL:    createReq.ImageURL,
		}
		createdProduct.ID = productID
		createdProduct.CreatedAt = time.Now()
		createdProduct.UpdatedAt = time.Now()

		// Setup mock expectations
		mockProductService.On("CreateProduct",
			createReq.Name,
			createReq.Description,
			createReq.SKU,
			createReq.Category,
			createReq.ImageURL).
			Return(createdProduct, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPost,
			URL:    "/api/products",
			Body:   createReq,
		})

		// Assert response
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	t.Run("CreateProduct_ValidationError", func(t *testing.T) {
		// Create invalid product request (missing required fields)
		createReq := requests.CreateProductRequest{
			// Name is required but missing
			Description: "This is a new product",
			// SKU is required but missing
			Category: "New Category",
			ImageURL: "https://example.com/new-image.jpg",
		}

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPost,
			URL:    "/api/products",
			Body:   createReq,
		})

		// Assert response
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UpdateProduct", func(t *testing.T) {
		// Create product ID
		productID := uuid.New()

		// Create update request
		updateReq := requests.UpdateProductRequest{
			Name:        "Updated Product",
			Description: "This is an updated product",
			SKU:         "UP001",
			Category:    "Updated Category",
			ImageURL:    "https://example.com/updated-image.jpg",
		}

		// Create updated product
		updatedProduct := product.Product{
			Name:        updateReq.Name,
			Description: updateReq.Description,
			SKU:         updateReq.SKU,
			Category:    updateReq.Category,
			ImageURL:    updateReq.ImageURL,
		}
		updatedProduct.ID = productID
		updatedProduct.CreatedAt = time.Now().Add(-24 * time.Hour)
		updatedProduct.UpdatedAt = time.Now()

		// Setup mock expectations
		mockProductService.On("UpdateProduct",
			productID,
			updateReq.Name,
			updateReq.Description,
			updateReq.SKU,
			updateReq.Category,
			updateReq.ImageURL).
			Return(updatedProduct, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/products/" + productID.String(),
			Body:   updateReq,
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	t.Run("UpdateProduct_NotFound", func(t *testing.T) {
		// Create product ID
		productID := uuid.New()

		// Create update request
		updateReq := requests.UpdateProductRequest{
			Name:        "Updated Product",
			Description: "This is an updated product",
			SKU:         "UP001",
			Category:    "Updated Category",
			ImageURL:    "https://example.com/updated-image.jpg",
		}

		// Setup mock expectations
		mockProductService.On("UpdateProduct",
			productID,
			updateReq.Name,
			updateReq.Description,
			updateReq.SKU,
			updateReq.Category,
			updateReq.ImageURL).
			Return(product.Product{}, errors.New("product not found"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/products/" + productID.String(),
			Body:   updateReq,
		})

		// Assert response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	t.Run("DeleteProduct", func(t *testing.T) {
		// Create product ID
		productID := uuid.New()

		// Setup mock expectations
		mockProductService.On("DeleteProduct", productID).
			Return(nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodDelete,
			URL:    "/api/products/" + productID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify expected response structure
		testutil.AssertJSONContains(t, resp, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Product deleted successfully",
		})

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})

	t.Run("DeleteProduct_NotFound", func(t *testing.T) {
		// Create product ID
		productID := uuid.New()

		// Setup mock expectations
		mockProductService.On("DeleteProduct", productID).
			Return(errors.New("product not found"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodDelete,
			URL:    "/api/products/" + productID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify mock expectations
		mockProductService.AssertExpectations(t)
	})
}
