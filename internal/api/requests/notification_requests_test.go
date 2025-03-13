package requests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMarkNotificationAsReadRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request MarkNotificationAsReadRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: MarkNotificationAsReadRequest{
				NotificationID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "Invalid request - nil UUID",
			request: MarkNotificationAsReadRequest{
				NotificationID: uuid.Nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateNotificationRequest_Validate(t *testing.T) {
	recipientID := uuid.New()
	tests := []struct {
		name    string
		request CreateNotificationRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: CreateNotificationRequest{
				RecipientID:   &recipientID,
				RecipientType: "user",
				Title:         "Test Notification",
				Message:       "This is a test notification",
				Metadata:      map[string]interface{}{"key": "value"},
				Channels:      []string{"email", "push"},
			},
			wantErr: false,
		},
		{
			name: "Invalid request - empty title",
			request: CreateNotificationRequest{
				RecipientID:   &recipientID,
				RecipientType: "user",
				Title:         "",
				Message:       "This is a test notification",
			},
			wantErr: true,
		},
		{
			name: "Invalid request - empty message",
			request: CreateNotificationRequest{
				RecipientID:   &recipientID,
				RecipientType: "user",
				Title:         "Test Notification",
				Message:       "",
			},
			wantErr: true,
		},
		{
			name: "Invalid request - empty recipient type",
			request: CreateNotificationRequest{
				RecipientID:   &recipientID,
				RecipientType: "",
				Title:         "Test Notification",
				Message:       "This is a test notification",
			},
			wantErr: true,
		},
		{
			name: "Valid request - nil recipient ID",
			request: CreateNotificationRequest{
				RecipientID:   nil,
				RecipientType: "system",
				Title:         "Test Notification",
				Message:       "This is a test notification",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetNotificationsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request GetNotificationsRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: GetNotificationsRequest{
				Page:       1,
				PageSize:   10,
				UnreadOnly: false,
			},
			wantErr: false,
		},
		{
			name: "Valid request - unread only",
			request: GetNotificationsRequest{
				Page:       1,
				PageSize:   10,
				UnreadOnly: true,
			},
			wantErr: false,
		},
		{
			name: "Invalid request - negative page",
			request: GetNotificationsRequest{
				Page:       -1,
				PageSize:   10,
				UnreadOnly: false,
			},
			wantErr: true,
		},
		{
			name: "Invalid request - zero page size",
			request: GetNotificationsRequest{
				Page:       1,
				PageSize:   0,
				UnreadOnly: false,
			},
			wantErr: true,
		},
		{
			name: "Invalid request - negative page size",
			request: GetNotificationsRequest{
				Page:       1,
				PageSize:   -5,
				UnreadOnly: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
