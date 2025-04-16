package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/api/responses"
	"github.com/ybds/internal/services"
	"github.com/ybds/pkg/upload"
	"gorm.io/gorm"
)

// ProductImage represents a product image in Swagger documentation
// @Description Product image information
// @Schema ProductImage
type ProductImage struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductID string `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL       string `json:"url" example:"/uploads/products/product_20230101_120000_abcdef12.jpg"`
	Filename  string `json:"filename" example:"product_20230101_120000_abcdef12.jpg"`
	IsPrimary bool   `json:"is_primary" example:"true"`
	SortOrder int    `json:"sort_order" example:"0"`
	CreatedAt string `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2023-01-01T12:00:00Z"`
}

// ReorderRequest represents the request to reorder product images
// @Description Request to reorder product images
// @Schema ReorderRequest
type ReorderRequest struct {
	ImageIDs []string `json:"imageIds" example:"['550e8400-e29b-41d4-a716-446655440000','550e8400-e29b-41d4-a716-446655440001']"`
}

// ProductHandler handles HTTP requests related to products
type ProductHandler struct {
	productService *services.ProductService
}

// NewProductHandler creates a new instance of ProductHandler
func NewProductHandler(db *gorm.DB, notificationService *services.NotificationService, uploadService *upload.Service) *ProductHandler {
	return &ProductHandler{
		productService: services.NewProductService(db, notificationService, uploadService),
	}
}

// RegisterRoutes registers all routes related to products
func (h *ProductHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	products := router.Group("/products")
	products.Use(authMiddleware)

	// Product routes
	products.Post("/", h.CreateProduct)
	products.Get("/", h.GetProducts)
	products.Get("/:id", h.GetProductByID)
	products.Put("/:id", h.UpdateProduct)
	products.Delete("/:id", h.DeleteProduct)

	// Inventory routes
	products.Post("/:id/inventories", h.CreateInventory)
	products.Post("/:id/inventories/batch", h.CreateInventory)
	products.Put("/inventories/:id", h.UpdateInventory)
	products.Delete("/inventories/:id", h.DeleteInventory)

	// Price routes
	products.Post("/:id/prices", h.CreatePrice)
	products.Put("/prices/:id", h.UpdatePrice)
	products.Delete("/prices/:id", h.DeletePrice)

	// Image routes
	products.Get("/:id/images", h.GetProductImages)
	products.Post("/:id/images", h.UploadProductImage)
	products.Post("/:id/images/multiple", h.UploadMultipleProductImages)
	products.Put("/:id/images/:imageId/primary", h.SetPrimaryProductImage)
	products.Put("/:id/images/reorder", h.ReorderProductImages)
	products.Delete("/:id/images/:imageId", h.DeleteProductImage)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with optional inventories, prices, and images. Images are stored at /uploads/products/ path.
// @Tags products
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Product name"
// @Param description formData string false "Product description"
// @Param sku formData string true "Product SKU (unique identifier)"
// @Param category formData string true "Product category"
// @Param inventories formData string false "JSON array of inventory objects [{\"size\":\"M\",\"color\":\"Red\",\"quantity\":10,\"location\":\"Warehouse A\"}]"
// @Param prices formData string false "JSON array of price objects [{\"price\":99.99,\"currency\":\"USD\",\"endDate\":\"2023-12-31T23:59:59Z\"}]"
// @Param images formData file false "Product images (can upload multiple, first image will be set as primary)"
// @Success 201 {object} responses.ProductDetailResponse "Returns the created product with all related data"
// @Failure 400 {object} responses.ErrorResponse "Invalid request data"
// @Failure 500 {object} responses.ErrorResponse "Server error"
// @Router /api/products [post]
// @Security ApiKeyAuth
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	// Parse form fields
	name := c.FormValue("name")
	description := c.FormValue("description")
	sku := c.FormValue("sku")
	category := c.FormValue("category")

	// Validate required fields
	if name == "" || sku == "" || category == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   "Name, SKU, and category are required",
		})
	}

	// Create product
	result, err := h.productService.CreateProduct(
		name,
		description,
		sku,
		category,
		"", // Empty image URL, will be updated if images are uploaded
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to create product",
			Error:   err.Error(),
		})
	}

	// Parse and create inventories if provided
	inventoriesJSON := c.FormValue("inventories")
	if inventoriesJSON != "" {
		var inventories []requests.InventoryRequest
		if err := json.Unmarshal([]byte(inventoriesJSON), &inventories); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Message: "Invalid inventories data",
				Error:   err.Error(),
			})
		}

		for _, inv := range inventories {
			_, err := h.productService.CreateInventory(
				result.ProductID,
				inv.Size,
				inv.Color,
				inv.Quantity,
				inv.Location,
			)

			if err != nil {
				// Log error but continue
				log.Printf("Error creating inventory: %v", err)
			}
		}
	}

	// Parse and create prices if provided
	pricesJSON := c.FormValue("prices")
	if pricesJSON != "" {
		var prices []requests.PriceRequest
		if err := json.Unmarshal([]byte(pricesJSON), &prices); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Message: "Invalid prices data",
				Error:   err.Error(),
			})
		}

		for _, price := range prices {
			startDate := time.Now()
			var endDate *time.Time

			if price.EndDate != nil {
				ed := *price.EndDate
				endDate = &ed
			}

			_, err := h.productService.CreatePrice(
				result.ProductID,
				price.Price,
				price.Currency,
				startDate,
				endDate,
			)

			if err != nil {
				// Log error but continue
				log.Printf("Error creating price: %v", err)
			}
		}
	}

	// Handle image uploads
	form, err := c.MultipartForm()
	if err == nil && form.File["images"] != nil {
		for i, file := range form.File["images"] {
			// First image is primary
			isPrimary := i == 0

			// Upload the image
			_, err := h.productService.UploadProductImage(result.ProductID, file, isPrimary)
			if err != nil {
				// Log error but continue
				log.Printf("Error uploading image: %v", err)
			}
		}
	}

	// Get the updated product with all relations
	product, err := h.productService.GetProductByID(result.ProductID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Product created but failed to retrieve details",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product created successfully",
		"data":    product,
	})
}

