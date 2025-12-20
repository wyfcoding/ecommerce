package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/message/v1"
	"github.com/wyfcoding/ecommerce/internal/message/application"
	"github.com/wyfcoding/ecommerce/internal/message/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedMessageServiceServer
	app *application.MessageService
}

func NewServer(app *application.MessageService) *Server {
	return &Server{app: app}
}

func (s *Server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	mType := domain.MessageType(req.Type)

	msg, err := s.app.SendMessage(ctx, req.SenderId, req.ReceiverId, mType, req.Title, req.Content, req.Link)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to send message: %v", err))
	}

	return &pb.SendMessageResponse{
		MessageId: uint64(msg.ID),
	}, nil
}

func (s *Server) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	var filterStatus *int
	if !req.IncludeRead {
		st := int(domain.MessageStatusUnread)
		filterStatus = &st
	}

	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	msgs, total, err := s.app.ListMessages(ctx, req.UserId, filterStatus, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list messages: %v", err))
	}

	pbMsgs := make([]*pb.Message, len(msgs))
	for i, m := range msgs {
		pbMsgs[i] = convertMessageToProto(m)
	}

	return &pb.ListMessagesResponse{
		Messages:   pbMsgs,
		TotalCount: uint64(total),
	}, nil
}

func (s *Server) MarkMessageAsRead(ctx context.Context, req *pb.MarkMessageAsReadRequest) (*pb.MarkMessageAsReadResponse, error) {
	if err := s.app.MarkAsRead(ctx, req.MessageId, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to mark message as read: %v", err))
	}
	return &pb.MarkMessageAsReadResponse{}, nil
}

func (s *Server) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	count, err := s.app.GetUnreadCount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get unread count: %v", err))
	}
	return &pb.GetUnreadCountResponse{
		Count: count,
	}, nil
}

func convertMessageToProto(m *domain.Message) *pb.Message {
	if m == nil {
		return nil
	}
	return &pb.Message{
		Id:         uint64(m.ID),
		SenderId:   m.SenderID,
		ReceiverId: m.ReceiverID,
		Type:       string(m.MessageType),
		Title:      m.Title,
		Content:    m.Content,
		Link:       m.Link,
		IsRead:     m.Status == domain.MessageStatusRead,
		CreatedAt:  timestamppb.New(m.CreatedAt),
	}
}
