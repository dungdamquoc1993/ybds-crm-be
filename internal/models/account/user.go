package account

import (
	"github.com/ybds/internal/models"
)

// User represents a registered user in the system
type User struct {
	models.Base
	Username     string `gorm:"column:username;type:varchar(255);not null;uniqueIndex" json:"username"`
	Email        string `gorm:"column:email;type:varchar(255);not null;uniqueIndex" json:"email"`
	Phone        string `gorm:"column:phone;type:varchar(20);index" json:"phone"`
	PasswordHash string `gorm:"column:password_hash;type:text;not null" json:"-"`
	Salt         string `gorm:"column:salt;type:text;not null" json:"-"`
	IsActive     bool   `gorm:"column:is_active;not null;default:true;index" json:"is_active"`
	Roles        []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}