// GetProducts godoc
// @Summary Get all products
// @Description Get a list of all products with pagination, filtering and search
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search term"
// @Param category query string false "Filter by category"
// @Success 200 {object} responses.ProductsResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products [get]
// @Security ApiKeyAuth
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	// Ensure page and pageSize are valid
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Parse filters
	filters := make(map[string]interface{})

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}

	// First, get the total count to calculate total pages
	_, total, err := h.productService.GetAllProducts(1, 1, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve products count",
			Error:   err.Error(),
		})
	}

	// Calculate total pages
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	// Adjust page if it exceeds total pages
	if totalPages > 0 && int64(page) > totalPages {
		page = int(totalPages)
	}

	// Get products
	products, _, err := h.productService.GetAllProducts(page, pageSize, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve products",
			Error:   err.Error(),
		})
	}

	// Convert products to response objects
	productResponses := responses.ConvertToProductDetailResponses(products)

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Products retrieved successfully",
		"data": fiber.Map{
			"products":    productResponses,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// GetProductByID godoc
// @Summary Get a product by ID
// @Description Get detailed information about a product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} responses.ProductDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/{id} [get]
// @Security ApiKeyAuth
func (h *ProductHandler) GetProductByID(c *fiber.Ctx) error {
	// Parse product ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Get product
	product, err := h.productService.GetProductByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Product not found",
			Error:   err.Error(),
		})
	}

	// Convert product to response object
	productResponse := responses.ConvertToProductDetailResponse(*product)

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product retrieved successfully",
		"data":    productResponse,
	})
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update a product's information and optionally upload new images. Images are stored at /uploads/products/ path.
// @Tags products
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Product ID"
// @Param name formData string false "Product name"
// @Param description formData string false "Product description"
// @Param sku formData string false "Product SKU (unique identifier)"
// @Param category formData string false "Product category"
// @Param images formData file false "Product images to add (can upload multiple, first image will be set as primary if no existing images)"
// @Success 200 {object} responses.ProductDetailResponse "Returns the updated product with all related data"
// @Failure 400 {object} responses.ErrorResponse "Invalid request data"
// @Failure 404 {object} responses.ErrorResponse "Product not found"
// @Failure 500 {object} responses.ErrorResponse "Server error"
// @Router /api/products/{id} [put]
// @Security ApiKeyAuth
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	// Parse product ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Parse form fields
	name := c.FormValue("name")
	description := c.FormValue("description")
	sku := c.FormValue("sku")
	category := c.FormValue("category")

	// Get existing product to check if it exists
	_, err = h.productService.GetProductByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Product not found",
			Error:   err.Error(),
		})
	}

	// Update product
	_, err = h.productService.UpdateProduct(
		id,
		name,
		description,
		sku,
		category,
		"", // Empty image URL, will be updated if a primary image exists
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update product",
			Error:   err.Error(),
		})
	}

	// Handle image uploads
	form, err := c.MultipartForm()
	if err == nil && form.File["images"] != nil {
		// Get existing images to determine if this is the first image
		existingImages, _ := h.productService.GetProductImages(id)
		hasExistingImages := len(existingImages) > 0

		for i, file := range form.File["images"] {
			// First image is primary only if there are no existing images
			isPrimary := !hasExistingImages && i == 0

			// Upload the image
			_, err := h.productService.UploadProductImage(id, file, isPrimary)
			if err != nil {
				// Log error but continue
				log.Printf("Error uploading image: %v", err)
			}
		}
	}

	// Get the updated product with all relations
	product, err := h.productService.GetProductByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Product updated but failed to retrieve details",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product updated successfully",
		"data":    product,
	})
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product and all associated inventories and prices
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/{id} [delete]
// @Security ApiKeyAuth
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	// Parse product ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Delete product
	result, err := h.productService.DeleteProduct(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to delete product",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product deleted successfully",
		"data":    result,
	})
}

