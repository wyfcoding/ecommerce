package domain

import "context"

// LocalOrderReadRepository 定义渠道订单读模型仓储接口（Redis）。
type LocalOrderReadRepository interface {
	Save(ctx context.Context, order *LocalOrder) error
	GetByChannelOrderID(ctx context.Context, channelID uint64, channelOrderID string) (*LocalOrder, error)
	Delete(ctx context.Context, channelID uint64, channelOrderID string) error
}
