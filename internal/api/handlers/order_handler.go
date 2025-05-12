package handlers

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ybds/internal/api/requests"
	"github.com/ybds/internal/api/responses"
	"github.com/ybds/internal/models/order"
	"github.com/ybds/internal/services"
	"gorm.io/gorm"
)

// OrderHandler handles HTTP requests related to orders
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler creates a new instance of OrderHandler
func NewOrderHandler(db *gorm.DB, productService *services.ProductService, userService *services.UserService, notificationService *services.NotificationService) *OrderHandler {
	return &OrderHandler{
		orderService: services.NewOrderService(db, productService, userService, notificationService),
	}
}

// RegisterRoutes registers all routes related to orders
func (h *OrderHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	orders := router.Group("/orders")
	orders.Use(authMiddleware)

	// Order routes - accessible by admin or agent
	orders.Post("/", h.CreateOrder)
	orders.Get("/", h.GetOrders)
	orders.Get("/:id", h.GetOrderByID)
	orders.Get("/tracking/:number", h.GetOrderByTrackingNumber)
	orders.Get("/phone/:phone", h.GetOrdersByPhoneNumber)
	orders.Put("/:id/details", h.UpdateOrderDetails)
	orders.Put("/:id/shipment", h.UpdateShipment)
	orders.Put("/:id/status", h.UpdateOrderStatus)
	orders.Delete("/:id", h.DeleteOrder)
	orders.Get("/:id/debug", h.DebugOrder) // Debug endpoint

	// Order item routes - accessible by admin or agent
	orders.Post("/:id/items", h.AddOrderItem)
	orders.Put("/items/:id", h.UpdateOrderItem)
	orders.Delete("/items/:id", h.DeleteOrderItem)

	// Admin-only routes can be added here if needed
	// If we need to separate admin-only routes, we can modify this method to accept an adminRouter parameter
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with items and optional shipment information. Only customer_name and items are required, all other fields are optional. Customer phone number must be a valid Vietnamese number.
// @Tags orders
// @Accept json
// @Produce json
// @Param order body requests.CreateOrderRequest true "Order details"
// @Success 201 {object} responses.OrderResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders [post]
// @Security ApiKeyAuth
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	// Parse request
	var req requests.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		errorMessage := err.Error()

		// Add more context for phone validation errors
		if errorMessage == "invalid Vietnamese phone number format" {
			errorMessage = "Phone number must be a valid Vietnamese mobile or landline number (e.g., 0912345678, 0281234567)"
		}

		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   errorMessage,
		})
	}

	// Set default values for optional fields
	paymentMethod := order.PaymentMethod("cash")
	if req.PaymentMethod != "" {
		paymentMethod = order.PaymentMethod(req.PaymentMethod)
	}

	// Convert request items to service items
	items := make([]services.OrderItemInfo, len(req.Items))
	for i, item := range req.Items {
		items[i] = services.OrderItemInfo{
			InventoryID: item.InventoryID,
			Quantity:    item.Quantity,
		}
	}

	// Create order
	result, err := h.orderService.CreateOrder(
		paymentMethod,
		items,
		req.DiscountAmount,
		req.DiscountReason,
		&userID, // CreatedBy (staff member)
		req.ShippingAddress,
		req.ShippingWard,
		req.ShippingDistrict,
		req.ShippingCity,
		req.ShippingCountry,
		req.CustomerName,
		req.CustomerEmail,
		req.CustomerPhone,
		req.Notes,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to create order",
			Error:   err.Error(),
		})
	}

	// Create shipment if tracking number or carrier is provided
	if req.ShipmentTrackingNumber != "" || req.ShipmentCarrier != "" {
		err = h.orderService.UpdateShipment(result.OrderID, req.ShipmentTrackingNumber, req.ShipmentCarrier)
		if err != nil {
			// Log error but continue, as the order was created successfully
			log.Printf("Failed to create shipment for order %s: %v", result.OrderID, err)
		}
	}

	// Get the complete order to return in the response
	createdOrder, err := h.orderService.GetOrderByID(result.OrderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order created but failed to retrieve complete details",
			Error:   err.Error(),
		})
	}

	// Get creator information if available
	var creatorName string
	if createdOrder.CreatedBy != nil {
		// Get user information
		user, err := h.orderService.UserService.GetUserByID(*createdOrder.CreatedBy)
		if err == nil {
			creatorName = user.Username
		}
	}

	// Convert items to response format
	responseItems := make([]responses.OrderItemResponse, len(createdOrder.Items))
	for i, item := range createdOrder.Items {
		responseItems[i] = responses.OrderItemResponse{
			ID:          item.ID,
			OrderID:     item.OrderID,
			InventoryID: item.InventoryID,
			Quantity:    item.Quantity,
			Price:       item.PriceAtOrder,
			Subtotal:    item.PriceAtOrder * float64(item.Quantity),
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}

		// Get inventory details if available
		inventory, err := h.orderService.ProductService.GetInventoryByID(item.InventoryID)
		if err == nil && inventory != nil {
			// Add inventory details
			responseItems[i].Size = inventory.Size
			responseItems[i].Color = inventory.Color

			// Get product details if available
			product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
			if err == nil && product != nil {
				responseItems[i].ProductID = product.ID
				responseItems[i].ProductName = product.Name

				// Get price details if available
				price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
				if err == nil && price != nil {
					responseItems[i].PriceID = price.ID
					responseItems[i].Currency = price.Currency
				}
			}
		}
	}

	// Create shipment response if available
	var shipmentResponse *responses.ShipmentResponse
	if createdOrder.Shipment != nil {
		shipmentResponse = &responses.ShipmentResponse{
			ID:             createdOrder.Shipment.ID,
			OrderID:        createdOrder.Shipment.OrderID,
			TrackingNumber: createdOrder.Shipment.TrackingNumber,
			Carrier:        createdOrder.Shipment.Carrier,
			CreatedAt:      createdOrder.Shipment.CreatedAt,
			UpdatedAt:      createdOrder.Shipment.UpdatedAt,
		}
	}

	// Return response with complete order information
	return c.Status(fiber.StatusCreated).JSON(responses.OrderResponse{
		Success: true,
		Message: "Order created successfully",
		Data: responses.OrderDetail{
			ID:               createdOrder.ID,
			CustomerName:     createdOrder.CustomerName,
			CustomerEmail:    createdOrder.CustomerEmail,
			CustomerPhone:    createdOrder.CustomerPhone,
			ShippingAddress:  createdOrder.ShippingAddress,
			ShippingWard:     createdOrder.ShippingWard,
			ShippingDistrict: createdOrder.ShippingDistrict,
			ShippingCity:     createdOrder.ShippingCity,
			ShippingCountry:  createdOrder.ShippingCountry,
			PaymentMethod:    string(createdOrder.PaymentMethod),
			Status:           string(createdOrder.OrderStatus),
			Notes:            createdOrder.Notes,
			Total:            createdOrder.TotalAmount,
			DiscountAmount:   createdOrder.DiscountAmount,
			DiscountReason:   createdOrder.DiscountReason,
			FinalTotal:       createdOrder.FinalTotalAmount,
			CreatedBy:        *createdOrder.CreatedBy,
			CreatedByName:    creatorName,
			Items:            responseItems,
			Shipment:         shipmentResponse,
			CreatedAt:        createdOrder.CreatedAt,
			UpdatedAt:        createdOrder.UpdatedAt,
		},
	})
}

