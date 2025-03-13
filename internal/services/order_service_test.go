package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ybds/internal/models/order"
	"github.com/ybds/internal/services"
)

// TestOrderService tests the OrderService functionality
func TestOrderService(t *testing.T) {
	// This is an integration test that would require a database
	// In a real-world scenario, you would use a test database or mock the database
	t.Skip("Skipping integration test")
}

// TestOrderResult tests the OrderResult struct
func TestOrderResult(t *testing.T) {
	// Create an OrderResult
	orderID := uuid.New()
	createdBy := uuid.New()
	result := services.OrderResult{
		Success:   true,
		Message:   "Order created successfully",
		OrderID:   orderID,
		Status:    order.OrderConfirmed,
		Total:     1000.0,
		CreatedBy: &createdBy,
	}

	// Test the fields
	assert.True(t, result.Success)
	assert.Equal(t, "Order created successfully", result.Message)
	assert.Equal(t, orderID, result.OrderID)
	assert.Equal(t, order.OrderConfirmed, result.Status)
	assert.Equal(t, 1000.0, result.Total)
	assert.Equal(t, &createdBy, result.CreatedBy)
}

// TestOrderItemInfo tests the OrderItemInfo struct
func TestOrderItemInfo(t *testing.T) {
	// Create an OrderItemInfo
	inventoryID := uuid.New()
	itemInfo := services.OrderItemInfo{
		InventoryID: inventoryID,
		Quantity:    2,
	}

	// Test the fields
	assert.Equal(t, inventoryID, itemInfo.InventoryID)
	assert.Equal(t, 2, itemInfo.Quantity)
}

// TestOrder tests the Order model
func TestOrder(t *testing.T) {
	// Create an Order
	orderID := uuid.New()
	customerID := uuid.New()
	createdBy := uuid.New()
	o := order.Order{
		CustomerID:    customerID,
		CustomerType:  order.CustomerUser,
		PaymentMethod: order.PaymentCash,
		PaymentStatus: "pending",
		TotalAmount:   1000.0,
		PaidAmount:    0.0,
		OrderStatus:   order.OrderPendingConfirmation,
	}
	o.ID = orderID
	o.CreatedBy = &createdBy

	// Test the fields
	assert.Equal(t, orderID, o.ID)
	assert.Equal(t, customerID, o.CustomerID)
	assert.Equal(t, order.CustomerUser, o.CustomerType)
	assert.Equal(t, order.PaymentCash, o.PaymentMethod)
	assert.Equal(t, "pending", o.PaymentStatus)
	assert.Equal(t, 1000.0, o.TotalAmount)
	assert.Equal(t, 0.0, o.PaidAmount)
	assert.Equal(t, order.OrderPendingConfirmation, o.OrderStatus)
	assert.Equal(t, &createdBy, o.CreatedBy)
}

// TestOrderItem tests the OrderItem model
func TestOrderItem(t *testing.T) {
	// Create an OrderItem
	orderItemID := uuid.New()
	orderID := uuid.New()
	inventoryID := uuid.New()
	item := order.OrderItem{
		OrderID:      orderID,
		InventoryID:  inventoryID,
		Quantity:     2,
		PriceAtOrder: 500.0,
	}
	item.ID = orderItemID

	// Test the fields
	assert.Equal(t, orderItemID, item.ID)
	assert.Equal(t, orderID, item.OrderID)
	assert.Equal(t, inventoryID, item.InventoryID)
	assert.Equal(t, 2, item.Quantity)
	assert.Equal(t, 500.0, item.PriceAtOrder)
}
