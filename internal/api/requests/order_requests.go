package requests

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// OrderItemInfo represents an item to be added to an order
type OrderItemInfo struct {
	InventoryID uuid.UUID `json:"inventory_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Quantity    int       `json:"quantity" example:"2"`
}

// Validate validates the order item info
func (i *OrderItemInfo) Validate() error {
	if i.InventoryID == uuid.Nil {
		return errors.New("inventory ID is required")
	}
	if i.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

// CreateOrderRequest represents a request to create a new order
type CreateOrderRequest struct {
	PaymentMethod  string          `json:"payment_method" example:"cash"`
	Status         string          `json:"status" example:"pending_confirmation"`
	Notes          string          `json:"notes" example:"Please deliver in the morning"`
	DiscountAmount float64         `json:"discount_amount" example:"10.50"`
	DiscountReason string          `json:"discount_reason" example:"Loyalty discount"`
	Items          []OrderItemInfo `json:"items" required:"true"`
	// Shipping address information
	ShippingAddress  string `json:"shipping_address" example:"123 Main St"`
	ShippingWard     string `json:"shipping_ward" example:"Ward 1"`
	ShippingDistrict string `json:"shipping_district" example:"District 1"`
	ShippingCity     string `json:"shipping_city" example:"Ho Chi Minh City"`
	ShippingCountry  string `json:"shipping_country" example:"Vietnam"`
	// Customer information
	CustomerName  string `json:"customer_name" example:"John Doe" required:"true"`
	CustomerEmail string `json:"customer_email" example:"john@example.com"`
	CustomerPhone string `json:"customer_phone" example:"1234567890"`
}

// Validate validates the create order request
func (r *CreateOrderRequest) Validate() error {
	if r.CustomerName == "" {
		return errors.New("customer name is required")
	}

	if len(r.Items) == 0 {
		return errors.New("at least one item is required")
	}

	// Validate each item
	for i, item := range r.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("item %d: %s", i, err.Error())
		}
	}

	return nil
}

// UpdateOrderStatusRequest represents a request to update an order's status
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

// Validate validates the update order status request
func (r *UpdateOrderStatusRequest) Validate() error {
	if r.Status == "" {
		return errors.New("status is required")
	}
	return nil
}

// AddOrderItemRequest represents a request to add an item to an order
type AddOrderItemRequest struct {
	InventoryID uuid.UUID `json:"inventory_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Quantity    int       `json:"quantity" example:"2"`
}

// Validate validates the add order item request
func (r *AddOrderItemRequest) Validate() error {
	if r.InventoryID == uuid.Nil {
		return errors.New("inventory ID is required")
	}
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

// UpdateOrderItemRequest represents a request to update an order item
type UpdateOrderItemRequest struct {
	Quantity int `json:"quantity" example:"3"`
}

// Validate validates the update order item request
func (r *UpdateOrderItemRequest) Validate() error {
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

// UpdateOrderDetailsRequest represents a request to update order details
type UpdateOrderDetailsRequest struct {
	// Order information
	PaymentMethod  string  `json:"payment_method" example:"cash"`
	Notes          string  `json:"notes" example:"Please deliver in the morning"`
	DiscountAmount float64 `json:"discount_amount" example:"10.50"`
	DiscountReason string  `json:"discount_reason" example:"Free delivery"`
	// Shipping address information
	ShippingAddress  string `json:"shipping_address" example:"123 Main St"`
	ShippingWard     string `json:"shipping_ward" example:"Ward 1"`
	ShippingDistrict string `json:"shipping_district" example:"District 1"`
	ShippingCity     string `json:"shipping_city" example:"Ho Chi Minh City"`
	ShippingCountry  string `json:"shipping_country" example:"Vietnam"`
	// Customer information
	CustomerName  string `json:"customer_name" example:"John Doe"`
	CustomerEmail string `json:"customer_email" example:"john@example.com"`
	CustomerPhone string `json:"customer_phone" example:"1234567890"`
}

// Validate validates the update order details request
func (r *UpdateOrderDetailsRequest) Validate() error {
	// No validation needed as all fields are optional
	return nil
}

// UpdateShipmentRequest represents a request to update shipment details
type UpdateShipmentRequest struct {
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier"`
}

// Validate validates the UpdateShipmentRequest
func (r *UpdateShipmentRequest) Validate() error {
	if r.TrackingNumber == "" && r.Carrier == "" {
		return errors.New("at least one of tracking number or carrier is required")
	}
	return nil
}