// GetOrders godoc
// @Summary Get all orders
// @Description Get a list of all orders with pagination, filtering and search
// @Tags orders
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Filter by status"
// @Param created_by query string false "Filter by creator ID"
// @Param from_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param to_date query string false "Filter by end date (YYYY-MM-DD)"
// @Param phone_number query string false "Filter by customer phone number"
// @Param search query string false "Search term"
// @Success 200 {object} responses.OrdersResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders [get]
// @Security ApiKeyAuth
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("page_size", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Initialize filters
	filters := make(map[string]interface{})

	// Apply status filter if provided
	if status := c.Query("status"); status != "" {
		filters["order_status"] = status
	}

	// Apply creator ID filter if provided
	if createdBy := c.Query("created_by"); createdBy != "" {
		createdByID, err := uuid.Parse(createdBy)
		if err == nil {
			filters["created_by"] = createdByID
		}
	}

	// Apply from_date filter if provided
	if fromDate := c.Query("from_date"); fromDate != "" {
		// Parse date in format YYYY-MM-DD
		date, err := time.Parse("2006-01-02", fromDate)
		if err == nil {
			// Set time to beginning of day
			date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			filters["from_date"] = date
		}
	}

	// Apply to_date filter if provided
	if toDate := c.Query("to_date"); toDate != "" {
		// Parse date in format YYYY-MM-DD
		date, err := time.Parse("2006-01-02", toDate)
		if err == nil {
			// Set time to end of day
			date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
			filters["to_date"] = date
		}
	}

	// Apply phone number filter if provided
	if phoneNumber := c.Query("phone_number"); phoneNumber != "" {
		filters["phone_number"] = phoneNumber
	}

	// Get orders with filters
	orders, total, err := h.orderService.GetAllOrders(page, pageSize, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to get orders",
			Error:   err.Error(),
		})
	}

	// Format response data
	var orderList []responses.OrderDetail

	// Create user info cache to avoid repeated database queries
	userCache := make(map[uuid.UUID]string)

	for _, o := range orders {
		// Get creator name if available
		var creatorName string
		if o.CreatedBy != nil {
			// Check if we already have this user in cache
			if name, found := userCache[*o.CreatedBy]; found {
				creatorName = name
			} else {
				// If not in cache, fetch from database
				user, err := h.orderService.UserService.GetUserByID(*o.CreatedBy)
				if err == nil {
					creatorName = user.Username
					// Add to cache
					userCache[*o.CreatedBy] = creatorName
				}
			}
		}

		// Create basic order detail
		orderDetail := responses.OrderDetail{
			ID:               o.ID,
			CustomerName:     o.CustomerName,
			CustomerEmail:    o.CustomerEmail,
			CustomerPhone:    o.CustomerPhone,
			ShippingAddress:  o.ShippingAddress,
			ShippingWard:     o.ShippingWard,
			ShippingDistrict: o.ShippingDistrict,
			ShippingCity:     o.ShippingCity,
			ShippingCountry:  o.ShippingCountry,
			PaymentMethod:    string(o.PaymentMethod),
			Status:           string(o.OrderStatus),
			Notes:            o.Notes,
			Total:            o.TotalAmount,
			DiscountAmount:   o.DiscountAmount,
			DiscountReason:   o.DiscountReason,
			FinalTotal:       o.FinalTotalAmount,
			CreatedAt:        o.CreatedAt,
			UpdatedAt:        o.UpdatedAt,
		}

		// Set created by if available
		if o.CreatedBy != nil {
			orderDetail.CreatedBy = *o.CreatedBy
			orderDetail.CreatedByName = creatorName
		}

		// Add shipment info if available
		if o.Shipment != nil {
			orderDetail.Shipment = &responses.ShipmentResponse{
				ID:             o.Shipment.ID,
				OrderID:        o.Shipment.OrderID,
				TrackingNumber: o.Shipment.TrackingNumber,
				Carrier:        o.Shipment.Carrier,
				CreatedAt:      o.Shipment.CreatedAt,
				UpdatedAt:      o.Shipment.UpdatedAt,
			}
		}

		// Add items if available
		items := make([]responses.OrderItemResponse, len(o.Items))
		for i, item := range o.Items {
			// Create basic item
			items[i] = responses.OrderItemResponse{
				ID:          item.ID,
				OrderID:     item.OrderID,
				InventoryID: item.InventoryID,
				Quantity:    item.Quantity,
				Price:       item.PriceAtOrder,
				Subtotal:    item.PriceAtOrder * float64(item.Quantity),
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
			}

			// Get inventory details if needed
			inventory, err := h.orderService.ProductService.GetInventoryByID(item.InventoryID)
			if err == nil && inventory != nil {
				// Add inventory details
				items[i].Size = inventory.Size
				items[i].Color = inventory.Color

				// Get product details if available
				product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
				if err == nil && product != nil {
					items[i].ProductID = product.ID
					items[i].ProductName = product.Name

					// Get price details if available
					price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
					if err == nil && price != nil {
						items[i].PriceID = price.ID
						items[i].Currency = price.Currency
					}
				}
			}
		}

		orderDetail.Items = items
		orderList = append(orderList, orderDetail)
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrdersResponse{
		Success:    true,
		Message:    "Orders retrieved successfully",
		Data:       orderList,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int64(math.Ceil(float64(total) / float64(pageSize))),
	})
}

