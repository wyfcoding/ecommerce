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

// NewMultiChannelService creates a new MultiChannelService facade.
func NewMultiChannelService(manager *ChannelManager, query *ChannelQuery) *MultiChannelService {
	return &MultiChannelService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *MultiChannelService) RegisterChannel(ctx context.Context, channel *domain.Channel) error {
	return s.manager.RegisterChannel(ctx, channel)
}

func (s *MultiChannelService) SyncOrders(ctx context.Context, channelID uint64) error {
	return s.manager.SyncOrders(ctx, channelID)
}

// --- 读操作（委托给 Query）---

func (s *MultiChannelService) ListChannels(ctx context.Context) ([]*domain.Channel, error) {
	return s.query.ListChannels(ctx)
}

func (s *MultiChannelService) ListOrders(ctx context.Context, channelID uint64, status string, page, pageSize int) ([]*domain.LocalOrder, int64, error) {
	return s.query.ListOrders(ctx, channelID, status, page, pageSize)
}
