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
	db                  *gorm.DB
	userRepo            *repositories.UserRepository
	notificationService *NotificationService
}

// NewUserService creates a new instance of UserService
func NewUserService(db *gorm.DB, notificationService *NotificationService) *UserService {
	return &UserService{
		db:                  db,
		userRepo:            repositories.NewUserRepository(db),
		notificationService: notificationService,
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
	return s.userRepo.GetUserByID(id)
}

// GetUserByUsernameOrEmail retrieves a user by username or email
func (s *UserService) GetUserByUsernameOrEmail(usernameOrEmail string) (*account.User, error) {
	var user account.User
	if err := s.db.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).
		Preload("Roles").First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers retrieves all users with pagination
func (s *UserService) GetAllUsers(page, pageSize int) ([]account.User, int64, error) {
	return s.userRepo.GetAllUsers(page, pageSize)
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
	query := s.db.Model(&account.User{})
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

	// Generate username if email provided
	username := ""
	if email != "" {
		// Use part before @ as username
		for i, c := range email {
			if c == '@' {
				username = email[:i]
				break
			}
		}
	} else {
		// Use phone as username if no email
		username = phone
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

	// Create user in database
	user := account.User{
		Username:     username,
		Email:        email,
		Phone:        phone,
		PasswordHash: hash,
		Salt:         salt,
		IsActive:     true,
	}

	// Start transaction
	tx := s.db.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return &UserResult{
			Success: false,
			Message: "User creation failed",
			Error:   "Error creating user",
		}, err
	}

	// By default, assign customer role
	var customerRole account.Role
	if err := tx.Where("name = ?", account.RoleCustomer).First(&customerRole).Error; err != nil {
		// Create customer role if it doesn't exist
		customerRole = account.Role{
			Name: account.RoleCustomer,
		}
		if err := tx.Create(&customerRole).Error; err != nil {
			tx.Rollback()
			return &UserResult{
				Success: false,
				Message: "User creation failed",
				Error:   "Error creating role",
			}, err
		}
	}

	// Assign role to user
	userRole := account.UserRole{
		UserID: user.ID,
		RoleID: customerRole.ID,
	}
	if err := tx.Create(&userRole).Error; err != nil {
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
	if s.notificationService != nil {
		metadata := notification.Metadata{
			"user_id": user.ID.String(),
		}
		s.notificationService.CreateNotification(
			&user.ID,
			notification.RecipientUser,
			"Welcome to our platform!",
			"Thank you for registering. We're excited to have you on board.",
			metadata,
			[]notification.ChannelType{notification.ChannelWebsocket},
		)
	}

	return &UserResult{
		Success:  true,
		Message:  "User created successfully",
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(id uuid.UUID, email, phone, username string, isActive *bool) (*UserResult, error) {
	// Get the user
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return &UserResult{
			Success: false,
			Message: "User update failed",
			Error:   "User not found",
		}, err
	}

	// Update fields if provided
	if email != "" {
		user.Email = email
	}
	if phone != "" {
		user.Phone = phone
	}
	if username != "" {
		user.Username = username
	}
	if isActive != nil {
		user.IsActive = *isActive
	}

	// Save the user
	if err := s.userRepo.UpdateUser(user); err != nil {
		return &UserResult{
			Success: false,
			Message: "User update failed",
			Error:   "Error updating user",
		}, err
	}

	// Extract roles
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

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id uuid.UUID) (*UserResult, error) {
	// Get the user
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return &UserResult{
			Success: false,
			Message: "User deletion failed",
			Error:   "User not found",
		}, err
	}

	// Delete the user
	if err := s.userRepo.DeleteUser(id); err != nil {
		return &UserResult{
			Success: false,
			Message: "User deletion failed",
			Error:   "Error deleting user",
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

// AddUserAddress adds an address to a user
func (s *UserService) AddUserAddress(userID uuid.UUID, address *account.Address) error {
	address.UserID = &userID
	return s.userRepo.CreateAddress(address)
}

// UpdateUserAddress updates a user's address
func (s *UserService) UpdateUserAddress(addressID uuid.UUID, address *account.Address) error {
	address.ID = addressID
	return s.userRepo.UpdateAddress(address)
}

// DeleteUserAddress deletes a user's address
func (s *UserService) DeleteUserAddress(addressID uuid.UUID) error {
	return s.userRepo.DeleteAddress(addressID)
}

// GetUserAddresses gets all addresses for a user
func (s *UserService) GetUserAddresses(userID uuid.UUID) ([]account.Address, error) {
	return s.userRepo.GetUserAddresses(userID)
}

// GetGuestByID retrieves a guest by ID
func (s *UserService) GetGuestByID(id uuid.UUID) (*account.Guest, error) {
	return s.userRepo.GetGuestByID(id)
}

// CreateGuest creates a new guest
func (s *UserService) CreateGuest(name, email, phone string) (*account.Guest, error) {
	guest := &account.Guest{
		Name:  name,
		Email: email,
		Phone: phone,
	}
	err := s.userRepo.CreateGuest(guest)
	return guest, err
}

// UpdateGuest updates an existing guest
func (s *UserService) UpdateGuest(id uuid.UUID, name, email, phone string) (*account.Guest, error) {
	// Get the guest
	guest, err := s.userRepo.GetGuestByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if name != "" {
		guest.Name = name
	}
	if email != "" {
		guest.Email = email
	}
	if phone != "" {
		guest.Phone = phone
	}

	// Save the guest
	if err := s.userRepo.UpdateGuest(guest); err != nil {
		return nil, err
	}

	return guest, nil
}

// DeleteGuest deletes a guest by ID
func (s *UserService) DeleteGuest(id uuid.UUID) error {
	return s.userRepo.DeleteGuest(id)
}

// AddGuestAddress adds an address to a guest
func (s *UserService) AddGuestAddress(guestID uuid.UUID, address *account.Address) error {
	address.GuestID = &guestID
	return s.userRepo.CreateAddress(address)
}

// GetGuestAddresses gets all addresses for a guest
func (s *UserService) GetGuestAddresses(guestID uuid.UUID) ([]account.Address, error) {
	return s.userRepo.GetGuestAddresses(guestID)
}
