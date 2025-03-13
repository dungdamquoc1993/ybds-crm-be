package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/repositories"
	"github.com/ybds/pkg/websocket"
	"gorm.io/gorm"
)

// NotificationService handles notification-related business logic
type NotificationService struct {
	db               *gorm.DB
	notificationRepo *repositories.NotificationRepository
	websocketHub     *websocket.Hub
}

// NewNotificationService creates a new instance of NotificationService
func NewNotificationService(db *gorm.DB, websocketHub *websocket.Hub) *NotificationService {
	return &NotificationService{
		db:               db,
		notificationRepo: repositories.NewNotificationRepository(db),
		websocketHub:     websocketHub,
	}
}

// NotificationResult represents the result of a notification operation
type NotificationResult struct {
	Success        bool
	Message        string
	Error          string
	NotificationID uuid.UUID
}

// CreateNotification creates a new notification and sends it through appropriate channels
func (s *NotificationService) CreateNotification(
	recipientID *uuid.UUID,
	recipientType notification.RecipientType,
	title string,
	message string,
	metadata notification.Metadata,
	channels []notification.ChannelType,
) (*NotificationResult, error) {
	// Start a transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return &NotificationResult{
			Success: false,
			Message: "Failed to create notification",
			Error:   "Database transaction error",
		}, tx.Error
	}

	// Create the notification
	notif := notification.Notification{
		RecipientID:   recipientID,
		RecipientType: recipientType,
		Title:         title,
		Message:       message,
		Status:        notification.NotificationPending,
		Metadata:      metadata,
		IsRead:        false,
	}

	if err := tx.Create(&notif).Error; err != nil {
		tx.Rollback()
		return &NotificationResult{
			Success: false,
			Message: "Failed to create notification",
			Error:   "Database error",
		}, err
	}

	// Create channels for the notification
	for _, channelType := range channels {
		channel := notification.Channel{
			NotificationID: notif.ID,
			Channel:        channelType,
			Status:         notification.ChannelPending,
			Attempts:       0,
		}

		if err := tx.Create(&channel).Error; err != nil {
			tx.Rollback()
			return &NotificationResult{
				Success: false,
				Message: "Failed to create notification channel",
				Error:   "Database error",
			}, err
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return &NotificationResult{
			Success: false,
			Message: "Failed to create notification",
			Error:   "Database commit error",
		}, err
	}

	// Send the notification through websocket if applicable
	if recipientID != nil {
		for _, channelType := range channels {
			if channelType == notification.ChannelWebsocket {
				s.sendWebsocketNotification(notif)
				break
			}
		}
	}

	return &NotificationResult{
		Success:        true,
		Message:        "Notification created successfully",
		NotificationID: notif.ID,
	}, nil
}

// sendWebsocketNotification sends a notification through websocket
func (s *NotificationService) sendWebsocketNotification(notif notification.Notification) {
	// Skip if websocketHub is nil
	if s.websocketHub == nil {
		return
	}

	// Prepare the notification payload
	payload := map[string]interface{}{
		"id":             notif.ID,
		"title":          notif.Title,
		"message":        notif.Message,
		"created_at":     notif.CreatedAt,
		"recipient_id":   notif.RecipientID,
		"recipient_type": notif.RecipientType,
		"metadata":       notif.Metadata,
	}

	// Create the websocket message
	wsMessage := map[string]interface{}{
		"type":    "notification",
		"payload": payload,
	}

	// Convert the message to JSON
	jsonMessage, err := json.Marshal(wsMessage)
	if err != nil {
		fmt.Printf("Error marshaling websocket message: %v\n", err)
		return
	}

	// Broadcast to the user if it's a user notification
	if notif.RecipientType == notification.RecipientUser && notif.RecipientID != nil {
		s.websocketHub.BroadcastToUser(notif.RecipientID.String(), jsonMessage)
	} else if notif.RecipientID != nil {
		// Use BroadcastToAll for now as a workaround
		// TODO: Implement proper topic-based broadcasting
		s.websocketHub.BroadcastToAll(jsonMessage)
	}

	// Update the channel status
	go func() {
		var channel notification.Channel
		if err := s.db.Where("notification_id = ? AND channel = ?", notif.ID, notification.ChannelWebsocket).First(&channel).Error; err == nil {
			channel.Status = notification.ChannelSent
			channel.Response = notification.Response{"sent_at": time.Now()}
			s.db.Save(&channel)
		}
	}()
}

