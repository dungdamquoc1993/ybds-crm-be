package account

import (
	"github.com/ybds/internal/models"
)

// Guest represents a non-registered customer
type Guest struct {
	models.Base
	Name      string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Phone     string    `gorm:"column:phone;type:varchar(20);not null" json:"phone"`
	Email     string    `gorm:"column:email;type:varchar(255)" json:"email"`
	Addresses []Address `gorm:"foreignKey:GuestID" json:"addresses,omitempty"`
}

// TableName specifies the table name for Guest
func (Guest) TableName() string {
	return "guests"
}
