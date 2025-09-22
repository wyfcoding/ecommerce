package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	v1 "ecommerce/api/customer_service/v1"
	"ecommerce/internal/customer_service/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CustomerService is the gRPC service implementation for customer service.
type CustomerService struct {
	v1.UnimplementedCustomerServiceServer
	uc *biz.CustomerServiceUsecase
}

// NewCustomerService creates a new CustomerService.
func NewCustomerService(uc *biz.CustomerServiceUsecase) *CustomerService {
	return &CustomerService{uc: uc}
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

// bizTicketToProto converts biz.Ticket to v1.Ticket.
func bizTicketToProto(ticket *biz.Ticket) *v1.Ticket {
	if ticket == nil {
		return nil
	}
	protoMessages := make([]*v1.TicketMessage, len(ticket.Messages))
	for i, msg := range ticket.Messages {
		protoMessages[i] = bizTicketMessageToProto(msg)
	}
	return &v1.Ticket{
		TicketId:    ticket.TicketID,
		UserId:      ticket.UserID,
		Subject:     ticket.Subject,
		Description: ticket.Description,
		Status:      ticket.Status,
		CreatedAt:   timestamppb.New(ticket.CreatedAt),
		UpdatedAt:   timestamppb.New(ticket.UpdatedAt),
		Messages:    protoMessages,
	}
}

// bizTicketMessageToProto converts biz.TicketMessage to v1.TicketMessage.
func bizTicketMessageToProto(msg *biz.TicketMessage) *v1.TicketMessage {
	if msg == nil {
		return nil
	}
	return &v1.TicketMessage{
		MessageId:  msg.MessageID,
		TicketId:   msg.TicketID,
		SenderId:   msg.SenderID,
		SenderType: msg.SenderType,
		Content:    msg.Content,
		CreatedAt:  timestamppb.New(msg.CreatedAt),
	}
}

// CreateTicket implements the CreateTicket RPC.
func (s *CustomerService) CreateTicket(ctx context.Context, req *v1.CreateTicketRequest) (*v1.Ticket, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Subject == "" || req.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "subject and description are required")
	}

	bizTicket, err := s.uc.CreateTicket(ctx, userID, req.Subject, req.Description)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create ticket: %v", err)
	}

	return bizTicketToProto(bizTicket), nil
}

// GetTicket implements the GetTicket RPC.
func (s *CustomerService) GetTicket(ctx context.Context, req *v1.GetTicketRequest) (*v1.Ticket, error) {
	if req.TicketId == 0 {
		return nil, status.Error(codes.InvalidArgument, "ticket_id is required")
	}

	bizTicket, err := s.uc.GetTicket(ctx, strconv.FormatUint(req.TicketId, 10))
	if err != nil {
		if errors.Is(err, biz.ErrTicketNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to get ticket: %v", err)
	}

	return bizTicketToProto(bizTicket), nil
}

// AddTicketMessage implements the AddTicketMessage RPC.
func (s *CustomerService) AddTicketMessage(ctx context.Context, req *v1.AddTicketMessageRequest) (*v1.TicketMessage, error) {
	// In a real system, sender_id would come from context (user/agent ID)
	// For simplicity, using req.SenderId directly.

	if req.TicketId == 0 || req.SenderId == 0 || req.SenderType == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_id, sender_id, sender_type, and content are required")
	}

	bizMessage, err := s.uc.AddTicketMessage(ctx, strconv.FormatUint(req.TicketId, 10), req.SenderId, req.SenderType, req.Content)
	if err != nil {
		if errors.Is(err, biz.ErrTicketNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to add ticket message: %v", err)
	}

	return bizTicketMessageToProto(bizMessage), nil
}

// ListTickets implements the ListTickets RPC.
func (s *CustomerService) ListTickets(ctx context.Context, req *v1.ListTicketsRequest) (*v1.ListTicketsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bizTickets, err := s.uc.ListTickets(ctx, userID, req.Status)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tickets: %v", err)
	}

	protoTickets := make([]*v1.Ticket, len(bizTickets))
	for i, t := range bizTickets {
		protoTickets[i] = bizTicketToProto(t)
	}

	return &v1.ListTicketsResponse{Tickets: protoTickets}, nil
}
