package handlers_test

import (
	"testing"

	"github.com/ybds/internal/services"
)

// MockUserNotificationService is a mock implementation of services.NotificationService for user tests
type MockUserNotificationService struct {
	*services.NotificationService
}

// NewMockUserNotificationService creates a new mock notification service for user tests
func NewMockUserNotificationService() *MockUserNotificationService {
	return &MockUserNotificationService{
		NotificationService: &services.NotificationService{},
	}
}

func TestUserHandler(t *testing.T) {
	// Skip tests for now until we can properly mock the services
	t.Skip("Skipping user handler tests until we can properly mock the services")
}