// CreateInventory godoc
// @Summary Create inventory for a product
// @Description Add inventory information for a specific product. Can create a single inventory or multiple inventories at once.
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param inventory body requests.CreateInventoryRequest true "Single inventory information"
// @Success 201 {object} responses.InventoryResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/{id}/inventories [post]
// @Security ApiKeyAuth
func (h *ProductHandler) CreateInventory(c *fiber.Ctx) error {
	// Parse product ID
	idStr := c.Params("id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Check if the request is for multiple inventories
	contentType := c.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		// Try to parse as multiple inventories first
		var multiReq requests.CreateMultipleInventoriesRequest
		if err := c.BodyParser(&multiReq); err == nil && len(multiReq.Inventories) > 0 {
			return h.createMultipleInventories(c, productID, multiReq)
		}
	}

	// If not multiple inventories, process as a single inventory
	var req requests.CreateInventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Create inventory
	result, err := h.productService.CreateInventory(
		productID,
		req.Size,
		req.Color,
		req.Quantity,
		req.Location,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to create inventory",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Inventory created successfully",
		"data":    result,
	})
}

// CreateMultipleInventories godoc
// @Summary Create multiple inventories for a product
// @Description Add multiple inventory items for a specific product at once
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param inventories body requests.CreateMultipleInventoriesRequest true "Multiple inventory information"
// @Success 201 {object} responses.SuccessResponse{data=[]responses.InventoryResponse}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/{id}/inventories/batch [post]
// @Security ApiKeyAuth
func (h *ProductHandler) createMultipleInventories(c *fiber.Ctx, productID uuid.UUID, req requests.CreateMultipleInventoriesRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Create inventories
	var results []interface{}
	var failedCount int

	for _, inv := range req.Inventories {
		result, err := h.productService.CreateInventory(
			productID,
			inv.Size,
			inv.Color,
			inv.Quantity,
			inv.Location,
		)

		if err != nil {
			failedCount++
			// Log error but continue with other inventories
			log.Printf("Failed to create inventory: %v", err)
		} else {
			results = append(results, result)
		}
	}

	// Return response
	message := fmt.Sprintf("%d inventories created successfully", len(results))
	if failedCount > 0 {
		message += fmt.Sprintf(", %d failed", failedCount)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    results,
	})
}

