package account

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ybds/internal/models"
	"gorm.io/gorm"
)

// Address represents a shipping address for a user or guest
type Address struct {
	models.Base
	UserID    *uuid.UUID `gorm:"column:user_id;type:uuid;null;index" json:"user_id,omitempty"`
	GuestID   *uuid.UUID `gorm:"column:guest_id;type:uuid;null;index" json:"guest_id,omitempty"`
	Address   string     `gorm:"column:address;type:text;not null" json:"address"`
	Ward      string     `gorm:"column:ward;type:varchar(100)" json:"ward"`
	District  string     `gorm:"column:district;type:varchar(100)" json:"district"`
	City      string     `gorm:"column:city;type:varchar(100);not null" json:"city"`
	Country   string     `gorm:"column:country;type:varchar(100);not null;default:'Vietnam'" json:"country"`
	IsDefault bool       `gorm:"column:is_default;not null;default:false;index" json:"is_default"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Guest     *Guest     `gorm:"foreignKey:GuestID" json:"guest,omitempty"`
}

// TableName specifies the table name for Address
func (Address) TableName() string {
	return "user_addresses"
}

// BeforeCreate validates the address data before creating
func (a *Address) BeforeCreate(tx *gorm.DB) error {
	// Ensure address is associated with either a User or a Guest, but not both or neither
	if (a.UserID == nil && a.GuestID == nil) || (a.UserID != nil && a.GuestID != nil) {
		return errors.New("address must be associated with either a user or a guest, but not both or neither")
	}
	return nil
}

// BeforeUpdate validates the address data before updating
func (a *Address) BeforeUpdate(tx *gorm.DB) error {
	// Ensure address is associated with either a User or a Guest, but not both or neither
	if (a.UserID == nil && a.GuestID == nil) || (a.UserID != nil && a.GuestID != nil) {
		return errors.New("address must be associated with either a user or a guest, but not both or neither")
	}
	return nil
}
