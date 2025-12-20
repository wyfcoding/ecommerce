package domain

import (
	"context"
)

// MultiChannelRepository 是多渠道模块的仓储接口。
type MultiChannelRepository interface {
	// Channel
	SaveChannel(ctx context.Context, channel *Channel) error
	GetChannel(ctx context.Context, id uint64) (*Channel, error)
	ListChannels(ctx context.Context, activeOnly bool) ([]*Channel, error)

	// LocalOrder
	SaveOrder(ctx context.Context, order *LocalOrder) error
	GetOrderByChannelID(ctx context.Context, channelID uint64, channelOrderID string) (*LocalOrder, error)
	ListOrders(ctx context.Context, channelID uint64, status string, offset, limit int) ([]*LocalOrder, int64, error)

	// ChannelSyncLog
	SaveSyncLog(ctx context.Context, log *ChannelSyncLog) error
}
