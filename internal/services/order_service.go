package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ybds/internal/models/order"
	"github.com/ybds/internal/models/product"
	"github.com/ybds/internal/repositories"
	"gorm.io/gorm"
)

// OrderService handles order-related business logic
type OrderService struct {
	db                  *gorm.DB
	orderRepo           *repositories.OrderRepository
	productRepo         *repositories.ProductRepository
	userService         *UserService
	notificationService *NotificationService
}

// NewOrderService creates a new instance of OrderService
func NewOrderService(db *gorm.DB, userService *UserService, notificationService *NotificationService) *OrderService {
	return &OrderService{
		db:                  db,
		orderRepo:           repositories.NewOrderRepository(db),
		productRepo:         repositories.NewProductRepository(db),
		userService:         userService,
		notificationService: notificationService,
	}
}

// OrderResult represents the result of an order operation
type OrderResult struct {
	Success   bool
	Message   string
	Error     string
	OrderID   uuid.UUID
	Status    order.OrderStatus
	Total     float64
	CreatedBy *uuid.UUID
}

// OrderItemInfo represents information about an order item
type OrderItemInfo struct {
	InventoryID uuid.UUID
	Quantity    int
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(id uuid.UUID) (*order.Order, error) {
	return s.orderRepo.GetOrderByID(id)
}

// GetAllOrders retrieves all orders with pagination and filtering
func (s *OrderService) GetAllOrders(page, pageSize int, filters map[string]interface{}) ([]order.Order, int64, error) {
	return s.orderRepo.GetAllOrders(page, pageSize, filters)
}

// GetOrdersByCustomer retrieves all orders for a customer
func (s *OrderService) GetOrdersByCustomer(customerID uuid.UUID, customerType order.CustomerType) ([]order.Order, error) {
	return s.orderRepo.GetOrdersByCustomer(customerID, customerType)
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(
	customerID uuid.UUID,
	customerType order.CustomerType,
	paymentMethod order.PaymentMethod,
	items []OrderItemInfo,
	createdByID *uuid.UUID,
) (*OrderResult, error) {
	// Validate input
	if customerID == uuid.Nil {
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "Customer ID is required",
		}, fmt.Errorf("customer ID is required")
	}

	if len(items) == 0 {
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "At least one item is required",
		}, fmt.Errorf("at least one item is required")
	}

	// Verify customer exists using UserService
	var customerExists bool
	if customerType == order.CustomerUser {
		user, err := s.userService.GetUserByID(customerID)
		customerExists = err == nil && user != nil
	} else if customerType == order.CustomerGuest {
		guest, err := s.userService.GetGuestByID(customerID)
		customerExists = err == nil && guest != nil
	}

	if !customerExists {
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "Customer not found",
		}, fmt.Errorf("customer not found")
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "Database transaction error",
		}, tx.Error
	}

	// Create order
	o := &order.Order{
		CustomerID:    customerID,
		CustomerType:  customerType,
		PaymentMethod: paymentMethod,
		PaymentStatus: "pending",
		TotalAmount:   0,
		PaidAmount:    0,
		OrderStatus:   order.OrderPendingConfirmation,
	}

	// Set CreatedBy field if provided
	if createdByID != nil {
		o.CreatedBy = createdByID
	}

	if err := tx.Create(o).Error; err != nil {
		tx.Rollback()
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "Error creating order",
		}, err
	}

	// Add items to order
	var totalAmount float64
	for _, item := range items {
		// Get inventory
		inventory, err := s.productRepo.GetInventoryByID(item.InventoryID)
		if err != nil {
			tx.Rollback()
			return &OrderResult{
				Success: false,
				Message: "Order creation failed",
				Error:   "Inventory not found",
			}, err
		}

		// Check if inventory has enough quantity
		if inventory.Quantity < item.Quantity {
			tx.Rollback()
			return &OrderResult{
				Success: false,
				Message: "Order creation failed",
				Error:   fmt.Sprintf("Not enough inventory for product %s", inventory.ProductID),
			}, fmt.Errorf("not enough inventory for product %s", inventory.ProductID)
		}

		// Get current price
		price, err := s.productRepo.GetCurrentPrice(inventory.ProductID)
		if err != nil {
			tx.Rollback()
			return &OrderResult{
				Success: false,
				Message: "Order creation failed",
				Error:   fmt.Sprintf("No valid price found for product %s", inventory.ProductID),
			}, fmt.Errorf("no valid price found for product %s", inventory.ProductID)
		}

		// Create order item
		orderItem := &order.OrderItem{
			OrderID:      o.ID,
			InventoryID:  item.InventoryID,
			Quantity:     item.Quantity,
			PriceAtOrder: price.Price,
		}

		if err := tx.Create(orderItem).Error; err != nil {
			tx.Rollback()
			return &OrderResult{
				Success: false,
				Message: "Order creation failed",
				Error:   "Error creating order item",
			}, err
		}

		// Update total amount
		totalAmount += price.Price * float64(item.Quantity)
	}

	// Update order total
	o.TotalAmount = totalAmount
	if err := tx.Save(o).Error; err != nil {
		tx.Rollback()
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "Error updating order total",
		}, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "Error committing transaction",
		}, err
	}

	// Send notification
	if s.notificationService != nil {
		metadata := map[string]interface{}{
			"order_id":       o.ID.String(),
			"customer_id":    customerID.String(),
			"customer_type":  string(customerType),
			"payment_method": string(paymentMethod),
			"total_amount":   totalAmount,
			"items_count":    len(items),
		}

		// Add created_by to metadata if available
		if o.CreatedBy != nil {
			metadata["created_by"] = o.CreatedBy.String()
		}

		s.notificationService.CreateOrderNotification(o.ID, customerID, "created", metadata)
	}

	return &OrderResult{
		Success:   true,
		Message:   "Order created successfully",
		OrderID:   o.ID,
		Status:    o.OrderStatus,
		Total:     totalAmount,
		CreatedBy: o.CreatedBy,
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(id uuid.UUID, status order.OrderStatus) (*OrderResult, error) {
	// Get the order
	o, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		return &OrderResult{
			Success: false,
			Message: "Order status update failed",
			Error:   "Order not found",
		}, err
	}

	// Check if status transition is valid
	if !isValidStatusTransition(o.OrderStatus, status) {
		return &OrderResult{
			Success: false,
			Message: "Order status update failed",
			Error:   fmt.Sprintf("Invalid status transition from %s to %s", o.OrderStatus, status),
		}, fmt.Errorf("invalid status transition from %s to %s", o.OrderStatus, status)
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return &OrderResult{
			Success: false,
			Message: "Order status update failed",
			Error:   "Database transaction error",
		}, tx.Error
	}

	oldStatus := o.OrderStatus

	// Update order status
	if err := tx.Model(o).Update("order_status", status).Error; err != nil {
		tx.Rollback()
		return &OrderResult{
			Success: false,
			Message: "Order status update failed",
			Error:   "Error updating order status",
		}, err
	}

	// Handle inventory updates based on status change
	if err := s.handleInventoryForStatusChange(tx, o, oldStatus, status); err != nil {
		tx.Rollback()
		return &OrderResult{
			Success: false,
			Message: "Order status update failed",
			Error:   "Error updating inventory",
		}, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return &OrderResult{
			Success: false,
			Message: "Order status update failed",
			Error:   "Error committing transaction",
		}, err
	}

	// Send notification
	if s.notificationService != nil {
		metadata := map[string]interface{}{
			"order_id":      o.ID.String(),
			"customer_id":   o.CustomerID.String(),
			"customer_type": string(o.CustomerType),
			"old_status":    string(oldStatus),
			"new_status":    string(status),
		}

		var event string
		switch status {
		case order.OrderConfirmed:
			event = "confirmed"
		case order.OrderShipped:
			event = "shipped"
		case order.OrderDelivered:
			event = "delivered"
		case order.OrderCanceled:
			event = "canceled"
		default:
			event = "updated"
		}

		s.notificationService.CreateOrderNotification(o.ID, o.CustomerID, event, metadata)
	}

	return &OrderResult{
		Success:   true,
		Message:   "Order status updated successfully",
		OrderID:   o.ID,
		Status:    status,
		Total:     o.TotalAmount,
		CreatedBy: o.CreatedBy,
	}, nil
}

