package notification

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/ybds/internal/models"
)

// RecipientType defines the type of notification recipient
type RecipientType string

const (
	// RecipientUser represents a registered user
	RecipientUser RecipientType = "user"
	// RecipientGuest represents a guest customer
	RecipientGuest RecipientType = "guest"
	// RecipientPotentialCustomer represents a potential customer
	RecipientPotentialCustomer RecipientType = "potential_customer"
	// RecipientPartner represents a partner
	RecipientPartner RecipientType = "partner"
	// RecipientOther represents other types of recipients
	RecipientOther RecipientType = "other"
)

// NotificationStatus defines the status of a notification
type NotificationStatus string

const (
	// NotificationPending means the notification is waiting to be sent
	NotificationPending NotificationStatus = "pending"
	// NotificationSent means the notification has been sent
	NotificationSent NotificationStatus = "sent"
	// NotificationFailed means the notification failed to send
	NotificationFailed NotificationStatus = "failed"
)

// Metadata represents additional data for a notification
type Metadata map[string]interface{}

// Value implements the driver.Valuer interface for Metadata
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for Metadata
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(Metadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &m)
}

// Notification represents a notification in the system
type Notification struct {
	models.Base
	RecipientID   *uuid.UUID         `gorm:"column:recipient_id;type:uuid;null;index" json:"recipient_id,omitempty"`
	RecipientType RecipientType      `gorm:"column:recipient_type;type:varchar(50);not null;index" json:"recipient_type"`
	Title         string             `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Message       string             `gorm:"column:message;type:text;not null" json:"message"`
	Status        NotificationStatus `gorm:"column:status;type:varchar(50);not null;default:'pending';index" json:"status"`
	Metadata      Metadata           `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	IsRead        bool               `gorm:"column:is_read;not null;default:false;index" json:"is_read"`
	Channels      []Channel          `gorm:"foreignKey:NotificationID" json:"channels,omitempty"`
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}
