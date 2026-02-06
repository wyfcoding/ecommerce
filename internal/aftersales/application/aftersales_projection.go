// 生成摘要：新增售后读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
)

// AfterSalesProjectionService 负责将事件转换为读模型更新。
type AfterSalesProjectionService struct {
	repo                           domain.AfterSalesRepository
	readRepo                       domain.AfterSalesReadRepository
	searchRepo                     domain.AfterSalesSearchRepository
	supportTicketReadRepo          domain.SupportTicketReadRepository
	supportTicketSearchRepo        domain.SupportTicketSearchRepository
	supportTicketMessageSearchRepo domain.SupportTicketMessageSearchRepository
	configReadRepo                 domain.AfterSalesConfigReadRepository
	logger                         *slog.Logger
}

// NewAfterSalesProjectionService 创建售后投影服务。
func NewAfterSalesProjectionService(
	repo domain.AfterSalesRepository,
	readRepo domain.AfterSalesReadRepository,
	searchRepo domain.AfterSalesSearchRepository,
	supportTicketReadRepo domain.SupportTicketReadRepository,
	supportTicketSearchRepo domain.SupportTicketSearchRepository,
	supportTicketMessageSearchRepo domain.SupportTicketMessageSearchRepository,
	configReadRepo domain.AfterSalesConfigReadRepository,
	logger *slog.Logger,
) *AfterSalesProjectionService {
	return &AfterSalesProjectionService{
		repo:                           repo,
		readRepo:                       readRepo,
		searchRepo:                     searchRepo,
		supportTicketReadRepo:          supportTicketReadRepo,
		supportTicketSearchRepo:        supportTicketSearchRepo,
		supportTicketMessageSearchRepo: supportTicketMessageSearchRepo,
		configReadRepo:                 configReadRepo,
		logger:                         logger,
	}
}

// OnAfterSalesCreated 处理售后创建事件。
func (s *AfterSalesProjectionService) OnAfterSalesCreated(ctx context.Context, event *domain.AfterSalesCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshAfterSales(ctx, event.AfterSalesID, event.AfterSalesNo)
}

// OnAfterSalesStatusUpdated 处理售后状态更新事件。
func (s *AfterSalesProjectionService) OnAfterSalesStatusUpdated(ctx context.Context, event *domain.AfterSalesStatusUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshAfterSales(ctx, event.AfterSalesID, event.AfterSalesNo)
}

// OnSupportTicketCreated 处理客服工单创建事件。
func (s *AfterSalesProjectionService) OnSupportTicketCreated(ctx context.Context, event *domain.SupportTicketCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshSupportTicket(ctx, event.TicketID)
}

// OnSupportTicketUpdated 处理客服工单更新事件。
func (s *AfterSalesProjectionService) OnSupportTicketUpdated(ctx context.Context, event *domain.SupportTicketUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshSupportTicket(ctx, event.TicketID)
}

// OnSupportTicketMessageCreated 处理客服工单消息创建事件。
func (s *AfterSalesProjectionService) OnSupportTicketMessageCreated(ctx context.Context, event *domain.SupportTicketMessageCreatedEvent) error {
	if event == nil || s.supportTicketMessageSearchRepo == nil {
		return nil
	}
	msg, err := s.repo.GetSupportTicketMessage(ctx, event.MessageID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load support ticket message for projection", "message_id", event.MessageID, "error", err)
		return err
	}
	if msg == nil {
		_ = s.supportTicketMessageSearchRepo.Delete(ctx, event.MessageID)
		return nil
	}
	if err := s.supportTicketMessageSearchRepo.Index(ctx, msg); err != nil {
		s.logger.ErrorContext(ctx, "failed to index support ticket message", "message_id", event.MessageID, "error", err)
		return err
	}
	return nil
}

// OnConfigUpdated 处理配置更新事件。
func (s *AfterSalesProjectionService) OnConfigUpdated(ctx context.Context, event *domain.AfterSalesConfigUpdatedEvent) error {
	if event == nil {
		return nil
	}
	if s.configReadRepo == nil {
		return nil
	}
	cfg, err := s.repo.GetConfig(ctx, event.Key)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load aftersales config", "key", event.Key, "error", err)
		return err
	}
	if cfg == nil {
		_ = s.configReadRepo.Delete(ctx, event.Key)
		return nil
	}
	return s.configReadRepo.Save(ctx, cfg)
}

func (s *AfterSalesProjectionService) refreshAfterSales(ctx context.Context, id uint64, no string) error {
	if id == 0 {
		return nil
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load aftersales for projection", "id", id, "error", err)
		return err
	}
	if item == nil {
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, id, no)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.Delete(ctx, id)
		}
		return nil
	}
	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, item); err != nil {
			s.logger.ErrorContext(ctx, "failed to save aftersales read model", "id", id, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, item); err != nil {
			s.logger.ErrorContext(ctx, "failed to index aftersales", "id", id, "error", err)
			return err
		}
	}
	return nil
}

func (s *AfterSalesProjectionService) refreshSupportTicket(ctx context.Context, id uint64) error {
	if id == 0 {
		return nil
	}
	ticket, err := s.repo.GetSupportTicket(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load support ticket for projection", "id", id, "error", err)
		return err
	}
	if ticket == nil {
		if s.supportTicketReadRepo != nil {
			_ = s.supportTicketReadRepo.Delete(ctx, id)
		}
		if s.supportTicketSearchRepo != nil {
			_ = s.supportTicketSearchRepo.Delete(ctx, id)
		}
		return nil
	}
	if s.supportTicketReadRepo != nil {
		if err := s.supportTicketReadRepo.Save(ctx, ticket); err != nil {
			s.logger.ErrorContext(ctx, "failed to save support ticket read model", "id", id, "error", err)
			return err
		}
	}
	if s.supportTicketSearchRepo != nil {
		if err := s.supportTicketSearchRepo.Index(ctx, ticket); err != nil {
			s.logger.ErrorContext(ctx, "failed to index support ticket", "id", id, "error", err)
			return err
		}
	}
	return nil
}
