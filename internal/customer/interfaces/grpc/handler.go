package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/customer/v1"
	"github.com/wyfcoding/ecommerce/internal/customer/application"
	"github.com/wyfcoding/ecommerce/internal/customer/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 Customer 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedCustomerServiceServer
	app *application.Customer
}

// NewServer 创建并返回一个新的 Customer gRPC 服务端实例。
func NewServer(app *application.Customer) *Server {
	return &Server{app: app}
}

// CreateTicket 处理创建工单的gRPC请求。
func (s *Server) CreateTicket(ctx context.Context, req *pb.CreateTicketRequest) (*pb.TicketResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateTicket received", "user_id", req.UserId, "subject", req.Subject)

	ticket, err := s.app.CreateTicket(ctx, req.UserId, req.Subject, req.Description, "general", domain.TicketPriorityMedium)
	if err != nil {
		slog.Error("gRPC CreateTicket failed", "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create ticket: %v", err))
	}

	slog.Info("gRPC CreateTicket successful", "ticket_id", ticket.ID, "user_id", req.UserId, "duration", time.Since(start))
	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

// GetTicketByID 处理根据ID获取工单信息的gRPC请求。
func (s *Server) GetTicketByID(ctx context.Context, req *pb.GetTicketByIDRequest) (*pb.TicketResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetTicketByID received", "ticket_id", req.TicketId)

	ticket, err := s.app.GetTicket(ctx, req.TicketId)
	if err != nil {
		slog.Error("gRPC GetTicketByID failed", "ticket_id", req.TicketId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.NotFound, fmt.Sprintf("ticket not found: %v", err))
	}

	slog.Debug("gRPC GetTicketByID successful", "ticket_id", req.TicketId, "duration", time.Since(start))
	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

// UpdateTicketStatus 处理更新工单状态的gRPC请求。
func (s *Server) UpdateTicketStatus(ctx context.Context, req *pb.UpdateTicketStatusRequest) (*pb.TicketResponse, error) {
	start := time.Now()
	slog.Info("gRPC UpdateTicketStatus received", "ticket_id", req.TicketId, "status", req.Status)

	st := strings.ToUpper(req.Status)
	var err error

	switch st {
	case "CLOSED":
		err = s.app.CloseTicket(ctx, req.TicketId)
	case "RESOLVED":
		err = s.app.ResolveTicket(ctx, req.TicketId)
	default:
		slog.Warn("gRPC UpdateTicketStatus unsupported status", "ticket_id", req.TicketId, "status", req.Status)
		return nil, status.Errorf(codes.Unimplemented, "status transition to %s not supported via gRPC yet", req.Status)
	}

	if err != nil {
		slog.Error("gRPC UpdateTicketStatus failed", "ticket_id", req.TicketId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update ticket status: %v", err))
	}

	ticket, err := s.app.GetTicket(ctx, req.TicketId)
	if err != nil {
		slog.Error("gRPC GetTicket after update failed", "ticket_id", req.TicketId, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to fetch updated ticket: %v", err))
	}

	slog.Info("gRPC UpdateTicketStatus successful", "ticket_id", req.TicketId, "duration", time.Since(start))
	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

// AddMessageToTicket 处理向工单添加消息的gRPC请求。
func (s *Server) AddMessageToTicket(ctx context.Context, req *pb.AddMessageToTicketRequest) (*pb.TicketMessageResponse, error) {
	start := time.Now()
	slog.Info("gRPC AddMessageToTicket received", "ticket_id", req.TicketId, "sender_id", req.SenderId)

	msg, err := s.app.ReplyTicket(ctx, req.TicketId, req.SenderId, req.SenderType, req.MessageBody, domain.MessageTypeText)
	if err != nil {
		slog.Error("gRPC AddMessageToTicket failed", "ticket_id", req.TicketId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add message to ticket: %v", err))
	}

	slog.Info("gRPC AddMessageToTicket successful", "ticket_id", req.TicketId, "message_id", msg.ID, "duration", time.Since(start))
	return &pb.TicketMessageResponse{
		Message: convertMessageToProto(msg),
	}, nil
}

// GetTicketMessages 处理获取工单消息列表的gRPC请求。
func (s *Server) GetTicketMessages(ctx context.Context, req *pb.GetTicketMessagesRequest) (*pb.GetTicketMessagesResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetTicketMessages received", "ticket_id", req.TicketId)

	msgs, _, err := s.app.ListMessages(ctx, req.TicketId, 1, 100)
	if err != nil {
		slog.Error("gRPC GetTicketMessages failed", "ticket_id", req.TicketId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list ticket messages: %v", err))
	}

	pbMsgs := make([]*pb.TicketMessage, len(msgs))
	for i, m := range msgs {
		pbMsgs[i] = convertMessageToProto(m)
	}

	slog.Debug("gRPC GetTicketMessages successful", "ticket_id", req.TicketId, "count", len(pbMsgs), "duration", time.Since(start))
	return &pb.GetTicketMessagesResponse{
		Messages: pbMsgs,
	}, nil
}

func convertTicketToProto(t *domain.Ticket) *pb.TicketInfo {
	if t == nil {
		return nil
	}

	statusStr := "UNKNOWN"
	switch t.Status {
	case domain.TicketStatusOpen:
		statusStr = "OPEN"
	case domain.TicketStatusInProgress:
		statusStr = "IN_PROGRESS"
	case domain.TicketStatusResolved:
		statusStr = "RESOLVED"
	case domain.TicketStatusClosed:
		statusStr = "CLOSED"
	}

	return &pb.TicketInfo{
		TicketId:    uint64(t.ID),
		UserId:      t.UserID,
		Subject:     t.Subject,
		Description: t.Description,
		Status:      statusStr,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}

func convertMessageToProto(m *domain.Message) *pb.TicketMessage {
	if m == nil {
		return nil
	}
	return &pb.TicketMessage{
		MessageId:   uint64(m.ID),
		TicketId:    m.TicketID,
		SenderId:    m.SenderID,
		SenderType:  m.SenderType,
		MessageBody: m.Content,
		SentAt:      timestamppb.New(m.CreatedAt),
	}
}