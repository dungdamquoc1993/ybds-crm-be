package repositories

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models/account"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetUserByID retrieves a user by ID with all relations
func (r *UserRepository) GetUserByID(id uuid.UUID) (*account.User, error) {
	var user account.User
	err := r.db.Where("id = ?", id).
		Preload("Roles").
		First(&user).Error
	return &user, err
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*account.User, error) {
	var user account.User
	err := r.db.Where("username = ?", username).
		Preload("Roles").
		First(&user).Error
	return &user, err
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*account.User, error) {
	var user account.User
	err := r.db.Where("email = ?", email).
		Preload("Roles").
		First(&user).Error
	return &user, err
}

// GetUserByPhone retrieves a user by phone
func (r *UserRepository) GetUserByPhone(phone string) (*account.User, error) {
	var user account.User
	err := r.db.Where("phone = ?", phone).
		Preload("Roles").
		First(&user).Error
	return &user, err
}

// GetUserByEmailOrPhone retrieves a user by email or phone
func (r *UserRepository) GetUserByEmailOrPhone(email, phone string) (*account.User, error) {
	var user account.User
	query := r.db.Model(&account.User{})
	if email != "" {
		query = query.Where("email = ?", email)
	}
	if phone != "" {
		query = query.Or("phone = ?", phone)
	}
	err := query.Preload("Roles").First(&user).Error
	return &user, err
}

// GetAllUsers retrieves all users with pagination
func (r *UserRepository) GetAllUsers(page, pageSize int) ([]account.User, int64, error) {
	var users []account.User
	var total int64

	// Count total records
	if err := r.db.Model(&account.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).
		Preload("Roles").
		Find(&users).Error

	return users, total, err
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(user *account.User) error {
	return r.db.Create(user).Error
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(user *account.User) error {
	return r.db.Save(user).Error
}

// DeleteUser deletes a user by ID
func (r *UserRepository) DeleteUser(id uuid.UUID) error {
	return r.db.Delete(&account.User{}, id).Error
}

// GetRoleByName retrieves a role by name
func (r *UserRepository) GetRoleByName(name account.RoleType) (*account.Role, error) {
	var role account.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, err
}

// CreateRole creates a new role
func (r *UserRepository) CreateRole(role *account.Role) error {
	return r.db.Create(role).Error
}

// AssignRoleToUser assigns a role to a user
func (r *UserRepository) AssignRoleToUser(userID, roleID uuid.UUID) error {
	userRole := account.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.Create(&userRole).Error
}

// RemoveRoleFromUser removes a role from a user
func (r *UserRepository) RemoveRoleFromUser(userID, roleID uuid.UUID) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&account.UserRole{}).Error
}