// GetOrderByID godoc
// @Summary Get an order by ID
// @Description Get a specific order with all its items and details
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} responses.OrderDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/{id} [get]
// @Security ApiKeyAuth
func (h *OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	// Parse order ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order ID format",
			Error:   err.Error(),
		})
	}

	// Get order
	o, err := h.orderService.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order not found",
			Error:   err.Error(),
		})
	}

	// Convert items to response format
	items := make([]responses.OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = responses.OrderItemResponse{
			ID:          item.ID,
			OrderID:     item.OrderID,
			InventoryID: item.InventoryID,
			Quantity:    item.Quantity,
			Price:       item.PriceAtOrder,
			Subtotal:    item.PriceAtOrder * float64(item.Quantity),
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			// Other fields would need to be fetched from related services
			ProductID:    uuid.Nil, // Will be set below if product is found
			ProductName:  "",       // Will be set below if product is found
			ProductImage: "",       // Will be set below if product is found
			Size:         "",       // Will be set below if inventory is found
			Color:        "",       // Will be set below if inventory is found
			PriceID:      uuid.Nil, // Will be set below if price is found
			Currency:     "",       // Will be set below if price is found
			Notes:        "",       // Not in the model, would need to add
		}

		// Get inventory details if available
		inventory, err := h.orderService.ProductService.GetInventoryByID(item.InventoryID)
		if err == nil && inventory != nil {
			// Add inventory details
			items[i].Size = inventory.Size
			items[i].Color = inventory.Color

			// Get product details if available
			product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
			if err == nil && product != nil {
				items[i].ProductID = product.ID
				items[i].ProductName = product.Name
				items[i].ProductImage = h.orderService.ProductService.GetPrimaryImageURL(product.ID)

				// Get price details if available
				price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
				if err == nil && price != nil {
					items[i].PriceID = price.ID
					items[i].Currency = price.Currency
				}
			}
		}
	}

	// Get creator information if available
	var creatorName string
	if o.CreatedBy != nil {
		// Get user information
		user, err := h.orderService.UserService.GetUserByID(*o.CreatedBy)
		if err == nil {
			creatorName = user.Username
		}
	}

	// Create shipment response if available
	var shipmentResponse *responses.ShipmentResponse
	if o.Shipment != nil {
		shipmentResponse = &responses.ShipmentResponse{
			ID:             o.Shipment.ID,
			OrderID:        o.Shipment.OrderID,
			TrackingNumber: o.Shipment.TrackingNumber,
			Carrier:        o.Shipment.Carrier,
			CreatedAt:      o.Shipment.CreatedAt,
			UpdatedAt:      o.Shipment.UpdatedAt,
		}
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrderDetailResponse{
		Success: true,
		Message: "Order retrieved successfully",
		Data: responses.OrderDetail{
			ID:               o.ID,
			CustomerName:     o.CustomerName,
			CustomerEmail:    o.CustomerEmail,
			CustomerPhone:    o.CustomerPhone,
			ShippingAddress:  o.ShippingAddress,
			ShippingWard:     o.ShippingWard,
			ShippingDistrict: o.ShippingDistrict,
			ShippingCity:     o.ShippingCity,
			ShippingCountry:  o.ShippingCountry,
			PaymentMethod:    string(o.PaymentMethod),
			Status:           string(o.OrderStatus),
			Notes:            o.Notes,
			Total:            o.TotalAmount,
			DiscountAmount:   o.DiscountAmount,
			DiscountReason:   o.DiscountReason,
			FinalTotal:       o.FinalTotalAmount,
			CreatedBy:        *o.CreatedBy,
			CreatedByName:    creatorName,
			Items:            items,
			Shipment:         shipmentResponse,
			CreatedAt:        o.CreatedAt,
			UpdatedAt:        o.UpdatedAt,
		},
	})
}

// UpdateOrderStatus godoc
// @Summary Update an order's status
// @Description Update the status of an order. Admins can change to any status. Agents can only change orders with status 'pending_confirmation', 'confirmed', or 'shipment_requested'.
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body requests.UpdateOrderStatusRequest true "Order status"
// @Success 200 {object} responses.OrderResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/{id}/status [put]
// @Security ApiKeyAuth
func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	// Get user roles from context (set by auth middleware)
	_, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	userRoles, ok := c.Locals("roles").([]string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user roles",
		})
	}

	// Parse order ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order ID format",
			Error:   err.Error(),
		})
	}

	// Get current order to check status
	currentOrder, err := h.orderService.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order not found",
			Error:   err.Error(),
		})
	}

	// Check if user is an admin
	isAdmin := false
	for _, role := range userRoles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	// For non-admin users, check if the order is in an allowed status for agents
	if !isAdmin {
		currentStatus := string(currentOrder.OrderStatus)
		// Check if current status is one of the allowed statuses for agents
		allowedStatuses := []string{"pending_confirmation", "confirmed", "shipment_requested"}
		isAllowed := false
		for _, status := range allowedStatuses {
			if currentStatus == status {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(responses.ErrorResponse{
				Success: false,
				Message: "Permission denied",
				Error:   "Agents can only update orders with status: pending_confirmation, confirmed, or shipment_requested",
			})
		}
	}

	// Parse request
	var req requests.UpdateOrderStatusRequest
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

	// Update order status
	_, err = h.orderService.UpdateOrderStatus(id, order.OrderStatus(req.Status))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update order status",
			Error:   err.Error(),
		})
	}

	// Get the updated order to return complete information
	updatedOrder, err := h.orderService.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve updated order",
			Error:   err.Error(),
		})
	}

	// Get creator information if available
	var creatorName string
	if updatedOrder.CreatedBy != nil {
		// Get user information
		user, err := h.orderService.UserService.GetUserByID(*updatedOrder.CreatedBy)
		if err == nil {
			creatorName = user.Username
		}
	}

	// Convert items to response format
	items := make([]responses.OrderItemResponse, len(updatedOrder.Items))
	for i, item := range updatedOrder.Items {
		items[i] = responses.OrderItemResponse{
			ID:          item.ID,
			OrderID:     item.OrderID,
			InventoryID: item.InventoryID,
			Quantity:    item.Quantity,
			Price:       item.PriceAtOrder,
			Subtotal:    item.PriceAtOrder * float64(item.Quantity),
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}

		// Get inventory details if available
		inventory, err := h.orderService.ProductService.GetInventoryByID(item.InventoryID)
		if err == nil && inventory != nil {
			// Add inventory details
			items[i].Size = inventory.Size
			items[i].Color = inventory.Color

			// Get product details if available
			product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
			if err == nil && product != nil {
				items[i].ProductID = product.ID
				items[i].ProductName = product.Name

				// Get price details if available
				price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
				if err == nil && price != nil {
					items[i].PriceID = price.ID
					items[i].Currency = price.Currency
				}
			}
		}
	}

	// Create shipment response if available
	var shipmentResponse *responses.ShipmentResponse
	if updatedOrder.Shipment != nil {
		shipmentResponse = &responses.ShipmentResponse{
			ID:             updatedOrder.Shipment.ID,
			OrderID:        updatedOrder.Shipment.OrderID,
			TrackingNumber: updatedOrder.Shipment.TrackingNumber,
			Carrier:        updatedOrder.Shipment.Carrier,
			CreatedAt:      updatedOrder.Shipment.CreatedAt,
			UpdatedAt:      updatedOrder.Shipment.UpdatedAt,
		}
	}

	// Return response with complete order information
	return c.Status(fiber.StatusOK).JSON(responses.OrderResponse{
		Success: true,
		Message: "Order status updated successfully",
		Data: responses.OrderDetail{
			ID:               updatedOrder.ID,
			CustomerName:     updatedOrder.CustomerName,
			CustomerEmail:    updatedOrder.CustomerEmail,
			CustomerPhone:    updatedOrder.CustomerPhone,
			ShippingAddress:  updatedOrder.ShippingAddress,
			ShippingWard:     updatedOrder.ShippingWard,
			ShippingDistrict: updatedOrder.ShippingDistrict,
			ShippingCity:     updatedOrder.ShippingCity,
			ShippingCountry:  updatedOrder.ShippingCountry,
			PaymentMethod:    string(updatedOrder.PaymentMethod),
			Status:           string(updatedOrder.OrderStatus),
			Notes:            updatedOrder.Notes,
			Total:            updatedOrder.TotalAmount,
			DiscountAmount:   updatedOrder.DiscountAmount,
			DiscountReason:   updatedOrder.DiscountReason,
			FinalTotal:       updatedOrder.FinalTotalAmount,
			CreatedBy:        *updatedOrder.CreatedBy,
			CreatedByName:    creatorName,
			Items:            items,
			Shipment:         shipmentResponse,
			CreatedAt:        updatedOrder.CreatedAt,
			UpdatedAt:        updatedOrder.UpdatedAt,
		},
	})
}

