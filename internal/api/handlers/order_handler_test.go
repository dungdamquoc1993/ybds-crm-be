package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ybds/internal/models/order"
	"github.com/ybds/internal/testutil"
)

// orderMockJWTMiddleware creates a simple JWT middleware for testing
func orderMockJWTMiddleware(c *fiber.Ctx) error {
	// Set user ID in locals for testing
	c.Locals("user_id", "test-user-id")
	c.Locals("roles", []string{"admin"})
	return c.Next()
}

func TestOrderHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Create a mock order service
	mockOrderService := new(testutil.MockOrderService)

	// Register routes with mock middleware
	api := app.Group("/api")
	orders := api.Group("/orders")
	orders.Use(orderMockJWTMiddleware)

	// Register the GetOrders endpoint
	orders.Get("/", func(c *fiber.Ctx) error {
		// Parse pagination parameters
		page := 1
		pageSize := 10

		// Get orders from service
		orders, total, err := mockOrderService.GetAllOrders(page, pageSize, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to retrieve orders",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Orders retrieved successfully",
			"data": fiber.Map{
				"orders":      orders,
				"total":       total,
				"page":        page,
				"page_size":   pageSize,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	})

	// Register the GetOrderByID endpoint
	orders.Get("/:id", func(c *fiber.Ctx) error {
		// Parse order ID from path
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid order ID format",
				"error":   err.Error(),
			})
		}

		// Get order from service
		order, err := mockOrderService.GetOrderByID(id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Order not found",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Order retrieved successfully",
			"data":    order,
		})
	})

	// Register the CreateOrder endpoint
	orders.Post("/", func(c *fiber.Ctx) error {
		// Parse request body
		var request struct {
			UserID          *string                   `json:"user_id"`
			GuestID         *string                   `json:"guest_id"`
			Items           []testutil.OrderItemInput `json:"items"`
			ShippingAddress *testutil.AddressInput    `json:"shipping_address"`
			BillingAddress  *testutil.AddressInput    `json:"billing_address"`
			PaymentMethod   string                    `json:"payment_method"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		}

		// Validate request
		if (request.UserID == nil && request.GuestID == nil) || len(request.Items) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "User ID or Guest ID and at least one item are required",
			})
		}

		// Parse UUIDs
		var userID, guestID *uuid.UUID
		if request.UserID != nil {
			parsedUserID, err := uuid.Parse(*request.UserID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "Invalid user ID format",
					"error":   err.Error(),
				})
			}
			userID = &parsedUserID
		}

		if request.GuestID != nil {
			parsedGuestID, err := uuid.Parse(*request.GuestID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "Invalid guest ID format",
					"error":   err.Error(),
				})
			}
			guestID = &parsedGuestID
		}

		// Create order
		result, err := mockOrderService.CreateOrder(
			userID,
			guestID,
			request.Items,
			request.ShippingAddress,
			request.BillingAddress,
			request.PaymentMethod,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create order",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Order created successfully",
			"data":    result,
		})
	})

	// Register the UpdateOrderStatus endpoint
	orders.Patch("/:id/status", func(c *fiber.Ctx) error {
		// Parse order ID from path
		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid order ID format",
				"error":   err.Error(),
			})
		}

		// Parse request body
		var request struct {
			Status string `json:"status"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
		}

		// Validate request
		if request.Status == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Status is required",
			})
		}

		// Update order status
		result, err := mockOrderService.UpdateOrderStatus(id, request.Status)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update order status",
				"error":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Order status updated successfully",
			"data":    result,
		})
	})

	// Test case 1: Get all orders
	t.Run("GetOrders", func(t *testing.T) {
		// Setup mock expectations
		orders := []order.Order{
			{
				PaymentMethod: order.PaymentCash,
				TotalAmount:   100.0,
				OrderStatus:   order.OrderShipmentRequested,
			},
			{
				PaymentMethod: order.PaymentCash,
				TotalAmount:   200.0,
				OrderStatus:   order.OrderDelivered,
			},
		}
		orders[0].ID = uuid.New()
		orders[1].ID = uuid.New()

		mockOrderService.On("GetAllOrders", 1, 10, mock.Anything).Return(orders, int64(2), nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
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
		assert.Equal(t, "Orders retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		assert.NotNil(t, data["orders"])
		assert.Equal(t, float64(2), data["total"])
		assert.Equal(t, float64(1), data["page"])
		assert.Equal(t, float64(10), data["page_size"])

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	// Test case 2: Get order by ID
	t.Run("GetOrderByID", func(t *testing.T) {
		// Setup mock expectations
		orderID := uuid.New()
		testOrder := &order.Order{
			PaymentMethod: order.PaymentCash,
			TotalAmount:   100.0,
			OrderStatus:   order.OrderShipmentRequested,
		}
		testOrder.ID = orderID
		testOrder.CreatedAt = time.Now()

		mockOrderService.On("GetOrderByID", orderID).Return(testOrder, nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/orders/"+orderID.String(), nil)
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
		assert.Equal(t, "Order retrieved successfully", response["message"])

		data := response["data"].(map[string]interface{})
		assert.Equal(t, "pending_confirmation", data["order_status"])
		assert.Equal(t, float64(100.0), data["total_amount"])

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	// Test case 3: Get order by ID - not found
	t.Run("GetOrderByID_NotFound", func(t *testing.T) {
		// Setup mock expectations
		orderID := uuid.New()
		mockOrderService.On("GetOrderByID", orderID).Return(nil, errors.New("order not found")).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/api/orders/"+orderID.String(), nil)
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
		assert.Equal(t, "Order not found", response["message"])
		assert.Equal(t, "order not found", response["error"])

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	// Test case 4: Get order by ID - invalid ID
	t.Run("GetOrderByID_InvalidID", func(t *testing.T) {
		// Create request with invalid UUID
		req := httptest.NewRequest(http.MethodGet, "/api/orders/invalid-uuid", nil)
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
		assert.Equal(t, "Invalid order ID format", response["message"])
	})

	// Test case 5: Create order
	t.Run("CreateOrder", func(t *testing.T) {
		// Setup mock expectations
		userID := uuid.New()
		orderID := uuid.New()

		result := &testutil.OrderResult{
			Success: true,
			Message: "Order created successfully",
			OrderID: orderID,
			UserID:  &userID,
			Status:  "pending",
			Total:   150.0,
		}

		// Create items
		items := []testutil.OrderItemInput{
			{
				ProductID: uuid.New(),
				Quantity:  2,
				UnitPrice: 75.0,
				Properties: map[string]interface{}{
					"color": "blue",
					"size":  "M",
				},
			},
		}

		// Create addresses
		shippingAddress := &testutil.AddressInput{
			Address:  "123 Main St",
			Ward:     "Ward 1",
			District: "District 1",
			City:     "City 1",
			Country:  "Country 1",
		}

		billingAddress := &testutil.AddressInput{
			Address:  "456 Billing St",
			Ward:     "Ward 2",
			District: "District 2",
			City:     "City 2",
			Country:  "Country 2",
		}

		mockOrderService.On(
			"CreateOrder",
			&userID,
			(*uuid.UUID)(nil),
			mock.Anything,
			mock.Anything,
			mock.Anything,
			"credit_card",
		).Return(result, nil).Once()

		// Create request body
		userIDStr := userID.String()
		requestBody := map[string]interface{}{
			"user_id":          userIDStr,
			"items":            items,
			"shipping_address": shippingAddress,
			"billing_address":  billingAddress,
			"payment_method":   "credit_card",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "Order created successfully", response["message"])

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	// Test case 6: Create order - validation error
	t.Run("CreateOrder_ValidationError", func(t *testing.T) {
		// Create request body with missing required fields
		requestBody := map[string]interface{}{
			"payment_method": "credit_card",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "User ID or Guest ID and at least one item are required", response["message"])
	})

	// Test case 7: Update order status
	t.Run("UpdateOrderStatus", func(t *testing.T) {
		// Setup mock expectations
		orderID := uuid.New()
		userID := uuid.New()

		result := &testutil.OrderResult{
			Success: true,
			Message: "Order status updated successfully",
			OrderID: orderID,
			UserID:  &userID,
			Status:  "shipped",
			Total:   150.0,
		}

		mockOrderService.On("UpdateOrderStatus", orderID, "shipped").Return(result, nil).Once()

		// Create request body
		requestBody := map[string]interface{}{
			"status": "shipped",
		}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPatch, "/api/orders/"+orderID.String()+"/status", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "Order status updated successfully", response["message"])

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	// Test case 8: Update order status - validation error
	t.Run("UpdateOrderStatus_ValidationError", func(t *testing.T) {
		// Setup mock expectations
		orderID := uuid.New()

		// Create request body with missing required fields
		requestBody := map[string]interface{}{}
		requestJSON, _ := json.Marshal(requestBody)

		// Create request
		req := httptest.NewRequest(http.MethodPatch, "/api/orders/"+orderID.String()+"/status", bytes.NewBuffer(requestJSON))
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
		assert.Equal(t, "Status is required", response["message"])
	})
}
