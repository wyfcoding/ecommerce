package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

// GroupbuyService 结构体定义了拼团活动相关的应用服务门面。
type GroupbuyService struct {
	manager *GroupbuyManager
	query   *GroupbuyQuery
}

// NewGroupbuyService 创建拼团服务门面实例。
func NewGroupbuyService(manager *GroupbuyManager, query *GroupbuyQuery) *GroupbuyService {
	return &GroupbuyService{
		manager: manager,
		query:   query,
	}
}

// CreateGroupbuy 创建一个新的拼团活动。
func (s *GroupbuyService) CreateGroupbuy(ctx context.Context, name string, productID, skuID, originalPrice, groupPrice uint64,
	minPeople, maxPeople, totalStock int32, startTime, endTime time.Time,
) (*domain.Groupbuy, error) {
	return s.manager.CreateGroupbuy(ctx, name, productID, skuID, originalPrice, groupPrice, minPeople, maxPeople, totalStock, startTime, endTime)
}

// ListGroupbuys 获取拼团活动列表（分页）。
func (s *GroupbuyService) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*domain.Groupbuy, int64, error) {
	return s.query.ListGroupbuys(ctx, page, pageSize)
}

// InitiateTeam 发起一个新的拼团队伍。
func (s *GroupbuyService) InitiateTeam(ctx context.Context, groupbuyID, userID uint64) (*domain.GroupbuyTeam, *domain.GroupbuyOrder, error) {
	return s.manager.InitiateTeam(ctx, groupbuyID, userID)
}

// JoinTeam 加入一个已存在的拼团队伍。
func (s *GroupbuyService) JoinTeam(ctx context.Context, teamNo string, userID uint64) (*domain.GroupbuyOrder, error) {
	return s.manager.JoinTeam(ctx, teamNo, userID)
}

// GetTeamDetails 获取指定ID的拼团队伍详细信息及其订单。
func (s *GroupbuyService) GetTeamDetails(ctx context.Context, teamID uint64) (*domain.GroupbuyTeam, []*domain.GroupbuyOrder, error) {
	return s.query.GetTeamDetails(ctx, teamID)
}

// GetGroupbuyByID 根据ID获取拼团活动详情。
func (s *GroupbuyService) GetGroupbuyByID(ctx context.Context, id uint64) (*domain.Groupbuy, error) {
	return s.query.GetGroupbuyByID(ctx, id)
}