// DeleteOrder godoc
// @Summary Delete an order
// @Description Delete an order and all its items
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/{id} [delete]
// @Security ApiKeyAuth
func (h *OrderHandler) DeleteOrder(c *fiber.Ctx) error {
	// Parse order ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order ID format",
			Error:   err.Error(),
		})
	}

	// Delete order
	result, err := h.orderService.DeleteOrder(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to delete order",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: result.Message,
	})
}

// AddOrderItem godoc
// @Summary Add an item to an order
// @Description Add a new item to an existing order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param item body requests.AddOrderItemRequest true "Order item details"
// @Success 201 {object} responses.OrderItemDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/{id}/items [post]
// @Security ApiKeyAuth
func (h *OrderHandler) AddOrderItem(c *fiber.Ctx) error {
	// Parse order ID
	idStr := c.Params("id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order ID format",
			Error:   err.Error(),
		})
	}

	// Parse request
	var req requests.AddOrderItemRequest
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

	// Add order item
	err = h.orderService.AddOrderItem(orderID, req.InventoryID, req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to add order item",
			Error:   err.Error(),
		})
	}

	// Get the newly added order item
	// Since we don't have the ID of the newly created item, we need to get all items for the order
	// and find the one with the matching inventory ID
	items, err := h.orderService.OrderRepo.GetOrderItemsByOrderID(orderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve added order item",
			Error:   err.Error(),
		})
	}

	// Find the newly added item (should be the last one with matching inventory ID)
	var newItem *order.OrderItem
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].InventoryID == req.InventoryID {
			newItem = &items[i]
			break
		}
	}

	if newItem == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve added order item",
			Error:   "Item not found after adding",
		})
	}

	// Get inventory details
	inventory, err := h.orderService.ProductService.GetInventoryByID(newItem.InventoryID)
	if err != nil {
		// Log the error but continue
		// We can still return the order item without the product details
	}

	// Create response with actual data
	response := responses.OrderItemResponse{
		ID:          newItem.ID,
		OrderID:     newItem.OrderID,
		InventoryID: newItem.InventoryID,
		Quantity:    newItem.Quantity,
		Price:       newItem.PriceAtOrder,
		Subtotal:    newItem.PriceAtOrder * float64(newItem.Quantity),
		CreatedAt:   newItem.CreatedAt,
		UpdatedAt:   newItem.UpdatedAt,
	}

	// Add product details if available
	if inventory != nil {
		product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
		if err == nil && product != nil {
			response.ProductID = product.ID
			response.ProductName = product.Name

			// Get current price information
			price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
			if err == nil && price != nil {
				response.PriceID = price.ID
				response.Currency = price.Currency
			}
		}

		// Add inventory details
		response.Size = inventory.Size
		response.Color = inventory.Color
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(responses.OrderItemDetailResponse{
		Success: true,
		Message: "Order item added successfully",
		Data:    response,
	})
}