// GetNotificationsByRecipient retrieves all notifications for a recipient
func (s *NotificationService) GetNotificationsByRecipient(recipientID uuid.UUID, recipientType notification.RecipientType) ([]notification.Notification, error) {
	return s.notificationRepo.GetNotificationsByRecipient(recipientID, recipientType)
}

// GetUnreadNotificationsByRecipient retrieves all unread notifications for a recipient
func (s *NotificationService) GetUnreadNotificationsByRecipient(recipientID uuid.UUID, recipientType notification.RecipientType) ([]notification.Notification, error) {
	return s.notificationRepo.GetUnreadNotificationsByRecipient(recipientID, recipientType)
}

// MarkNotificationAsRead marks a notification as read
func (s *NotificationService) MarkNotificationAsRead(id uuid.UUID) error {
	return s.notificationRepo.MarkNotificationAsRead(id)
}

// MarkAllNotificationsAsRead marks all notifications for a recipient as read
func (s *NotificationService) MarkAllNotificationsAsRead(recipientID uuid.UUID, recipientType notification.RecipientType) error {
	return s.notificationRepo.MarkAllNotificationsAsRead(recipientID, recipientType)
}

// CreateProductNotification creates a notification for a product event
func (s *NotificationService) CreateProductNotification(productID uuid.UUID, productName string, event string, metadata map[string]interface{}) (*NotificationResult, error) {
	// Create metadata
	notifMetadata := notification.Metadata{
		"product_id":   productID.String(),
		"product_name": productName,
		"event":        event,
	}

	// Add additional metadata
	for k, v := range metadata {
		notifMetadata[k] = v
	}

	// Create title and message based on event
	title := ""
	message := ""
	switch event {
	case "created":
		title = "New Product Added"
		message = fmt.Sprintf("A new product '%s' has been added to the catalog.", productName)
	case "updated":
		title = "Product Updated"
		message = fmt.Sprintf("The product '%s' has been updated.", productName)
	case "deleted":
		title = "Product Removed"
		message = fmt.Sprintf("The product '%s' has been removed from the catalog.", productName)
	case "low_stock":
		title = "Low Stock Alert"
		message = fmt.Sprintf("The product '%s' is running low on stock.", productName)
	case "out_of_stock":
		title = "Out of Stock Alert"
		message = fmt.Sprintf("The product '%s' is now out of stock.", productName)
	case "back_in_stock":
		title = "Back in Stock"
		message = fmt.Sprintf("The product '%s' is back in stock.", productName)
	default:
		title = "Product Notification"
		message = fmt.Sprintf("Notification for product '%s'.", productName)
	}

	// Create the notification for admin users
	// We're not specifying a recipient ID because this is a broadcast to all admins
	return s.CreateNotification(
		nil,
		notification.RecipientUser,
		title,
		message,
		notifMetadata,
		[]notification.ChannelType{notification.ChannelWebsocket},
	)
}

// CreateOrderNotification creates a notification for an order event
func (s *NotificationService) CreateOrderNotification(orderID uuid.UUID, customerID uuid.UUID, event string, metadata map[string]interface{}) (*NotificationResult, error) {
	// Create metadata
	notifMetadata := notification.Metadata{
		"order_id":    orderID.String(),
		"customer_id": customerID.String(),
		"event":       event,
	}

	// Add additional metadata
	for k, v := range metadata {
		notifMetadata[k] = v
	}

	// Create title and message based on event
	title := ""
	message := ""
	switch event {
	case "created":
		title = "New Order Received"
		message = fmt.Sprintf("A new order (#%s) has been received.", orderID.String()[:8])
	case "confirmed":
		title = "Order Confirmed"
		message = fmt.Sprintf("Your order (#%s) has been confirmed.", orderID.String()[:8])
	case "shipped":
		title = "Order Shipped"
		message = fmt.Sprintf("Your order (#%s) has been shipped.", orderID.String()[:8])
	case "delivered":
		title = "Order Delivered"
		message = fmt.Sprintf("Your order (#%s) has been delivered.", orderID.String()[:8])
	case "canceled":
		title = "Order Canceled"
		message = fmt.Sprintf("Your order (#%s) has been canceled.", orderID.String()[:8])
	default:
		title = "Order Update"
		message = fmt.Sprintf("Update for your order (#%s).", orderID.String()[:8])
	}

	// Create the notification for the customer
	return s.CreateNotification(
		&customerID,
		notification.RecipientUser,
		title,
		message,
		notifMetadata,
		[]notification.ChannelType{notification.ChannelWebsocket},
	)
}
