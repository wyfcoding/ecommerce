package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

// GroupbuyQueryService 负责处理 Groupbuy 相关的读操作和查询逻辑。
type GroupbuyQueryService struct {
	repo domain.GroupbuyRepository
}

// NewGroupbuyQueryService 负责处理 NewGroupbuy 相关的读操作和查询逻辑。
func NewGroupbuyQueryService(repo domain.GroupbuyRepository) *GroupbuyQueryService {
	return &GroupbuyQueryService{
		repo: repo,
	}
}

func (q *GroupbuyQueryService) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*domain.Groupbuy, int64, error) {
	return q.repo.ListGroupbuys(ctx, page, pageSize)
}

func (q *GroupbuyQueryService) GetGroupbuyByID(ctx context.Context, id uint64) (*domain.Groupbuy, error) {
	return q.repo.GetGroupbuyByID(ctx, id)
}

func (q *GroupbuyQueryService) GetTeamDetails(ctx context.Context, teamID uint64) (*domain.GroupbuyTeam, []*domain.GroupbuyOrder, error) {
	team, err := q.repo.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.repo.ListOrdersByTeamID(ctx, teamID)
	if err != nil {
		return nil, nil, err
	}

	return team, orders, nil
}
