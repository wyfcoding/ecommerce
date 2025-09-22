package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	v1 "ecommerce/api/notification/v1"
	"ecommerce/internal/notification/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NotificationService is the gRPC service implementation for notification.
type NotificationService struct {
	v1.UnimplementedNotificationServiceServer
	uc *biz.NotificationUsecase
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(uc *biz.NotificationUsecase) *NotificationService {
	return &NotificationService{uc: uc}
}

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "无法获取元数据")
	}
	// 兼容 gRPC-Gateway 在 HTTP 请求时注入的用户ID
	values := md.Get("x-md-global-user-id")
	if len(values) == 0 {
		// 兼容直接 gRPC 调用时注入的用户ID
		values = md.Get("x-user-id")
		if len(values) == 0 {
			return 0, status.Errorf(codes.Unauthenticated, "请求头中缺少 x-user-id 信息")
		}
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "x-user-id 格式无效")
	}
	return userID, nil
}

// bizNotificationToProto converts biz.Notification to v1.Notification.
func bizNotificationToProto(notif *biz.Notification) *v1.Notification {
	if notif == nil {
		return nil
	}
	return &v1.Notification{
		NotificationId: notif.NotificationID,
		UserId:         notif.UserID,
		Type:           notif.Type,
		Title:          notif.Title,
		Content:        notif.Content,
		IsRead:         notif.IsRead,
		CreatedAt:      timestamppb.New(notif.CreatedAt),
	}
}

// SendNotification implements the SendNotification RPC.
func (s *NotificationService) SendNotification(ctx context.Context, req *v1.SendNotificationRequest) (*v1.SendNotificationResponse, error) {
	if req.UserId == 0 || req.Type == "" || req.Title == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, type, title, and content are required")
	}

	notif, err := s.uc.SendNotification(ctx, req.UserId, req.Type, req.Title, req.Content)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send notification: %v", err)
	}

	return &v1.SendNotificationResponse{NotificationId: notif.NotificationID}, nil
}

// ListNotifications implements the ListNotifications RPC.
func (s *NotificationService) ListNotifications(ctx context.Context, req *v1.ListNotificationsRequest) (*v1.ListNotificationsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	notifs, total, err := s.uc.ListNotifications(ctx, userID, req.IncludeRead, req.PageSize, req.PageNum)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list notifications: %v", err)
	}

	protoNotifs := make([]*v1.Notification, len(notifs))
	for i, notif := range notifs {
		protoNotifs[i] = bizNotificationToProto(notif)
	}

	return &v1.ListNotificationsResponse{Notifications: protoNotifs, TotalCount: total}, nil
}

// MarkNotificationAsRead implements the MarkNotificationAsRead RPC.
func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, req *v1.MarkNotificationAsReadRequest) (*v1.MarkNotificationAsReadResponse, error) {
	if req.NotificationId == 0 || req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "notification_id and user_id are required")
	}

	// Optional: Verify user_id matches the notification's user_id in biz layer
	err := s.uc.MarkNotificationAsRead(ctx, strconv.FormatUint(req.NotificationId, 10))
	if err != nil {
		if errors.Is(err, biz.ErrNotificationNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to mark notification as read: %v", err)
	}

	return &v1.MarkNotificationAsReadResponse{}, nil
}
