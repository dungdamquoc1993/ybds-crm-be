package account

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	models.Base
	UserID uuid.UUID `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	RoleID uuid.UUID `gorm:"column:role_id;type:uuid;not null" json:"role_id"`
	User   User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role   Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}
