package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Default API URL format for Telegram
var telegramAPIURL = "https://api.telegram.org/bot%s/sendMessage"

// TelegramClient represents a client for the Telegram Bot API
type TelegramClient struct {
	BotToken string
}

// NewClient creates a new Telegram client with the given bot token
func NewClient(botToken string) *TelegramClient {
	return &TelegramClient{
		BotToken: botToken,
	}
}

// SendMessage sends a message to a specific chat ID
func (c *TelegramClient) SendMessage(chatID int64, message string) error {
	url := fmt.Sprintf(telegramAPIURL, c.BotToken)

	payload, err := json.Marshal(map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
	})
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Description string `json:"description"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return fmt.Errorf("telegram API error: %s (code: %d)", errorResponse.Description, resp.StatusCode)
		}
		return fmt.Errorf("telegram API returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}