// UpdateOrderItem godoc
// @Summary Update an order item
// @Description Update the quantity of an order item. Admins can update any order item. Agents can only update items if the order status is 'pending_confirmation', 'confirmed', or 'shipment_requested'.
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order Item ID"
// @Param item body requests.UpdateOrderItemRequest true "Order item details"
// @Success 200 {object} responses.OrderItemDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/items/{id} [put]
// @Security ApiKeyAuth
func (h *OrderHandler) UpdateOrderItem(c *fiber.Ctx) error {
	// Get user roles from context (set by auth middleware)
	_, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	userRoles, ok := c.Locals("roles").([]string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user roles",
		})
	}

	// Parse order item ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order item ID format",
			Error:   err.Error(),
		})
	}

	// Get the order item to get the order ID
	orderItem, err := h.orderService.OrderRepo.GetOrderItemByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order item not found",
			Error:   err.Error(),
		})
	}

	// Get the order to check its status
	order, err := h.orderService.GetOrderByID(orderItem.OrderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order not found",
			Error:   err.Error(),
		})
	}

	// Check if user is an admin
	isAdmin := false
	for _, role := range userRoles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	// For non-admin users, check if the order is in an allowed status for agents
	if !isAdmin {
		currentStatus := string(order.OrderStatus)
		// Check if current status is one of the allowed statuses for agents
		allowedStatuses := []string{"pending_confirmation", "confirmed", "shipment_requested"}
		isAllowed := false
		for _, status := range allowedStatuses {
			if currentStatus == status {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(responses.ErrorResponse{
				Success: false,
				Message: "Permission denied",
				Error:   "Agents can only update items for orders with status: pending_confirmation, confirmed, or shipment_requested",
			})
		}
	}

	// Parse request
	var req requests.UpdateOrderItemRequest
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

	// Update order item
	err = h.orderService.UpdateOrderItem(id, req.Quantity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update order item",
			Error:   err.Error(),
		})
	}

	// Get the updated order item from the database
	updatedItem, err := h.orderService.OrderRepo.GetOrderItemByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve updated order item",
			Error:   err.Error(),
		})
	}

	// Get inventory details if needed
	inventory, err := h.orderService.ProductService.GetInventoryByID(updatedItem.InventoryID)
	if err != nil {
		// Log the error but continue
		// We can still return the order item without the product details
	}

	// Create response with actual data
	response := responses.OrderItemResponse{
		ID:          updatedItem.ID,
		OrderID:     updatedItem.OrderID,
		InventoryID: updatedItem.InventoryID,
		Quantity:    updatedItem.Quantity,
		Price:       updatedItem.PriceAtOrder,
		Subtotal:    updatedItem.PriceAtOrder * float64(updatedItem.Quantity),
		CreatedAt:   updatedItem.CreatedAt,
		UpdatedAt:   updatedItem.UpdatedAt,
	}

	// Add product details if available
	if inventory != nil {
		product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
		if err == nil && product != nil {
			response.ProductID = product.ID
			response.ProductName = product.Name

			// Get current price information
			price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
			if err == nil && price != nil {
				response.PriceID = price.ID
				response.Currency = price.Currency
			}
		}

		// Add inventory details
		response.Size = inventory.Size
		response.Color = inventory.Color
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrderItemDetailResponse{
		Success: true,
		Message: "Order item updated successfully",
		Data:    response,
	})
}

// DeleteOrderItem godoc
// @Summary Delete an order item
// @Description Delete an existing order item
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order Item ID"
// @Success 200 {object} responses.SuccessResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/items/{id} [delete]
// @Security ApiKeyAuth
func (h *OrderHandler) DeleteOrderItem(c *fiber.Ctx) error {
	// Parse order item ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order item ID format",
			Error:   err.Error(),
		})
	}

	// Delete order item
	err = h.orderService.DeleteOrderItem(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to delete order item",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: "Order item deleted successfully",
	})
}

// UpdateOrderDetails godoc
// @Summary Update order details
// @Description Update the details of an order including payment details, shipping address, and customer information. Customer phone number must be a valid Vietnamese number. Admins can update any order. Agents can only update orders with status 'pending_confirmation', 'confirmed', or 'shipment_requested'.
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param details body requests.UpdateOrderDetailsRequest true "Order details"
// @Success 200 {object} responses.OrderResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/{id}/details [put]
// @Security ApiKeyAuth
func (h *OrderHandler) UpdateOrderDetails(c *fiber.Ctx) error {
	// Get user roles from context (set by auth middleware)
	_, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	userRoles, ok := c.Locals("roles").([]string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user roles",
		})
	}

	// Parse order ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order ID format",
			Error:   err.Error(),
		})
	}

	// Get current order to check status
	currentOrder, err := h.orderService.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order not found",
			Error:   err.Error(),
		})
	}

	// Check if user is an admin
	isAdmin := false
	for _, role := range userRoles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	// For non-admin users, check if the order is in an allowed status for agents
	if !isAdmin {
		currentStatus := string(currentOrder.OrderStatus)
		// Check if current status is one of the allowed statuses for agents
		allowedStatuses := []string{"pending_confirmation", "confirmed", "shipment_requested"}
		isAllowed := false
		for _, status := range allowedStatuses {
			if currentStatus == status {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(responses.ErrorResponse{
				Success: false,
				Message: "Permission denied",
				Error:   "Agents can only update orders with status: pending_confirmation, confirmed, or shipment_requested",
			})
		}
	}

	// Parse request
	var req requests.UpdateOrderDetailsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		errorMessage := err.Error()

		// Add more context for phone validation errors
		if errorMessage == "invalid Vietnamese phone number format" {
			errorMessage = "Phone number must be a valid Vietnamese mobile or landline number (e.g., 0912345678, 0281234567)"
		}

		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   errorMessage,
		})
	}

	// Convert payment method
	var paymentMethod order.PaymentMethod
	if req.PaymentMethod != "" {
		paymentMethod = order.PaymentMethod(req.PaymentMethod)
	}

	// Update order details
	_, err = h.orderService.UpdateOrderDetails(
		id,
		req.Notes,
		paymentMethod,
		req.DiscountAmount,
		req.DiscountReason,
		req.ShippingAddress,
		req.ShippingWard,
		req.ShippingDistrict,
		req.ShippingCity,
		req.ShippingCountry,
		req.CustomerName,
		req.CustomerEmail,
		req.CustomerPhone,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update order details",
			Error:   err.Error(),
		})
	}

	// Get the updated order to return complete information
	updatedOrder, err := h.orderService.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve updated order",
			Error:   err.Error(),
		})
	}

	// Get creator information if available
	var creatorName string
	if updatedOrder.CreatedBy != nil {
		// Get user information
		user, err := h.orderService.UserService.GetUserByID(*updatedOrder.CreatedBy)
		if err == nil {
			creatorName = user.Username
		}
	}

	// Convert items to response format
	items := make([]responses.OrderItemResponse, len(updatedOrder.Items))
	for i, item := range updatedOrder.Items {
		items[i] = responses.OrderItemResponse{
			ID:          item.ID,
			OrderID:     item.OrderID,
			InventoryID: item.InventoryID,
			Quantity:    item.Quantity,
			Price:       item.PriceAtOrder,
			Subtotal:    item.PriceAtOrder * float64(item.Quantity),
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}

		// Get inventory details if available
		inventory, err := h.orderService.ProductService.GetInventoryByID(item.InventoryID)
		if err == nil && inventory != nil {
			// Add inventory details
			items[i].Size = inventory.Size
			items[i].Color = inventory.Color

			// Get product details if available
			product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
			if err == nil && product != nil {
				items[i].ProductID = product.ID
				items[i].ProductName = product.Name

				// Get price details if available
				price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
				if err == nil && price != nil {
					items[i].PriceID = price.ID
					items[i].Currency = price.Currency
				}
			}
		}
	}

	// Create shipment response if available
	var shipmentResponse *responses.ShipmentResponse
	if updatedOrder.Shipment != nil {
		shipmentResponse = &responses.ShipmentResponse{
			ID:             updatedOrder.Shipment.ID,
			OrderID:        updatedOrder.Shipment.OrderID,
			TrackingNumber: updatedOrder.Shipment.TrackingNumber,
			Carrier:        updatedOrder.Shipment.Carrier,
			CreatedAt:      updatedOrder.Shipment.CreatedAt,
			UpdatedAt:      updatedOrder.Shipment.UpdatedAt,
		}
	}

	// Return response with complete order information
	return c.Status(fiber.StatusOK).JSON(responses.OrderResponse{
		Success: true,
		Message: "Order details updated successfully",
		Data: responses.OrderDetail{
			ID:               updatedOrder.ID,
			CustomerName:     updatedOrder.CustomerName,
			CustomerEmail:    updatedOrder.CustomerEmail,
			CustomerPhone:    updatedOrder.CustomerPhone,
			ShippingAddress:  updatedOrder.ShippingAddress,
			ShippingWard:     updatedOrder.ShippingWard,
			ShippingDistrict: updatedOrder.ShippingDistrict,
			ShippingCity:     updatedOrder.ShippingCity,
			ShippingCountry:  updatedOrder.ShippingCountry,
			PaymentMethod:    string(updatedOrder.PaymentMethod),
			Status:           string(updatedOrder.OrderStatus),
			Notes:            updatedOrder.Notes,
			Total:            updatedOrder.TotalAmount,
			DiscountAmount:   updatedOrder.DiscountAmount,
			DiscountReason:   updatedOrder.DiscountReason,
			FinalTotal:       updatedOrder.FinalTotalAmount,
			CreatedBy:        *updatedOrder.CreatedBy,
			CreatedByName:    creatorName,
			Items:            items,
			Shipment:         shipmentResponse,
			CreatedAt:        updatedOrder.CreatedAt,
			UpdatedAt:        updatedOrder.UpdatedAt,
		},
	})
}

