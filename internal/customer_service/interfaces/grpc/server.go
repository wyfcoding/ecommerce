package grpc

import (
	"context"
	pb "ecommerce/api/customer_service/v1"
	"ecommerce/internal/customer_service/application"
	"ecommerce/internal/customer_service/domain/entity"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedCustomerServiceServer
	app *application.CustomerService
}

func NewServer(app *application.CustomerService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateTicket(ctx context.Context, req *pb.CreateTicketRequest) (*pb.TicketResponse, error) {
	// Service CreateTicket(ctx, userID, subject, description, category, priority)
	// Proto missing category and priority.
	// Default category: "general"
	// Default priority: Medium (2)

	ticket, err := s.app.CreateTicket(ctx, req.UserId, req.Subject, req.Description, "general", entity.TicketPriorityMedium)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

func (s *Server) GetTicketByID(ctx context.Context, req *pb.GetTicketByIDRequest) (*pb.TicketResponse, error) {
	ticket, err := s.app.GetTicket(ctx, req.TicketId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

func (s *Server) UpdateTicketStatus(ctx context.Context, req *pb.UpdateTicketStatusRequest) (*pb.TicketResponse, error) {
	// Service only exposes CloseTicket and ResolveTicket explicitly.
	// We map status string to these methods.

	st := strings.ToUpper(req.Status)
	var err error

	switch st {
	case "CLOSED":
		err = s.app.CloseTicket(ctx, req.TicketId)
	case "RESOLVED":
		err = s.app.ResolveTicket(ctx, req.TicketId)
	default:
		// For other statuses, we might not have a direct transition method in service yet.
		// We could add one, or just return error.
		return nil, status.Errorf(codes.Unimplemented, "status transition to %s not supported via gRPC yet", req.Status)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch updated ticket
	ticket, err := s.app.GetTicket(ctx, req.TicketId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TicketResponse{
		Ticket: convertTicketToProto(ticket),
	}, nil
}

func (s *Server) AddMessageToTicket(ctx context.Context, req *pb.AddMessageToTicketRequest) (*pb.TicketMessageResponse, error) {
	// Service ReplyTicket(ctx, ticketID, senderID, senderType, content, msgType)
	// Proto missing msgType. Default to Text.

	msg, err := s.app.ReplyTicket(ctx, req.TicketId, req.SenderId, req.SenderType, req.MessageBody, entity.MessageTypeText)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.TicketMessageResponse{
		Message: convertMessageToProto(msg),
	}, nil
}

func (s *Server) GetTicketMessages(ctx context.Context, req *pb.GetTicketMessagesRequest) (*pb.GetTicketMessagesResponse, error) {
	// Service ListMessages(ctx, ticketID, page, pageSize)
	// Proto missing pagination. Default to 1, 100?

	msgs, _, err := s.app.ListMessages(ctx, req.TicketId, 1, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbMsgs := make([]*pb.TicketMessage, len(msgs))
	for i, m := range msgs {
		pbMsgs[i] = convertMessageToProto(m)
	}

	return &pb.GetTicketMessagesResponse{
		Messages: pbMsgs,
	}, nil
}

func convertTicketToProto(t *entity.Ticket) *pb.TicketInfo {
	if t == nil {
		return nil
	}

	statusStr := "UNKNOWN"
	switch t.Status {
	case entity.TicketStatusOpen:
		statusStr = "OPEN"
	case entity.TicketStatusInProgress:
		statusStr = "IN_PROGRESS"
	case entity.TicketStatusResolved:
		statusStr = "RESOLVED"
	case entity.TicketStatusClosed:
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

func convertMessageToProto(m *entity.Message) *pb.TicketMessage {
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
