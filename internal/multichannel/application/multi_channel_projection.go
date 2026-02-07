// 生成摘要：新增多渠道读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
)

// MultiChannelProjectionService 负责将多渠道事件投影到读模型。
type MultiChannelProjectionService struct {
	repo        domain.MultiChannelRepository
	channelRead domain.ChannelReadRepository
	orderRead   domain.LocalOrderReadRepository
	orderSearch domain.LocalOrderSearchRepository
	logger      *slog.Logger
}

// NewMultiChannelProjectionService 创建投影服务。
func NewMultiChannelProjectionService(
	repo domain.MultiChannelRepository,
	channelRead domain.ChannelReadRepository,
	orderRead domain.LocalOrderReadRepository,
	orderSearch domain.LocalOrderSearchRepository,
	logger *slog.Logger,
) *MultiChannelProjectionService {
	return &MultiChannelProjectionService{
		repo:        repo,
		channelRead: channelRead,
		orderRead:   orderRead,
		orderSearch: orderSearch,
		logger:      logger,
	}
}

func (s *MultiChannelProjectionService) OnChannelRegistered(ctx context.Context, _ *domain.ChannelRegisteredEvent) error {
	return s.refreshChannels(ctx)
}

func (s *MultiChannelProjectionService) OnChannelOrderCreated(ctx context.Context, event *domain.ChannelOrderCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshOrder(ctx, event.ChannelID, event.ExternalID, event.OrderID)
}

func (s *MultiChannelProjectionService) refreshChannels(ctx context.Context) error {
	if s.channelRead == nil {
		return nil
	}
	list, err := s.repo.ListChannels(ctx, false)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load channels for projection", "error", err)
		return err
	}
	if err := s.channelRead.SaveAll(ctx, list); err != nil {
		s.logger.ErrorContext(ctx, "failed to save channel cache", "error", err)
		return err
	}
	return nil
}

func (s *MultiChannelProjectionService) refreshOrder(ctx context.Context, channelID uint64, channelOrderID string, orderID uint64) error {
	if s.orderRead == nil && s.orderSearch == nil {
		return nil
	}
	order, err := s.repo.GetOrderByChannelID(ctx, channelID, channelOrderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load order for projection", "channel_id", channelID, "external_id", channelOrderID, "error", err)
		return err
	}
	if order == nil {
		if s.orderRead != nil {
			_ = s.orderRead.Delete(ctx, channelID, channelOrderID)
		}
		if s.orderSearch != nil && orderID > 0 {
			_ = s.orderSearch.Delete(ctx, orderID)
		}
		return nil
	}
	if s.orderRead != nil {
		if err := s.orderRead.Save(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to save order cache", "order_id", order.ID, "error", err)
			return err
		}
	}
	if s.orderSearch != nil {
		if err := s.orderSearch.Index(ctx, order); err != nil {
			s.logger.ErrorContext(ctx, "failed to index order", "order_id", order.ID, "error", err)
			return err
		}
	}
	return nil
}
