package product

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// TransactionType defines the type of inventory transaction
type TransactionType string

const (
	// TransactionInbound represents an inbound inventory transaction (stock increase)
	TransactionInbound TransactionType = "inbound"
	// TransactionOutbound represents an outbound inventory transaction (stock decrease)
	TransactionOutbound TransactionType = "outbound"
	// TransactionReservation represents a reservation of inventory (temporary hold)
	TransactionReservation TransactionType = "reservation"
	// TransactionRelease represents a release of reserved inventory
	TransactionRelease TransactionType = "release"
	// TransactionAdjustment represents an inventory adjustment
	TransactionAdjustment TransactionType = "adjustment"
)

// TransactionReason defines the reason for an inventory transaction
type TransactionReason string

const (
	// ReasonPurchase represents a purchase from supplier
	ReasonPurchase TransactionReason = "purchase"
	// ReasonSale represents a sale to customer
	ReasonSale TransactionReason = "sale"
	// ReasonReturn represents a return from customer
	ReasonReturn TransactionReason = "return"
	// ReasonDamage represents damaged inventory
	ReasonDamage TransactionReason = "damage"
	// ReasonExpired represents expired inventory
	ReasonExpired TransactionReason = "expired"
	// ReasonStockCount represents a stock count adjustment
	ReasonStockCount TransactionReason = "stock_count"
	// ReasonReservation represents a reservation for an order
	ReasonReservation TransactionReason = "reservation"
	// ReasonOrderCancellation represents a cancellation of an order
	ReasonOrderCancellation TransactionReason = "order_cancellation"
)

// InventoryTransaction represents a transaction affecting inventory
type InventoryTransaction struct {
	models.Base
	InventoryID   uuid.UUID         `gorm:"column:inventory_id;type:uuid;not null;index" json:"inventory_id"`
	Quantity      int               `gorm:"column:quantity;not null" json:"quantity"`
	Type          TransactionType   `gorm:"column:type;type:varchar(50);not null;index" json:"type"`
	Reason        TransactionReason `gorm:"column:reason;type:varchar(50);not null;index" json:"reason"`
	ReferenceID   *uuid.UUID        `gorm:"column:reference_id;type:uuid;null;index" json:"reference_id,omitempty"`
	ReferenceType string            `gorm:"column:reference_type;type:varchar(50);null;index" json:"reference_type,omitempty"`
	Notes         string            `gorm:"column:notes;type:text" json:"notes,omitempty"`
	Inventory     Inventory         `gorm:"foreignKey:InventoryID" json:"inventory,omitempty"`
}

// TableName specifies the table name for InventoryTransaction
func (InventoryTransaction) TableName() string {
	return "inventory_transactions"
}
