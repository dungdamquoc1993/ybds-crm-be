package responses

import (
	"time"

	"github.com/google/uuid"
)

// OrderItemResponse represents an order item in responses
type OrderItemResponse struct {
	ID           uuid.UUID `json:"id"`
	OrderID      uuid.UUID `json:"order_id"`
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	ProductImage string    `json:"product_image"`
	InventoryID  uuid.UUID `json:"inventory_id"`
	Size         string    `json:"size"`
	Color        string    `json:"color"`
	PriceID      uuid.UUID `json:"price_id"`
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	Quantity     int       `json:"quantity"`
	Subtotal     float64   `json:"subtotal"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ShipmentResponse represents a shipment in responses
type ShipmentResponse struct {
	ID             uuid.UUID `json:"id"`
	OrderID        uuid.UUID `json:"order_id"`
	TrackingNumber string    `json:"tracking_number"`
	Carrier        string    `json:"carrier"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OrderResponse represents an order in responses
type OrderResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    OrderDetail `json:"data"`
}

// OrderDetail represents the details of an order
type OrderDetail struct {
	ID               uuid.UUID           `json:"id"`
	CustomerName     string              `json:"customer_name"`
	CustomerEmail    string              `json:"customer_email"`
	CustomerPhone    string              `json:"customer_phone"`
	ShippingAddress  string              `json:"shipping_address"`
	ShippingWard     string              `json:"shipping_ward"`
	ShippingDistrict string              `json:"shipping_district"`
	ShippingCity     string              `json:"shipping_city"`
	ShippingCountry  string              `json:"shipping_country"`
	PaymentMethod    string              `json:"payment_method"`
	Status           string              `json:"status"`
	Notes            string              `json:"notes"`
	Total            float64             `json:"total"`
	DiscountAmount   float64             `json:"discount_amount"`
	DiscountReason   string              `json:"discount_reason"`
	FinalTotal       float64             `json:"final_total"`
	CreatedBy        uuid.UUID           `json:"created_by"`
	CreatedByName    string              `json:"created_by_name"`
	Items            []OrderItemResponse `json:"items,omitempty"`
	Shipment         *ShipmentResponse   `json:"shipment,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// OrdersResponse represents a list of orders in responses
type OrdersResponse struct {
	Success    bool          `json:"success"`
	Message    string        `json:"message"`
	Data       []OrderDetail `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int64         `json:"total_pages"`
}

// OrderDetailResponse represents a detailed order in responses
type OrderDetailResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    OrderDetail `json:"data"`
}

// OrderItemDetailResponse represents a detailed order item in responses
type OrderItemDetailResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    OrderItemResponse `json:"data"`
}
