package repositories

import (
	"github.com/google/uuid"
	"github.com/ybds/internal/models/notification"
	"gorm.io/gorm"
)

// NotificationRepository handles database operations for notifications
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new instance of NotificationRepository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

// GetNotificationByID retrieves a notification by ID with all relations
func (r *NotificationRepository) GetNotificationByID(id uuid.UUID) (*notification.Notification, error) {
	var n notification.Notification
	err := r.db.Where("id = ?", id).
		Preload("Channels").
		First(&n).Error
	return &n, err
}

// GetNotificationsByRecipient retrieves all notifications for a recipient
func (r *NotificationRepository) GetNotificationsByRecipient(recipientID uuid.UUID, recipientType notification.RecipientType) ([]notification.Notification, error) {
	var notifications []notification.Notification
	err := r.db.Where("recipient_id = ? AND recipient_type = ?", recipientID, recipientType).
		Preload("Channels").
		Find(&notifications).Error
	return notifications, err
}

// GetUnreadNotificationsByRecipient retrieves all unread notifications for a recipient
func (r *NotificationRepository) GetUnreadNotificationsByRecipient(recipientID uuid.UUID, recipientType notification.RecipientType) ([]notification.Notification, error) {
	var notifications []notification.Notification
	err := r.db.Where("recipient_id = ? AND recipient_type = ? AND is_read = ?", recipientID, recipientType, false).
		Preload("Channels").
		Find(&notifications).Error
	return notifications, err
}

// GetAllNotifications retrieves all notifications with pagination
func (r *NotificationRepository) GetAllNotifications(page, pageSize int) ([]notification.Notification, int64, error) {
	var notifications []notification.Notification
	var total int64

	// Count total records
	if err := r.db.Model(&notification.Notification{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	offset := (page - 1) * pageSize
	err := r.db.Offset(offset).Limit(pageSize).
		Preload("Channels").
		Find(&notifications).Error

	return notifications, total, err
}

// CreateNotification creates a new notification
func (r *NotificationRepository) CreateNotification(n *notification.Notification) error {
	return r.db.Create(n).Error
}

// UpdateNotification updates an existing notification
func (r *NotificationRepository) UpdateNotification(n *notification.Notification) error {
	return r.db.Save(n).Error
}

// DeleteNotification deletes a notification by ID
func (r *NotificationRepository) DeleteNotification(id uuid.UUID) error {
	return r.db.Delete(&notification.Notification{}, id).Error
}

// MarkNotificationAsRead marks a notification as read
func (r *NotificationRepository) MarkNotificationAsRead(id uuid.UUID) error {
	return r.db.Model(&notification.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}

// MarkAllNotificationsAsRead marks all notifications for a recipient as read
func (r *NotificationRepository) MarkAllNotificationsAsRead(recipientID uuid.UUID, recipientType notification.RecipientType) error {
	return r.db.Model(&notification.Notification{}).
		Where("recipient_id = ? AND recipient_type = ?", recipientID, recipientType).
		Update("is_read", true).Error
}

// GetChannelByID retrieves a channel by ID
func (r *NotificationRepository) GetChannelByID(id uuid.UUID) (*notification.Channel, error) {
	var channel notification.Channel
	err := r.db.Joins("JOIN notifications ON channels.notification_id = notifications.id").
		Where("channels.id = ? AND notifications.deleted_at IS NULL", id).
		First(&channel).Error
	return &channel, err
}

// GetChannelsByNotificationID retrieves all channels for a notification
func (r *NotificationRepository) GetChannelsByNotificationID(notificationID uuid.UUID) ([]notification.Channel, error) {
	// Check if notification exists and is not deleted
	var count int64
	if err := r.db.Model(&notification.Notification{}).Where("id = ? AND deleted_at IS NULL", notificationID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var channels []notification.Channel
	err := r.db.Where("notification_id = ?", notificationID).Find(&channels).Error
	return channels, err
}

// CreateChannel creates a new channel
func (r *NotificationRepository) CreateChannel(channel *notification.Channel) error {
	return r.db.Create(channel).Error
}

// UpdateChannel updates an existing channel
func (r *NotificationRepository) UpdateChannel(channel *notification.Channel) error {
	return r.db.Save(channel).Error
}

// DeleteChannel deletes a channel by ID
func (r *NotificationRepository) DeleteChannel(id uuid.UUID) error {
	return r.db.Delete(&notification.Channel{}, id).Error
}

// UpdateChannelStatus updates the status of a channel
func (r *NotificationRepository) UpdateChannelStatus(id uuid.UUID, status notification.ChannelStatus) error {
	return r.db.Model(&notification.Channel{}).Where("id = ?", id).Update("status", status).Error
}

// IncrementChannelAttempts increments the attempts count for a channel
func (r *NotificationRepository) IncrementChannelAttempts(id uuid.UUID) error {
	return r.db.Model(&notification.Channel{}).Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + ?", 1)).Error
}

// UpdateChannelResponse updates the response for a channel
func (r *NotificationRepository) UpdateChannelResponse(id uuid.UUID, response notification.Response) error {
	return r.db.Model(&notification.Channel{}).Where("id = ?", id).Update("response", response).Error
}
