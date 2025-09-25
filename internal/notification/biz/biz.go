package biz

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

// Notification represents a notification in the business logic layer.
type Notification struct {
	ID             uint
	NotificationID string
	UserID         uint64
	Type           string
	Title          string
	Content        string
	IsRead         bool
	CreatedAt      time.Time
}

// NotificationRepo defines the interface for notification data access.
type NotificationRepo interface {
	CreateNotification(ctx context.Context, notification *Notification) (*Notification, error)
	GetNotificationByID(ctx context.Context, notificationID string) (*Notification, error)
	ListNotificationsByUserID(ctx context.Context, userID uint64, includeRead bool, pageSize, pageNum uint32) ([]*Notification, uint64, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string) error
}

// NotificationUsecase is the business logic for notification management.
type NotificationUsecase struct {
	repo NotificationRepo
	// TODO: Add clients for external notification services (e.g., SMS, Email, Push)
}

// NewNotificationUsecase creates a new NotificationUsecase.
func NewNotificationUsecase(repo NotificationRepo) *NotificationUsecase {
	return &NotificationUsecase{repo: repo}
}

// SendNotification sends a new notification.
func (uc *NotificationUsecase) SendNotification(ctx context.Context, userID uint64, notifType, title, content string) (*Notification, error) {
	notificationID := uuid.New().String()
	notification := &Notification{
		NotificationID: notificationID,
		UserID:         userID,
		Type:           notifType,
		Title:          title,
		Content:        content,
		IsRead:         false, // New notifications are unread by default
	}
	return uc.repo.CreateNotification(ctx, notification)
}

// ListNotifications lists notifications for a specific user.
func (uc *NotificationUsecase) ListNotifications(ctx context.Context, userID uint64, includeRead bool, pageSize, pageNum uint32) ([]*Notification, uint64, error) {
	return uc.repo.ListNotificationsByUserID(ctx, userID, includeRead, pageSize, pageNum)
}

// MarkNotificationAsRead marks a notification as read.
func (uc *NotificationUsecase) MarkNotificationAsRead(ctx context.Context, notificationID string) error {
	// Optional: Check if notification belongs to the user before marking as read
	notification, err := uc.repo.GetNotificationByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if notification == nil {
		return ErrNotificationNotFound
	}
	if notification.IsRead {
		return nil // Already read
	}
	return uc.repo.MarkNotificationAsRead(ctx, notificationID)
}
