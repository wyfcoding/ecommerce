package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/usertier/domain"
)

// UserTierService 用户等级与积分门面服务，整合 Command 和 Query。
type UserTierService struct {
	Command *UserTierCommandService
	Query   *UserTierQueryService
}

// NewUserTierService 构造函数。
func NewUserTierService(repo domain.UserTierRepository, command *UserTierCommandService, query *UserTierQueryService) *UserTierService {
	return &UserTierService{
		Command: command,
		Query:   query,
	}
}

// --- Manager (Writes) ---

func (s *UserTierService) AddScore(ctx context.Context, userID uint64, score int64) error {
	return s.Command.AddScore(ctx, userID, score)
}

func (s *UserTierService) AddPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	return s.Command.AddPoints(ctx, userID, points, reason)
}

func (s *UserTierService) DeductPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	return s.Command.DeductPoints(ctx, userID, points, reason)
}

func (s *UserTierService) Exchange(ctx context.Context, userID uint64, exchangeID uint64) error {
	return s.Command.Exchange(ctx, userID, exchangeID)
}

// --- Query (Reads) ---

func (s *UserTierService) GetUserTier(ctx context.Context, userID uint64) (*domain.UserTier, error) {
	return s.Query.GetUserTier(ctx, userID)
}

func (s *UserTierService) GetPoints(ctx context.Context, userID uint64) (int64, error) {
	return s.Query.GetPoints(ctx, userID)
}

func (s *UserTierService) ListPointsLogs(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.PointsLog, int64, error) {
	return s.Query.ListPointsLogs(ctx, userID, page, pageSize)
}

func (s *UserTierService) ListExchanges(ctx context.Context, page, pageSize int) ([]*domain.Exchange, int64, error) {
	return s.Query.ListExchanges(ctx, page, pageSize)
}
