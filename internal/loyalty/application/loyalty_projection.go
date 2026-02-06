// 生成摘要：新增忠诚度读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以 user_id 与 benefit_level 为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
)

// LoyaltyProjectionService 负责将事件转换为读模型更新。
type LoyaltyProjectionService struct {
	repo              domain.LoyaltyRepository
	accountReadRepo   domain.MemberAccountReadRepository
	benefitReadRepo   domain.MemberBenefitReadRepository
	txSearchRepo      domain.PointsTransactionSearchRepository
	benefitSearchRepo domain.MemberBenefitSearchRepository
	logger            *slog.Logger
}

// NewLoyaltyProjectionService 创建忠诚度投影服务。
func NewLoyaltyProjectionService(
	repo domain.LoyaltyRepository,
	accountReadRepo domain.MemberAccountReadRepository,
	benefitReadRepo domain.MemberBenefitReadRepository,
	txSearchRepo domain.PointsTransactionSearchRepository,
	benefitSearchRepo domain.MemberBenefitSearchRepository,
	logger *slog.Logger,
) *LoyaltyProjectionService {
	return &LoyaltyProjectionService{
		repo:              repo,
		accountReadRepo:   accountReadRepo,
		benefitReadRepo:   benefitReadRepo,
		txSearchRepo:      txSearchRepo,
		benefitSearchRepo: benefitSearchRepo,
		logger:            logger,
	}
}

// OnAccountUpdated 处理会员账户更新事件。
func (s *LoyaltyProjectionService) OnAccountUpdated(ctx context.Context, event *domain.MemberAccountUpdatedEvent) error {
	if event == nil {
		return nil
	}
	account, err := s.repo.GetMemberAccount(ctx, event.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load account for projection", "user_id", event.UserID, "error", err)
		return err
	}
	if account == nil {
		if s.accountReadRepo != nil {
			_ = s.accountReadRepo.Delete(ctx, event.UserID)
		}
		return nil
	}
	if s.accountReadRepo != nil {
		if err := s.accountReadRepo.Save(ctx, account); err != nil {
			s.logger.ErrorContext(ctx, "failed to save account read model", "user_id", event.UserID, "error", err)
			return err
		}
	}
	return nil
}

// OnTransactionCreated 处理积分交易创建事件。
func (s *LoyaltyProjectionService) OnTransactionCreated(ctx context.Context, event *domain.PointsTransactionCreatedEvent) error {
	if event == nil {
		return nil
	}
	if s.txSearchRepo == nil {
		return nil
	}
	tx, err := s.repo.GetPointsTransaction(ctx, event.TransactionID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load transaction for projection", "transaction_id", event.TransactionID, "error", err)
		return err
	}
	if tx == nil {
		_ = s.txSearchRepo.DeleteTransaction(ctx, event.TransactionID)
		return nil
	}
	if err := s.txSearchRepo.IndexTransaction(ctx, tx); err != nil {
		s.logger.ErrorContext(ctx, "failed to index transaction", "transaction_id", event.TransactionID, "error", err)
		return err
	}
	return nil
}

// OnBenefitSaved 处理会员权益保存事件。
func (s *LoyaltyProjectionService) OnBenefitSaved(ctx context.Context, event *domain.MemberBenefitSavedEvent) error {
	if event == nil {
		return nil
	}
	benefit, err := s.repo.GetMemberBenefit(ctx, event.BenefitID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load benefit for projection", "benefit_id", event.BenefitID, "error", err)
		return err
	}
	if benefit == nil {
		if s.benefitReadRepo != nil {
			_ = s.benefitReadRepo.DeleteByLevel(ctx, event.Level)
		}
		if s.benefitSearchRepo != nil {
			_ = s.benefitSearchRepo.DeleteBenefit(ctx, event.BenefitID)
		}
		return nil
	}
	if s.benefitReadRepo != nil {
		if err := s.benefitReadRepo.Save(ctx, benefit); err != nil {
			s.logger.ErrorContext(ctx, "failed to save benefit read model", "benefit_id", event.BenefitID, "error", err)
			return err
		}
	}
	if s.benefitSearchRepo != nil {
		if err := s.benefitSearchRepo.IndexBenefit(ctx, benefit); err != nil {
			s.logger.ErrorContext(ctx, "failed to index benefit", "benefit_id", event.BenefitID, "error", err)
			return err
		}
	}
	return nil
}

// OnBenefitDeleted 处理会员权益删除事件。
func (s *LoyaltyProjectionService) OnBenefitDeleted(ctx context.Context, event *domain.MemberBenefitDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.benefitReadRepo != nil {
		_ = s.benefitReadRepo.DeleteByLevel(ctx, event.Level)
	}
	if s.benefitSearchRepo != nil {
		_ = s.benefitSearchRepo.DeleteBenefit(ctx, event.BenefitID)
	}
	return nil
}
