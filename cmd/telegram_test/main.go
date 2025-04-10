package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/ybds/internal/models/notification"
	"github.com/ybds/pkg/config"
	"github.com/ybds/pkg/database"
	"github.com/ybds/pkg/telegram"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Telegram client
	telegramClient := telegram.NewClient(cfg.Telegram.BotToken)
	log.Println("Telegram client initialized with token:", cfg.Telegram.BotToken)

	// Connect to notification database
	db, err := database.NewDatabase(&cfg.NotificationDB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to notification database")

	// Test sending a direct message with hard-coded chat ID
	chatID := int64(7452534928) // The chat_id from the existing test
	message := "🔔 TEST: This is a test notification from the YBDS system"

	log.Printf("Sending message to chat ID %d...", chatID)
	if err := telegramClient.SendMessage(chatID, message); err != nil {
		log.Fatalf("Failed to send Telegram message: %v", err)
	}
	log.Println("Message sent successfully!")

	// Create a notification in the database
	log.Println("Creating a test notification in the database...")
	notif := notification.Notification{
		Title:         "Test Notification",
		Message:       "This is a test notification from the YBDS system",
		Status:        notification.NotificationPending,
		RecipientType: notification.RecipientUser,
		IsRead:        false,
		Metadata: notification.Metadata{
			"test": true,
			"type": "telegram_test",
		},
	}

	// Create notification
	if err := db.Create(&notif).Error; err != nil {
		log.Fatalf("Failed to create notification: %v", err)
	}

	// Create channel for the notification
	channel := notification.Channel{
		NotificationID: notif.ID,
		Channel:        notification.ChannelTelegram,
		Status:         notification.ChannelPending,
		Attempts:       0,
	}

	if err := db.Create(&channel).Error; err != nil {
		log.Fatalf("Failed to create notification channel: %v", err)
	}

	log.Printf("Notification created with ID: %s", notif.ID)
	log.Printf("Notification channel created with ID: %s", channel.ID)

	log.Println("Test completed successfully!")
}
