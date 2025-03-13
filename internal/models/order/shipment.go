package order

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// Shipment represents a shipment for an order
type Shipment struct {
	models.Base
	OrderID        uuid.UUID `gorm:"column:order_id;type:uuid;not null;uniqueIndex" json:"order_id"`
	TrackingNumber string    `gorm:"column:tracking_number;type:varchar(100)" json:"tracking_number"`
	Carrier        string    `gorm:"column:carrier;type:varchar(50)" json:"carrier"`
	Order          Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// TableName specifies the table name for Shipment
func (Shipment) TableName() string {
	return "shipments"
}
