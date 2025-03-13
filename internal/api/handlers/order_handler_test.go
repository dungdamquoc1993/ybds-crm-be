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
	"github.com/ybds/internal/models/order"
	"github.com/ybds/internal/services"
	"github.com/ybds/internal/testutil"
)

// MockNotificationService is a mock implementation of services.NotificationService for testing
type MockNotificationService struct {
	*services.NotificationService
}

// NewMockNotificationService creates a new mock notification service
func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		NotificationService: &services.NotificationService{},
	}
}

// MockUserService is a mock implementation of services.UserService for testing
type MockUserService struct {
	*services.UserService
}

// NewMockUserService creates a new mock user service
func NewMockUserService() *MockUserService {
	return &MockUserService{
		UserService: &services.UserService{},
	}
}

func TestOrderHandler(t *testing.T) {
	// Create mock services
	mockOrderService := new(testutil.MockOrderService)
	mockUserService := NewMockUserService()
	mockNotificationService := NewMockNotificationService()

	// Create test app
	app := testutil.SetupTestApp()

	// Create a fixed user ID for testing
	userID := uuid.New()

	// Create auth middleware with fixed user ID
	authMiddleware := testutil.CreateAuthMiddlewareWithUserID(userID)

	// Create order handler
	orderHandler := handlers.NewOrderHandler(nil, mockUserService.UserService, mockNotificationService.NotificationService)

	// Register routes
	api := app.Group("/api")
	authenticated := api.Group("/")
	authenticated.Use(authMiddleware)
	orderHandler.RegisterRoutes(authenticated, authMiddleware)

	t.Run("GetOrders", func(t *testing.T) {
		// Create test orders
		orders := []order.Order{
			{
				CustomerID:    userID,
				CustomerType:  order.CustomerUser,
				PaymentMethod: order.PaymentCOD,
				PaymentStatus: "pending",
				TotalAmount:   99.99,
				PaidAmount:    0,
				OrderStatus:   order.OrderPendingConfirmation,
			},
			{
				CustomerID:    userID,
				CustomerType:  order.CustomerUser,
				PaymentMethod: order.PaymentBankTransfer,
				PaymentStatus: "completed",
				TotalAmount:   149.99,
				PaidAmount:    149.99,
				OrderStatus:   order.OrderDelivered,
			},
		}

		// Set IDs and timestamps manually
		now := time.Now()
		orders[0].ID = uuid.New()
		orders[0].CreatedAt = now.Add(-24 * time.Hour)
		orders[0].UpdatedAt = now.Add(-24 * time.Hour)

		orders[1].ID = uuid.New()
		orders[1].CreatedAt = now.Add(-48 * time.Hour)
		orders[1].UpdatedAt = now.Add(-24 * time.Hour)

		// Setup mock expectations
		mockOrderService.On("GetAllOrders", mock.Anything, mock.Anything, mock.Anything).
			Return(orders, 2, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/orders",
			QueryParams: map[string]string{
				"page":      "1",
				"page_size": "10",
			},
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	t.Run("GetOrders_Error", func(t *testing.T) {
		// Setup mock expectations
		mockOrderService.On("GetAllOrders", mock.Anything, mock.Anything, mock.Anything).
			Return([]order.Order{}, 0, errors.New("database error"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/orders",
		})

		// Assert response
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	t.Run("GetOrder", func(t *testing.T) {
		// Create test order
		orderID := uuid.New()
		testOrder := order.Order{
			CustomerID:    userID,
			CustomerType:  order.CustomerUser,
			PaymentMethod: order.PaymentCOD,
			PaymentStatus: "pending",
			TotalAmount:   99.99,
			PaidAmount:    0,
			OrderStatus:   order.OrderPendingConfirmation,
		}
		testOrder.ID = orderID
		testOrder.CreatedAt = time.Now().Add(-24 * time.Hour)
		testOrder.UpdatedAt = time.Now().Add(-24 * time.Hour)

		// Setup mock expectations
		mockOrderService.On("GetOrderByID", orderID).
			Return(&testOrder, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/orders/" + orderID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	t.Run("GetOrder_NotFound", func(t *testing.T) {
		// Create order ID
		orderID := uuid.New()

		// Setup mock expectations
		mockOrderService.On("GetOrderByID", orderID).
			Return(nil, errors.New("order not found"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodGet,
			URL:    "/api/orders/" + orderID.String(),
		})

		// Assert response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	t.Run("CreateOrder", func(t *testing.T) {
		// Create order request
		productID := uuid.New()
		shippingAddressID := uuid.New()

		createReq := requests.CreateOrderRequest{
			CustomerID:        userID,
			ShippingAddressID: shippingAddressID,
			PaymentMethod:     "cod",
			Status:            "pending_confirmation",
			Notes:             "Test order",
			Items: []requests.OrderItemInfo{
				{
					ProductID:   productID,
					InventoryID: uuid.New(),
					PriceID:     uuid.New(),
					Quantity:    2,
					Notes:       "Test item",
				},
			},
		}

		// Create order result
		orderID := uuid.New()
		orderResult := &testutil.OrderResult{
			Success: true,
			Message: "Order created successfully",
			OrderID: orderID,
			UserID:  &userID,
			Status:  "pending_confirmation",
			Total:   99.98,
		}

		// Setup mock expectations
		mockOrderService.On("CreateOrder",
			mock.AnythingOfType("*uuid.UUID"),
			mock.AnythingOfType("*uuid.UUID"),
			mock.AnythingOfType("[]requests.OrderItemInfo"),
			mock.AnythingOfType("*requests.AddressInput"),
			mock.AnythingOfType("*requests.AddressInput"),
			mock.AnythingOfType("string")).
			Return(orderResult, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPost,
			URL:    "/api/orders",
			Body:   createReq,
		})

		// Assert response
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	t.Run("CreateOrder_ValidationError", func(t *testing.T) {
		// Create invalid order request (missing required fields)
		createReq := requests.CreateOrderRequest{
			// CustomerID is required but missing
			ShippingAddressID: uuid.New(),
			PaymentMethod:     "cod",
			// Items is required but missing
		}

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPost,
			URL:    "/api/orders",
			Body:   createReq,
		})

		// Assert response
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("UpdateOrderStatus", func(t *testing.T) {
		// Create order ID
		orderID := uuid.New()

		// Create update request
		updateReq := requests.UpdateOrderStatusRequest{
			Status: string(order.OrderShipped),
		}

		// Create order result
		orderResult := &testutil.OrderResult{
			Success: true,
			Message: "Order status updated successfully",
			OrderID: orderID,
			UserID:  &userID,
			Status:  string(order.OrderShipped),
			Total:   99.99,
		}

		// Setup mock expectations
		mockOrderService.On("UpdateOrderStatus",
			orderID,
			updateReq.Status).
			Return(orderResult, nil)

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/orders/" + orderID.String() + "/status",
			Body:   updateReq,
		})

		// Assert response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})

	t.Run("UpdateOrderStatus_NotFound", func(t *testing.T) {
		// Create order ID
		orderID := uuid.New()

		// Create update request
		updateReq := requests.UpdateOrderStatusRequest{
			Status: string(order.OrderShipped),
		}

		// Setup mock expectations
		mockOrderService.On("UpdateOrderStatus",
			orderID,
			updateReq.Status).
			Return(nil, errors.New("order not found"))

		// Execute request
		resp := testutil.ExecuteRequest(t, app, testutil.TestRequest{
			Method: http.MethodPut,
			URL:    "/api/orders/" + orderID.String() + "/status",
			Body:   updateReq,
		})

		// Assert response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify mock expectations
		mockOrderService.AssertExpectations(t)
	})
}
