package repository

import (
	"context"
	"ecommerce/internal/groupbuy/domain/entity"
)

// GroupbuyRepository 拼团仓储接口
type GroupbuyRepository interface {
	// Groupbuy methods
	CreateGroupbuy(ctx context.Context, groupbuy *entity.Groupbuy) error
	GetGroupbuyByID(ctx context.Context, id uint64) (*entity.Groupbuy, error)
	UpdateGroupbuy(ctx context.Context, groupbuy *entity.Groupbuy) error
	ListGroupbuys(ctx context.Context, page, pageSize int) ([]*entity.Groupbuy, int64, error)

	// GroupbuyTeam methods
	CreateTeam(ctx context.Context, team *entity.GroupbuyTeam) error
	GetTeamByID(ctx context.Context, id uint64) (*entity.GroupbuyTeam, error)
	GetTeamByNo(ctx context.Context, teamNo string) (*entity.GroupbuyTeam, error)
	UpdateTeam(ctx context.Context, team *entity.GroupbuyTeam) error
	ListTeamsByGroupbuyID(ctx context.Context, groupbuyID uint64, page, pageSize int) ([]*entity.GroupbuyTeam, int64, error)

	// GroupbuyOrder methods
	CreateOrder(ctx context.Context, order *entity.GroupbuyOrder) error
	GetOrderByID(ctx context.Context, id uint64) (*entity.GroupbuyOrder, error)
	UpdateOrder(ctx context.Context, order *entity.GroupbuyOrder) error
	ListOrdersByTeamID(ctx context.Context, teamID uint64) ([]*entity.GroupbuyOrder, error)
	ListOrdersByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.GroupbuyOrder, int64, error)
}
