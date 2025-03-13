package requests

import (
	"errors"

	"github.com/google/uuid"
)

// MarkNotificationAsReadRequest represents a request to mark a notification as read
type MarkNotificationAsReadRequest struct {
	NotificationID uuid.UUID `json:"notification_id"`
}

// Validate validates the mark notification as read request
func (r *MarkNotificationAsReadRequest) Validate() error {
	if r.NotificationID == uuid.Nil {
		return errors.New("notification ID is required")
	}
	return nil
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	RecipientID   *uuid.UUID             `json:"recipient_id"`
	RecipientType string                 `json:"recipient_type"`
	Title         string                 `json:"title"`
	Message       string                 `json:"message"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Channels      []string               `json:"channels,omitempty"`
}

// Validate validates the create notification request
func (r *CreateNotificationRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Message == "" {
		return errors.New("message is required")
	}
	if r.RecipientType == "" {
		return errors.New("recipient type is required")
	}
	return nil
}

// GetNotificationsRequest represents a request to get notifications
type GetNotificationsRequest struct {
	Page       int  `json:"page" query:"page"`
	PageSize   int  `json:"page_size" query:"page_size"`
	UnreadOnly bool `json:"unread_only" query:"unread_only"`
}

// Validate validates the get notifications request
func (r *GetNotificationsRequest) Validate() error {
	if r.Page < 0 {
		return errors.New("page must be greater than or equal to 0")
	}
	if r.PageSize < 1 {
		return errors.New("page size must be greater than 0")
	}
	return nil
}
