package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
)

// Customer 作为客户服务操作的门面。
type Support struct {
	Command *SupportCommandService
	Query   *SupportQueryService
}

// NewSupport 创建并返回一个新的 Support 门面实例。
func NewSupport(command *SupportCommandService, query *SupportQueryService) *Support {
	return &Support{
		Command: command,
		Query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreateTicket 创建一个新的客服工单。
func (s *Support) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority domain.TicketPriority) (*domain.Ticket, error) {
	return s.Command.CreateTicket(ctx, userID, subject, description, category, priority)
}

// ReplyTicket 为指定工单添加一条新回复。
func (s *Support) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType domain.MessageType) (*domain.Message, error) {
	return s.Command.ReplyTicket(ctx, ticketID, senderID, senderType, content, msgType)
}

// CloseTicket 关闭指定的客服工单。
func (s *Support) CloseTicket(ctx context.Context, id uint64) error {
	return s.Command.CloseTicket(ctx, id)
}

// ResolveTicket 将工单状态标记为已解决。
func (s *Support) ResolveTicket(ctx context.Context, id uint64) error {
	return s.Command.ResolveTicket(ctx, id)
}

// --- 读操作（委托给 Query）---

// GetTicket 获取指定ID的工单详情。
func (s *Support) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	return s.Query.GetTicket(ctx, id)
}

// ListTickets 获取用户的工单列表。
func (s *Support) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, page, pageSize int) ([]*domain.Ticket, int64, error) {
	return s.Query.ListTickets(ctx, userID, status, page, pageSize)
}

// ListMessages 获取指定工单下的所有聊天消息。
func (s *Support) ListMessages(ctx context.Context, ticketID uint64, page, pageSize int) ([]*domain.Message, int64, error) {
	return s.Query.ListMessages(ctx, ticketID, page, pageSize)
}

// --- Conversation (P2P) 操作 ---

// StartConversation 开启一个新的私聊会话。
func (s *Support) StartConversation(ctx context.Context, user1ID, user2ID uint64) (*domain.Conversation, error) {
	return s.Command.StartConversation(ctx, user1ID, user2ID)
}

// SendConversationMessage 发送私聊消息。
func (s *Support) SendConversationMessage(ctx context.Context, convID, senderID, receiverID uint64, content string, msgType domain.MessageType) (*domain.ConversationMessage, error) {
	return s.Command.SendConversationMessage(ctx, convID, senderID, receiverID, content, msgType)
}

// ListConversationMessages 获取指定会话的消息列表。
func (s *Support) ListConversationMessages(ctx context.Context, convID uint64, page, pageSize int) ([]*domain.ConversationMessage, int64, error) {
	return s.Query.ListConversationMessages(ctx, convID, page, pageSize)
}
