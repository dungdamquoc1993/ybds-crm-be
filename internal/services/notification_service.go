package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/internal/repositories"
	"github.com/ybds/pkg/telegram"
	"github.com/ybds/pkg/websocket"
	"gorm.io/gorm"
)

// NotificationService handles notification-related business logic
type NotificationService struct {
	DB               *gorm.DB
	NotificationRepo *repositories.NotificationRepository
	WebsocketHub     *websocket.Hub
	TelegramClient   *telegram.TelegramClient
	UserRepo         *repositories.UserRepository
}

// NewNotificationService creates a new instance of NotificationService
func NewNotificationService(notificationDB *gorm.DB, accountDB *gorm.DB, websocketHub *websocket.Hub, telegramClient *telegram.TelegramClient) *NotificationService {
	return &NotificationService{
		DB:               notificationDB,
		NotificationRepo: repositories.NewNotificationRepository(notificationDB),
		WebsocketHub:     websocketHub,
		TelegramClient:   telegramClient,
		UserRepo:         repositories.NewUserRepository(accountDB),
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
	tx := s.DB.Begin()
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

	// Use repository to create notification
	if err := s.NotificationRepo.CreateNotification(&notif); err != nil {
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

		// Use repository to create channel
		if err := s.NotificationRepo.CreateChannel(&channel); err != nil {
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

	// Send notifications through the appropriate channels
	for _, channelType := range channels {
		switch channelType {
		case notification.ChannelWebsocket:
			if recipientID != nil {
				s.sendWebsocketNotification(notif)
			}
		case notification.ChannelTelegram:
			s.sendTelegramNotification(notif)
		case notification.ChannelEmail:
			s.sendEmailNotification(notif)
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
	if s.WebsocketHub == nil {
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
		s.WebsocketHub.BroadcastToUser(notif.RecipientID.String(), jsonMessage)
	} else if notif.RecipientID != nil {
		// Use BroadcastToAll for now as a workaround
		// TODO: Implement proper topic-based broadcasting
		s.WebsocketHub.BroadcastToAll(jsonMessage)
	}

	// Update the channel status
	s.updateChannelStatus(notif.ID, notification.ChannelWebsocket, notification.ChannelSent, "Websocket message sent")
}

// sendTelegramNotification sends a notification through Telegram
func (s *NotificationService) sendTelegramNotification(notif notification.Notification) {
	// Skip if TelegramClient is nil
	if s.TelegramClient == nil {
		return
	}

	// Only proceed if this is a user notification with a recipient ID
	if notif.RecipientType == notification.RecipientUser && notif.RecipientID != nil {
		// Get the user by ID using the repository
		user, err := s.UserRepo.GetUserByID(*notif.RecipientID)
		if err != nil {
			fmt.Printf("Error finding user for Telegram notification: %v\n", err)
			s.updateChannelStatus(notif.ID, notification.ChannelTelegram, notification.ChannelFailed, "User not found")
			return
		}

		// Check if user has telegram_id
		if user.TelegramID <= 0 {
			fmt.Printf("User %s does not have a valid Telegram ID\n", user.Username)
			s.updateChannelStatus(notif.ID, notification.ChannelTelegram, notification.ChannelFailed, "User has no Telegram ID")
			return
		}

		// Format the message
		message := fmt.Sprintf("%s\n\n%s", notif.Title, notif.Message)

		// Send the message
		if err := s.TelegramClient.SendMessage(user.TelegramID, message); err != nil {
			fmt.Printf("Error sending Telegram notification: %v\n", err)
			s.updateChannelStatus(notif.ID, notification.ChannelTelegram, notification.ChannelFailed, err.Error())
			return
		}

		// Update channel status
		s.updateChannelStatus(notif.ID, notification.ChannelTelegram, notification.ChannelSent, "Message sent successfully")
	} else {
		// Update channel status for non-user notifications
		s.updateChannelStatus(notif.ID, notification.ChannelTelegram, notification.ChannelFailed, "Unsupported recipient type")
	}
}

// sendEmailNotification sends notification through email
func (s *NotificationService) sendEmailNotification(notif notification.Notification) {
	// Email service is not implemented
	s.updateChannelStatus(notif.ID, notification.ChannelEmail, notification.ChannelFailed, "Email service not implemented")
}

// updateChannelStatus updates the status of a notification channel
func (s *NotificationService) updateChannelStatus(notificationID uuid.UUID, channelType notification.ChannelType, status notification.ChannelStatus, message string) {
	go func() {
		var channel notification.Channel
		// Get the channel using the repository instead of direct DB access
		channels, err := s.NotificationRepo.GetChannelsByNotificationID(notificationID)
		if err != nil {
			log.Printf("Error finding channel for notification %s and channel %s: %v", notificationID, channelType, err)
			return
		}

		// Find the specific channel by type
		var channelFound bool
		for _, ch := range channels {
			if ch.Channel == channelType {
				channel = ch
				channelFound = true
				break
			}
		}

		if !channelFound {
			log.Printf("Channel %s not found for notification %s", channelType, notificationID)
			return
		}

		// Update the channel status and response
		channel.Status = status
		channel.Response = notification.Response{
			"updated_at": time.Now(),
			"message":    message,
		}

		// Save the channel using the repository
		if err := s.NotificationRepo.UpdateChannel(&channel); err != nil {
			log.Printf("Error updating channel status: %v", err)
		}
	}()
}

// GetNotificationsByRecipient retrieves all notifications for a recipient
func (s *NotificationService) GetNotificationsByRecipient(recipientID uuid.UUID, recipientType notification.RecipientType) ([]notification.Notification, error) {
	return s.NotificationRepo.GetNotificationsByRecipient(recipientID, recipientType)
}

// GetUnreadNotificationsByRecipient retrieves all unread notifications for a recipient
func (s *NotificationService) GetUnreadNotificationsByRecipient(recipientID uuid.UUID, recipientType notification.RecipientType) ([]notification.Notification, error) {
	return s.NotificationRepo.GetUnreadNotificationsByRecipient(recipientID, recipientType)
}

// MarkNotificationAsRead marks a notification as read
func (s *NotificationService) MarkNotificationAsRead(id uuid.UUID) error {
	return s.NotificationRepo.MarkNotificationAsRead(id)
}

// MarkAllNotificationsAsRead marks all notifications for a recipient as read
func (s *NotificationService) MarkAllNotificationsAsRead(recipientID uuid.UUID, recipientType notification.RecipientType) error {
	return s.NotificationRepo.MarkAllNotificationsAsRead(recipientID, recipientType)
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

	// Find all admin users using the repository
	adminUsers, err := s.UserRepo.GetAdminUsers()
	if err != nil {
		return &NotificationResult{
			Success: false,
			Message: "Failed to find admin users for notification",
			Error:   err.Error(),
		}, err
	}

	if len(adminUsers) == 0 {
		return &NotificationResult{
			Success: false,
			Message: "No admin users found to notify",
			Error:   "No admin users found",
		}, fmt.Errorf("no admin users found")
	}

	// Track notification results
	var lastResult *NotificationResult
	var lastError error
	successCount := 0

	// Send notification to each admin user
	for _, admin := range adminUsers {
		result, err := s.CreateNotification(
			&admin.ID,
			notification.RecipientUser,
			title,
			message,
			notifMetadata,
			[]notification.ChannelType{notification.ChannelWebsocket, notification.ChannelTelegram},
		)

		if err == nil && result.Success {
			successCount++
		}

		// Store the last result for return value
		lastResult = result
		lastError = err
	}

	// Return success if at least one notification was sent successfully
	if successCount > 0 {
		return &NotificationResult{
			Success:        true,
			Message:        fmt.Sprintf("Product notification sent to %d admin users", successCount),
			NotificationID: lastResult.NotificationID,
		}, nil
	}

	// If all notifications failed, return the last error
	return lastResult, lastError
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
		message = fmt.Sprintf("Order (#%s) has been confirmed.", orderID.String()[:8])
	case "shipped":
		title = "Order Shipped"
		message = fmt.Sprintf("Order (#%s) has been shipped.", orderID.String()[:8])
	case "delivered":
		title = "Order Delivered"
		message = fmt.Sprintf("Order (#%s) has been delivered.", orderID.String()[:8])
	case "canceled":
		title = "Order Canceled"
		message = fmt.Sprintf("Order (#%s) has been canceled.", orderID.String()[:8])
	default:
		title = "Order Update"
		message = fmt.Sprintf("Update for order (#%s).", orderID.String()[:8])
	}

	// Find all admin users using the repository
	adminUsers, err := s.UserRepo.GetAdminUsers()
	if err != nil {
		return &NotificationResult{
			Success: false,
			Message: "Failed to find admin users for notification",
			Error:   err.Error(),
		}, err
	}

	if len(adminUsers) == 0 {
		return &NotificationResult{
			Success: false,
			Message: "No admin users found to notify",
			Error:   "No admin users found",
		}, fmt.Errorf("no admin users found")
	}

	// Track notification results
	var lastResult *NotificationResult
	var lastError error
	successCount := 0
	log.Println("day la tong so luong admin", len(adminUsers))
	// Send notification to each admin user
	for _, admin := range adminUsers {
		result, err := s.CreateNotification(
			&admin.ID,
			notification.RecipientUser,
			title,
			message,
			notifMetadata,
			[]notification.ChannelType{notification.ChannelWebsocket, notification.ChannelTelegram},
		)

		if err == nil && result.Success {
			successCount++
		}

		// Store the last result for return value
		lastResult = result
		lastError = err
	}

	// Return success if at least one notification was sent successfully
	if successCount > 0 {
		return &NotificationResult{
			Success:        true,
			Message:        fmt.Sprintf("Order notification sent to %d admin users", successCount),
			NotificationID: lastResult.NotificationID,
		}, nil
	}

	// If all notifications failed, return the last error
	return lastResult, lastError
}
