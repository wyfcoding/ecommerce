package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

// GroupbuyQueryService 负责处理 Groupbuy 相关的读操作和查询逻辑。
type GroupbuyQueryService struct {
	repo           domain.GroupbuyRepository
	groupbuyRead   domain.GroupbuyReadRepository
	teamRead       domain.GroupbuyTeamReadRepository
	orderRead      domain.GroupbuyOrderReadRepository
	groupbuySearch domain.GroupbuySearchRepository
	orderSearch    domain.GroupbuyOrderSearchRepository
	logger         *slog.Logger
}

// NewGroupbuyQueryService 负责处理 NewGroupbuy 相关的读操作和查询逻辑。
func NewGroupbuyQueryService(
	repo domain.GroupbuyRepository,
	groupbuyRead domain.GroupbuyReadRepository,
	teamRead domain.GroupbuyTeamReadRepository,
	orderRead domain.GroupbuyOrderReadRepository,
	groupbuySearch domain.GroupbuySearchRepository,
	orderSearch domain.GroupbuyOrderSearchRepository,
	logger *slog.Logger,
) *GroupbuyQueryService {
	return &GroupbuyQueryService{
		repo:           repo,
		groupbuyRead:   groupbuyRead,
		teamRead:       teamRead,
		orderRead:      orderRead,
		groupbuySearch: groupbuySearch,
		orderSearch:    orderSearch,
		logger:         logger,
	}
}

func (q *GroupbuyQueryService) ListGroupbuys(ctx context.Context, page, pageSize int) ([]*domain.Groupbuy, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	query := &domain.GroupbuyQuery{
		Page:     page,
		PageSize: pageSize,
	}
	offset := (page - 1) * pageSize
	if q.groupbuySearch != nil {
		list, total, err := q.groupbuySearch.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		if q.logger != nil {
			q.logger.WarnContext(ctx, "groupbuy search fallback to mysql", "error", err)
		}
	}
	return q.repo.ListGroupbuys(ctx, query)
}

func (q *GroupbuyQueryService) GetGroupbuyByID(ctx context.Context, id uint64) (*domain.Groupbuy, error) {
	if q.groupbuyRead != nil {
		if cached, err := q.groupbuyRead.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	groupbuy, err := q.repo.GetGroupbuyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if groupbuy != nil && q.groupbuyRead != nil {
		_ = q.groupbuyRead.Save(ctx, groupbuy)
	}
	return groupbuy, nil
}

func (q *GroupbuyQueryService) GetTeamDetails(ctx context.Context, teamID uint64) (*domain.GroupbuyTeam, []*domain.GroupbuyOrder, error) {
	var team *domain.GroupbuyTeam
	var err error
	if q.teamRead != nil {
		if cached, cerr := q.teamRead.GetByID(ctx, teamID); cerr == nil && cached != nil {
			team = cached
		}
	}
	if team == nil {
		team, err = q.repo.GetTeamByID(ctx, teamID)
		if err != nil {
			return nil, nil, err
		}
		if team != nil && q.teamRead != nil {
			_ = q.teamRead.Save(ctx, team)
		}
	}

	var orders []*domain.GroupbuyOrder
	if q.orderSearch != nil {
		list, _, serr := q.orderSearch.Search(ctx, &domain.GroupbuyOrderQuery{
			TeamID:   teamID,
			Page:     1,
			PageSize: 200,
		}, 0, 200)
		if serr == nil {
			orders = list
		} else if q.logger != nil {
			q.logger.WarnContext(ctx, "groupbuy order search fallback to mysql", "error", serr)
		}
	}
	if orders == nil {
		orders, err = q.repo.ListOrdersByTeamID(ctx, teamID)
		if err != nil {
			return nil, nil, err
		}
		if q.orderRead != nil {
			for _, o := range orders {
				_ = q.orderRead.Save(ctx, o)
			}
		}
	}

	return team, orders, nil
}
