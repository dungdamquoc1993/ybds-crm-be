package testutil

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/ybds/internal/models/account"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/models/order"
	"github.com/ybds/internal/models/product"
)

// Define service result types for testing
type UserResult struct {
	Success  bool
	Message  string
	Error    string
	UserID   uuid.UUID
	Username string
	Email    string
	Roles    []string
}

type ProductResult struct {
	Success   bool
	Message   string
	Error     string
	ProductID uuid.UUID
	Name      string
	SKU       string
}

type InventoryResult struct {
	Success     bool
	Message     string
	Error       string
	InventoryID uuid.UUID
	ProductID   uuid.UUID
	Size        string
	Color       string
	Quantity    int
	Location    string
}

type PriceResult struct {
	Success   bool
	Message   string
	Error     string
	PriceID   uuid.UUID
	ProductID uuid.UUID
	Price     float64
	Currency  string
}

type OrderResult struct {
	Success bool
	Message string
	Error   string
	OrderID uuid.UUID
	UserID  *uuid.UUID
	GuestID *uuid.UUID
	Status  string
	Total   float64
}

type NotificationResult struct {
	Success        bool
	Message        string
	Error          string
	NotificationID uuid.UUID
}

type AuthResult struct {
	Success  bool
	Message  string
	Error    string
	UserID   uuid.UUID
	Username string
	Email    string
	Token    string
	Roles    []string
}

// OrderItemInput represents an order item in a request
type OrderItemInput struct {
	ProductID  uuid.UUID
	Quantity   int
	UnitPrice  float64
	Properties map[string]interface{}
}

// AddressInput represents an address in a request
type AddressInput struct {
	Address  string
	Ward     string
	District string
	City     string
	Country  string
}

// MockUserService is a mock implementation of the UserService
type MockUserService struct {
	mock.Mock
}

// GetUserByID mocks the GetUserByID method
func (m *MockUserService) GetUserByID(id uuid.UUID) (*account.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.User), args.Error(1)
}

// GetAllUsers mocks the GetAllUsers method
func (m *MockUserService) GetAllUsers(page, pageSize int) ([]account.User, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]account.User), args.Get(1).(int64), args.Error(2)
}

// CreateUser mocks the CreateUser method
func (m *MockUserService) CreateUser(username, email, password, phone, firstName, lastName string, roles []string) (*UserResult, error) {
	args := m.Called(username, email, password, phone, firstName, lastName, roles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResult), args.Error(1)
}

// UpdateUser mocks the UpdateUser method
func (m *MockUserService) UpdateUser(id uuid.UUID, email, phone, firstName, lastName *string, isActive *bool, roles []string) (*UserResult, error) {
	args := m.Called(id, email, phone, firstName, lastName, isActive, roles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResult), args.Error(1)
}

// DeleteUser mocks the DeleteUser method
func (m *MockUserService) DeleteUser(id uuid.UUID) (*UserResult, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserResult), args.Error(1)
}

// MockProductService is a mock implementation of the ProductService
type MockProductService struct {
	mock.Mock
}

// GetProductByID mocks the GetProductByID method
func (m *MockProductService) GetProductByID(id uuid.UUID) (*product.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

// GetAllProducts mocks the GetAllProducts method
func (m *MockProductService) GetAllProducts(page, pageSize int, filters map[string]interface{}) ([]product.Product, int64, error) {
	args := m.Called(page, pageSize, filters)
	return args.Get(0).([]product.Product), args.Get(1).(int64), args.Error(2)
}

// CreateProduct mocks the CreateProduct method
func (m *MockProductService) CreateProduct(name, description, sku, category, imageURL string) (*ProductResult, error) {
	args := m.Called(name, description, sku, category, imageURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProductResult), args.Error(1)
}

// UpdateProduct mocks the UpdateProduct method
func (m *MockProductService) UpdateProduct(id uuid.UUID, name, description, sku, category, imageURL string) (*ProductResult, error) {
	args := m.Called(id, name, description, sku, category, imageURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProductResult), args.Error(1)
}

// DeleteProduct mocks the DeleteProduct method
func (m *MockProductService) DeleteProduct(id uuid.UUID) (*ProductResult, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProductResult), args.Error(1)
}

// GetInventoryByID mocks the GetInventoryByID method
func (m *MockProductService) GetInventoryByID(id uuid.UUID) (*product.Inventory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Inventory), args.Error(1)
}

// CreateInventory mocks the CreateInventory method
func (m *MockProductService) CreateInventory(productID uuid.UUID, size, color string, quantity int, location string) (*InventoryResult, error) {
	args := m.Called(productID, size, color, quantity, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*InventoryResult), args.Error(1)
}

// UpdateInventory mocks the UpdateInventory method
func (m *MockProductService) UpdateInventory(id uuid.UUID, size, color string, quantity *int, location string) (*InventoryResult, error) {
	args := m.Called(id, size, color, quantity, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*InventoryResult), args.Error(1)
}

