package domain

import (
	"context"
)

// GroupbuyRepository 是拼团模块的仓储接口。
type GroupbuyRepository interface {
	// --- tx helpers ---
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- Groupbuy methods ---

	// CreateGroupbuy 在数据存储中创建一个新的拼团活动实体。
	CreateGroupbuy(ctx context.Context, groupbuy *Groupbuy) error
	CreateGroupbuyInTx(ctx context.Context, tx any, groupbuy *Groupbuy) error
	// GetGroupbuyByID 根据ID获取拼团活动实体。
	GetGroupbuyByID(ctx context.Context, id uint64) (*Groupbuy, error)
	// UpdateGroupbuy 更新拼团活动实体的信息。
	UpdateGroupbuy(ctx context.Context, groupbuy *Groupbuy) error
	UpdateGroupbuyInTx(ctx context.Context, tx any, groupbuy *Groupbuy) error
	// ListGroupbuys 列出所有拼团活动实体，支持分页。
	ListGroupbuys(ctx context.Context, query *GroupbuyQuery) ([]*Groupbuy, int64, error)

	// --- GroupbuyTeam methods ---

	// CreateTeam 在数据存储中创建一个新的拼团团队实体。
	CreateTeam(ctx context.Context, team *GroupbuyTeam) error
	CreateTeamInTx(ctx context.Context, tx any, team *GroupbuyTeam) error
	// GetTeamByID 根据ID获取拼团团队实体。
	GetTeamByID(ctx context.Context, id uint64) (*GroupbuyTeam, error)
	// GetTeamByNo 根据团队编号获取拼团团队实体。
	GetTeamByNo(ctx context.Context, teamNo string) (*GroupbuyTeam, error)
	// UpdateTeam 更新拼团团队实体的信息。
	UpdateTeam(ctx context.Context, team *GroupbuyTeam) error
	UpdateTeamInTx(ctx context.Context, tx any, team *GroupbuyTeam) error
	// ListTeamsByGroupbuyID 列出指定拼团活动ID的所有拼团团队实体，支持分页。
	ListTeamsByGroupbuyID(ctx context.Context, query *GroupbuyTeamQuery) ([]*GroupbuyTeam, int64, error)

	// --- GroupbuyOrder methods ---

	// CreateOrder 在数据存储中创建一个新的拼团订单实体。
	CreateOrder(ctx context.Context, order *GroupbuyOrder) error
	CreateOrderInTx(ctx context.Context, tx any, order *GroupbuyOrder) error
	// GetOrderByID 根据ID获取拼团订单实体。
	GetOrderByID(ctx context.Context, id uint64) (*GroupbuyOrder, error)
	// UpdateOrder 更新拼团订单实体的信息。
	UpdateOrder(ctx context.Context, order *GroupbuyOrder) error
	UpdateOrderInTx(ctx context.Context, tx any, order *GroupbuyOrder) error
	// ListOrdersByTeamID 列出指定团队ID的所有拼团订单实体。
	ListOrdersByTeamID(ctx context.Context, teamID uint64) ([]*GroupbuyOrder, error)
	// ListOrdersByUserID 列出指定用户ID的所有拼团订单实体，支持分页。
	ListOrdersByUserID(ctx context.Context, query *GroupbuyOrderQuery) ([]*GroupbuyOrder, int64, error)
}

// GroupbuyQuery 拼团活动查询条件。
type GroupbuyQuery struct {
	Status    *GroupbuyStatus
	ProductID uint64
	Page      int
	PageSize  int
}

// GroupbuyTeamQuery 拼团团队查询条件。
type GroupbuyTeamQuery struct {
	GroupbuyID uint64
	Status     *GroupbuyTeamStatus
	Page       int
	PageSize   int
}

// GroupbuyOrderQuery 拼团订单查询条件。
type GroupbuyOrderQuery struct {
	UserID   uint64
	TeamID   uint64
	Status   *GroupbuyOrderStatus
	Page     int
	PageSize int
}