// handleInventoryForStatusChange handles inventory updates based on order status changes
func (s *OrderService) handleInventoryForStatusChange(tx *gorm.DB, o *order.Order, oldStatus, newStatus order.OrderStatus) error {
	// Get order items
	var items []order.OrderItem
	if err := tx.Where("order_id = ?", o.ID).Find(&items).Error; err != nil {
		return err
	}

	// Handle inventory updates based on status change
	// Packing => Deduct from inventory
	if (oldStatus != order.OrderPacking && newStatus == order.OrderPacking) ||
		(oldStatus != order.OrderShipped && newStatus == order.OrderShipped) {
		for _, item := range items {
			// Deduct from inventory
			inventory := &product.Inventory{}
			if err := tx.Where("id = ?", item.InventoryID).First(inventory).Error; err != nil {
				return err
			}

			// Update inventory quantity
			inventory.Quantity -= item.Quantity
			if err := tx.Save(inventory).Error; err != nil {
				return err
			}

			// Create inventory transaction
			transaction := &product.InventoryTransaction{
				InventoryID:   item.InventoryID,
				Quantity:      -item.Quantity,
				Type:          product.TransactionOutbound,
				Reason:        product.ReasonSale,
				ReferenceID:   &o.ID,
				ReferenceType: "order",
				Notes:         fmt.Sprintf("Order status changed to %s", newStatus),
			}
			if err := tx.Create(transaction).Error; err != nil {
				return err
			}
		}
	}

	// Returned => Add back to inventory
	if oldStatus != order.OrderReturned && newStatus == order.OrderReturned {
		for _, item := range items {
			// Add back to inventory
			inventory := &product.Inventory{}
			if err := tx.Where("id = ?", item.InventoryID).First(inventory).Error; err != nil {
				return err
			}

			// Update inventory quantity
			inventory.Quantity += item.Quantity
			if err := tx.Save(inventory).Error; err != nil {
				return err
			}

			// Create inventory transaction
			transaction := &product.InventoryTransaction{
				InventoryID:   item.InventoryID,
				Quantity:      item.Quantity,
				Type:          product.TransactionInbound,
				Reason:        product.ReasonReturn,
				ReferenceID:   &o.ID,
				ReferenceType: "order",
				Notes:         "Order returned",
			}
			if err := tx.Create(transaction).Error; err != nil {
				return err
			}
		}
	}

	// Canceled => Add back to inventory if previously deducted
	if (oldStatus == order.OrderPacking || oldStatus == order.OrderShipped) && newStatus == order.OrderCanceled {
		for _, item := range items {
			// Add back to inventory
			inventory := &product.Inventory{}
			if err := tx.Where("id = ?", item.InventoryID).First(inventory).Error; err != nil {
				return err
			}

			// Update inventory quantity
			inventory.Quantity += item.Quantity
			if err := tx.Save(inventory).Error; err != nil {
				return err
			}

			// Create inventory transaction
			transaction := &product.InventoryTransaction{
				InventoryID:   item.InventoryID,
				Quantity:      item.Quantity,
				Type:          product.TransactionInbound,
				Reason:        product.ReasonOrderCancellation,
				ReferenceID:   &o.ID,
				ReferenceType: "order",
				Notes:         "Order canceled",
			}
			if err := tx.Create(transaction).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(oldStatus, newStatus order.OrderStatus) bool {
	// Define valid transitions
	validTransitions := map[order.OrderStatus][]order.OrderStatus{
		order.OrderPendingConfirmation: {
			order.OrderConfirmed,
			order.OrderCanceled,
		},
		order.OrderConfirmed: {
			order.OrderShipmentRequested,
			order.OrderPacking,
			order.OrderCanceled,
		},
		order.OrderShipmentRequested: {
			order.OrderPacking,
			order.OrderCanceled,
		},
		order.OrderPacking: {
			order.OrderShipped,
			order.OrderCanceled,
		},
		order.OrderShipped: {
			order.OrderDelivered,
			order.OrderReturnRequested,
		},
		order.OrderDelivered: {
			order.OrderReturnRequested,
		},
		order.OrderReturnRequested: {
			order.OrderReturnProcessing,
			order.OrderCanceled,
		},
		order.OrderReturnProcessing: {
			order.OrderReturned,
			order.OrderCanceled,
		},
	}

	// Check if transition is valid
	for _, validStatus := range validTransitions[oldStatus] {
		if validStatus == newStatus {
			return true
		}
	}

	return false
}

// UpdatePaymentStatus updates the payment status of an order
func (s *OrderService) UpdatePaymentStatus(id uuid.UUID, status string, paidAmount *float64) (*OrderResult, error) {
	// Get the order
	o, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		return &OrderResult{
			Success: false,
			Message: "Payment status update failed",
			Error:   "Order not found",
		}, err
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return &OrderResult{
			Success: false,
			Message: "Payment status update failed",
			Error:   "Database transaction error",
		}, tx.Error
	}

	// Update payment status
	if err := tx.Model(o).Update("payment_status", status).Error; err != nil {
		tx.Rollback()
		return &OrderResult{
			Success: false,
			Message: "Payment status update failed",
			Error:   "Error updating payment status",
		}, err
	}

	// Update paid amount if provided
	if paidAmount != nil {
		if err := tx.Model(o).Update("paid_amount", *paidAmount).Error; err != nil {
			tx.Rollback()
			return &OrderResult{
				Success: false,
				Message: "Payment status update failed",
				Error:   "Error updating paid amount",
			}, err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return &OrderResult{
			Success: false,
			Message: "Payment status update failed",
			Error:   "Error committing transaction",
		}, err
	}

	// Send notification
	if s.notificationService != nil {
		metadata := map[string]interface{}{
			"order_id":       o.ID.String(),
			"customer_id":    o.CustomerID.String(),
			"customer_type":  string(o.CustomerType),
			"payment_status": status,
		}
		if paidAmount != nil {
			metadata["paid_amount"] = *paidAmount
		}

		s.notificationService.CreateOrderNotification(o.ID, o.CustomerID, "payment_updated", metadata)
	}

	return &OrderResult{
		Success:   true,
		Message:   "Payment status updated successfully",
		OrderID:   o.ID,
		Status:    o.OrderStatus,
		Total:     o.TotalAmount,
		CreatedBy: o.CreatedBy,
	}, nil
}

// DeleteOrder deletes an order by ID
func (s *OrderService) DeleteOrder(id uuid.UUID) (*OrderResult, error) {
	// Get the order
	o, err := s.orderRepo.GetOrderByID(id)
	if err != nil {
		return &OrderResult{
			Success: false,
			Message: "Order deletion failed",
			Error:   "Order not found",
		}, err
	}

	// Only allow deletion of pending or canceled orders
	if o.OrderStatus != order.OrderPendingConfirmation && o.OrderStatus != order.OrderCanceled {
		return &OrderResult{
			Success: false,
			Message: "Order deletion failed",
			Error:   "Only pending or canceled orders can be deleted",
		}, fmt.Errorf("only pending or canceled orders can be deleted")
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return &OrderResult{
			Success: false,
			Message: "Order deletion failed",
			Error:   "Database transaction error",
		}, tx.Error
	}

	// Delete order items
	if err := tx.Where("order_id = ?", id).Delete(&order.OrderItem{}).Error; err != nil {
		tx.Rollback()
		return &OrderResult{
			Success: false,
			Message: "Order deletion failed",
			Error:   "Error deleting order items",
		}, err
	}

	// Delete shipment if exists
	if err := tx.Where("order_id = ?", id).Delete(&order.Shipment{}).Error; err != nil {
		tx.Rollback()
		return &OrderResult{
			Success: false,
			Message: "Order deletion failed",
			Error:   "Error deleting shipment",
		}, err
	}

	// Delete order
	if err := tx.Delete(o).Error; err != nil {
		tx.Rollback()
		return &OrderResult{
			Success: false,
			Message: "Order deletion failed",
			Error:   "Error deleting order",
		}, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return &OrderResult{
			Success: false,
			Message: "Order deletion failed",
			Error:   "Error committing transaction",
		}, err
	}

	return &OrderResult{
		Success:   true,
		Message:   "Order deleted successfully",
		OrderID:   id,
		Status:    o.OrderStatus,
		Total:     o.TotalAmount,
		CreatedBy: o.CreatedBy,
	}, nil
}

// CreateShipment creates a shipment for an order
func (s *OrderService) CreateShipment(orderID uuid.UUID, trackingNumber, carrier string) error {
	// Get the order
	o, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	// Check if order status allows shipment
	if o.OrderStatus != order.OrderConfirmed && o.OrderStatus != order.OrderShipmentRequested && o.OrderStatus != order.OrderPacking {
		return fmt.Errorf("order status does not allow shipment creation")
	}

	// Check if shipment already exists
	existingShipment, err := s.orderRepo.GetShipmentByOrderID(orderID)
	if err == nil && existingShipment != nil && existingShipment.ID != uuid.Nil {
		return fmt.Errorf("shipment already exists for this order")
	}

	// Create shipment
	shipment := &order.Shipment{
		OrderID:        orderID,
		TrackingNumber: trackingNumber,
		Carrier:        carrier,
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Save shipment
	if err := tx.Create(shipment).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update order status to shipment requested if not already
	if o.OrderStatus == order.OrderConfirmed {
		if err := tx.Model(o).Update("order_status", order.OrderShipmentRequested).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Send notification
	if s.notificationService != nil {
		metadata := map[string]interface{}{
			"order_id":        o.ID.String(),
			"customer_id":     o.CustomerID.String(),
			"customer_type":   string(o.CustomerType),
			"tracking_number": trackingNumber,
			"carrier":         carrier,
		}

		s.notificationService.CreateOrderNotification(o.ID, o.CustomerID, "shipment_created", metadata)
	}

	return nil
}

// UpdateShipment updates a shipment
func (s *OrderService) UpdateShipment(orderID uuid.UUID, trackingNumber, carrier string) error {
	// Get the shipment
	shipment, err := s.orderRepo.GetShipmentByOrderID(orderID)
	if err != nil {
		return err
	}

	// Update fields
	if trackingNumber != "" {
		shipment.TrackingNumber = trackingNumber
	}
	if carrier != "" {
		shipment.Carrier = carrier
	}

	// Save shipment
	if err := s.orderRepo.UpdateShipment(shipment); err != nil {
		return err
	}

	// Send notification
	if s.notificationService != nil {
		// Get the order
		o, err := s.orderRepo.GetOrderByID(orderID)
		if err == nil {
			metadata := map[string]interface{}{
				"order_id":        o.ID.String(),
				"customer_id":     o.CustomerID.String(),
				"customer_type":   string(o.CustomerType),
				"tracking_number": shipment.TrackingNumber,
				"carrier":         shipment.Carrier,
			}

			s.notificationService.CreateOrderNotification(o.ID, o.CustomerID, "shipment_updated", metadata)
		}
	}

	return nil
}

// DeleteShipment deletes a shipment
func (s *OrderService) DeleteShipment(orderID uuid.UUID) error {
	// Get the shipment
	shipment, err := s.orderRepo.GetShipmentByOrderID(orderID)
	if err != nil {
		return err
	}

	// Delete shipment
	return s.orderRepo.DeleteShipment(shipment.ID)
}

// AddOrderItem adds an item to an order
func (s *OrderService) AddOrderItem(orderID uuid.UUID, inventoryID uuid.UUID, quantity int) error {
	// Get the order
	o, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	// Check if order status allows adding items
	if o.OrderStatus != order.OrderPendingConfirmation && o.OrderStatus != order.OrderConfirmed {
		return fmt.Errorf("order status does not allow adding items")
	}

	// Get inventory
	inventory, err := s.productRepo.GetInventoryByID(inventoryID)
	if err != nil {
		return err
	}

	// Check if inventory has enough quantity
	if inventory.Quantity < quantity {
		return fmt.Errorf("not enough inventory")
	}

	// Get current price
	price, err := s.productRepo.GetCurrentPrice(inventory.ProductID)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Create order item
	orderItem := &order.OrderItem{
		OrderID:      orderID,
		InventoryID:  inventoryID,
		Quantity:     quantity,
		PriceAtOrder: price.Price,
	}

	if err := tx.Create(orderItem).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update order total
	o.TotalAmount += price.Price * float64(quantity)
	if err := tx.Save(o).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// UpdateOrderItem updates an order item
func (s *OrderService) UpdateOrderItem(id uuid.UUID, quantity int) error {
	// Get the order item
	item, err := s.orderRepo.GetOrderItemByID(id)
	if err != nil {
		return err
	}

	// Get the order
	o, err := s.orderRepo.GetOrderByID(item.OrderID)
	if err != nil {
		return err
	}

	// Check if order status allows updating items
	if o.OrderStatus != order.OrderPendingConfirmation && o.OrderStatus != order.OrderConfirmed {
		return fmt.Errorf("order status does not allow updating items")
	}

	// Get inventory
	inventory, err := s.productRepo.GetInventoryByID(item.InventoryID)
	if err != nil {
		return err
	}

	// Check if inventory has enough quantity for the increase
	if quantity > item.Quantity && inventory.Quantity < (quantity-item.Quantity) {
		return fmt.Errorf("not enough inventory")
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Calculate price difference
	priceDifference := item.PriceAtOrder * float64(quantity-item.Quantity)

	// Update order item
	item.Quantity = quantity
	if err := tx.Save(item).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update order total
	o.TotalAmount += priceDifference
	if err := tx.Save(o).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

// DeleteOrderItem deletes an order item
func (s *OrderService) DeleteOrderItem(id uuid.UUID) error {
	// Get the order item
	item, err := s.orderRepo.GetOrderItemByID(id)
	if err != nil {
		return err
	}

	// Get the order
	o, err := s.orderRepo.GetOrderByID(item.OrderID)
	if err != nil {
		return err
	}

	// Check if order status allows deleting items
	if o.OrderStatus != order.OrderPendingConfirmation && o.OrderStatus != order.OrderConfirmed {
		return fmt.Errorf("order status does not allow deleting items")
	}

	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Update order total
	o.TotalAmount -= item.PriceAtOrder * float64(item.Quantity)
	if err := tx.Save(o).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete order item
	if err := tx.Delete(item).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}
