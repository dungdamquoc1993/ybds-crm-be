package services

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ybds/internal/models/order"
	"github.com/ybds/internal/repositories"
	"gorm.io/gorm"
)

// OrderService handles order-related business logic
type OrderService struct {
	DB                  *gorm.DB
	OrderRepo           *repositories.OrderRepository
	ProductService      *ProductService
	UserService         *UserService
	NotificationService *NotificationService
}

// NewOrderService creates a new instance of OrderService
func NewOrderService(db *gorm.DB, productService *ProductService, userService *UserService, notificationService *NotificationService) *OrderService {
	return &OrderService{
		DB:                  db,
		OrderRepo:           repositories.NewOrderRepository(db),
		ProductService:      productService,
		UserService:         userService,
		NotificationService: notificationService,
	}
}

// OrderResult represents the result of an order operation
type OrderResult struct {
	Success        bool
	Message        string
	Error          string
	OrderID        uuid.UUID
	Status         order.OrderStatus
	Total          float64
	DiscountAmount float64
	DiscountReason string
	FinalTotal     float64
	CreatedBy      *uuid.UUID
}

// OrderItemInfo represents information about an order item
type OrderItemInfo struct {
	InventoryID uuid.UUID
	Quantity    int
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(id uuid.UUID) (*order.Order, error) {
	return s.OrderRepo.GetOrderByID(id)
}

// GetAllOrders retrieves all orders with pagination and filtering
func (s *OrderService) GetAllOrders(page, pageSize int, filters map[string]interface{}) ([]order.Order, int64, error) {
	return s.OrderRepo.GetAllOrders(page, pageSize, filters)
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(
	paymentMethod order.PaymentMethod,
	items []OrderItemInfo,
	discountAmount float64,
	discountReason string,
	createdByID *uuid.UUID,
	shippingAddress string,
	shippingWard string,
	shippingDistrict string,
	shippingCity string,
	shippingCountry string,
	customerName string,
	customerEmail string,
	customerPhone string,
	notes string,
) (*OrderResult, error) {
	// Validate input
	if createdByID == nil {
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "Created by ID is required",
		}, fmt.Errorf("created by ID is required")
	}

	if len(items) == 0 {
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "At least one item is required",
		}, fmt.Errorf("at least one item is required")
	}

	// Check inventory availability for all items
	for _, item := range items {
		available, err := s.ProductService.CheckInventoryAvailability(item.InventoryID, item.Quantity)
		if err != nil {
			return &OrderResult{
				Success: false,
				Message: "Order creation failed",
				Error:   "Inventory not found",
			}, err
		}

		if !available {
			inventory, _ := s.ProductService.GetInventoryByID(item.InventoryID)
			return &OrderResult{
				Success: false,
				Message: "Order creation failed",
				Error:   fmt.Sprintf("Not enough inventory for product %s", inventory.ProductID),
			}, fmt.Errorf("not enough inventory for product %s", inventory.ProductID)
		}
	}

	// Start transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return &OrderResult{
			Success: false,
			Message: "Order creation failed",
			Error:   "Database transaction error",
		}, tx.Error
	}

	// Create order
	o := &order.Order{
		PaymentMethod:    paymentMethod,
		OrderStatus:      order.OrderPendingConfirmation,
		TotalAmount:      0,
		DiscountAmount:   discountAmount,
		DiscountReason:   discountReason,
		FinalTotalAmount: 0, // Will be calculated later
		Notes:            notes,
		// Shipping address fields
		ShippingAddress:  shippingAddress,
		ShippingWard:     shippingWard,
		ShippingDistrict: shippingDistrict,
		ShippingCity:     shippingCity,
		ShippingCountry:  shippingCountry,
		// Customer information
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		CustomerPhone: customerPhone,
	}

	// Set created by if provided
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
	totalAmount := 0.0
	for _, item := range items {
		// Get inventory for product ID
		inventory, err := s.ProductService.GetInventoryByID(item.InventoryID)
		if err != nil {
			tx.Rollback()
			return &OrderResult{
				Success: false,
				Message: "Order creation failed",
				Error:   "Inventory not found",
			}, err
		}

		// Get current price
		price, err := s.ProductService.GetCurrentPrice(inventory.ProductID)
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

	// Calculate final total amount (after discount)
	o.FinalTotalAmount = totalAmount - discountAmount
	if o.FinalTotalAmount < 0 {
		o.FinalTotalAmount = 0 // Ensure final amount is not negative
	}

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

	// Create a default shipment for the order
	shipment := &order.Shipment{
		OrderID: o.ID,
		// TrackingNumber and Carrier will be empty initially
	}
	if err := s.DB.Create(shipment).Error; err != nil {
		// Log the error but don't fail the order creation
		log.Printf("Failed to create default shipment for order %s: %v", o.ID, err)
	}

	// Send notification
	if s.NotificationService != nil {
		metadata := map[string]interface{}{
			"order_id":        o.ID.String(),
			"created_by":      createdByID.String(),
			"payment_method":  string(paymentMethod),
			"order_status":    string(o.OrderStatus),
			"total_amount":    totalAmount,
			"discount_amount": discountAmount,
			"final_amount":    o.FinalTotalAmount,
			"number_of_items": len(items),
		}

		notificationResult, err := s.NotificationService.CreateOrderNotification(o.ID, *createdByID, "created", metadata)
		if err != nil {
			log.Printf("Failed to create order notification: %v", err)
		}
		log.Println("CreateOrderNotification result day ne ma", notificationResult)
	}

	return &OrderResult{
		Success:        true,
		Message:        "Order created successfully",
		OrderID:        o.ID,
		Status:         o.OrderStatus,
		Total:          totalAmount,
		DiscountAmount: discountAmount,
		DiscountReason: discountReason,
		FinalTotal:     o.FinalTotalAmount,
		CreatedBy:      createdByID,
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(id uuid.UUID, status order.OrderStatus) (*OrderResult, error) {
	// Get the order
	o, err := s.OrderRepo.GetOrderByID(id)
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
	tx := s.DB.Begin()
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
	if s.NotificationService != nil {
		metadata := map[string]interface{}{
			"order_id":   o.ID.String(),
			"created_by": o.CreatedBy.String(),
			"old_status": string(oldStatus),
			"new_status": string(status),
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

		s.NotificationService.CreateOrderNotification(o.ID, *o.CreatedBy, event, metadata)
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

// handleInventoryForStatusChange handles inventory changes based on order status changes
func (s *OrderService) handleInventoryForStatusChange(tx *gorm.DB, o *order.Order, oldStatus, newStatus order.OrderStatus) error {
	// Get order items
	items, err := s.OrderRepo.GetOrderItemsByOrderID(o.ID)
	if err != nil {
		return err
	}

	// Handle inventory changes based on status transition
	switch {
	// When transitioning to packed status, reduce inventory
	case newStatus == order.OrderPacked:
		for _, item := range items {
			if err := s.ProductService.ReserveInventory(item.InventoryID, item.Quantity); err != nil {
				return err
			}
		}

	// When transitioning to returned status, increase inventory
	case newStatus == order.OrderReturned:
		for _, item := range items {
			if err := s.ProductService.ReleaseInventory(item.InventoryID, item.Quantity); err != nil {
				return err
			}
		}

	// When canceling an order that was in packed or shipped status, increase inventory
	case newStatus == order.OrderCanceled && (oldStatus == order.OrderPacked || oldStatus == order.OrderShipped):
		for _, item := range items {
			if err := s.ProductService.ReleaseInventory(item.InventoryID, item.Quantity); err != nil {
				return err
			}
		}
	}

	return nil
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(oldStatus, newStatus order.OrderStatus) bool {
	// Allow transition to canceled from most statuses except a few
	if newStatus == order.OrderCanceled {
		// Cannot cancel if already returned, in return processing, or delivered
		return oldStatus != order.OrderReturned &&
			oldStatus != order.OrderReturnProcessing &&
			oldStatus != order.OrderDelivered
	}

	// Define valid transitions for normal flow
	validTransitions := map[order.OrderStatus][]order.OrderStatus{
		order.OrderPendingConfirmation: {
			order.OrderConfirmed,
			order.OrderCanceled,
		},
		order.OrderConfirmed: {
			order.OrderShipmentRequested,
			order.OrderPacked,
			order.OrderCanceled,
		},
		order.OrderShipmentRequested: {
			order.OrderPacked,
			order.OrderCanceled,
		},
		order.OrderPacked: {
			order.OrderShipped,
			order.OrderCanceled,
		},
		order.OrderShipped: {
			order.OrderDelivered,
			order.OrderCanceled,
		},
		order.OrderDelivered: {
			order.OrderReturnProcessing,
		},
		order.OrderReturnProcessing: {
			order.OrderReturned,
		},
		order.OrderReturned: {
			// No further transitions allowed
		},
		order.OrderCanceled: {
			// No further transitions allowed
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

// DeleteOrder deletes an order
func (s *OrderService) DeleteOrder(id uuid.UUID) (*OrderResult, error) {
	// Get the order
	o, err := s.OrderRepo.GetOrderByID(id)
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
	tx := s.DB.Begin()
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
	o, err := s.OrderRepo.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	// Check if order status allows shipment
	if o.OrderStatus != order.OrderConfirmed && o.OrderStatus != order.OrderShipmentRequested && o.OrderStatus != order.OrderPacked {
		return fmt.Errorf("order status does not allow shipment creation")
	}

	// Check if shipment already exists
	existingShipment, err := s.OrderRepo.GetShipmentByOrderID(orderID)
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
	tx := s.DB.Begin()
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
	if s.NotificationService != nil {
		metadata := map[string]interface{}{
			"order_id":        o.ID.String(),
			"created_by":      o.CreatedBy.String(),
			"tracking_number": trackingNumber,
			"carrier":         carrier,
		}

		s.NotificationService.CreateOrderNotification(o.ID, *o.CreatedBy, "shipment_created", metadata)
	}

	return nil
}

// UpdateShipment updates the shipment details for an order
func (s *OrderService) UpdateShipment(orderID uuid.UUID, trackingNumber, carrier string) error {
	// Get the shipment
	shipment, err := s.OrderRepo.GetShipmentByOrderID(orderID)
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
	if err := s.OrderRepo.UpdateShipment(shipment); err != nil {
		return err
	}

	// Send notification
	if s.NotificationService != nil {
		// Get the order
		o, err := s.OrderRepo.GetOrderByID(orderID)
		if err == nil && o.CreatedBy != nil {
			metadata := map[string]interface{}{
				"order_id":        o.ID.String(),
				"created_by":      o.CreatedBy.String(),
				"tracking_number": shipment.TrackingNumber,
				"carrier":         shipment.Carrier,
			}

			s.NotificationService.CreateOrderNotification(o.ID, *o.CreatedBy, "shipment_updated", metadata)
		}
	}

	return nil
}

// DeleteShipment deletes a shipment
func (s *OrderService) DeleteShipment(orderID uuid.UUID) error {
	// Get the shipment
	shipment, err := s.OrderRepo.GetShipmentByOrderID(orderID)
	if err != nil {
		return err
	}

	// Delete shipment
	return s.OrderRepo.DeleteShipment(shipment.ID)
}

// AddOrderItem adds an item to an order
func (s *OrderService) AddOrderItem(orderID uuid.UUID, inventoryID uuid.UUID, quantity int) error {
	// Get the order
	o, err := s.OrderRepo.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	// Check if order status allows adding items
	if o.OrderStatus != order.OrderPendingConfirmation && o.OrderStatus != order.OrderConfirmed {
		return fmt.Errorf("order status does not allow adding items")
	}

	// Check inventory availability
	available, err := s.ProductService.CheckInventoryAvailability(inventoryID, quantity)
	if err != nil {
		return err
	}

	if !available {
		return fmt.Errorf("not enough inventory")
	}

	// Get inventory for product ID
	inventory, err := s.ProductService.GetInventoryByID(inventoryID)
	if err != nil {
		return err
	}

	// Get current price
	price, err := s.ProductService.GetCurrentPrice(inventory.ProductID)
	if err != nil {
		return err
	}

	// Start transaction
	tx := s.DB.Begin()
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
	item, err := s.OrderRepo.GetOrderItemByID(id)
	if err != nil {
		return err
	}

	// Get the order
	o, err := s.OrderRepo.GetOrderByID(item.OrderID)
	if err != nil {
		return err
	}

	// Check if order status allows updating items
	if o.OrderStatus != order.OrderPendingConfirmation && o.OrderStatus != order.OrderConfirmed {
		return fmt.Errorf("order status does not allow updating items")
	}

	// If quantity is increasing, check inventory availability
	if quantity > item.Quantity {
		available, err := s.ProductService.CheckInventoryAvailability(item.InventoryID, quantity-item.Quantity)
		if err != nil {
			return err
		}

		if !available {
			return fmt.Errorf("not enough inventory")
		}
	}

	// Start transaction
	tx := s.DB.Begin()
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
	item, err := s.OrderRepo.GetOrderItemByID(id)
	if err != nil {
		return err
	}

	// Get the order
	o, err := s.OrderRepo.GetOrderByID(item.OrderID)
	if err != nil {
		return err
	}

	// Check if order status allows deleting items
	if o.OrderStatus != order.OrderPendingConfirmation && o.OrderStatus != order.OrderConfirmed {
		return fmt.Errorf("order status does not allow deleting items")
	}

	// Start transaction
	tx := s.DB.Begin()
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

// UpdateOrderDetails updates the details of an order
func (s *OrderService) UpdateOrderDetails(
	id uuid.UUID,
	notes string,
	paymentMethod order.PaymentMethod,
	discountAmount float64,
	discountReason string,
	shippingAddress string,
	shippingWard string,
	shippingDistrict string,
	shippingCity string,
	shippingCountry string,
	customerName string,
	customerEmail string,
	customerPhone string,
) (*OrderResult, error) {
	// Get the order
	o, err := s.OrderRepo.GetOrderByID(id)
	if err != nil {
		return &OrderResult{
			Success: false,
			Message: "Order details update failed",
			Error:   "Order not found",
		}, err
	}

	// Update fields if provided
	if paymentMethod != "" {
		o.PaymentMethod = paymentMethod
	}

	// Update notes if provided
	if notes != "" {
		o.Notes = notes
	}

	// Update discount if provided
	if discountAmount >= 0 {
		o.DiscountAmount = discountAmount
		o.DiscountReason = discountReason
		// Recalculate final total
		o.FinalTotalAmount = o.TotalAmount - o.DiscountAmount
		if o.FinalTotalAmount < 0 {
			o.FinalTotalAmount = 0 // Ensure final amount is not negative
		}
	}

	// Update shipping address if provided
	if shippingAddress != "" {
		o.ShippingAddress = shippingAddress
	}
	if shippingWard != "" {
		o.ShippingWard = shippingWard
	}
	if shippingDistrict != "" {
		o.ShippingDistrict = shippingDistrict
	}
	if shippingCity != "" {
		o.ShippingCity = shippingCity
	}
	if shippingCountry != "" {
		o.ShippingCountry = shippingCountry
	}

	// Update customer information if provided
	if customerName != "" {
		o.CustomerName = customerName
	}
	if customerEmail != "" {
		o.CustomerEmail = customerEmail
	}
	if customerPhone != "" {
		o.CustomerPhone = customerPhone
	}

	// Save the order
	if err := s.OrderRepo.UpdateOrder(o); err != nil {
		return &OrderResult{
			Success: false,
			Message: "Order details update failed",
			Error:   "Error updating order details",
		}, err
	}

	// Send notification
	if s.NotificationService != nil {
		metadata := map[string]interface{}{
			"order_id":        o.ID.String(),
			"payment_method":  string(o.PaymentMethod),
			"discount_amount": o.DiscountAmount,
			"discount_reason": o.DiscountReason,
			"final_amount":    o.FinalTotalAmount,
		}

		if o.CreatedBy != nil {
			s.NotificationService.CreateOrderNotification(o.ID, *o.CreatedBy, "details_updated", metadata)
		}
	}

	return &OrderResult{
		Success:        true,
		Message:        "Order details updated successfully",
		OrderID:        o.ID,
		Status:         o.OrderStatus,
		Total:          o.TotalAmount,
		DiscountAmount: o.DiscountAmount,
		DiscountReason: o.DiscountReason,
		FinalTotal:     o.FinalTotalAmount,
		CreatedBy:      o.CreatedBy,
	}, nil
}

// GetOrderByTrackingNumber retrieves an order by shipment tracking number
func (s *OrderService) GetOrderByTrackingNumber(trackingNumber string) (*order.Order, error) {
	if trackingNumber == "" {
		return nil, fmt.Errorf("tracking number is required")
	}
	return s.OrderRepo.GetOrderByTrackingNumber(trackingNumber)
}