// UpdateInventory godoc
// @Summary Update an inventory
// @Description Update inventory information
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Inventory ID"
// @Param inventory body requests.UpdateInventoryRequest true "Updated inventory information"
// @Success 200 {object} responses.InventoryResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/inventories/{id} [put]
// @Security ApiKeyAuth
func (h *ProductHandler) UpdateInventory(c *fiber.Ctx) error {
	// Parse inventory ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid inventory ID format",
			Error:   err.Error(),
		})
	}

	// Parse request
	var req requests.UpdateInventoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Get inventory to check if it exists
	_, err = h.productService.GetInventoryByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Inventory not found",
			Error:   err.Error(),
		})
	}

	// Create a pointer to the quantity value
	var quantityPtr *int
	if req.Quantity != 0 {
		quantity := req.Quantity
		quantityPtr = &quantity
	}

	// Update inventory
	result, err := h.productService.UpdateInventory(
		id,
		req.Size,
		req.Color,
		quantityPtr,
		req.Location,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update inventory",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Inventory updated successfully",
		"data":    result,
	})
}

// DeleteInventory godoc
// @Summary Delete an inventory
// @Description Delete an inventory record
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Inventory ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/inventories/{id} [delete]
// @Security ApiKeyAuth
func (h *ProductHandler) DeleteInventory(c *fiber.Ctx) error {
	// Parse inventory ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid inventory ID format",
			Error:   err.Error(),
		})
	}

	// Delete inventory
	result, err := h.productService.DeleteInventory(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to delete inventory",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Inventory deleted successfully",
		"data":    result,
	})
}

// CreatePrice godoc
// @Summary Create a new price for a product
// @Description Add price information for a specific product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param price body requests.CreatePriceRequest true "Price information"
// @Success 200 {object} responses.PriceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/{id}/prices [post]
// @Security ApiKeyAuth
func (h *ProductHandler) CreatePrice(c *fiber.Ctx) error {
	// Parse product ID
	idStr := c.Params("id")
	productID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Parse request
	var req requests.CreatePriceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Create price
	startDate := time.Now()
	var endDate *time.Time

	if req.EndDate != nil {
		ed := *req.EndDate
		endDate = &ed
	}

	result, err := h.productService.CreatePrice(
		productID,
		req.Price,
		req.Currency,
		startDate,
		endDate,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to create price",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Price created successfully",
		"data":    result,
	})
}

// UpdatePrice godoc
// @Summary Update a price
// @Description Update price information
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Price ID"
// @Param price body requests.UpdatePriceRequest true "Updated price information"
// @Success 200 {object} responses.PriceResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/prices/{id} [put]
// @Security ApiKeyAuth
func (h *ProductHandler) UpdatePrice(c *fiber.Ctx) error {
	// Parse price ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid price ID format",
			Error:   err.Error(),
		})
	}

	// Parse request
	var req requests.UpdatePriceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Get price to check if it exists
	_, err = h.productService.GetPriceByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Price not found",
			Error:   err.Error(),
		})
	}

	// Create pointers for optional fields
	var pricePtr *float64
	if req.Price > 0 {
		price := req.Price
		pricePtr = &price
	}

	var endDatePtr *time.Time
	if req.EndDate != nil {
		endDate := *req.EndDate
		endDatePtr = &endDate
	}

	// Update price
	result, err := h.productService.UpdatePrice(
		id,
		pricePtr,
		req.Currency,
		nil, // We don't allow updating start date
		endDatePtr,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update price",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Price updated successfully",
		"data":    result,
	})
}

