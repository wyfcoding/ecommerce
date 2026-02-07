package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

// FlashSaleProjectionService 负责将秒杀事件投影到读模型。
type FlashSaleProjectionService struct {
	repo            domain.FlashSaleRepository
	flashsaleRead   domain.FlashsaleReadRepository
	orderRead       domain.FlashsaleOrderReadRepository
	flashsaleSearch domain.FlashsaleSearchRepository
	orderSearch     domain.FlashsaleOrderSearchRepository
	logger          *slog.Logger
}

// NewFlashSaleProjectionService 创建投影服务。
func NewFlashSaleProjectionService(
	repo domain.FlashSaleRepository,
	flashsaleRead domain.FlashsaleReadRepository,
	orderRead domain.FlashsaleOrderReadRepository,
	flashsaleSearch domain.FlashsaleSearchRepository,
	orderSearch domain.FlashsaleOrderSearchRepository,
	logger *slog.Logger,
) *FlashSaleProjectionService {
	return &FlashSaleProjectionService{
		repo:            repo,
		flashsaleRead:   flashsaleRead,
		orderRead:       orderRead,
		flashsaleSearch: flashsaleSearch,
		orderSearch:     orderSearch,
		logger:          logger,
	}
}

func (s *FlashSaleProjectionService) OnFlashsaleCreated(ctx context.Context, event *domain.FlashSaleEventCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshFlashsale(ctx, uint64(event.EventID))
}

func (s *FlashSaleProjectionService) OnOrderCreated(ctx context.Context, event *domain.FlashSaleOrderCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshOrder(ctx, event.OrderID)
}

func (s *FlashSaleProjectionService) OnOrderCancelled(ctx context.Context, event *domain.FlashSaleOrderCancelledEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshOrder(ctx, event.OrderID)
}

func (s *FlashSaleProjectionService) OnOrderPaid(ctx context.Context, event *domain.FlashSaleOrderPaidEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshOrder(ctx, event.OrderID)
}

func (s *FlashSaleProjectionService) refreshFlashsale(ctx context.Context, id uint64) error {
	if s.flashsaleRead == nil && s.flashsaleSearch == nil {
		return nil
	}
	flashsale, err := s.repo.GetFlashsale(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load flashsale for projection", "flashsale_id", id, "error", err)
		return err
	}
	if flashsale == nil {
		if s.flashsaleRead != nil {
			_ = s.flashsaleRead.Delete(ctx, id)
		}
		if s.flashsaleSearch != nil {
			_ = s.flashsaleSearch.Delete(ctx, id)
		}
		return nil
	}
	if s.flashsaleRead != nil {
		if err := s.flashsaleRead.Save(ctx, flashsale); err != nil {
			s.logger.ErrorContext(ctx, "failed to save flashsale cache", "flashsale_id", id, "error", err)
			return err
		}
	}
	if s.flashsaleSearch != nil {
		if err := s.flashsaleSearch.Index(ctx, flashsale); err != nil {
			s.logger.ErrorContext(ctx, "failed to index flashsale", "flashsale_id", id, "error", err)
			return err
		}
	}
	return nil
}

func (s *FlashSaleProjectionService) refreshOrder(ctx context.Context, orderID uint64) error {
	if s.orderRead == nil && s.orderSearch == nil {
		return nil
	}
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load flashsale order for projection", "order_id", orderID, "error", err)
		return err
	}
	if order == nil {
		if s.orderRead != nil {
			_ = s.orderRead.Delete(ctx, orderID)
		}
		if s.orderSearch != nil {
			_ = s.orderSearch.Delete(ctx, orderID)
		}
		return nil
	}
	if s.orderRead != nil {
		if err := s.orderRead.Save(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to save order cache", "order_id", orderID, "error", err)
			return err
		}
	}
	if s.orderSearch != nil {
		if err := s.orderSearch.Index(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to index order", "order_id", orderID, "error", err)
			return err
		}
	}
	return nil
}
