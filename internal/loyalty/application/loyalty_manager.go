package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// LoyaltyManager 负责处理 Loyalty 相关的写操作和业务逻辑。
type LoyaltyManager struct {
	repo     domain.LoyaltyRepository
	logger   *slog.Logger
	rankList *algorithm.SkipList[int64, uint64] // 内存积分排行榜 (Points -> UserID)
}

// NewLoyaltyManager 创建并返回一个新的 LoyaltyManager 实例。
func NewLoyaltyManager(repo domain.LoyaltyRepository, logger *slog.Logger) *LoyaltyManager {
	return &LoyaltyManager{
		repo:     repo,
		logger:   logger,
		rankList: algorithm.NewSkipList[int64, uint64](),
	}
}

// GetTopUsers 获取积分排名前 N 的用户
func (m *LoyaltyManager) GetTopUsers(limit int) []uint64 {
	it := m.rankList.Iterator()
	results := make([]uint64, 0, limit)

	// 由于 SkipList 默认是升序，我们需要收集全部后取末尾并反转
	for {
		_, val, ok := it.Next()
		if !ok {
			break
		}
		results = append(results, val) // 泛型直接返回 uint64，无需类型断言
	}

	if len(results) > limit {
		results = results[len(results)-limit:]
	}

	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results
}

// CalculateOrderPoints 计算订单应得积分。
func (m *LoyaltyManager) CalculateOrderPoints(ctx context.Context, userID uint64, orderAmount int64, items []struct {
	Category string
	Amount   int64
}) (int64, error) {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return 0, err
	}
	level := domain.MemberLevelBronze
	if account != nil {
		level = account.Level
	}

	benefit, err := m.repo.GetMemberBenefitByLevel(ctx, level)
	if err != nil {
		m.logger.WarnContext(ctx, "failed to get member benefit, using default", "level", level)
		benefit = &domain.MemberBenefit{PointsRate: 1.0}
	}

	var totalPoints float64
	// 基础积分：每 100 分 (1元) 积 1 分
	baseRate := 0.01

	for _, item := range items {
		itemPoints := float64(item.Amount) * baseRate * benefit.PointsRate

		// 检查类目特定加倍
		if benefit.Multipliers != nil {
			if multiplier, ok := benefit.Multipliers[item.Category]; ok {
				itemPoints *= multiplier
			}
		}
		totalPoints += itemPoints
	}

	return int64(totalPoints), nil
}

func (m *LoyaltyManager) AddPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		account = domain.NewMemberAccount(userID)
		if err := m.repo.SaveMemberAccount(ctx, account); err != nil {
			m.logger.ErrorContext(ctx, "failed to create member account", "user_id", userID, "error", err)
			return err
		}
	}

	account.AddPoints(points)
	if err := m.repo.SaveMemberAccount(ctx, account); err != nil {
		return err
	}

	// 同步更新跳表排行榜：泛型支持直接传入 int64
	m.rankList.Insert(account.AvailablePoints, userID)

	tx := domain.NewPointsTransaction(userID, transactionType, points, account.AvailablePoints, orderID, description, nil)
	return m.repo.SavePointsTransaction(ctx, tx)
}

func (m *LoyaltyManager) DeductPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		return domain.ErrInsufficientPoints
	}

	oldPoints := account.AvailablePoints
	if err := account.DeductPoints(points); err != nil {
		return err
	}

	if err := m.repo.SaveMemberAccount(ctx, account); err != nil {
		return err
	}

	// 同步更新跳表排行榜：先删旧值，再插新值
	m.rankList.Delete(oldPoints)
	m.rankList.Insert(account.AvailablePoints, userID)

	tx := domain.NewPointsTransaction(userID, transactionType, -points, account.AvailablePoints, orderID, description, nil)
	return m.repo.SavePointsTransaction(ctx, tx)
}

func (m *LoyaltyManager) AddSpent(ctx context.Context, userID uint64, amount uint64) error {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		account = domain.NewMemberAccount(userID)
	}

	account.AddSpent(amount)
	return m.repo.SaveMemberAccount(ctx, account)
}

func (m *LoyaltyManager) AddBenefit(ctx context.Context, level domain.MemberLevel, name, description string, discountRate, pointsRate float64) (*domain.MemberBenefit, error) {
	benefit := domain.NewMemberBenefit(level, name, description, discountRate, pointsRate)
	if err := m.repo.SaveMemberBenefit(ctx, benefit); err != nil {
		m.logger.ErrorContext(ctx, "failed to save member benefit", "level", level, "error", err)
		return nil, err
	}
	return benefit, nil
}

func (m *LoyaltyManager) DeleteBenefit(ctx context.Context, id uint64) error {
	return m.repo.DeleteMemberBenefit(ctx, id)
}
