package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity"
)

// MultiChannelRepository 多渠道仓储接口
type MultiChannelRepository interface {
	// 渠道
	SaveChannel(ctx context.Context, channel *entity.Channel) error
	GetChannel(ctx context.Context, id uint64) (*entity.Channel, error)
	ListChannels(ctx context.Context, activeOnly bool) ([]*entity.Channel, error)

	// 订单
	SaveOrder(ctx context.Context, order *entity.LocalOrder) error
	GetOrderByChannelID(ctx context.Context, channelID uint64, channelOrderID string) (*entity.LocalOrder, error)
	ListOrders(ctx context.Context, channelID uint64, status string, offset, limit int) ([]*entity.LocalOrder, int64, error)

	// 日志
	SaveSyncLog(ctx context.Context, log *entity.ChannelSyncLog) error
}
