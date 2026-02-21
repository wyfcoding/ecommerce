package domain

import (
	"context"
)

// TicketRepository 定义了 customer 模块的仓储接口。
type TicketRepository interface {
	// 事务管理
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// Ticket
	SaveTicket(ctx context.Context, ticket *Ticket) error
	SaveTicketInTx(ctx context.Context, tx any, ticket *Ticket) error
	GetTicket(ctx context.Context, id uint64) (*Ticket, error)
	GetTicketByNo(ctx context.Context, ticketNo string) (*Ticket, error)
	UpdateTicket(ctx context.Context, ticket *Ticket) error
	UpdateTicketInTx(ctx context.Context, tx any, ticket *Ticket) error
	ListTickets(ctx context.Context, userID uint64, status TicketStatus, offset, limit int) ([]*Ticket, int64, error)
	GetCustomerSegmentationStats(ctx context.Context) ([]struct {
		UserID      uint64
		TicketCount float64
		AvgPriority float64
	}, error)
	GetTicketsForAutomation(ctx context.Context) ([]*Ticket, error)

	// Ticket Message
	SaveMessage(ctx context.Context, message *Message) error
	SaveMessageInTx(ctx context.Context, tx any, message *Message) error
	GetMessage(ctx context.Context, id uint64) (*Message, error)
	ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*Message, int64, error)

	// Conversation (P2P Chat)
	SaveConversation(ctx context.Context, conversation *Conversation) error
	SaveConversationInTx(ctx context.Context, tx any, conversation *Conversation) error
	GetConversation(ctx context.Context, id uint64) (*Conversation, error)
	SaveConversationMessage(ctx context.Context, message *ConversationMessage) error
	SaveConversationMessageInTx(ctx context.Context, tx any, message *ConversationMessage) error
	GetConversationMessage(ctx context.Context, id uint64) (*ConversationMessage, error)
	ListConversationMessages(ctx context.Context, conversationID uint64, offset, limit int) ([]*ConversationMessage, int64, error)

	// --- Intelligent Support (KB & AI Chat) ---

	// Knowledge Base
	SaveKnowledgeBase(ctx context.Context, kb *KnowledgeBase) error
	GetKnowledgeBase(ctx context.Context, id string) (*KnowledgeBase, error)
	GetKnowledgeArticle(ctx context.Context, id string) (*KnowledgeArticle, error)
	SearchKnowledgeArticles(ctx context.Context, query string, limit int) ([]*KnowledgeArticle, error)

	// AI Conversation
	SaveAIConversation(ctx context.Context, conversation *AIConversation) error
	GetAIConversation(ctx context.Context, id string) (*AIConversation, error)
	SaveAIMessage(ctx context.Context, message *AIMessage) error
	ListAIMessages(ctx context.Context, conversationID string) ([]*AIMessage, error)
}
