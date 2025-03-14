package order

import (
	"github.com/ybds/internal/models"
)

// PaymentMethod defines the payment method for an order
type PaymentMethod string

const (
	// PaymentCash represents a cash payment
	PaymentCash PaymentMethod = "cash"
	// PaymentCOD represents a cash on delivery payment
	PaymentCOD PaymentMethod = "cod"
	// PaymentBankTransfer represents a bank transfer payment
	PaymentBankTransfer PaymentMethod = "bank_transfer"
)

// OrderStatus defines the status of an order
type OrderStatus string

const (
	// OrderPendingConfirmation means the order is waiting for confirmation
	OrderPendingConfirmation OrderStatus = "pending_confirmation"
	// OrderConfirmed means the order has been confirmed
	OrderConfirmed OrderStatus = "confirmed"
	// OrderShipmentRequested means a shipment has been requested for the order
	OrderShipmentRequested OrderStatus = "shipment_requested"
	// OrderPacking means the order is being packed
	OrderPacking OrderStatus = "packing"
	// OrderShipped means the order has been shipped
	OrderShipped OrderStatus = "shipped"
	// OrderDelivered means the order has been delivered
	OrderDelivered OrderStatus = "delivered"
	// OrderReturnRequested means a return has been requested for the order
	OrderReturnRequested OrderStatus = "return_requested"
	// OrderReturnProcessing means the order return is being processed
	OrderReturnProcessing OrderStatus = "return_processing"
	// OrderReturned means the order has been returned
	OrderReturned OrderStatus = "returned"
	// OrderCanceled means the order has been canceled
	OrderCanceled OrderStatus = "canceled"
)

// Order represents an order in the system
type Order struct {
	models.Base
	PaymentMethod    PaymentMethod `gorm:"column:payment_method;type:varchar(50);not null;index" json:"payment_method"`
	TotalAmount      float64       `gorm:"column:total_amount;type:decimal(10,2);not null" json:"total_amount"`
	DiscountAmount   float64       `gorm:"column:discount_amount;type:decimal(10,2);not null;default:0" json:"discount_amount"`
	DiscountReason   string        `gorm:"column:discount_reason;type:varchar(255)" json:"discount_reason"`
	FinalTotalAmount float64       `gorm:"column:final_total_amount;type:decimal(10,2);not null" json:"final_total_amount"`
	OrderStatus      OrderStatus   `gorm:"column:order_status;type:varchar(50);not null;default:'pending_confirmation';index" json:"order_status"`
	// Shipping address fields
	ShippingAddress  string `gorm:"column:shipping_address;type:text" json:"shipping_address"`
	ShippingWard     string `gorm:"column:shipping_ward;type:varchar(100)" json:"shipping_ward"`
	ShippingDistrict string `gorm:"column:shipping_district;type:varchar(100)" json:"shipping_district"`
	ShippingCity     string `gorm:"column:shipping_city;type:varchar(100)" json:"shipping_city"`
	ShippingCountry  string `gorm:"column:shipping_country;type:varchar(100);default:'Vietnam'" json:"shipping_country"`
	// Customer contact information
	CustomerName  string `gorm:"column:customer_name;type:varchar(255)" json:"customer_name"`
	CustomerEmail string `gorm:"column:customer_email;type:varchar(255)" json:"customer_email"`
	CustomerPhone string `gorm:"column:customer_phone;type:varchar(20)" json:"customer_phone"`
	// Relationships
	Items    []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Shipment *Shipment   `gorm:"foreignKey:OrderID" json:"shipment,omitempty"`
}

// TableName specifies the table name for Order
func (Order) TableName() string {
	return "orders"
}
