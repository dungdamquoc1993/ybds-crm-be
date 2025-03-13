package notification

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// ChannelType defines the type of notification channel
type ChannelType string

const (
	// ChannelWebsocket represents a WebSocket notification channel
	ChannelWebsocket ChannelType = "websocket"
	// ChannelEmail represents an email notification channel
	ChannelEmail ChannelType = "email"
	// ChannelTelegram represents a Telegram notification channel
	ChannelTelegram ChannelType = "telegram"
	// ChannelSMS represents an SMS notification channel
	ChannelSMS ChannelType = "sms"
)

// ChannelStatus defines the status of a notification channel
type ChannelStatus string

const (
	// ChannelPending means the notification is waiting to be sent through this channel
	ChannelPending ChannelStatus = "pending"
	// ChannelSent means the notification has been sent through this channel
	ChannelSent ChannelStatus = "sent"
	// ChannelFailed means the notification failed to send through this channel
	ChannelFailed ChannelStatus = "failed"
)

// Response represents the response from the notification sending system
type Response map[string]interface{}

// Value implements the driver.Valuer interface for Response
func (r Response) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}

// Scan implements the sql.Scanner interface for Response
func (r *Response) Scan(value interface{}) error {
	if value == nil {
		*r = make(Response)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &r)
}

// Channel represents a notification channel in the system
type Channel struct {
	models.Base
	NotificationID uuid.UUID     `gorm:"column:notification_id;type:uuid;not null" json:"notification_id"`
	Channel        ChannelType   `gorm:"column:channel;type:varchar(50);not null" json:"channel"`
	Status         ChannelStatus `gorm:"column:status;type:varchar(50);not null;default:'pending'" json:"status"`
	Attempts       int           `gorm:"column:attempts;not null;default:0" json:"attempts"`
	Response       Response      `gorm:"column:response;type:jsonb" json:"response,omitempty"`
	Notification   Notification  `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`
}

// TableName specifies the table name for Channel
func (Channel) TableName() string {
	return "notification_channels"
}
