package domain

import (
	"context"
)

// MultiChannelRepository 是多渠道模块的仓储接口。
type MultiChannelRepository interface {
	// --- tx helpers ---
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// Channel
	SaveChannel(ctx context.Context, channel *Channel) error
	SaveChannelInTx(ctx context.Context, tx any, channel *Channel) error
	GetChannel(ctx context.Context, id uint64) (*Channel, error)
	ListChannels(ctx context.Context, activeOnly bool) ([]*Channel, error)

	// LocalOrder
	SaveOrder(ctx context.Context, order *LocalOrder) error
	SaveOrderInTx(ctx context.Context, tx any, order *LocalOrder) error
	GetOrderByChannelID(ctx context.Context, channelID uint64, channelOrderID string) (*LocalOrder, error)
	ListOrders(ctx context.Context, query *LocalOrderQuery) ([]*LocalOrder, int64, error)

	// ChannelSyncLog
	SaveSyncLog(ctx context.Context, log *ChannelSyncLog) error
	SaveSyncLogInTx(ctx context.Context, tx any, log *ChannelSyncLog) error
}

// LocalOrderQuery 订单查询条件。
type LocalOrderQuery struct {
	ChannelID      uint64
	Status         string
	ChannelOrderID string
	Page           int
	PageSize       int
}
