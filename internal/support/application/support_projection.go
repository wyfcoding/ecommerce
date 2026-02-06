// 生成摘要：新增客服读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
)

// SupportProjectionService 负责将事件转换为读模型更新。
type SupportProjectionService struct {
	repo                          domain.TicketRepository
	ticketReadRepo                domain.TicketReadRepository
	ticketSearchRepo              domain.TicketSearchRepository
	ticketMessageSearchRepo       domain.TicketMessageSearchRepository
	conversationReadRepo          domain.ConversationReadRepository
	conversationMessageSearchRepo domain.ConversationMessageSearchRepository
	logger                        *slog.Logger
}

// NewSupportProjectionService 创建客服投影服务。
func NewSupportProjectionService(
	repo domain.TicketRepository,
	ticketReadRepo domain.TicketReadRepository,
	ticketSearchRepo domain.TicketSearchRepository,
	ticketMessageSearchRepo domain.TicketMessageSearchRepository,
	conversationReadRepo domain.ConversationReadRepository,
	conversationMessageSearchRepo domain.ConversationMessageSearchRepository,
	logger *slog.Logger,
) *SupportProjectionService {
	return &SupportProjectionService{
		repo:                          repo,
		ticketReadRepo:                ticketReadRepo,
		ticketSearchRepo:              ticketSearchRepo,
		ticketMessageSearchRepo:       ticketMessageSearchRepo,
		conversationReadRepo:          conversationReadRepo,
		conversationMessageSearchRepo: conversationMessageSearchRepo,
		logger:                        logger,
	}
}

// OnTicketCreated 处理工单创建事件。
func (s *SupportProjectionService) OnTicketCreated(ctx context.Context, event *domain.TicketCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshTicket(ctx, event.TicketID, event.TicketNo)
}

// OnTicketUpdated 处理工单更新事件。
func (s *SupportProjectionService) OnTicketUpdated(ctx context.Context, event *domain.TicketUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshTicket(ctx, event.TicketID, "")
}

// OnTicketMessageCreated 处理工单消息创建事件。
func (s *SupportProjectionService) OnTicketMessageCreated(ctx context.Context, event *domain.TicketMessageCreatedEvent) error {
	if event == nil || s.ticketMessageSearchRepo == nil {
		return nil
	}
	msg, err := s.repo.GetMessage(ctx, event.MessageID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load ticket message for projection", "message_id", event.MessageID, "error", err)
		return err
	}
	if msg == nil {
		_ = s.ticketMessageSearchRepo.Delete(ctx, event.MessageID)
		return nil
	}
	if err := s.ticketMessageSearchRepo.Index(ctx, msg); err != nil {
		s.logger.ErrorContext(ctx, "failed to index ticket message", "message_id", event.MessageID, "error", err)
		return err
	}
	return nil
}

// OnConversationCreated 处理会话创建事件。
func (s *SupportProjectionService) OnConversationCreated(ctx context.Context, event *domain.ConversationCreatedEvent) error {
	if event == nil {
		return nil
	}
	if s.conversationReadRepo == nil {
		return nil
	}
	conv, err := s.repo.GetConversation(ctx, event.ConversationID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load conversation for projection", "conversation_id", event.ConversationID, "error", err)
		return err
	}
	if conv == nil {
		_ = s.conversationReadRepo.Delete(ctx, event.ConversationID)
		return nil
	}
	if err := s.conversationReadRepo.Save(ctx, conv); err != nil {
		s.logger.ErrorContext(ctx, "failed to save conversation read model", "conversation_id", event.ConversationID, "error", err)
		return err
	}
	return nil
}

// OnConversationMessageCreated 处理会话消息创建事件。
func (s *SupportProjectionService) OnConversationMessageCreated(ctx context.Context, event *domain.ConversationMessageCreatedEvent) error {
	if event == nil || s.conversationMessageSearchRepo == nil {
		return nil
	}
	msg, err := s.repo.GetConversationMessage(ctx, event.MessageID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load conversation message for projection", "message_id", event.MessageID, "error", err)
		return err
	}
	if msg == nil {
		_ = s.conversationMessageSearchRepo.Delete(ctx, event.MessageID)
		return nil
	}
	if err := s.conversationMessageSearchRepo.Index(ctx, msg); err != nil {
		s.logger.ErrorContext(ctx, "failed to index conversation message", "message_id", event.MessageID, "error", err)
		return err
	}
	return nil
}

func (s *SupportProjectionService) refreshTicket(ctx context.Context, ticketID uint64, ticketNo string) error {
	if ticketID == 0 {
		return nil
	}
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load ticket for projection", "ticket_id", ticketID, "error", err)
		return err
	}
	if ticket == nil {
		if s.ticketReadRepo != nil {
			_ = s.ticketReadRepo.Delete(ctx, ticketID, ticketNo)
		}
		if s.ticketSearchRepo != nil {
			_ = s.ticketSearchRepo.Delete(ctx, ticketID)
		}
		return nil
	}
	if s.ticketReadRepo != nil {
		if err := s.ticketReadRepo.Save(ctx, ticket); err != nil {
			s.logger.ErrorContext(ctx, "failed to save ticket read model", "ticket_id", ticketID, "error", err)
			return err
		}
	}
	if s.ticketSearchRepo != nil {
		if err := s.ticketSearchRepo.Index(ctx, ticket); err != nil {
			s.logger.ErrorContext(ctx, "failed to index ticket", "ticket_id", ticketID, "error", err)
			return err
		}
	}
	return nil
}
