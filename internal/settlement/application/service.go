package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/ledger"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/repository"
)

// SettlementService 结算应用服务
type SettlementService struct {
	repo          repository.SettlementRepository
	ledgerService *ledger.LedgerService
	logger        *slog.Logger
}

// NewSettlementService 构造函数
func NewSettlementService(
	repo repository.SettlementRepository,
	ledgerService *ledger.LedgerService,
	logger *slog.Logger,
) *SettlementService {
	return &SettlementService{
		repo:          repo,
		ledgerService: ledgerService,
		logger:        logger,
	}
}

// RecordPaymentSuccess 记录支付成功事件 (核心清分与记账逻辑)。
func (s *SettlementService) RecordPaymentSuccess(ctx context.Context, orderID uint64, orderNo string, merchantID uint64, amount int64, channelCost int64) error {
	s.logger.InfoContext(ctx, "processing payment success for settlement", "order_no", orderNo, "amount", amount)

	// 1. 获取商户费率配置
	account, err := s.repo.GetMerchantAccount(ctx, merchantID)
	if err != nil {
		return err
	}
	feeRate := 0.0
	if account != nil {
		feeRate = account.FeeRate
	} else {
		// 自动开户逻辑应在此处或更早处理，这里简化
		feeRate = 0.006 // 默认0.6%
	}

	// 2. 清分计算 (Clearing)
	// 平台收入 = 订单金额 * 商户费率
	platformFee := int64(float64(amount) * feeRate)
	// 商户应收 = 订单金额 - platform收入
	merchantReceivable := amount - platformFee

	// 3. 构造会计分录 (Accounting) - 复式记账
	entry := &ledger.JournalEntry{
		TransactionID: orderNo,
		EventType:     "PAYMENT_SUCCESS",
		Description:   fmt.Sprintf("Payment for Order %s", orderNo),
		PostingDate:   time.Now(),
		Lines: []ledger.EntryLine{
			// 借：应收渠道款 (增加资产)
			{
				SubjectCode: "1001",
				AccountID:   s.getAccountID("1001", "CHANNEL_GLOBAL"),
				Direction:   ledger.Debit,
				Amount:      amount,
			},
			// 贷：应付商户款 (增加负债)
			{
				SubjectCode: "2001",
				AccountID:   s.getAccountID("2001", fmt.Sprintf("MERCH_%d", merchantID)),
				Direction:   ledger.Credit,
				Amount:      merchantReceivable,
			},
			// 贷：手续费收入 (增加收入)
			{
				SubjectCode: "6001",
				AccountID:   s.getAccountID("6001", "PLATFORM_MAIN"),
				Direction:   ledger.Credit,
				Amount:      platformFee,
			},
		},
	}

	// 4. 调用账务核心记账
	if err := s.ledgerService.PostEntry(ctx, entry); err != nil {
		s.logger.ErrorContext(ctx, "failed to post ledger entry", "order_id", orderID, "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "payment recorded in ledger", "entry_no", entry.EntryNo)
	return nil
}

// getAccountID 辅助方法，用于获取或创建账务账户。
func (s *SettlementService) getAccountID(subjectCode, entityID string) uint64 {
	acc, err := s.ledgerService.CreateAccount(context.Background(), subjectCode, entityID)
	if err != nil {
		s.logger.Error("failed to get/create account", "subject", subjectCode, "entity", entityID, "error", err)
		return 0
	}
	return uint64(acc.ID)
}

// CreateSettlement 创建结算单。
func (s *SettlementService) CreateSettlement(ctx context.Context, merchantID uint64, cycle string, startDate, endDate time.Time) (*entity.Settlement, error) {
	settlementNo := fmt.Sprintf("S%d%d", merchantID, time.Now().UnixNano())
	settlement := &entity.Settlement{
		SettlementNo: settlementNo,
		MerchantID:   merchantID,
		Cycle:        entity.SettlementCycle(cycle),
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       entity.SettlementStatusPending,
	}
	if err := s.repo.SaveSettlement(ctx, settlement); err != nil {
		return nil, err
	}
	return settlement, nil
}

// AddOrderToSettlement 添加订单到结算单。
func (s *SettlementService) AddOrderToSettlement(ctx context.Context, settlementID uint64, orderID uint64, orderNo string, amount uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != entity.SettlementStatusPending {
		return errors.New("settlement is not pending")
	}

	account, err := s.repo.GetMerchantAccount(ctx, settlement.MerchantID)
	if err != nil {
		return err
	}
	feeRate := 0.0
	if account != nil {
		feeRate = account.FeeRate
	}

	platformFee := uint64(float64(amount) * feeRate / 100)
	settlementAmount := amount - platformFee

	detail := &entity.SettlementDetail{
		SettlementID:     settlementID,
		OrderID:          orderID,
		OrderNo:          orderNo,
		OrderAmount:      amount,
		PlatformFee:      platformFee,
		SettlementAmount: settlementAmount,
	}

	if err := s.repo.SaveSettlementDetail(ctx, detail); err != nil {
		return err
	}

	settlement.OrderCount++
	settlement.TotalAmount += amount
	settlement.PlatformFee += platformFee
	settlement.SettlementAmount += settlementAmount

	return s.repo.SaveSettlement(ctx, settlement)
}

// ProcessSettlement 处理结算单（开始处理）。
func (s *SettlementService) ProcessSettlement(ctx context.Context, id uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != entity.SettlementStatusPending {
		return errors.New("settlement is not pending")
	}

	settlement.Status = entity.SettlementStatusProcessing
	return s.repo.SaveSettlement(ctx, settlement)
}

// CompleteSettlement 完成结算单。
func (s *SettlementService) CompleteSettlement(ctx context.Context, id uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != entity.SettlementStatusProcessing {
		return errors.New("settlement is not processing")
	}

	account, err := s.repo.GetMerchantAccount(ctx, settlement.MerchantID)
	if err != nil {
		return err
	}
	if account == nil {
		account = &entity.MerchantAccount{
			MerchantID: settlement.MerchantID,
			FeeRate:    0.0,
		}
	}

	account.Balance += settlement.SettlementAmount
	account.TotalIncome += settlement.SettlementAmount
	if err := s.repo.SaveMerchantAccount(ctx, account); err != nil {
		return err
	}

	now := time.Now()
	settlement.Status = entity.SettlementStatusCompleted
	settlement.SettledAt = &now
	return s.repo.SaveSettlement(ctx, settlement)
}

// GetMerchantAccount 获取商户账户信息。
func (s *SettlementService) GetMerchantAccount(ctx context.Context, merchantID uint64) (*entity.MerchantAccount, error) {
	return s.repo.GetMerchantAccount(ctx, merchantID)
}

// ListSettlements 获取结算单列表。
func (s *SettlementService) ListSettlements(ctx context.Context, merchantID uint64, status *int, page, pageSize int) ([]*entity.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	var st *entity.SettlementStatus
	if status != nil {
		s := entity.SettlementStatus(*status)
		st = &s
	}
	return s.repo.ListSettlements(ctx, merchantID, st, offset, pageSize)
}
