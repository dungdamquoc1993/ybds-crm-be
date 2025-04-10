package requests

import (
	"errors"
)

// Since the current handlers (GetUsers and GetUserByID) don't require any request body validation,
// we don't need any request structs. The handlers only use query parameters and path parameters
// which are parsed directly in the handler functions.

// This file is kept as a placeholder for future request types that might be needed.

// UpdateTelegramIDRequest defines the request for updating a user's Telegram ID
type UpdateTelegramIDRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

// Validate validates the UpdateTelegramIDRequest
func (r *UpdateTelegramIDRequest) Validate() error {
	if r.TelegramID <= 0 {
		return errors.New("telegram_id must be a positive number")
	}
	return nil
}
