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
	// OrderShipmentRequested means a shipment has been requested for the order
	// This is the initial status when an order is created
	OrderShipmentRequested OrderStatus = "shipment_requested"

	// OrderPacked means the order has been packed and is ready for shipping
	OrderPacked OrderStatus = "packed"

	// OrderPicked means the order has been picked up by the shipping company corresponding to the picked status in GHN
	OrderPicked OrderStatus = "picked"

	// OrderDelivering means the order is being delivered by the shipping company
	// corresponding to the status in GHN: storing, transporting, delivering, delivery_fail, waiting_to_return
	OrderDelivering OrderStatus = "delivering" // this status is now active in the application

	// OrderDelivered means the order has been delivered by the shipping company corresponding to the delivered status in GHN
	OrderDelivered OrderStatus = "delivered"

	// OrderReturnProcessing means the order return is being processed by the shipping company
	// corresponding to the status in GHN: return, return_transporting, returning, retrun_fail
	OrderReturnProcessing OrderStatus = "return_processing"

	// OrderReturned means the order has been returned by the shipping company corresponding to the returned status in GHN
	OrderReturned OrderStatus = "returned"

	// OrderCanceled means the order has been canceled corresponding to the canceled status in GHN
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
	OrderStatus      OrderStatus   `gorm:"column:order_status;type:varchar(50);not null;default:'shipment_requested';index" json:"order_status"`
	Notes            string        `gorm:"column:notes;type:text" json:"notes"`
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
