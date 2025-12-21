package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain"
)

// MultiChannelService 作为多渠道操作的门面。
type MultiChannelService struct {
	manager *ChannelManager
	query   *ChannelQuery
}

// NewMultiChannelService 创建多渠道服务门面实例。
func NewMultiChannelService(manager *ChannelManager, query *ChannelQuery) *MultiChannelService {
	return &MultiChannelService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// RegisterChannel 注册一个新的外部渠道（如第三方电商平台）。
func (s *MultiChannelService) RegisterChannel(ctx context.Context, channel *domain.Channel) error {
	return s.manager.RegisterChannel(ctx, channel)
}

// SyncOrders 触发与指定外部渠道的订单同步操作。
func (s *MultiChannelService) SyncOrders(ctx context.Context, channelID uint64) error {
	return s.manager.SyncOrders(ctx, channelID)
}

// --- 读操作（委托给 Query）---

// ListChannels 获取所有已注册的渠道列表。
func (s *MultiChannelService) ListChannels(ctx context.Context) ([]*domain.Channel, error) {
	return s.query.ListChannels(ctx)
}

// ListOrders 获取特定渠道同步过来的订单列表（分页）。
func (s *MultiChannelService) ListOrders(ctx context.Context, channelID uint64, status string, page, pageSize int) ([]*domain.LocalOrder, int64, error) {
	return s.query.ListOrders(ctx, channelID, status, page, pageSize)
}
