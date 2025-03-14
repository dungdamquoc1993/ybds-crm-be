package product

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// ProductImage represents an image associated with a product
type ProductImage struct {
	models.Base
	ProductID uuid.UUID `gorm:"column:product_id;type:uuid;not null;index" json:"product_id"`
	URL       string    `gorm:"column:url;type:text;not null" json:"url"`
	Filename  string    `gorm:"column:filename;type:varchar(255);not null" json:"filename"`
	IsPrimary bool      `gorm:"column:is_primary;type:boolean;default:false" json:"is_primary"`
	SortOrder int       `gorm:"column:sort_order;type:int;default:0" json:"sort_order"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"-"`
}

// TableName specifies the table name for ProductImage
func (ProductImage) TableName() string {
	return "product_images"
}
