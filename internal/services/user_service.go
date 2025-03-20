package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ybds/internal/models/account"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/repositories"
	passwordpkg "github.com/ybds/pkg/password"
	"gorm.io/gorm"
)

// UserService handles user-related business logic
type UserService struct {
	DB                  *gorm.DB
	UserRepo            *repositories.UserRepository
	NotificationService *NotificationService
}

// NewUserService creates a new instance of UserService
func NewUserService(db *gorm.DB, notificationService *NotificationService) *UserService {
	return &UserService{
		DB:                  db,
		UserRepo:            repositories.NewUserRepository(db),
		NotificationService: notificationService,
	}
}

// UserResult represents the result of a user operation
type UserResult struct {
	Success  bool
	Message  string
	Error    string
	UserID   uuid.UUID
	Username string
	Email    string
	Roles    []string
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id uuid.UUID) (*account.User, error) {
	return s.UserRepo.GetUserByID(id)
}

// GetUserByUsernameOrEmail retrieves a user by username or email
func (s *UserService) GetUserByUsernameOrEmail(usernameOrEmail string) (*account.User, error) {
	var user account.User
	if err := s.DB.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).
		Preload("Roles").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByCredentials retrieves a user by username, email, or phone
func (s *UserService) GetUserByCredentials(credential string) (*account.User, error) {
	var user account.User
	if err := s.DB.Where("username = ? OR email = ? OR phone = ?", credential, credential, credential).
		Preload("Roles").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers retrieves all users with pagination
func (s *UserService) GetAllUsers(page, pageSize int) ([]account.User, int64, error) {
	return s.UserRepo.GetAllUsers(page, pageSize)
}

// CreateUser creates a new user
func (s *UserService) CreateUser(email, phone, password string) (*UserResult, error) {
	// Validate input
	if email == "" && phone == "" {
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Email or phone number is required",
		}, fmt.Errorf("email or phone number is required")
	}

	if password == "" || len(password) < 6 {
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Password must be at least 6 characters long",
		}, fmt.Errorf("password must be at least 6 characters long")
	}

	// Check if user already exists
	var existingCount int64
	query := s.DB.Model(&account.User{})
	if email != "" {
		query = query.Where("email = ?", email)
	}
	if phone != "" {
		query = query.Or("phone = ?", phone)
	}
	if err := query.Count(&existingCount).Error; err != nil {
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Error checking existing user",
		}, err
	}
	if existingCount > 0 {
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Email or phone number already registered",
		}, fmt.Errorf("email or phone number already registered")
	}

	// Generate username if not provided
	username := fmt.Sprintf("user_%s", uuid.New().String()[:8])
	if email != "" {
		// Use part of email as username
		username = email[:min(len(email), 8)]
	}

	// Hash password
	hash, salt, err := passwordpkg.GenerateHashAndSalt(password)
	if err != nil {
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Error hashing password",
		}, err
	}

	// Create user
	user := &account.User{
		Username:     username,
		Email:        email,
		Phone:        phone,
		PasswordHash: hash,
		Salt:         salt,
		IsActive:     true,
	}

	// Start transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Database transaction error",
		}, tx.Error
	}

	// Save user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Error creating user",
		}, err
	}

	// Assign default role (staff)
	role, err := s.UserRepo.GetRoleByName(account.RoleStaff)
	if err != nil {
		// Create the role if it doesn't exist
		role = &account.Role{
			Name: account.RoleStaff,
		}
		if err := tx.Create(role).Error; err != nil {
			tx.Rollback()
			return &UserResult{
				Success: false,
				Message: "User creation failed",
				Error:   "Error creating role",
			}, err
		}
	}

	// Assign role to user
	userRole := &account.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}
	if err := tx.Create(userRole).Error; err != nil {
		tx.Rollback()
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Error assigning role",
		}, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Error committing transaction",
		}, err
	}

	// Send welcome notification
	if s.NotificationService != nil {
		metadata := notification.Metadata{
			"user_id":  user.ID.String(),
			"username": user.Username,
			"email":    user.Email,
		}

		s.NotificationService.CreateNotification(
			&user.ID,
			notification.RecipientUser,
			"Welcome to our platform!",
			"Thank you for registering with us.",
			metadata,
			[]notification.ChannelType{notification.ChannelEmail},
		)
	}

	return &UserResult{
		Success:  true,
		Message:  "User created successfully",
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    []string{string(role.Name)},
	}, nil
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(id uuid.UUID, email, phone, username string, isActive *bool) (*UserResult, error) {
	// Get the user
	user, err := s.UserRepo.GetUserByID(id)
	if err != nil {
		return &UserResult{
			Success: false,
			Message: "User update failed",
			Error:   "User not found",
		}, err
	}

	// Update fields if provided
	if email != "" && email != user.Email {
		// Check if email is already in use
		var count int64
		if err := s.DB.Model(&account.User{}).Where("email = ? AND id != ?", email, id).Count(&count).Error; err != nil {
			return &UserResult{
				Success: false,
				Message: "User update failed",
				Error:   "Error checking email uniqueness",
			}, err
		}
		if count > 0 {
			return &UserResult{
				Success: false,
				Message: "User update failed",
				Error:   "Email already in use",
			}, fmt.Errorf("email already in use")
		}
		user.Email = email
	}

	if phone != "" && phone != user.Phone {
		// Check if phone is already in use
		var count int64
		if err := s.DB.Model(&account.User{}).Where("phone = ? AND id != ?", phone, id).Count(&count).Error; err != nil {
			return &UserResult{
				Success: false,
				Message: "User update failed",
				Error:   "Error checking phone uniqueness",
			}, err
		}
		if count > 0 {
			return &UserResult{
				Success: false,
				Message: "User update failed",
				Error:   "Phone already in use",
			}, fmt.Errorf("phone already in use")
		}
		user.Phone = phone
	}

	if username != "" && username != user.Username {
		// Check if username is already in use
		var count int64
		if err := s.DB.Model(&account.User{}).Where("username = ? AND id != ?", username, id).Count(&count).Error; err != nil {
			return &UserResult{
				Success: false,
				Message: "User update failed",
				Error:   "Error checking username uniqueness",
			}, err
		}
		if count > 0 {
			return &UserResult{
				Success: false,
				Message: "User update failed",
				Error:   "Username already in use",
			}, fmt.Errorf("username already in use")
		}
		user.Username = username
	}

	if isActive != nil {
		user.IsActive = *isActive
	}

	// Save the user
	if err := s.UserRepo.UpdateUser(user); err != nil {
		return &UserResult{
			Success: false,
			Message: "User update failed",
			Error:   "Error updating user",
		}, err
	}

	// Get user roles
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, string(role.Name))
	}

	return &UserResult{
		Success:  true,
		Message:  "User updated successfully",
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    roles,
	}, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id uuid.UUID) (*UserResult, error) {
	// Get the user
	user, err := s.UserRepo.GetUserByID(id)
	if err != nil {
		return &UserResult{
			Success: false,
			Message: "User deletion failed",
			Error:   "User not found",
		}, err
	}

	// Start transaction
	tx := s.DB.Begin()
	if tx.Error != nil {
		return &UserResult{
			Success: false,
			Message: "User deletion failed",
			Error:   "Database transaction error",
		}, tx.Error
	}

	// Delete user roles
	if err := tx.Where("user_id = ?", id).Delete(&account.UserRole{}).Error; err != nil {
		tx.Rollback()
		return &UserResult{
			Success: false,
			Message: "User deletion failed",
			Error:   "Error deleting user roles",
		}, err
	}

	// Delete user
	if err := tx.Delete(user).Error; err != nil {
		tx.Rollback()
		return &UserResult{
			Success: false,
			Message: "User deletion failed",
			Error:   "Error deleting user",
		}, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return &UserResult{
			Success: false,
			Message: "User deletion failed",
			Error:   "Error committing transaction",
		}, err
	}

	return &UserResult{
		Success:  true,
		Message:  "User deleted successfully",
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
