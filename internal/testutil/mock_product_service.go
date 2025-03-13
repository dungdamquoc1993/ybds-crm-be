package testutil

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/ybds/internal/models/product"
	"github.com/ybds/internal/services"
)

// MockProductRepository is a mock implementation of the ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetInventoryByID(id uuid.UUID) (*product.Inventory, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Inventory), args.Error(1)
}

func (m *MockProductRepository) UpdateInventory(inventory *product.Inventory) error {
	args := m.Called(inventory)
	return args.Error(0)
}

func (m *MockProductRepository) GetProductByID(id uuid.UUID) (*product.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductRepository) GetProductBySKU(sku string) (*product.Product, error) {
	args := m.Called(sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductRepository) GetAllProducts(page, pageSize int, filters map[string]interface{}) ([]product.Product, int64, error) {
	args := m.Called(page, pageSize, filters)
	return args.Get(0).([]product.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) CreateProduct(p *product.Product) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *MockProductRepository) UpdateProduct(p *product.Product) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteProduct(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) GetInventoriesByProductID(productID uuid.UUID) ([]product.Inventory, error) {
	args := m.Called(productID)
	return args.Get(0).([]product.Inventory), args.Error(1)
}

func (m *MockProductRepository) CreateInventory(inventory *product.Inventory) error {
	args := m.Called(inventory)
	return args.Error(0)
}

func (m *MockProductRepository) DeleteInventory(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductRepository) GetPriceByID(id uuid.UUID) (*product.Price, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Price), args.Error(1)
}

func (m *MockProductRepository) GetPricesByProductID(productID uuid.UUID) ([]product.Price, error) {
	args := m.Called(productID)
	return args.Get(0).([]product.Price), args.Error(1)
}

func (m *MockProductRepository) GetCurrentPrice(productID uuid.UUID) (*product.Price, error) {
	args := m.Called(productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Price), args.Error(1)
}

func (m *MockProductRepository) CreatePrice(price *product.Price) error {
	args := m.Called(price)
	return args.Error(0)
}

func (m *MockProductRepository) UpdatePrice(price *product.Price) error {
	args := m.Called(price)
	return args.Error(0)
}

func (m *MockProductRepository) DeletePrice(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockProductNotificationService is a mock implementation of the NotificationService for product tests
type MockProductNotificationService struct {
	mock.Mock
}

func (m *MockProductNotificationService) CreateProductNotification(productID uuid.UUID, productName string, event string, metadata map[string]interface{}) (*services.NotificationResult, error) {
	args := m.Called(productID, productName, event, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.NotificationResult), args.Error(1)
}

func (m *MockProductNotificationService) CreateOrderNotification(orderID uuid.UUID, customerID uuid.UUID, event string, metadata map[string]interface{}) (*services.NotificationResult, error) {
	args := m.Called(orderID, customerID, event, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.NotificationResult), args.Error(1)
}

func (m *MockProductNotificationService) CreateNotification(recipientID *uuid.UUID, recipientType interface{}, title string, message string, metadata interface{}, channels interface{}) (*services.NotificationResult, error) {
	args := m.Called(recipientID, recipientType, title, message, metadata, channels)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.NotificationResult), args.Error(1)
}

func (m *MockProductNotificationService) GetNotificationsByRecipient(recipientID uuid.UUID, recipientType interface{}) (interface{}, error) {
	args := m.Called(recipientID, recipientType)
	return args.Get(0), args.Error(1)
}

func (m *MockProductNotificationService) GetUnreadNotificationsByRecipient(recipientID uuid.UUID, recipientType interface{}) (interface{}, error) {
	args := m.Called(recipientID, recipientType)
	return args.Get(0), args.Error(1)
}

func (m *MockProductNotificationService) MarkNotificationAsRead(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProductNotificationService) MarkAllNotificationsAsRead(recipientID uuid.UUID, recipientType interface{}) error {
	args := m.Called(recipientID, recipientType)
	return args.Error(0)
}
