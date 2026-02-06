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
