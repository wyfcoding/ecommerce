package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/usertier/domain"
)

// UserTierQueryService 处理所有用户等级与积分相关的查询操作（Queries）。
type UserTierQueryService struct {
	repo    domain.UserTierRepository
	command *UserTierCommandService
}

// NewUserTierQueryService 构造函数。
func NewUserTierQueryService(repo domain.UserTierRepository, command *UserTierCommandService) *UserTierQueryService {
	return &UserTierQueryService{
		repo:    repo,
		command: command,
	}
}

func (q *UserTierQueryService) GetUserTier(ctx context.Context, userID uint64) (*domain.UserTier, error) {
	tier, err := q.repo.GetUserTier(ctx, userID)
	if err != nil {
		return nil, err
	}
	if tier == nil {
		// 查询时如果不存在则回退到 manager 进行创建（带副作用的查询）
		return q.command.EnsureUserTier(ctx, userID)
	}
	return tier, nil
}

func (q *UserTierQueryService) GetPoints(ctx context.Context, userID uint64) (int64, error) {
	acc, err := q.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return 0, err
	}
	if acc == nil {
		return 0, nil
	}
	return acc.Balance, nil
}

func (q *UserTierQueryService) ListPointsLogs(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.PointsLog, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListPointsLogs(ctx, userID, offset, pageSize)
}

func (q *UserTierQueryService) ListExchanges(ctx context.Context, page, pageSize int) ([]*domain.Exchange, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListExchanges(ctx, offset, pageSize)
}
