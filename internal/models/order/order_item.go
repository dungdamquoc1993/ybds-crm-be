package order

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// OrderItem represents an item in an order
type OrderItem struct {
	models.Base
	OrderID      uuid.UUID `gorm:"column:order_id;type:uuid;not null" json:"order_id"`
	InventoryID  uuid.UUID `gorm:"column:inventory_id;type:uuid;not null" json:"inventory_id"`
	Quantity     int       `gorm:"column:quantity;not null" json:"quantity"`
	PriceAtOrder float64   `gorm:"column:price_at_order;type:decimal(10,2);not null" json:"price_at_order"`
	Order        Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// TableName specifies the table name for OrderItem
func (OrderItem) TableName() string {
	return "order_items"
}
