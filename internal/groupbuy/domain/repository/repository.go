package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/entity" // 导入拼团领域的实体定义。
)

// GroupbuyRepository 是拼团模块的仓储接口。
// 它定义了对拼团活动、拼团团队和拼团订单实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type GroupbuyRepository interface {
	// --- Groupbuy methods ---

	// CreateGroupbuy 在数据存储中创建一个新的拼团活动实体。
	// ctx: 上下文。
	// groupbuy: 待创建的拼团活动实体。
	CreateGroupbuy(ctx context.Context, groupbuy *entity.Groupbuy) error
	// GetGroupbuyByID 根据ID获取拼团活动实体。
	GetGroupbuyByID(ctx context.Context, id uint64) (*entity.Groupbuy, error)
	// UpdateGroupbuy 更新拼团活动实体的信息。
	UpdateGroupbuy(ctx context.Context, groupbuy *entity.Groupbuy) error
	// ListGroupbuys 列出所有拼团活动实体，支持分页。
	ListGroupbuys(ctx context.Context, page, pageSize int) ([]*entity.Groupbuy, int64, error)

	// --- GroupbuyTeam methods ---

	// CreateTeam 在数据存储中创建一个新的拼团团队实体。
	CreateTeam(ctx context.Context, team *entity.GroupbuyTeam) error
	// GetTeamByID 根据ID获取拼团团队实体。
	GetTeamByID(ctx context.Context, id uint64) (*entity.GroupbuyTeam, error)
	// GetTeamByNo 根据团队编号获取拼团团队实体。
	GetTeamByNo(ctx context.Context, teamNo string) (*entity.GroupbuyTeam, error)
	// UpdateTeam 更新拼团团队实体的信息。
	UpdateTeam(ctx context.Context, team *entity.GroupbuyTeam) error
	// ListTeamsByGroupbuyID 列出指定拼团活动ID的所有拼团团队实体，支持分页。
	ListTeamsByGroupbuyID(ctx context.Context, groupbuyID uint64, page, pageSize int) ([]*entity.GroupbuyTeam, int64, error)

	// --- GroupbuyOrder methods ---

	// CreateOrder 在数据存储中创建一个新的拼团订单实体。
	CreateOrder(ctx context.Context, order *entity.GroupbuyOrder) error
	// GetOrderByID 根据ID获取拼团订单实体。
	GetOrderByID(ctx context.Context, id uint64) (*entity.GroupbuyOrder, error)
	// UpdateOrder 更新拼团订单实体的信息。
	UpdateOrder(ctx context.Context, order *entity.GroupbuyOrder) error
	// ListOrdersByTeamID 列出指定团队ID的所有拼团订单实体。
	ListOrdersByTeamID(ctx context.Context, teamID uint64) ([]*entity.GroupbuyOrder, error)
	// ListOrdersByUserID 列出指定用户ID的所有拼团订单实体，支持分页。
	ListOrdersByUserID(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.GroupbuyOrder, int64, error)
}
