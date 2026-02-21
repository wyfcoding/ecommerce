package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/support/v1"
	"github.com/wyfcoding/ecommerce/internal/support/application"
	"github.com/wyfcoding/ecommerce/internal/support/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 Customer 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedSupportServiceServer
	cmd   *application.SupportCommandService
	query *application.SupportQueryService
}

// NewServer 创建并返回一个新的 Customer gRPC 服务端实例。
func NewServer(cmd *application.SupportCommandService, query *application.SupportQueryService) *Server {
	return &Server{cmd: cmd, query: query}
}

// CreateTicket 处理创建工单的gRPC请求。
func (s *Server) CreateTicket(ctx context.Context, req *pb.CreateTicketRequest) (*pb.TicketResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateTicket received", "user_id", req.UserId, "subject", req.Subject)

	ticket, err := s.cmd.CreateTicket(ctx, req.UserId, req.Subject, req.Description, "general", domain.TicketPriorityMedium)
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

	ticket, err := s.query.GetTicket(ctx, req.TicketId)
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
		err = s.cmd.CloseTicket(ctx, req.TicketId)
	case "RESOLVED":
		err = s.cmd.ResolveTicket(ctx, req.TicketId)
	default:
		slog.Warn("gRPC UpdateTicketStatus unsupported status", "ticket_id", req.TicketId, "status", req.Status)
		return nil, status.Errorf(codes.InvalidArgument, "status transition to %s is not supported", req.Status)
	}

	if err != nil {
		slog.Error("gRPC UpdateTicketStatus failed", "ticket_id", req.TicketId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update ticket status: %v", err))
	}

	ticket, err := s.query.GetTicket(ctx, req.TicketId)
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

	msg, err := s.cmd.ReplyTicket(ctx, req.TicketId, req.SenderId, req.SenderType, req.MessageBody, domain.MessageTypeText)
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

	msgs, _, err := s.query.ListMessages(ctx, req.TicketId, 1, 100)
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

// --- Knowledge Base Methods ---

func (s *Server) CreateKnowledgeBase(ctx context.Context, req *pb.CreateKnowledgeBaseRequest) (*pb.KnowledgeBase, error) {
	kb, err := s.cmd.CreateKnowledgeBase(ctx, req.Name, req.Description, req.Language)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return convertKnowledgeBaseToProto(kb), nil
}

func (s *Server) GetKnowledgeBase(ctx context.Context, req *pb.GetKnowledgeBaseRequest) (*pb.KnowledgeBase, error) {
	kb, err := s.query.GetKnowledgeBase(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return convertKnowledgeBaseToProto(kb), nil
}

func (s *Server) SearchKnowledgeArticles(ctx context.Context, req *pb.SearchKnowledgeArticlesRequest) (*pb.SearchKnowledgeArticlesResponse, error) {
	articles, err := s.query.SearchKnowledgeArticles(ctx, req.Query, int(req.Limit))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	pbArticles := make([]*pb.KnowledgeArticle, len(articles))
	for i, a := range articles {
		pbArticles[i] = &pb.KnowledgeArticle{
			Id:      a.ID,
			Title:   a.Title,
			Content: a.Content,
			Score:   float32(a.Score),
		}
	}
	return &pb.SearchKnowledgeArticlesResponse{Articles: pbArticles}, nil
}

// --- AI Conversation Methods ---

func (s *Server) StartAIConversation(ctx context.Context, req *pb.StartAIConversationRequest) (*pb.Conversation, error) {
	conv, err := s.cmd.StartAIConversation(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Conversation{
		Id:        conv.ID,
		UserId:    conv.UserID,
		Status:    pb.ConversationStatus(conv.Status),
		StartedAt: timestamppb.New(conv.StartedAt),
	}, nil
}

func (s *Server) SendAIMessage(ctx context.Context, req *pb.SendAIMessageRequest) (*pb.AIMessage, error) {
	// 简单转发给 GetChatbotResponse 逻辑或独立保存
	result, err := s.cmd.GetChatbotResponse(ctx, req.ConversationId, req.Content)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AIMessage{
		ConversationId: req.ConversationId,
		Content:        result.SuggestedResponse,
		Sender:         pb.MessageSender_BOT,
		CreatedAt:      timestamppb.Now(),
	}, nil
}

func (s *Server) GetChatbotResponse(ctx context.Context, req *pb.GetChatbotResponseRequest) (*pb.ChatbotResponse, error) {
	result, err := s.cmd.GetChatbotResponse(ctx, req.ConversationId, req.UserMessage)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ChatbotResponse{
		ResponseText:   result.SuggestedResponse,
		DetectedIntent: pb.IntentCategory(result.Intent),
		Confidence:     float32(result.Confidence),
	}, nil
}

// --- AI Analytics Methods ---

func (s *Server) AnalyzeSentiment(ctx context.Context, req *pb.AnalyzeSentimentRequest) (*pb.SentimentAnalysis, error) {
	// 预留模拟逻辑
	return &pb.SentimentAnalysis{
		Sentiment:  pb.Sentiment_NEUTRAL,
		Confidence: 0.9,
	}, nil
}

func (s *Server) GetIntent(ctx context.Context, req *pb.GetIntentRequest) (*pb.IntentResult, error) {
	// 预留模拟逻辑
	return &pb.IntentResult{
		Intent:            pb.IntentCategory_GENERAL_INQUIRY,
		Confidence:        0.85,
		SuggestedResponse: "我已经理解您的意图，正在处理中。",
	}, nil
}

// --- Conversion Helpers ---

func convertKnowledgeBaseToProto(kb *domain.KnowledgeBase) *pb.KnowledgeBase {
	if kb == nil {
		return nil
	}
	return &pb.KnowledgeBase{
		Id:          kb.ID,
		Name:        kb.Name,
		Description: kb.Description,
		Language:    kb.Language,
		IsActive:    kb.IsActive,
		CreatedAt:   timestamppb.New(kb.CreatedAt),
	}
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
