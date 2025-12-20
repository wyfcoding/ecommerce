package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain"
)

// ChannelQuery handles read operations for channels.
type ChannelQuery struct {
	repo domain.MultiChannelRepository
}

// NewChannelQuery creates a new ChannelQuery instance.
func NewChannelQuery(repo domain.MultiChannelRepository) *ChannelQuery {
	return &ChannelQuery{
		repo: repo,
	}
}

// ListChannels 获取销售渠道列表。
func (q *ChannelQuery) ListChannels(ctx context.Context) ([]*domain.Channel, error) {
	return q.repo.ListChannels(ctx, false)
}

// ListOrders 获取本地化存储的外部渠道订单列表。
func (q *ChannelQuery) ListOrders(ctx context.Context, channelID uint64, status string, page, pageSize int) ([]*domain.LocalOrder, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListOrders(ctx, channelID, status, offset, pageSize)
}
