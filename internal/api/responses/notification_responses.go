package responses

import (
	"time"

	"github.com/google/uuid"
)

// NotificationResponse represents a single notification in the response
type NotificationResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	IsRead      bool      `json:"is_read"`
	RedirectURL string    `json:"redirect_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NotificationsResponse represents a list of notifications in the response
type NotificationsResponse struct {
	Success    bool                   `json:"success"`
	Message    string                 `json:"message"`
	Data       []NotificationResponse `json:"data"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// NotificationReadResponse represents the response after marking a notification as read
type NotificationReadResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Data    NotificationResponse `json:"data"`
}
