package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
)

// MultiChannelQueryService 处理渠道的读操作。
type MultiChannelQueryService struct {
	repo        domain.MultiChannelRepository
	channelRead domain.ChannelReadRepository
	orderRead   domain.LocalOrderReadRepository
	orderSearch domain.LocalOrderSearchRepository
	logger      *slog.Logger
}

// NewMultiChannelQueryService creates a new MultiChannelQueryService instance.
func NewMultiChannelQueryService(
	repo domain.MultiChannelRepository,
	channelRead domain.ChannelReadRepository,
	orderRead domain.LocalOrderReadRepository,
	orderSearch domain.LocalOrderSearchRepository,
	logger *slog.Logger,
) *MultiChannelQueryService {
	return &MultiChannelQueryService{
		repo:        repo,
		channelRead: channelRead,
		orderRead:   orderRead,
		orderSearch: orderSearch,
		logger:      logger,
	}
}

// ListChannels 获取销售渠道列表。
func (q *MultiChannelQueryService) ListChannels(ctx context.Context) ([]*domain.Channel, error) {
	if q.channelRead != nil {
		if cached, err := q.channelRead.GetAll(ctx); err == nil && cached != nil {
			return cached, nil
		}
	}
	list, err := q.repo.ListChannels(ctx, false)
	if err != nil {
		return nil, err
	}
	if q.channelRead != nil {
		_ = q.channelRead.SaveAll(ctx, list)
	}
	return list, nil
}

// ListOrders 获取本地化存储的外部渠道订单列表。
func (q *MultiChannelQueryService) ListOrders(ctx context.Context, channelID uint64, status string, page, pageSize int) ([]*domain.LocalOrder, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	query := &domain.LocalOrderQuery{
		ChannelID: channelID,
		Status:    status,
		Page:      page,
		PageSize:  pageSize,
	}

	offset := (page - 1) * pageSize
	if q.orderSearch != nil {
		list, total, err := q.orderSearch.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		if q.logger != nil {
			q.logger.WarnContext(ctx, "local order search fallback to mysql", "error", err)
		}
	}
	return q.repo.ListOrders(ctx, query)
}
