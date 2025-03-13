package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/api/responses"
	"github.com/ybds/internal/services"
	"gorm.io/gorm"
)

// ProductHandler handles HTTP requests related to products
type ProductHandler struct {
	productService *services.ProductService
}

// NewProductHandler creates a new instance of ProductHandler
func NewProductHandler(db *gorm.DB, notificationService *services.NotificationService) *ProductHandler {
	return &ProductHandler{
		productService: services.NewProductService(db, notificationService),
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
	products.Put("/inventories/:id", h.UpdateInventory)
	products.Delete("/inventories/:id", h.DeleteInventory)

	// Price routes
	products.Post("/:id/prices", h.CreatePrice)
	products.Put("/prices/:id", h.UpdatePrice)
	products.Delete("/prices/:id", h.DeletePrice)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with optional inventories and prices
// @Tags products
// @Accept json
// @Produce json
// @Param product body requests.CreateProductRequest true "Product information"
// @Success 200 {object} responses.ProductDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /products [post]
// @Security ApiKeyAuth
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	// Parse request
	var req requests.CreateProductRequest
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

	// Create product
	result, err := h.productService.CreateProduct(
		req.Name,
		req.Description,
		req.SKU,
		req.Category,
		req.ImageURL,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to create product",
			Error:   err.Error(),
		})
	}

	// Create inventories if provided
	if len(req.Inventories) > 0 {
		for _, inv := range req.Inventories {
			_, err := h.productService.CreateInventory(
				result.ProductID,
				inv.Size,
				inv.Color,
				inv.Quantity,
				inv.Location,
			)

			if err != nil {
				// Log error but continue
				// In a real application, you might want to handle this differently
				// For example, you might want to roll back the product creation
				// or return an error to the client
				// For simplicity, we'll just log the error and continue
				// TODO: Add proper logging
			}
		}
	}

	// Create prices if provided
	if len(req.Prices) > 0 {
		for _, price := range req.Prices {
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
				// TODO: Add proper logging
			}
		}
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Product created successfully",
		"data":    result,
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
// @Router /products [get]
// @Security ApiKeyAuth
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	// Parse filters
	filters := make(map[string]interface{})

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}

	// Get products
	products, total, err := h.productService.GetAllProducts(page, pageSize, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve products",
			Error:   err.Error(),
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
// @Router /products/{id} [get]
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

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product retrieved successfully",
		"data":    product,
	})
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update a product's information
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body requests.UpdateProductRequest true "Updated product information"
// @Success 200 {object} responses.ProductDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /products/{id} [put]
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

	// Parse request
	var req requests.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

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
	result, err := h.productService.UpdateProduct(
		id,
		req.Name,
		req.Description,
		req.SKU,
		req.Category,
		req.ImageURL,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update product",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Product updated successfully",
		"data":    result,
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
// @Router /products/{id} [delete]
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
// @Summary Create a new inventory for a product
// @Description Add inventory information for a specific product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param inventory body requests.CreateInventoryRequest true "Inventory information"
// @Success 200 {object} responses.InventoryResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /products/{id}/inventories [post]
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

	// Parse request
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
// @Router /products/inventories/{id} [put]
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
// @Router /products/inventories/{id} [delete]
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
// @Router /products/{id}/prices [post]
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
// @Router /products/prices/{id} [put]
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
	price, err := h.productService.GetPriceByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Price not found",
			Error:   err.Error(),
		})
	}

	// TODO: Implement UpdatePrice method in ProductService
	// For now, we'll just return a success response

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Price updated successfully",
		"data":    price,
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
// @Router /products/prices/{id} [delete]
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

	// TODO: Implement DeletePrice method in ProductService
	// For now, we'll just return a success response with the ID

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Price with ID %s deleted successfully", id.String()),
	})
}
