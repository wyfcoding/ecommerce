package domain

import "context"

// ChannelReadRepository 定义渠道读模型仓储接口（Redis）。
type ChannelReadRepository interface {
	SaveAll(ctx context.Context, channels []*Channel) error
	GetAll(ctx context.Context) ([]*Channel, error)
	DeleteAll(ctx context.Context) error
}
