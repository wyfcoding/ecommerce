package application

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
)

// SupportQueryService 处理客户服务的读操作。
type SupportQueryService struct {
	repo                          domain.TicketRepository
	ticketReadRepo                domain.TicketReadRepository
	ticketSearchRepo              domain.TicketSearchRepository
	ticketMessageSearchRepo       domain.TicketMessageSearchRepository
	conversationMessageSearchRepo domain.ConversationMessageSearchRepository
}

// NewSupportQueryService 创建并返回一个新的 SupportQueryService 实例。
func NewSupportQueryService(
	repo domain.TicketRepository,
	ticketReadRepo domain.TicketReadRepository,
	ticketSearchRepo domain.TicketSearchRepository,
	ticketMessageSearchRepo domain.TicketMessageSearchRepository,
	conversationMessageSearchRepo domain.ConversationMessageSearchRepository,
) *SupportQueryService {
	return &SupportQueryService{
		repo:                          repo,
		ticketReadRepo:                ticketReadRepo,
		ticketSearchRepo:              ticketSearchRepo,
		ticketMessageSearchRepo:       ticketMessageSearchRepo,
		conversationMessageSearchRepo: conversationMessageSearchRepo,
	}
}

// GetTicket 获取指定ID的工单详情。
func (q *SupportQueryService) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	if q.ticketReadRepo != nil {
		if cached, err := q.ticketReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	ticket, err := q.repo.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}
	if ticket != nil && q.ticketReadRepo != nil {
		_ = q.ticketReadRepo.Save(ctx, ticket)
	}
	return ticket, nil
}

// ListTickets 获取工单列表，支持通过用户ID和状态过滤。
func (q *SupportQueryService) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, page, pageSize int) ([]*domain.Ticket, int64, error) {
	offset := (page - 1) * pageSize
	var statusPtr *domain.TicketStatus
	if status != 0 {
		statusPtr = &status
	}
	if q.ticketSearchRepo != nil {
		list, total, err := q.ticketSearchRepo.Search(ctx, userID, statusPtr, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
	}
	return q.repo.ListTickets(ctx, userID, status, offset, pageSize)
}

// ListMessages 获取指定工单的所有消息列表。
func (q *SupportQueryService) ListMessages(ctx context.Context, ticketID uint64, page, pageSize int) ([]*domain.Message, int64, error) {
	offset := (page - 1) * pageSize
	if q.ticketMessageSearchRepo != nil {
		list, total, err := q.ticketMessageSearchRepo.Search(ctx, ticketID, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
	}
	return q.repo.ListMessages(ctx, ticketID, offset, pageSize)
}

// ListConversationMessages 获取指定会话的消息列表。
func (q *SupportQueryService) ListConversationMessages(ctx context.Context, convID uint64, page, pageSize int) ([]*domain.ConversationMessage, int64, error) {
	offset := (page - 1) * pageSize
	if q.conversationMessageSearchRepo != nil {
		list, total, err := q.conversationMessageSearchRepo.Search(ctx, convID, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
	}
	return q.repo.ListConversationMessages(ctx, convID, offset, pageSize)
}

// --- Intelligent Support Query Methods ---

// GetKnowledgeBase 获取知识库信息。
func (q *SupportQueryService) GetKnowledgeBase(ctx context.Context, id string) (*domain.KnowledgeBase, error) {
	kb, err := q.repo.GetKnowledgeBase(ctx, id)
	if err != nil {
		return nil, err
	}
	if kb == nil {
		return nil, errors.New("knowledge base not found")
	}
	return kb, nil
}

// GetKnowledgeArticle 获取文章详情。
func (q *SupportQueryService) GetKnowledgeArticle(ctx context.Context, id string) (*domain.KnowledgeArticle, error) {
	article, err := q.repo.GetKnowledgeArticle(ctx, id)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, errors.New("knowledge article not found")
	}
	return article, nil
}

// SearchKnowledgeArticles 搜索知识库文章。
func (q *SupportQueryService) SearchKnowledgeArticles(ctx context.Context, query string, limit int) ([]*domain.KnowledgeArticle, error) {
	if limit <= 0 {
		limit = 10
	}
	return q.repo.SearchKnowledgeArticles(ctx, query, limit)
}

// GetAIConversation 获取 AI 会话详情。
func (q *SupportQueryService) GetAIConversation(ctx context.Context, id string) (*domain.AIConversation, error) {
	conv, err := q.repo.GetAIConversation(ctx, id)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, errors.New("AI conversation not found")
	}
	return conv, nil
}

// ListAIMessages 获取 AI 会话的消息记录。
func (q *SupportQueryService) ListAIMessages(ctx context.Context, conversationID string) ([]*domain.AIMessage, error) {
	return q.repo.ListAIMessages(ctx, conversationID)
}