// UpdateShipment godoc
// @Summary Update shipment details
// @Description Update the shipment details of an order. Admins can update any order's shipment. Agents can only update shipments for orders with status 'pending_confirmation', 'confirmed', or 'shipment_requested'.
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param shipment body requests.UpdateShipmentRequest true "Shipment details"
// @Success 200 {object} responses.OrderResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/{id}/shipment [put]
// @Security ApiKeyAuth
func (h *OrderHandler) UpdateShipment(c *fiber.Ctx) error {
	// Get user roles from context (set by auth middleware)
	_, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user ID",
		})
	}

	userRoles, ok := c.Locals("roles").([]string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
			Error:   "Invalid user roles",
		})
	}

	// Parse order ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order ID format",
			Error:   err.Error(),
		})
	}

	// Get current order to check status
	currentOrder, err := h.orderService.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order not found",
			Error:   err.Error(),
		})
	}

	// Check if user is an admin
	isAdmin := false
	for _, role := range userRoles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	// For non-admin users, check if the order is in an allowed status for agents
	if !isAdmin {
		currentStatus := string(currentOrder.OrderStatus)
		// Check if current status is one of the allowed statuses for agents
		allowedStatuses := []string{"pending_confirmation", "confirmed", "shipment_requested"}
		isAllowed := false
		for _, status := range allowedStatuses {
			if currentStatus == status {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(responses.ErrorResponse{
				Success: false,
				Message: "Permission denied",
				Error:   "Agents can only update shipments for orders with status: pending_confirmation, confirmed, or shipment_requested",
			})
		}
	}

	// Parse request
	var req requests.UpdateShipmentRequest
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

	// Update shipment details
	err = h.orderService.UpdateShipment(id, req.TrackingNumber, req.Carrier)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update shipment details",
			Error:   err.Error(),
		})
	}

	// Get updated order to return in response
	updatedOrder, err := h.orderService.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve updated order",
			Error:   err.Error(),
		})
	}

	// Get creator information if available
	var creatorName string
	if updatedOrder.CreatedBy != nil {
		// Get user information
		user, err := h.orderService.UserService.GetUserByID(*updatedOrder.CreatedBy)
		if err == nil {
			creatorName = user.Username
		}
	}

	// Convert items to response format
	items := make([]responses.OrderItemResponse, len(updatedOrder.Items))
	for i, item := range updatedOrder.Items {
		items[i] = responses.OrderItemResponse{
			ID:          item.ID,
			OrderID:     item.OrderID,
			InventoryID: item.InventoryID,
			Quantity:    item.Quantity,
			Price:       item.PriceAtOrder,
			Subtotal:    item.PriceAtOrder * float64(item.Quantity),
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}

		// Get inventory details if available
		inventory, err := h.orderService.ProductService.GetInventoryByID(item.InventoryID)
		if err == nil && inventory != nil {
			// Add inventory details
			items[i].Size = inventory.Size
			items[i].Color = inventory.Color

			// Get product details if available
			product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
			if err == nil && product != nil {
				items[i].ProductID = product.ID
				items[i].ProductName = product.Name

				// Get price details if available
				price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
				if err == nil && price != nil {
					items[i].PriceID = price.ID
					items[i].Currency = price.Currency
				}
			}
		}
	}

	// Create shipment response if available
	var shipmentResponse *responses.ShipmentResponse
	if updatedOrder.Shipment != nil {
		shipmentResponse = &responses.ShipmentResponse{
			ID:             updatedOrder.Shipment.ID,
			OrderID:        updatedOrder.Shipment.OrderID,
			TrackingNumber: updatedOrder.Shipment.TrackingNumber,
			Carrier:        updatedOrder.Shipment.Carrier,
			CreatedAt:      updatedOrder.Shipment.CreatedAt,
			UpdatedAt:      updatedOrder.Shipment.UpdatedAt,
		}
	}

	// Return response with complete order information
	return c.Status(fiber.StatusOK).JSON(responses.OrderResponse{
		Success: true,
		Message: "Shipment details updated successfully",
		Data: responses.OrderDetail{
			ID:               updatedOrder.ID,
			CustomerName:     updatedOrder.CustomerName,
			CustomerEmail:    updatedOrder.CustomerEmail,
			CustomerPhone:    updatedOrder.CustomerPhone,
			ShippingAddress:  updatedOrder.ShippingAddress,
			ShippingWard:     updatedOrder.ShippingWard,
			ShippingDistrict: updatedOrder.ShippingDistrict,
			ShippingCity:     updatedOrder.ShippingCity,
			ShippingCountry:  updatedOrder.ShippingCountry,
			PaymentMethod:    string(updatedOrder.PaymentMethod),
			Status:           string(updatedOrder.OrderStatus),
			Notes:            updatedOrder.Notes,
			Total:            updatedOrder.TotalAmount,
			DiscountAmount:   updatedOrder.DiscountAmount,
			DiscountReason:   updatedOrder.DiscountReason,
			FinalTotal:       updatedOrder.FinalTotalAmount,
			CreatedBy:        *updatedOrder.CreatedBy,
			CreatedByName:    creatorName,
			Items:            items,
			Shipment:         shipmentResponse,
			CreatedAt:        updatedOrder.CreatedAt,
			UpdatedAt:        updatedOrder.UpdatedAt,
		},
	})
}

