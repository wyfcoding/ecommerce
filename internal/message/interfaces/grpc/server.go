package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/message/v1"           // 导入消息模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/message/application"   // 导入消息模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/message/domain/entity" // 导入消息模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 MessageService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedMessageServiceServer                             // 嵌入生成的UnimplementedMessageServiceServer，确保前向兼容性。
	app                                  *application.MessageService // 依赖Message应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Message gRPC 服务端实例。
func NewServer(app *application.MessageService) *Server {
	return &Server{app: app}
}

// SendMessage 处理发送消息的gRPC请求。
// req: 包含发送者ID、接收者ID、消息类型、标题、内容和链接的请求体。
// 返回发送成功的消息ID响应和可能发生的gRPC错误。
func (s *Server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// 将Proto的Type（字符串）转换为实体MessageType。
	// 注意：这里进行了直接转换，如果req.Type是未知类型，可能导致错误或默认值。
	// 应用服务层没有验证MessageType的合法性。
	mType := entity.MessageType(req.Type)

	// 调用应用服务层发送消息。
	msg, err := s.app.SendMessage(ctx, req.SenderId, req.ReceiverId, mType, req.Title, req.Content, req.Link)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to send message: %v", err))
	}

	return &pb.SendMessageResponse{
		MessageId: uint64(msg.ID), // 返回消息ID。
	}, nil
}

// ListMessages 处理列出用户消息的gRPC请求。
// req: 包含用户ID和是否包含已读消息的请求体。
// 返回消息列表响应和可能发生的gRPC错误。
func (s *Server) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	// 根据 req.IncludeRead 字段构建消息状态过滤器。
	var filterStatus *int // 指向 int 的指针，如果为nil表示不按状态过滤。
	if !req.IncludeRead {
		st := int(entity.MessageStatusUnread) // 只查询未读消息。
		filterStatus = &st
	}

	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取消息列表。
	msgs, total, err := s.app.ListMessages(ctx, req.UserId, filterStatus, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list messages: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbMsgs := make([]*pb.Message, len(msgs))
	for i, m := range msgs {
		pbMsgs[i] = convertMessageToProto(m)
	}

	return &pb.ListMessagesResponse{
		Messages:   pbMsgs,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// MarkMessageAsRead 处理标记消息为已读的gRPC请求。
// req: 包含消息ID和用户ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) MarkMessageAsRead(ctx context.Context, req *pb.MarkMessageAsReadRequest) (*pb.MarkMessageAsReadResponse, error) {
	if err := s.app.MarkAsRead(ctx, req.MessageId, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to mark message as read: %v", err))
	}
	return &pb.MarkMessageAsReadResponse{}, nil
}

// GetUnreadCount 处理获取未读消息数量的gRPC请求。
// req: 包含用户ID的请求体。
// 返回未读消息数量响应和可能发生的gRPC错误。
func (s *Server) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	count, err := s.app.GetUnreadCount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get unread count: %v", err))
	}
	return &pb.GetUnreadCountResponse{
		Count: count, // 未读消息数量。
	}, nil
}

// convertMessageToProto 是一个辅助函数，将领域层的 Message 实体转换为 protobuf 的 Message 消息。
func convertMessageToProto(m *entity.Message) *pb.Message {
	if m == nil {
		return nil
	}
	return &pb.Message{
		Id:         uint64(m.ID),                         // 消息ID。
		SenderId:   m.SenderID,                           // 发送者ID。
		ReceiverId: m.ReceiverID,                         // 接收者ID。
		Type:       string(m.MessageType),                // 消息类型。
		Title:      m.Title,                              // 标题。
		Content:    m.Content,                            // 内容。
		Link:       m.Link,                               // 链接。
		IsRead:     m.Status == entity.MessageStatusRead, // 是否已读。
		CreatedAt:  timestamppb.New(m.CreatedAt),         // 创建时间。
	}
}
