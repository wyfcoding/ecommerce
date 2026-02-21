package domain

import "context"

// BlacklistReadRepository 定义黑名单读模型仓储接口（Redis）。
type BlacklistReadRepository interface {
	Save(ctx context.Context, entry *Blacklist) error
	GetByTypeValue(ctx context.Context, bType BlacklistType, value string) (*Blacklist, error)
	DeleteByTypeValue(ctx context.Context, bType BlacklistType, value string) error
	DeleteByID(ctx context.Context, id uint64) error
}