// DebugOrder is a debug endpoint to check if an order exists
func (h *OrderHandler) DebugOrder(c *fiber.Ctx) error {
	// Parse order ID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order ID format",
			Error:   err.Error(),
		})
	}

	// Get order
	o, err := h.orderService.GetOrderByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order not found",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Order found",
		"data": fiber.Map{
			"id":           o.ID,
			"created_by":   o.CreatedBy,
			"order_status": o.OrderStatus,
			"created_at":   o.CreatedAt,
		},
	})
}

// GetOrderByTrackingNumber godoc
// @Summary Get order by tracking number
// @Description Get a specific order by its shipment tracking number
// @Tags orders
// @Accept json
// @Produce json
// @Param number path string true "Tracking Number"
// @Success 200 {object} responses.OrderDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/tracking/{number} [get]
// @Security ApiKeyAuth
func (h *OrderHandler) GetOrderByTrackingNumber(c *fiber.Ctx) error {
	// Get tracking number from params
	trackingNumber := c.Params("number")
	if trackingNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Tracking number is required",
			Error:   "Missing tracking number",
		})
	}

	// Get order by tracking number
	o, err := h.orderService.GetOrderByTrackingNumber(trackingNumber)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Order not found",
			Error:   err.Error(),
		})
	}

	// Get creator information if available
	var creatorName string
	if o.CreatedBy != nil {
		// Get user information
		user, err := h.orderService.UserService.GetUserByID(*o.CreatedBy)
		if err == nil {
			creatorName = user.Username
		}
	}

	// Convert items to response format
	items := make([]responses.OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = responses.OrderItemResponse{
			ID:          item.ID,
			OrderID:     item.OrderID,
			InventoryID: item.InventoryID,
			Quantity:    item.Quantity,
			Price:       item.PriceAtOrder,
			Subtotal:    item.PriceAtOrder * float64(item.Quantity),
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}

		// Get inventory details if available
		inventory, err := h.orderService.ProductService.GetInventoryByID(item.InventoryID)
		if err == nil && inventory != nil {
			// Add inventory details
			items[i].Size = inventory.Size
			items[i].Color = inventory.Color

			// Get product details if available
			product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
			if err == nil && product != nil {
				items[i].ProductID = product.ID
				items[i].ProductName = product.Name

				// Get price details if available
				price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
				if err == nil && price != nil {
					items[i].PriceID = price.ID
					items[i].Currency = price.Currency
				}
			}
		}
	}

	// Create shipment response if available
	var shipmentResponse *responses.ShipmentResponse
	if o.Shipment != nil {
		shipmentResponse = &responses.ShipmentResponse{
			ID:             o.Shipment.ID,
			OrderID:        o.Shipment.OrderID,
			TrackingNumber: o.Shipment.TrackingNumber,
			Carrier:        o.Shipment.Carrier,
			CreatedAt:      o.Shipment.CreatedAt,
			UpdatedAt:      o.Shipment.UpdatedAt,
		}
	}

	// Return response with order details
	return c.Status(fiber.StatusOK).JSON(responses.OrderDetailResponse{
		Success: true,
		Message: "Order found",
		Data: responses.OrderDetail{
			ID:               o.ID,
			CustomerName:     o.CustomerName,
			CustomerEmail:    o.CustomerEmail,
			CustomerPhone:    o.CustomerPhone,
			ShippingAddress:  o.ShippingAddress,
			ShippingWard:     o.ShippingWard,
			ShippingDistrict: o.ShippingDistrict,
			ShippingCity:     o.ShippingCity,
			ShippingCountry:  o.ShippingCountry,
			PaymentMethod:    string(o.PaymentMethod),
			Status:           string(o.OrderStatus),
			Notes:            o.Notes,
			Total:            o.TotalAmount,
			DiscountAmount:   o.DiscountAmount,
			DiscountReason:   o.DiscountReason,
			FinalTotal:       o.FinalTotalAmount,
			CreatedBy:        *o.CreatedBy,
			CreatedByName:    creatorName,
			Items:            items,
			Shipment:         shipmentResponse,
			CreatedAt:        o.CreatedAt,
			UpdatedAt:        o.UpdatedAt,
		},
	})
}

