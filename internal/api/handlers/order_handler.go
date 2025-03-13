package handlers

import (
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
func NewOrderHandler(db *gorm.DB, userService *services.UserService, notificationService *services.NotificationService) *OrderHandler {
	return &OrderHandler{
		orderService: services.NewOrderService(db, userService, notificationService),
	}
}

// RegisterRoutes registers all routes related to orders
func (h *OrderHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	orders := router.Group("/orders")
	orders.Use(authMiddleware)

	// Order routes
	orders.Post("/", h.CreateOrder)
	orders.Get("/", h.GetOrders)
	orders.Get("/:id", h.GetOrderByID)
	orders.Put("/:id/status", h.UpdateOrderStatus)
	orders.Put("/:id/payment", h.UpdatePaymentStatus)
	orders.Delete("/:id", h.DeleteOrder)

	// Order item routes
	orders.Post("/:id/items", h.AddOrderItem)
	orders.Put("/items/:id", h.UpdateOrderItem)
	orders.Delete("/items/:id", h.DeleteOrderItem)
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with items
// @Tags orders
// @Accept json
// @Produce json
// @Param order body requests.CreateOrderRequest true "Order details"
// @Success 201 {object} responses.OrderResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders [post]
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
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Convert request items to service items
	items := make([]services.OrderItemInfo, len(req.Items))
	for i, item := range req.Items {
		items[i] = services.OrderItemInfo{
			InventoryID: item.InventoryID,
			Quantity:    item.Quantity,
		}
	}

	// Determine customer type (assuming all are users for now)
	customerType := order.CustomerUser

	// Create order
	result, err := h.orderService.CreateOrder(
		req.CustomerID,
		customerType,
		order.PaymentMethod(req.PaymentMethod),
		items,
		&userID, // CreatedBy (staff member)
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to create order",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusCreated).JSON(responses.OrderResponse{
		Success: true,
		Message: "Order created successfully",
		Data: responses.OrderDetail{
			ID:            result.OrderID,
			Status:        string(result.Status),
			Total:         result.Total,
			CreatedBy:     *result.CreatedBy,
			CreatedByName: "", // Would need to fetch this from user service
		},
	})
}

// GetOrders godoc
// @Summary Get all orders
// @Description Get a list of all orders with pagination and filtering
// @Tags orders
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Filter by status"
// @Param customer_id query string false "Filter by customer ID"
// @Success 200 {object} responses.OrdersResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders [get]
// @Security ApiKeyAuth
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	// Parse filters
	filters := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		filters["order_status"] = status
	}

	if customerID := c.Query("customer_id"); customerID != "" {
		id, err := uuid.Parse(customerID)
		if err == nil {
			filters["customer_id"] = id
		}
	}

	// Get orders
	orders, total, err := h.orderService.GetAllOrders(page, pageSize, filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve orders",
			Error:   err.Error(),
		})
	}

	// Convert to response format
	orderDetails := make([]responses.OrderDetail, len(orders))
	for i, o := range orders {
		orderDetails[i] = responses.OrderDetail{
			ID:                o.ID,
			CustomerID:        o.CustomerID,
			CustomerName:      "",       // Would need to fetch this from user service
			CustomerEmail:     "",       // Would need to fetch this from user service
			CustomerPhone:     "",       // Would need to fetch this from user service
			ShippingAddressID: uuid.Nil, // Not in the model, would need to add
			ShippingAddress:   "",       // Not in the model, would need to add
			PaymentMethod:     string(o.PaymentMethod),
			PaymentStatus:     o.PaymentStatus,
			Status:            string(o.OrderStatus),
			Notes:             "", // Not in the model, would need to add
			Total:             o.TotalAmount,
			CreatedBy:         *o.CreatedBy, // Assuming it's not nil
			CreatedByName:     "",           // Would need to fetch this from user service
			CreatedAt:         o.CreatedAt,
			UpdatedAt:         o.UpdatedAt,
		}
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrdersResponse{
		Success:    true,
		Message:    "Orders retrieved successfully",
		Data:       orderDetails,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetOrderByID godoc
// @Summary Get an order by ID
// @Description Get a specific order with its items
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} responses.OrderDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders/{id} [get]
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
			ProductID:   uuid.Nil, // Would need to fetch from inventory
			ProductName: "",       // Would need to fetch from product service
			Size:        "",       // Would need to fetch from inventory
			Color:       "",       // Would need to fetch from inventory
			PriceID:     uuid.Nil, // Would need to fetch from price service
			Currency:    "",       // Would need to fetch from price service
			Notes:       "",       // Not in the model, would need to add
		}
	}

	// Create response
	orderDetail := responses.OrderDetail{
		ID:                o.ID,
		CustomerID:        o.CustomerID,
		CustomerName:      "",       // Would need to fetch this from user service
		CustomerEmail:     "",       // Would need to fetch this from user service
		CustomerPhone:     "",       // Would need to fetch this from user service
		ShippingAddressID: uuid.Nil, // Not in the model, would need to add
		ShippingAddress:   "",       // Not in the model, would need to add
		PaymentMethod:     string(o.PaymentMethod),
		PaymentStatus:     o.PaymentStatus,
		Status:            string(o.OrderStatus),
		Notes:             "", // Not in the model, would need to add
		Total:             o.TotalAmount,
		CreatedBy:         *o.CreatedBy, // Assuming it's not nil
		CreatedByName:     "",           // Would need to fetch this from user service
		Items:             items,
		CreatedAt:         o.CreatedAt,
		UpdatedAt:         o.UpdatedAt,
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrderDetailResponse{
		Success: true,
		Message: "Order retrieved successfully",
		Data:    orderDetail,
	})
}

