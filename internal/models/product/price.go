package product

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ybds/internal/models"
	"gorm.io/gorm"
)

// Price represents a product price entry
type Price struct {
	models.Base
	ProductID uuid.UUID  `gorm:"column:product_id;type:uuid;not null;index" json:"product_id"`
	Price     float64    `gorm:"column:price;type:decimal(10,2);not null" json:"price"`
	Currency  string     `gorm:"column:currency;type:varchar(10);not null;default:'VND'" json:"currency"`
	StartDate time.Time  `gorm:"column:start_date;not null;index" json:"start_date"`
	EndDate   *time.Time `gorm:"column:end_date;index" json:"end_date,omitempty"`
	Product   Product    `gorm:"foreignKey:ProductID" json:"-"`
}

// TableName specifies the table name for Price
func (Price) TableName() string {
	return "prices"
}

// BeforeCreate validates the price data before creating
func (p *Price) BeforeCreate(tx *gorm.DB) error {
	// Validate that EndDate is after StartDate if EndDate is provided
	if p.EndDate != nil && !p.EndDate.After(p.StartDate) {
		return errors.New("end date must be after start date")
	}
	return nil
}

// BeforeUpdate validates the price data before updating
func (p *Price) BeforeUpdate(tx *gorm.DB) error {
	// Validate that EndDate is after StartDate if EndDate is provided
	if p.EndDate != nil && !p.EndDate.After(p.StartDate) {
		return errors.New("end date must be after start date")
	}
	return nil
}
