package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/usertier/domain"
)

// UserTierService 用户等级与积分门面服务，整合 Manager 和 Query。
type UserTierService struct {
	manager *UserTierManager
	query   *UserTierQuery
}

// NewUserTierService 构造函数。
func NewUserTierService(repo domain.UserTierRepository, logger *slog.Logger) *UserTierService {
	m := NewUserTierManager(repo, logger)
	return &UserTierService{
		manager: m,
		query:   NewUserTierQuery(repo, m),
	}
}

// --- Manager (Writes) ---

func (s *UserTierService) AddScore(ctx context.Context, userID uint64, score int64) error {
	return s.manager.AddScore(ctx, userID, score)
}

func (s *UserTierService) AddPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	return s.manager.AddPoints(ctx, userID, points, reason)
}

func (s *UserTierService) DeductPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	return s.manager.DeductPoints(ctx, userID, points, reason)
}

func (s *UserTierService) Exchange(ctx context.Context, userID uint64, exchangeID uint64) error {
	return s.manager.Exchange(ctx, userID, exchangeID)
}

// --- Query (Reads) ---

func (s *UserTierService) GetUserTier(ctx context.Context, userID uint64) (*domain.UserTier, error) {
	return s.query.GetUserTier(ctx, userID)
}

func (s *UserTierService) GetPoints(ctx context.Context, userID uint64) (int64, error) {
	return s.query.GetPoints(ctx, userID)
}

func (s *UserTierService) ListPointsLogs(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.PointsLog, int64, error) {
	return s.query.ListPointsLogs(ctx, userID, page, pageSize)
}

func (s *UserTierService) ListExchanges(ctx context.Context, page, pageSize int) ([]*domain.Exchange, int64, error) {
	return s.query.ListExchanges(ctx, page, pageSize)
}