// DeletePrice godoc
// @Summary Delete a price
// @Description Delete a price record
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Price ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/products/prices/{id} [delete]
// @Security ApiKeyAuth
func (h *ProductHandler) DeletePrice(c *fiber.Ctx) error {
	// Parse price ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid price ID format",
			Error:   err.Error(),
		})
	}

	// Delete price
	result, err := h.productService.DeletePrice(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to delete price",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Price deleted successfully",
		"data":    result,
	})
}

// GetProductImages godoc
// @Summary Get all images for a product
// @Description Get a list of all images associated with a product. Images are stored at /uploads/products/ path.
// @Tags product-images
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} responses.SuccessResponse{data=[]ProductImage} "Returns an array of product images with URLs, filenames, and metadata"
// @Failure 400 {object} responses.ErrorResponse "Invalid product ID format"
// @Failure 404 {object} responses.ErrorResponse "Product not found"
// @Failure 500 {object} responses.ErrorResponse "Server error"
// @Router /api/products/{id}/images [get]
// @Security ApiKeyAuth
func (h *ProductHandler) GetProductImages(c *fiber.Ctx) error {
	// Parse product ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Get product images
	images, err := h.productService.GetProductImages(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve product images",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product images retrieved successfully",
		"data":    images,
	})
}

// UploadProductImage godoc
// @Summary Upload an image for a product
// @Description Upload a new image for a product. Images are stored at /uploads/products/ path.
// @Tags product-images
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Product ID"
// @Param file formData file true "Image file (supported formats: JPG, PNG, GIF)"
// @Param is_primary formData boolean false "Set as primary image (default: false)"
// @Success 200 {object} responses.SuccessResponse{data=ProductImage} "Returns the uploaded image details including URL and metadata"
// @Failure 400 {object} responses.ErrorResponse "Invalid request or file format"
// @Failure 404 {object} responses.ErrorResponse "Product not found"
// @Failure 500 {object} responses.ErrorResponse "Server error"
// @Router /api/products/{id}/images [post]
// @Security ApiKeyAuth
func (h *ProductHandler) UploadProductImage(c *fiber.Ctx) error {
	// Parse product ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to get file from request",
			Error:   err.Error(),
		})
	}

	// Parse is_primary parameter
	isPrimary := c.FormValue("is_primary", "false") == "true"

	// Upload the image
	result, err := h.productService.UploadProductImage(id, file, isPrimary)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to upload product image",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product image uploaded successfully",
		"data":    result,
	})
}

// SetPrimaryProductImage godoc
// @Summary Set an image as the primary image for a product
// @Description Set an existing image as the primary image for a product. The primary image URL will be used as the main product image.
// @Tags product-images
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param imageId path string true "Image ID"
// @Success 200 {object} responses.SuccessResponse{data=ProductImage} "Returns the updated image details"
// @Failure 400 {object} responses.ErrorResponse "Invalid ID format"
// @Failure 404 {object} responses.ErrorResponse "Product or image not found"
// @Failure 500 {object} responses.ErrorResponse "Server error"
// @Router /api/products/{id}/images/{imageId}/primary [put]
// @Security ApiKeyAuth
func (h *ProductHandler) SetPrimaryProductImage(c *fiber.Ctx) error {
	// Parse product ID
	productIDStr := c.Params("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Parse image ID
	imageIDStr := c.Params("imageId")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid image ID format",
			Error:   err.Error(),
		})
	}

	// Set as primary
	result, err := h.productService.SetPrimaryProductImage(imageID, productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to set primary image",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Primary image set successfully",
		"data":    result,
	})
}

