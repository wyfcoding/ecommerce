package grpc

import (
	"context"
	pb "ecommerce/api/message/v1"
	"ecommerce/internal/message/application"
	"ecommerce/internal/message/domain/entity"

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
	// Service SendMessage(ctx, senderID, receiverID, messageType, title, content, link)
	// Proto: sender_id, receiver_id, type, title, content, link

	mType := entity.MessageType(req.Type)
	// Default to SYSTEM if empty or invalid? Service doesn't validate.
	// Let's assume caller provides valid type or service handles it.

	msg, err := s.app.SendMessage(ctx, req.SenderId, req.ReceiverId, mType, req.Title, req.Content, req.Link)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendMessageResponse{
		MessageId: uint64(msg.ID),
	}, nil
}

func (s *Server) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	// Service ListMessages(ctx, userID, status, page, pageSize)

	var filterStatus *int
	if !req.IncludeRead {
		st := int(entity.MessageStatusUnread)
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
		return nil, status.Error(codes.Internal, err.Error())
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
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.MarkMessageAsReadResponse{}, nil
}

func (s *Server) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	count, err := s.app.GetUnreadCount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetUnreadCountResponse{
		Count: count,
	}, nil
}

func convertMessageToProto(m *entity.Message) *pb.Message {
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
		IsRead:     m.Status == entity.MessageStatusRead,
		CreatedAt:  timestamppb.New(m.CreatedAt),
	}
}
