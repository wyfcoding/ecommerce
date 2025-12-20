package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

// GroupbuyQuery 负责处理 Groupbuy 相关的读操作和查询逻辑。
type GroupbuyQuery struct {
	repo domain.GroupbuyRepository
}

// NewGroupbuyQuery 负责处理 NewGroupbuy 相关的读操作和查询逻辑。
func NewGroupbuyQuery(repo domain.GroupbuyRepository) *GroupbuyQuery {
	return &GroupbuyQuery{
		repo: repo,
	}
}

func (q *GroupbuyQuery) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*domain.Groupbuy, int64, error) {
	return q.repo.ListGroupbuys(ctx, page, pageSize)
}

func (q *GroupbuyQuery) GetGroupbuyByID(ctx context.Context, id uint64) (*domain.Groupbuy, error) {
	return q.repo.GetGroupbuyByID(ctx, id)
}

func (q *GroupbuyQuery) GetTeamDetails(ctx context.Context, teamID uint64) (*domain.GroupbuyTeam, []*domain.GroupbuyOrder, error) {
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
