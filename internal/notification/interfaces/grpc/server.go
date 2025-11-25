package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/notification/v1"
	"github.com/wyfcoding/ecommerce/internal/notification/application"
	"github.com/wyfcoding/ecommerce/internal/notification/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedNotificationServiceServer
	app *application.NotificationService
}

func NewServer(app *application.NotificationService) *Server {
	return &Server{app: app}
}

func (s *Server) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	// Service SendNotification(ctx, userID, notifType, channel, title, content, data)
	// Proto: user_id, type, title, content. Missing channel and data.

	// Map type string to entity.NotificationType
	nType := entity.NotificationType(req.Type)
	// Validate or default? Service doesn't validate strictly in signature but repo might.
	// Default channel to SYSTEM or APP if not provided.
	// Since this is internal gRPC, maybe we assume SYSTEM channel?
	// Default channel to APP
	channel := entity.NotificationChannelApp

	notif, err := s.app.SendNotification(ctx, req.UserId, nType, channel, req.Title, req.Content, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendNotificationResponse{
		NotificationId: uint64(notif.ID),
	}, nil
}

func (s *Server) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	// Service ListNotifications(ctx, userID, status, page, pageSize)
	// Proto: user_id, include_read, page_size, page_num

	var filterStatus *int
	if !req.IncludeRead {
		// If include_read is false, we only want unread.
		st := int(entity.NotificationStatusUnread)
		filterStatus = &st
	}
	// If include_read is true, we want all? Service takes status pointer.
	// If status is nil, does it return all?
	// Checking service.go: "if status != nil { ... } return s.repo.ListNotifications(..., notifStatus, ...)"
	// If status is nil, it passes nil to repo. Repo likely returns all if nil.

	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	notifs, total, err := s.app.ListNotifications(ctx, req.UserId, filterStatus, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbNotifs := make([]*pb.Notification, len(notifs))
	for i, n := range notifs {
		pbNotifs[i] = convertNotificationToProto(n)
	}

	return &pb.ListNotificationsResponse{
		Notifications: pbNotifs,
		TotalCount:    uint64(total),
	}, nil
}

func (s *Server) MarkNotificationAsRead(ctx context.Context, req *pb.MarkNotificationAsReadRequest) (*pb.MarkNotificationAsReadResponse, error) {
	if err := s.app.MarkAsRead(ctx, req.NotificationId, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.MarkNotificationAsReadResponse{}, nil
}

func convertNotificationToProto(n *entity.Notification) *pb.Notification {
	if n == nil {
		return nil
	}
	return &pb.Notification{
		NotificationId: uint64(n.ID),
		UserId:         n.UserID,
		Type:           string(n.NotifType),
		Title:          n.Title,
		Content:        n.Content,
		IsRead:         n.Status == entity.NotificationStatusRead,
		CreatedAt:      timestamppb.New(n.CreatedAt),
	}
}
