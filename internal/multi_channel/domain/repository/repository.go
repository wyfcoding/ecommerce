package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity" // 导入多渠道领域的实体定义。
)

// MultiChannelRepository 是多渠道模块的仓储接口。
// 它定义了对销售渠道、本地订单和渠道同步日志实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type MultiChannelRepository interface {
	// --- 渠道 (Channel methods) ---

	// SaveChannel 将销售渠道实体保存到数据存储中。
	// 如果渠道已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// channel: 待保存的渠道实体。
	SaveChannel(ctx context.Context, channel *entity.Channel) error
	// GetChannel 根据ID获取销售渠道实体。
	GetChannel(ctx context.Context, id uint64) (*entity.Channel, error)
	// ListChannels 列出所有销售渠道实体。
	// activeOnly: 布尔值，如果为true，则只列出启用的渠道。
	ListChannels(ctx context.Context, activeOnly bool) ([]*entity.Channel, error)

	// --- 订单 (LocalOrder methods) ---

	// SaveOrder 将本地订单实体保存到数据存储中。
	SaveOrder(ctx context.Context, order *entity.LocalOrder) error
	// GetOrderByChannelID 根据渠道ID和渠道订单ID获取本地订单实体。
	GetOrderByChannelID(ctx context.Context, channelID uint64, channelOrderID string) (*entity.LocalOrder, error)
	// ListOrders 列出所有本地订单实体，支持通过渠道ID和状态过滤，并支持分页。
	ListOrders(ctx context.Context, channelID uint64, status string, offset, limit int) ([]*entity.LocalOrder, int64, error)

	// --- 日志 (ChannelSyncLog methods) ---

	// SaveSyncLog 将渠道同步日志实体保存到数据存储中。
	SaveSyncLog(ctx context.Context, log *entity.ChannelSyncLog) error
}
