package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
)

// LoyaltyService 结构体定义了忠诚度计划相关的应用服务门面。
type LoyaltyService struct {
	manager *LoyaltyManager
	query   *LoyaltyQuery
}

// NewLoyaltyService 创建忠诚度服务门面实例。
func NewLoyaltyService(manager *LoyaltyManager, query *LoyaltyQuery) *LoyaltyService {
	return &LoyaltyService{
		manager: manager,
		query:   query,
	}
}

// GetOrCreateAccount 获取或创建会员账户。
func (s *LoyaltyService) GetOrCreateAccount(ctx context.Context, userID uint64) (*domain.MemberAccount, error) {
	return s.query.GetOrCreateAccount(ctx, userID)
}

// AddPoints 增加用户会员积分。
func (s *LoyaltyService) AddPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	return s.manager.AddPoints(ctx, userID, points, transactionType, description, orderID)
}

// DeductPoints 扣减用户会员积分（如积分抵扣、兑换等）。
func (s *LoyaltyService) DeductPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	return s.manager.DeductPoints(ctx, userID, points, transactionType, description, orderID)
}

// AddSpent 增加用户累计消费金额，可能触发等级提升。
func (s *LoyaltyService) AddSpent(ctx context.Context, userID uint64, amount uint64) error {
	return s.manager.AddSpent(ctx, userID, amount)
}

// GetPointsTransactions 获取用户的积分流水记录（分页）。
func (s *LoyaltyService) GetPointsTransactions(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.PointsTransaction, int64, error) {
	return s.query.GetPointsTransactions(ctx, userID, page, pageSize)
}

// AddBenefit 为特定会员等级添加权益配置。
func (s *LoyaltyService) AddBenefit(ctx context.Context, level domain.MemberLevel, name, description string, discountRate, pointsRate float64) (*domain.MemberBenefit, error) {
	return s.manager.AddBenefit(ctx, level, name, description, discountRate, pointsRate)
}

// ListBenefits 列出指定会员等级的所有权益。
func (s *LoyaltyService) ListBenefits(ctx context.Context, level domain.MemberLevel) ([]*domain.MemberBenefit, error) {
	return s.query.ListBenefits(ctx, level)
}

// DeleteBenefit 删除指定的会员权益项。
func (s *LoyaltyService) DeleteBenefit(ctx context.Context, id uint64) error {
	return s.manager.DeleteBenefit(ctx, id)
}
