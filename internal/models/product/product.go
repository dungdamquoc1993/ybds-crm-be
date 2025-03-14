package product

import (
	"github.com/ybds/internal/models"
)

// Product represents a product in the system
type Product struct {
	models.Base
	Name        string         `gorm:"column:name;type:varchar(255);not null;index" json:"name"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	SKU         string         `gorm:"column:sku;type:varchar(50);not null;uniqueIndex" json:"sku"`
	Category    string         `gorm:"column:category;type:varchar(100);not null;index" json:"category"`
	ImageURL    string         `gorm:"column:image_url;type:text" json:"image_url"`
	Inventory   []Inventory    `gorm:"foreignKey:ProductID" json:"inventory,omitempty"`
	Prices      []Price        `gorm:"foreignKey:ProductID" json:"prices,omitempty"`
	Images      []ProductImage `gorm:"foreignKey:ProductID" json:"images,omitempty"`
}

// TableName specifies the table name for Product
func (Product) TableName() string {
	return "products"
}