// GetOrdersByPhoneNumber godoc
// @Summary Get orders by phone number
// @Description Get a list of all orders with a specific customer phone number
// @Tags orders
// @Accept json
// @Produce json
// @Param phone path string true "Customer phone number"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Filter by status"
// @Param from_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param to_date query string false "Filter by end date (YYYY-MM-DD)"
// @Success 200 {object} responses.OrdersResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/orders/phone/{phone} [get]
// @Security ApiKeyAuth
func (h *OrderHandler) GetOrdersByPhoneNumber(c *fiber.Ctx) error {
	// Get phone number from path parameter
	phoneNumber := c.Params("phone")
	if phoneNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Phone number is required",
			Error:   "Missing phone number",
		})
	}

	// Parse pagination parameters
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("page_size", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// Initialize additional filters
	additionalFilters := make(map[string]interface{})

	// Apply status filter if provided
	if status := c.Query("status"); status != "" {
		additionalFilters["order_status"] = status
	}

	// Apply from_date filter if provided
	if fromDate := c.Query("from_date"); fromDate != "" {
		// Parse date in format YYYY-MM-DD
		date, err := time.Parse("2006-01-02", fromDate)
		if err == nil {
			// Set time to beginning of day
			date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			additionalFilters["from_date"] = date
		}
	}

	// Apply to_date filter if provided
	if toDate := c.Query("to_date"); toDate != "" {
		// Parse date in format YYYY-MM-DD
		date, err := time.Parse("2006-01-02", toDate)
		if err == nil {
			// Set time to end of day
			date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())
			additionalFilters["to_date"] = date
		}
	}

	// Get orders by phone number using dedicated method
	orders, total, err := h.orderService.GetOrdersByPhoneNumber(phoneNumber, page, pageSize, additionalFilters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to get orders",
			Error:   err.Error(),
		})
	}

	// Format response data
	var orderList []responses.OrderDetail

	// Create user info cache to avoid repeated database queries
	userCache := make(map[uuid.UUID]string)

	for _, o := range orders {
		// Get creator name if available
		var creatorName string
		if o.CreatedBy != nil {
			// Check if we already have this user in cache
			if name, found := userCache[*o.CreatedBy]; found {
				creatorName = name
			} else {
				// If not in cache, fetch from database
				user, err := h.orderService.UserService.GetUserByID(*o.CreatedBy)
				if err == nil {
					creatorName = user.Username
					// Add to cache
					userCache[*o.CreatedBy] = creatorName
				}
			}
		}

		// Create basic order detail
		orderDetail := responses.OrderDetail{
			ID:               o.ID,
			CustomerName:     o.CustomerName,
			CustomerEmail:    o.CustomerEmail,
			CustomerPhone:    o.CustomerPhone,
			ShippingAddress:  o.ShippingAddress,
			ShippingWard:     o.ShippingWard,
			ShippingDistrict: o.ShippingDistrict,
			ShippingCity:     o.ShippingCity,
			ShippingCountry:  o.ShippingCountry,
			PaymentMethod:    string(o.PaymentMethod),
			Status:           string(o.OrderStatus),
			Notes:            o.Notes,
			Total:            o.TotalAmount,
			DiscountAmount:   o.DiscountAmount,
			DiscountReason:   o.DiscountReason,
			FinalTotal:       o.FinalTotalAmount,
			CreatedAt:        o.CreatedAt,
			UpdatedAt:        o.UpdatedAt,
		}

		// Set created by if available
		if o.CreatedBy != nil {
			orderDetail.CreatedBy = *o.CreatedBy
			orderDetail.CreatedByName = creatorName
		}

		// Add shipment info if available
		if o.Shipment != nil {
			orderDetail.Shipment = &responses.ShipmentResponse{
				ID:             o.Shipment.ID,
				OrderID:        o.Shipment.OrderID,
				TrackingNumber: o.Shipment.TrackingNumber,
				Carrier:        o.Shipment.Carrier,
				CreatedAt:      o.Shipment.CreatedAt,
				UpdatedAt:      o.Shipment.UpdatedAt,
			}
		}

		// Add items if available
		items := make([]responses.OrderItemResponse, len(o.Items))
		for i, item := range o.Items {
			// Create basic item
			items[i] = responses.OrderItemResponse{
				ID:          item.ID,
				OrderID:     item.OrderID,
				InventoryID: item.InventoryID,
				Quantity:    item.Quantity,
				Price:       item.PriceAtOrder,
				Subtotal:    item.PriceAtOrder * float64(item.Quantity),
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
			}

			// Get inventory details if needed
			inventory, err := h.orderService.ProductService.GetInventoryByID(item.InventoryID)
			if err == nil && inventory != nil {
				// Add inventory details
				items[i].Size = inventory.Size
				items[i].Color = inventory.Color

				// Get product details if available
				product, err := h.orderService.ProductService.GetProductByID(inventory.ProductID)
				if err == nil && product != nil {
					items[i].ProductID = product.ID
					items[i].ProductName = product.Name

					// Get price details if available
					price, err := h.orderService.ProductService.GetCurrentPrice(product.ID)
					if err == nil && price != nil {
						items[i].PriceID = price.ID
						items[i].Currency = price.Currency
					}
				}
			}
		}

		orderDetail.Items = items
		orderList = append(orderList, orderDetail)
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrdersResponse{
		Success:    true,
		Message:    "Orders retrieved successfully",
		Data:       orderList,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int64(math.Ceil(float64(total) / float64(pageSize))),
	})
}

// ORDER STATUS WEBHOOK FOR GHN NOT USE YET
func (h *OrderHandler) HandleGHNOrderStatusWebhook(c *fiber.Ctx) error {
	// 1. Optional: validate GHN token via header or query
	ghnToken := c.Get("X-GHN-Token")
	if ghnToken != os.Getenv("GHN_WEBHOOK_SECRET") {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// 2. Parse JSON body
	var payload GHNWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid payload",
			"error":   err.Error(),
		})
	}

	// 3. Log & process
	fmt.Printf("GHN webhook: %+v\n", payload)

	// TODO: update order by payload.OrderCode, set status = payload.Status

	return c.SendStatus(fiber.StatusOK)
}

type GHNWebhookPayload struct {
	CODAmount         float64      `json:"CODAmount"`
	CODTransferDate   *string      `json:"CODTransferDate"` // dùng *string để chấp nhận null
	ClientOrderCode   string       `json:"ClientOrderCode"`
	ConvertedWeight   float64      `json:"ConvertedWeight"`
	Description       string       `json:"Description"`
	Fee               GHNFeeDetail `json:"Fee"`
	Height            float64      `json:"Height"`
	IsPartialReturn   bool         `json:"IsPartialReturn"`
	Length            float64      `json:"Length"`
	OrderCode         string       `json:"OrderCode"`
	PartialReturnCode string       `json:"PartialReturnCode"`
	PaymentType       int          `json:"PaymentType"`
	Reason            string       `json:"Reason"`
	ReasonCode        string       `json:"ReasonCode"`
	ShopID            int64        `json:"ShopID"`
	Status            string       `json:"Status"`
	Time              string       `json:"Time"` // ISO8601, nếu muốn thì có thể dùng time.Time và custom unmarshal
	TotalFee          float64      `json:"TotalFee"`
	Type              string       `json:"Type"`
	Warehouse         string       `json:"Warehouse"`
	Weight            float64      `json:"Weight"`
	Width             float64      `json:"Width"`
}

type GHNFeeDetail struct {
	CODFailedFee          float64 `json:"CODFailedFee"`
	CODFee                float64 `json:"CODFee"`
	Coupon                float64 `json:"Coupon"`
	DeliverRemoteAreasFee float64 `json:"DeliverRemoteAreasFee"`
	DocumentReturn        float64 `json:"DocumentReturn"`
	DoubleCheck           float64 `json:"DoubleCheck"`
	Insurance             float64 `json:"Insurance"`
	MainService           float64 `json:"MainService"`
	PickRemoteAreasFee    float64 `json:"PickRemoteAreasFee"`
	R2S                   float64 `json:"R2S"`
	Return                float64 `json:"Return"`
	StationDO             float64 `json:"StationDO"`
	StationPU             float64 `json:"StationPU"`
	Total                 float64 `json:"Total"`
}
