package order

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// CustomerType defines the type of customer
type CustomerType string

const (
	// CustomerUser represents a registered user
	CustomerUser CustomerType = "user"
	// CustomerGuest represents a guest customer
	CustomerGuest CustomerType = "guest"
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
	CustomerID    uuid.UUID     `gorm:"column:customer_id;type:uuid;not null;index" json:"customer_id"`
	CustomerType  CustomerType  `gorm:"column:customer_type;type:varchar(50);not null;index" json:"customer_type"`
	PaymentMethod PaymentMethod `gorm:"column:payment_method;type:varchar(50);not null;index" json:"payment_method"`
	PaymentStatus string        `gorm:"column:payment_status;type:varchar(50);not null;default:'pending';index" json:"payment_status"`
	TotalAmount   float64       `gorm:"column:total_amount;type:decimal(10,2);not null" json:"total_amount"`
	PaidAmount    float64       `gorm:"column:paid_amount;type:decimal(10,2);default:0" json:"paid_amount"`
	OrderStatus   OrderStatus   `gorm:"column:order_status;type:varchar(50);not null;default:'pending_confirmation';index" json:"order_status"`
	Items         []OrderItem   `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Shipment      *Shipment     `gorm:"foreignKey:OrderID" json:"shipment,omitempty"`
}

// TableName specifies the table name for Order
func (Order) TableName() string {
	return "orders"
}
