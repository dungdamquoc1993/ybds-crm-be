package account

import (
	"github.com/ybds/internal/models"
)

// RoleType defines the type of role
type RoleType string

const (
	// RoleAdmin represents an admin role
	RoleAdmin RoleType = "admin"
	// RoleStaff represents a staff role
	RoleStaff RoleType = "staff"
	// RoleAgent represents an AI agent role
	RoleAgent RoleType = "agent"
)

// Role represents a role in the system
type Role struct {
	models.Base
	Name RoleType `gorm:"column:name;type:varchar(50);not null;uniqueIndex" json:"name"`
	// Users is a many-to-many relationship with User
	Users []User `gorm:"many2many:user_roles;" json:"users,omitempty"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}