// DeleteInventory mocks the DeleteInventory method
func (m *MockProductService) DeleteInventory(id uuid.UUID) (*InventoryResult, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*InventoryResult), args.Error(1)
}

// GetPriceByID mocks the GetPriceByID method
func (m *MockProductService) GetPriceByID(id uuid.UUID) (*product.Price, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Price), args.Error(1)
}

// CreatePrice mocks the CreatePrice method
func (m *MockProductService) CreatePrice(productID uuid.UUID, price float64, currency string, startDate time.Time, endDate *time.Time) (*PriceResult, error) {
	args := m.Called(productID, price, currency, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PriceResult), args.Error(1)
}

// MockOrderService is a mock implementation of the OrderService
type MockOrderService struct {
	mock.Mock
}

// GetOrderByID mocks the GetOrderByID method
func (m *MockOrderService) GetOrderByID(id uuid.UUID) (*order.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.Order), args.Error(1)
}

// GetAllOrders mocks the GetAllOrders method
func (m *MockOrderService) GetAllOrders(page, pageSize int, filters map[string]interface{}) ([]order.Order, int64, error) {
	args := m.Called(page, pageSize, filters)
	return args.Get(0).([]order.Order), args.Get(1).(int64), args.Error(2)
}

// CreateOrder mocks the CreateOrder method
func (m *MockOrderService) CreateOrder(userID, guestID *uuid.UUID, items []OrderItemInput, shippingAddress, billingAddress *AddressInput, paymentMethod string) (*OrderResult, error) {
	args := m.Called(userID, guestID, items, shippingAddress, billingAddress, paymentMethod)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrderResult), args.Error(1)
}

// UpdateOrderStatus mocks the UpdateOrderStatus method
func (m *MockOrderService) UpdateOrderStatus(id uuid.UUID, status string) (*OrderResult, error) {
	args := m.Called(id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrderResult), args.Error(1)
}

// MockNotificationService is a mock implementation of the NotificationService
type MockNotificationService struct {
	mock.Mock
}

// GetNotificationsByRecipient mocks the GetNotificationsByRecipient method
func (m *MockNotificationService) GetNotificationsByRecipient(recipientID uuid.UUID, recipientType notification.RecipientType) ([]notification.Notification, error) {
	args := m.Called(recipientID, recipientType)
	return args.Get(0).([]notification.Notification), args.Error(1)
}

// GetUnreadNotificationsByRecipient mocks the GetUnreadNotificationsByRecipient method
func (m *MockNotificationService) GetUnreadNotificationsByRecipient(recipientID uuid.UUID, recipientType notification.RecipientType) ([]notification.Notification, error) {
	args := m.Called(recipientID, recipientType)
	return args.Get(0).([]notification.Notification), args.Error(1)
}

// MarkNotificationAsRead mocks the MarkNotificationAsRead method
func (m *MockNotificationService) MarkNotificationAsRead(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// MarkAllNotificationsAsRead mocks the MarkAllNotificationsAsRead method
func (m *MockNotificationService) MarkAllNotificationsAsRead(recipientID uuid.UUID, recipientType notification.RecipientType) error {
	args := m.Called(recipientID, recipientType)
	return args.Error(0)
}

// CreateNotification mocks the CreateNotification method
func (m *MockNotificationService) CreateNotification(recipientID *uuid.UUID, recipientType notification.RecipientType, title, message string, metadata notification.Metadata, channels []notification.ChannelType) (*NotificationResult, error) {
	args := m.Called(recipientID, recipientType, title, message, metadata, channels)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*NotificationResult), args.Error(1)
}

// MockAuthService is a mock implementation of the AuthService
type MockAuthService struct {
	mock.Mock
}

// Login mocks the Login method
func (m *MockAuthService) Login(usernameOrEmail, password string) (*AuthResult, error) {
	args := m.Called(usernameOrEmail, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

// Register mocks the Register method
func (m *MockAuthService) Register(email, phone, password string) (*AuthResult, error) {
	args := m.Called(email, phone, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

// VerifyEmail mocks the VerifyEmail method
func (m *MockAuthService) VerifyEmail(token string) (*AuthResult, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

// ResetPassword mocks the ResetPassword method
func (m *MockAuthService) ResetPassword(token, newPassword string) (*AuthResult, error) {
	args := m.Called(token, newPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

// RequestPasswordReset mocks the RequestPasswordReset method
func (m *MockAuthService) RequestPasswordReset(email string) (*AuthResult, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResult), args.Error(1)
}

// MockJWTService is a mock implementation of the JWTService
type MockJWTService struct {
	mock.Mock
}

// GenerateToken mocks the GenerateToken method
func (m *MockJWTService) GenerateToken(userID string, roles []string) (string, error) {
	args := m.Called(userID, roles)
	return args.String(0), args.Error(1)
}

// ValidateToken mocks the ValidateToken method
func (m *MockJWTService) ValidateToken(token string) (map[string]interface{}, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// AuthMiddleware mocks the AuthMiddleware method
func (m *MockJWTService) AuthMiddleware(c *fiber.Ctx) error {
	args := m.Called(c)
	return args.Error(0)
}
