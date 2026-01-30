package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
)

// MultiChannelQueryService 处理渠道的读操作。
type MultiChannelQueryService struct {
	repo domain.MultiChannelRepository
}

// NewMultiChannelQueryService creates a new MultiChannelQueryService instance.
func NewMultiChannelQueryService(repo domain.MultiChannelRepository) *MultiChannelQueryService {
	return &MultiChannelQueryService{
		repo: repo,
	}
}

// ListChannels 获取销售渠道列表。
func (q *MultiChannelQueryService) ListChannels(ctx context.Context) ([]*domain.Channel, error) {
	return q.repo.ListChannels(ctx, false)
}

// ListOrders 获取本地化存储的外部渠道订单列表。
func (q *MultiChannelQueryService) ListOrders(ctx context.Context, channelID uint64, status string, page, pageSize int) ([]*domain.LocalOrder, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListOrders(ctx, channelID, status, offset, pageSize)
}