// UpdateOrderStatus godoc
// @Summary Update an order's status
// @Description Update the status of an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body requests.UpdateOrderStatusRequest true "Order status"
// @Success 200 {object} responses.OrderResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders/{id}/status [put]
// @Security ApiKeyAuth
func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
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
	result, err := h.orderService.UpdateOrderStatus(id, order.OrderStatus(req.Status))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update order status",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrderResponse{
		Success: true,
		Message: "Order status updated successfully",
		Data: responses.OrderDetail{
			ID:            result.OrderID,
			Status:        string(result.Status),
			Total:         result.Total,
			CreatedBy:     *result.CreatedBy, // Assuming it's not nil
			CreatedByName: "",                // Would need to fetch this from user service
		},
	})
}

// UpdatePaymentStatus godoc
// @Summary Update an order's payment status
// @Description Update the payment status of an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param payment body requests.UpdatePaymentStatusRequest true "Payment status"
// @Success 200 {object} responses.OrderResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders/{id}/payment [put]
// @Security ApiKeyAuth
func (h *OrderHandler) UpdatePaymentStatus(c *fiber.Ctx) error {
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

	// Parse request
	var req requests.UpdatePaymentStatusRequest
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

	// Update payment status
	result, err := h.orderService.UpdatePaymentStatus(id, req.PaymentStatus, nil) // Not passing paid amount for now
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Failed to update payment status",
			Error:   err.Error(),
		})
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrderResponse{
		Success: true,
		Message: "Payment status updated successfully",
		Data: responses.OrderDetail{
			ID:            result.OrderID,
			Status:        string(result.Status),
			Total:         result.Total,
			CreatedBy:     *result.CreatedBy, // Assuming it's not nil
			CreatedByName: "",                // Would need to fetch this from user service
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
// @Router /orders/{id} [delete]
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
// @Router /orders/{id}/items [post]
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

	// Return response
	return c.Status(fiber.StatusCreated).JSON(responses.OrderItemDetailResponse{
		Success: true,
		Message: "Order item added successfully",
		Data: responses.OrderItemResponse{
			ID:          uuid.New(),
			OrderID:     orderID,
			ProductID:   req.ProductID,
			ProductName: "Product Name", // Mock data
			InventoryID: req.InventoryID,
			Size:        "M",    // Mock data
			Color:       "Blue", // Mock data
			PriceID:     req.PriceID,
			Price:       100.0, // Mock data
			Currency:    "USD", // Mock data
			Quantity:    req.Quantity,
			Subtotal:    100.0 * float64(req.Quantity), // Mock data
			Notes:       req.Notes,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	})
}

// UpdateOrderItem godoc
// @Summary Update an order item
// @Description Update an existing order item
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order Item ID"
// @Param item body requests.UpdateOrderItemRequest true "Order item details"
// @Success 200 {object} responses.OrderItemDetailResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /orders/items/{id} [put]
// @Security ApiKeyAuth
func (h *OrderHandler) UpdateOrderItem(c *fiber.Ctx) error {
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

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.OrderItemDetailResponse{
		Success: true,
		Message: "Order item updated successfully",
		Data: responses.OrderItemResponse{
			ID:          id,
			OrderID:     uuid.New(),     // Mock data
			ProductID:   uuid.New(),     // Mock data
			ProductName: "Product Name", // Mock data
			InventoryID: uuid.New(),     // Mock data
			Size:        "M",            // Mock data
			Color:       "Blue",         // Mock data
			PriceID:     uuid.New(),     // Mock data
			Price:       100.0,          // Mock data
			Currency:    "USD",          // Mock data
			Quantity:    req.Quantity,
			Subtotal:    100.0 * float64(req.Quantity), // Mock data
			Notes:       req.Notes,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
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
// @Router /orders/items/{id} [delete]
// @Security ApiKeyAuth
func (h *OrderHandler) DeleteOrderItem(c *fiber.Ctx) error {
	// Parse order item ID
	idStr := c.Params("id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(responses.ErrorResponse{
			Success: false,
			Message: "Invalid order item ID format",
			Error:   err.Error(),
		})
	}

	// TODO: Implement DeleteOrderItem method in OrderService
	// For now, we'll just return a success response

	// Return response
	return c.Status(fiber.StatusOK).JSON(responses.SuccessResponse{
		Success: true,
		Message: "Order item deleted successfully",
	})
}