// ReorderProductImages godoc
// @Summary Reorder product images
// @Description Update the sort order of product images to control their display order
// @Tags product-images
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param body body ReorderRequest true "Array of image IDs in the desired display order"
// @Success 200 {object} responses.SuccessResponse "Images reordered successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid request or ID format"
// @Failure 404 {object} responses.ErrorResponse "Product or image not found"
// @Failure 500 {object} responses.ErrorResponse "Server error"
// @Router /api/products/{id}/images/reorder [put]
// @Security ApiKeyAuth
func (h *ProductHandler) ReorderProductImages(c *fiber.Ctx) error {
	// Parse product ID
	productIDStr := c.Params("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Parse request body
	var req ReorderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Convert string IDs to UUIDs
	var imageIDs []uuid.UUID
	for _, idStr := range req.ImageIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
				Success: false,
				Message: "Invalid image ID format",
				Error:   err.Error(),
			})
		}
		imageIDs = append(imageIDs, id)
	}

	// Reorder images
	if err := h.productService.ReorderProductImages(productID, imageIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to reorder images",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: "Images reordered successfully",
	})
}

// DeleteProductImage godoc
// @Summary Delete a product image
// @Description Delete an image associated with a product. This will remove both the database record and the physical file.
// @Tags product-images
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param imageId path string true "Image ID"
// @Success 200 {object} responses.SuccessResponse{data=services.ProductImageResult} "Image deleted successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid ID format"
// @Failure 404 {object} responses.ErrorResponse "Product or image not found"
// @Failure 500 {object} responses.ErrorResponse "Server error"
// @Router /api/products/{id}/images/{imageId} [delete]
// @Security ApiKeyAuth
func (h *ProductHandler) DeleteProductImage(c *fiber.Ctx) error {
	// Parse product ID
	productIDStr := c.Params("id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Parse image ID
	imageIDStr := c.Params("imageId")
	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid image ID format",
			Error:   err.Error(),
		})
	}

	// Delete the image
	result, err := h.productService.DeleteProductImage(imageID, productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to delete image",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Image deleted successfully",
		"data":    result,
	})
}

// UploadMultipleProductImages godoc
// @Summary Upload multiple images for a product
// @Description Upload multiple images for a product at once. Images are stored at /uploads/products/ path.
// @Tags product-images
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Product ID"
// @Param files formData file true "Image files (supported formats: JPG, PNG, GIF) - can upload multiple files"
// @Param primary_index formData integer false "Index of the image to set as primary (0-based, default: -1 which means don't set any as primary)"
// @Success 200 {object} responses.SuccessResponse{data=services.MultipleProductImageResult} "Returns details of all uploaded images"
// @Failure 400 {object} responses.ErrorResponse "Invalid request or file format"
// @Failure 404 {object} responses.ErrorResponse "Product not found"
// @Failure 500 {object} responses.ErrorResponse "Server error"
// @Router /api/products/{id}/images/multiple [post]
// @Security ApiKeyAuth
func (h *ProductHandler) UploadMultipleProductImages(c *fiber.Ctx) error {
	// Parse product ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid product ID format",
			Error:   err.Error(),
		})
	}

	// Get the files from the request
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to parse multipart form",
			Error:   err.Error(),
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "No files provided",
			Error:   "No files were uploaded",
		})
	}

	// Parse primary_index parameter
	primaryIndexStr := c.FormValue("primary_index", "-1")
	primaryIndex, err := strconv.Atoi(primaryIndexStr)
	if err != nil {
		primaryIndex = -1 // Default to not setting any as primary
	}

	// Validate primary index is within bounds
	if primaryIndex >= len(files) {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid primary_index",
			Error:   "primary_index is out of bounds",
		})
	}

	// Upload the images
	result, err := h.productService.UploadMultipleProductImages(id, files, primaryIndex)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to upload product images",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product images uploaded successfully",
		"data":    result,
	})
}
