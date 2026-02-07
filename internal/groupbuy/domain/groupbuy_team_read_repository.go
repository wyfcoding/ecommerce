package domain

import "context"

// GroupbuyTeamReadRepository 定义拼团团队读模型仓储接口（Redis）。
type GroupbuyTeamReadRepository interface {
	Save(ctx context.Context, team *GroupbuyTeam) error
	GetByID(ctx context.Context, id uint64) (*GroupbuyTeam, error)
	Delete(ctx context.Context, id uint64) error
}
