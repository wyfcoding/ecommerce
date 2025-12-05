package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/notification/v1"              // 导入通知模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/notification/application"   // 导入通知模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/notification/domain/entity" // 导入通知模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 NotificationService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedNotificationServiceServer                                  // 嵌入生成的UnimplementedNotificationServiceServer，确保前向兼容性。
	app                                       *application.NotificationService // 依赖Notification应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Notification gRPC 服务端实例。
func NewServer(app *application.NotificationService) *Server {
	return &Server{app: app}
}

// SendNotification 处理发送通知的gRPC请求。
// req: 包含用户ID、类型、标题、内容和附加数据的请求体。
// 返回发送成功的通知ID响应和可能发生的gRPC错误。
func (s *Server) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	// 将Proto的Type（字符串）转换为实体NotificationType。
	// 注意：这里进行了直接转换，如果req.Type是未知类型，可能导致错误或默认值。
	// 应用服务层没有验证NotificationType的合法性。
	nType := entity.NotificationType(req.Type)

	// Proto请求中没有明确的通知渠道字段。这里默认使用应用内通知渠道。
	channel := entity.NotificationChannelApp
	// Proto请求中没有附加数据（data）字段。这里传递nil。
	// 实际生产中，可能需要扩展Proto以支持更多字段。
	notif, err := s.app.SendNotification(ctx, req.UserId, nType, channel, req.Title, req.Content, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to send notification: %v", err))
	}

	return &pb.SendNotificationResponse{
		NotificationId: uint64(notif.ID), // 返回通知ID。
	}, nil
}

// ListNotifications 处理列出用户通知的gRPC请求。
// req: 包含用户ID、是否包含已读消息和分页参数的请求体。
// 返回通知列表响应和可能发生的gRPC错误。
func (s *Server) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	// 根据 req.IncludeRead 字段构建消息状态过滤器。
	var filterStatus *int // 指向 int 的指针，如果为nil表示不按状态过滤。
	if !req.IncludeRead {
		// 如果 IncludeRead 为 false，则只查询未读消息。
		st := int(entity.NotificationStatusUnread)
		filterStatus = &st
	}
	// 如果 IncludeRead 为 true，则 filterStatus 保持为 nil，应用服务层可能会返回所有消息。

	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取通知列表。
	notifs, total, err := s.app.ListNotifications(ctx, req.UserId, filterStatus, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list notifications: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbNotifs := make([]*pb.Notification, len(notifs))
	for i, n := range notifs {
		pbNotifs[i] = convertNotificationToProto(n)
	}

	return &pb.ListNotificationsResponse{
		Notifications: pbNotifs,
		TotalCount:    uint64(total), // 总记录数。
	}, nil
}

// MarkNotificationAsRead 处理标记通知为已读的gRPC请求。
// req: 包含通知ID和用户ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) MarkNotificationAsRead(ctx context.Context, req *pb.MarkNotificationAsReadRequest) (*pb.MarkNotificationAsReadResponse, error) {
	if err := s.app.MarkAsRead(ctx, req.NotificationId, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to mark notification as read: %v", err))
	}
	return &pb.MarkNotificationAsReadResponse{}, nil
}

// GetUnreadCount 处理获取未读通知数量的gRPC请求。
// req: 包含用户ID的请求体。
// 返回未读通知数量响应和可能发生的gRPC错误。
/*
func (s *Server) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	count, err := s.app.GetUnreadCount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get unread count: %v", err))
	}
	return &pb.GetUnreadCountResponse{
		Count: count, // 未读通知数量。
	}, nil
}
*/

// convertNotificationToProto 是一个辅助函数，将领域层的 Notification 实体转换为 protobuf 的 Notification 消息。
func convertNotificationToProto(n *entity.Notification) *pb.Notification {
	if n == nil {
		return nil
	}
	return &pb.Notification{
		NotificationId: uint64(n.ID),                              // 通知ID。
		UserId:         n.UserID,                                  // 用户ID。
		Type:           string(n.NotifType),                       // 类型。
		Title:          n.Title,                                   // 标题。
		Content:        n.Content,                                 // 内容。
		IsRead:         n.Status == entity.NotificationStatusRead, // 是否已读。
		CreatedAt:      timestamppb.New(n.CreatedAt),              // 创建时间。
	}
}
