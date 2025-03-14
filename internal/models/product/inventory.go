package product

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// Inventory represents a product inventory entry
type Inventory struct {
	models.Base
	ProductID uuid.UUID `gorm:"column:product_id;type:uuid;not null;index" json:"product_id"`
	Size      string    `gorm:"column:size;type:varchar(10);index" json:"size"`
	Color     string    `gorm:"column:color;type:varchar(50);index" json:"color"`
	Quantity  int       `gorm:"column:quantity;not null;default:0;index" json:"quantity"`
	Location  string    `gorm:"column:location;type:varchar(255);index" json:"location"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"-"`
}

// TableName specifies the table name for Inventory
func (Inventory) TableName() string {
	return "inventory"
}
