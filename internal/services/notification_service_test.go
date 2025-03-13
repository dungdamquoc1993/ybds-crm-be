package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/services"
)

// TestNotificationService tests the NotificationService functionality
func TestNotificationService(t *testing.T) {
	// This is an integration test that would require a database
	// In a real-world scenario, you would use a test database or mock the database
	t.Skip("Skipping integration test")
}

// TestNotificationResult tests the NotificationResult struct
func TestNotificationResult(t *testing.T) {
	// Create a NotificationResult
	notificationID := uuid.New()
	result := services.NotificationResult{
		Success:        true,
		Message:        "Notification created successfully",
		NotificationID: notificationID,
	}

	// Test the fields
	assert.True(t, result.Success)
	assert.Equal(t, "Notification created successfully", result.Message)
	assert.Equal(t, notificationID, result.NotificationID)
}

// TestCreateNotification tests the CreateNotification method
func TestCreateNotification(t *testing.T) {
	// This is an integration test that would require a database
	// In a real-world scenario, you would use a test database or mock the database
	t.Skip("Skipping integration test")
}

// TestMetadata tests the Metadata struct
func TestMetadata(t *testing.T) {
	// Create a Metadata
	metadata := notification.Metadata{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	// Test the fields
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, 123, metadata["key2"])
	assert.Equal(t, true, metadata["key3"])
}
