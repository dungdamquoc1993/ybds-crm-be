package requests

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// OrderItemInfo represents an item to be added to an order
type OrderItemInfo struct {
	ProductID   uuid.UUID `json:"product_id"`
	InventoryID uuid.UUID `json:"inventory_id"`
	PriceID     uuid.UUID `json:"price_id"`
	Quantity    int       `json:"quantity"`
	Notes       string    `json:"notes"`
}

// Validate validates the order item info
func (i *OrderItemInfo) Validate() error {
	if i.ProductID == uuid.Nil {
		return errors.New("product ID is required")
	}
	if i.InventoryID == uuid.Nil {
		return errors.New("inventory ID is required")
	}
	if i.PriceID == uuid.Nil {
		return errors.New("price ID is required")
	}
	if i.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

// CreateOrderRequest represents a request to create a new order
type CreateOrderRequest struct {
	CustomerID        uuid.UUID       `json:"customer_id"`
	ShippingAddressID uuid.UUID       `json:"shipping_address_id"`
	PaymentMethod     string          `json:"payment_method"`
	Status            string          `json:"status"`
	Notes             string          `json:"notes"`
	Items             []OrderItemInfo `json:"items"`
}

// Validate validates the create order request
func (r *CreateOrderRequest) Validate() error {
	if r.CustomerID == uuid.Nil {
		return errors.New("customer ID is required")
	}
	if r.ShippingAddressID == uuid.Nil {
		return errors.New("shipping address ID is required")
	}
	if r.PaymentMethod == "" {
		return errors.New("payment method is required")
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

// UpdatePaymentStatusRequest represents a request to update an order's payment status
type UpdatePaymentStatusRequest struct {
	PaymentStatus    string `json:"payment_status"`
	PaymentReference string `json:"payment_reference"`
}

// Validate validates the update payment status request
func (r *UpdatePaymentStatusRequest) Validate() error {
	if r.PaymentStatus == "" {
		return errors.New("payment status is required")
	}
	return nil
}

// AddOrderItemRequest represents a request to add an item to an order
type AddOrderItemRequest struct {
	ProductID   uuid.UUID `json:"product_id"`
	InventoryID uuid.UUID `json:"inventory_id"`
	PriceID     uuid.UUID `json:"price_id"`
	Quantity    int       `json:"quantity"`
	Notes       string    `json:"notes"`
}

// Validate validates the add order item request
func (r *AddOrderItemRequest) Validate() error {
	if r.ProductID == uuid.Nil {
		return errors.New("product ID is required")
	}
	if r.InventoryID == uuid.Nil {
		return errors.New("inventory ID is required")
	}
	if r.PriceID == uuid.Nil {
		return errors.New("price ID is required")
	}
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}

// UpdateOrderItemRequest represents a request to update an order item
type UpdateOrderItemRequest struct {
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes"`
}

// Validate validates the update order item request
func (r *UpdateOrderItemRequest) Validate() error {
	if r.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	return nil
}
