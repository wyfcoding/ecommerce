package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/structures"
	"github.com/wyfcoding/pkg/messagequeue"
)

// LoyaltyCommandService 负责处理 Loyalty 相关的写操作和业务逻辑。
type LoyaltyCommandService struct {
	repo      domain.LoyaltyRepository
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
	rankList  *algorithm.SkipList[int64, uint64] // 内存积分排行榜 (Points -> UserID)
}

// NewLoyaltyCommandService 创建并返回一个新的 LoyaltyCommandService 实例。
func NewLoyaltyCommandService(repo domain.LoyaltyRepository, publisher messagequeue.EventPublisher, logger *slog.Logger) *LoyaltyCommandService {
	return &LoyaltyCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		rankList:  algorithm.NewSkipList[int64, uint64](),
	}
}

// GetTopUsers 获取积分排名前 N 的用户。
func (m *LoyaltyCommandService) GetTopUsers(limit int) []uint64 {
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
func (m *LoyaltyCommandService) CalculateOrderPoints(ctx context.Context, userID uint64, orderAmount int64, items []struct {
	Category string
	Amount   int64
},
) (int64, error) {
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

func (m *LoyaltyCommandService) AddPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		account = domain.NewMemberAccount(userID)
	}

	oldPoints := account.AvailablePoints
	account.AddPoints(points)

	tx := domain.NewPointsTransaction(userID, transactionType, points, account.AvailablePoints, orderID, description, nil)

	if err := m.repo.WithTx(ctx, func(trx any) error {
		if err := m.repo.SaveMemberAccountInTx(ctx, trx, account); err != nil {
			m.logger.ErrorContext(ctx, "failed to save member account", "user_id", userID, "error", err)
			return err
		}
		if err := m.repo.SavePointsTransactionInTx(ctx, trx, tx); err != nil {
			return err
		}
		if m.publisher != nil {
			if err := m.publisher.PublishInTx(ctx, trx, domain.MemberAccountUpdatedEventType, fmt.Sprintf("%d", userID), &domain.MemberAccountUpdatedEvent{
				UserID:    userID,
				Timestamp: time.Now(),
			}); err != nil {
				return err
			}
			if err := m.publisher.PublishInTx(ctx, trx, domain.PointsTransactionCreatedEventType, fmt.Sprintf("%d", tx.ID), &domain.PointsTransactionCreatedEvent{
				TransactionID: tx.ID,
				UserID:        userID,
				Timestamp:     time.Now(),
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if oldPoints > 0 {
		m.rankList.Delete(oldPoints)
	}
	m.rankList.Insert(account.AvailablePoints, userID)

	return nil
}

func (m *LoyaltyCommandService) DeductPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
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

	tx := domain.NewPointsTransaction(userID, transactionType, -points, account.AvailablePoints, orderID, description, nil)

	if err := m.repo.WithTx(ctx, func(trx any) error {
		if err := m.repo.SaveMemberAccountInTx(ctx, trx, account); err != nil {
			return err
		}
		if err := m.repo.SavePointsTransactionInTx(ctx, trx, tx); err != nil {
			return err
		}
		if m.publisher != nil {
			if err := m.publisher.PublishInTx(ctx, trx, domain.MemberAccountUpdatedEventType, fmt.Sprintf("%d", userID), &domain.MemberAccountUpdatedEvent{
				UserID:    userID,
				Timestamp: time.Now(),
			}); err != nil {
				return err
			}
			if err := m.publisher.PublishInTx(ctx, trx, domain.PointsTransactionCreatedEventType, fmt.Sprintf("%d", tx.ID), &domain.PointsTransactionCreatedEvent{
				TransactionID: tx.ID,
				UserID:        userID,
				Timestamp:     time.Now(),
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if oldPoints > 0 {
		m.rankList.Delete(oldPoints)
	}
	m.rankList.Insert(account.AvailablePoints, userID)

	return nil
}

func (m *LoyaltyCommandService) AddSpent(ctx context.Context, userID uint64, amount uint64) error {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		account = domain.NewMemberAccount(userID)
	}

	account.AddSpent(amount)
	return m.repo.WithTx(ctx, func(trx any) error {
		if err := m.repo.SaveMemberAccountInTx(ctx, trx, account); err != nil {
			return err
		}
		if m.publisher != nil {
			if err := m.publisher.PublishInTx(ctx, trx, domain.MemberAccountUpdatedEventType, fmt.Sprintf("%d", userID), &domain.MemberAccountUpdatedEvent{
				UserID:    userID,
				Timestamp: time.Now(),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *LoyaltyCommandService) AddBenefit(ctx context.Context, level domain.MemberLevel, name, description string, discountRate, pointsRate float64) (*domain.MemberBenefit, error) {
	benefit := domain.NewMemberBenefit(level, name, description, discountRate, pointsRate)
	if err := m.repo.WithTx(ctx, func(trx any) error {
		if err := m.repo.SaveMemberBenefitInTx(ctx, trx, benefit); err != nil {
			m.logger.ErrorContext(ctx, "failed to save member benefit", "level", level, "error", err)
			return err
		}
		if m.publisher != nil {
			if err := m.publisher.PublishInTx(ctx, trx, domain.MemberBenefitSavedEventType, fmt.Sprintf("%d", benefit.ID), &domain.MemberBenefitSavedEvent{
				BenefitID: benefit.ID,
				Level:     benefit.Level,
				Timestamp: time.Now(),
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return benefit, nil
}

func (m *LoyaltyCommandService) DeleteBenefit(ctx context.Context, id uint64) error {
	benefit, err := m.repo.GetMemberBenefit(ctx, id)
	if err != nil {
		return err
	}
	if benefit == nil {
		return nil
	}

	return m.repo.WithTx(ctx, func(trx any) error {
		if err := m.repo.DeleteMemberBenefitInTx(ctx, trx, id); err != nil {
			return err
		}
		if m.publisher != nil {
			if err := m.publisher.PublishInTx(ctx, trx, domain.MemberBenefitDeletedEventType, fmt.Sprintf("%d", id), &domain.MemberBenefitDeletedEvent{
				BenefitID: id,
				Level:     benefit.Level,
				Timestamp: time.Now(),
			}); err != nil {
				return err
			}
		}
		return nil
	})
}
