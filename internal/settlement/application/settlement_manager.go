package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	accountv1 "github.com/wyfcoding/financialtrading/goapi/account/v1"
)

// SettlementManager 处理所有结算相关的写入操作（Commands）。
type SettlementManager struct {
	repo             domain.SettlementRepository
	ledgerService    *domain.LedgerService
	logger           *slog.Logger
	remoteAccountCli accountv1.AccountServiceClient
}

// NewSettlementManager 构造函数。
func NewSettlementManager(repo domain.SettlementRepository, ledgerService *domain.LedgerService, logger *slog.Logger) *SettlementManager {
	return &SettlementManager{
		repo:          repo,
		ledgerService: ledgerService,
		logger:        logger,
	}
}

func (m *SettlementManager) SetRemoteAccountClient(cli accountv1.AccountServiceClient) {
	m.remoteAccountCli = cli
}

// RecordPaymentSuccess 记录支付成功事件 (核心清分与记账逻辑)。
func (m *SettlementManager) RecordPaymentSuccess(ctx context.Context, orderID uint64, orderNo string, merchantID uint64, amount int64, channelCost int64) error {
	m.logger.InfoContext(ctx, "processing payment success for settlement", "order_no", orderNo, "amount", amount)

	// 1. 获取商户费率配置
	account, err := m.repo.GetMerchantAccount(ctx, merchantID)
	if err != nil {
		return err
	}
	feeRate := 0.0
	if account != nil {
		feeRate = account.FeeRate
	} else {
		feeRate = 0.006 // 默认0.6%
	}

	// 2. 清分计算 (Clearing)
	platformFee := int64(float64(amount) * feeRate)
	merchantReceivable := amount - platformFee

	// 3. 构造会计分录 (Accounting)
	entry := &domain.JournalEntry{
		TransactionID: orderNo,
		EventType:     "PAYMENT_SUCCESS",
		Description:   fmt.Sprintf("Payment for Order %s", orderNo),
		PostingDate:   time.Now(),
		Lines: []domain.EntryLine{
			{
				SubjectCode: "1001",
				AccountID:   m.getAccountID("1001", "CHANNEL_GLOBAL"),
				Direction:   domain.Debit,
				Amount:      amount,
			},
			{
				SubjectCode: "2001",
				AccountID:   m.getAccountID("2001", fmt.Sprintf("MERCH_%d", merchantID)),
				Direction:   domain.Credit,
				Amount:      merchantReceivable,
			},
			{
				SubjectCode: "6001",
				AccountID:   m.getAccountID("6001", "PLATFORM_MAIN"),
				Direction:   domain.Credit,
				Amount:      platformFee,
			},
		},
	}

	// 4. 调用账务核心记账
	if err := m.ledgerService.PostEntry(ctx, entry); err != nil {
		m.logger.ErrorContext(ctx, "failed to post ledger entry", "order_id", orderID, "error", err)
		return err
	}

	// 5. 跨项目同步 (Cross-Project Interaction)
	// 假设商户在 FinancialTrading 系统中也有对应的交易账户 (UserID = MerchantID)
	if m.remoteAccountCli != nil {
		_, err := m.remoteAccountCli.Deposit(ctx, &accountv1.DepositRequest{
			UserId:   fmt.Sprintf("%d", merchantID),
			Amount:   fmt.Sprintf("%d", merchantReceivable),
			Currency: "USD",
		})
		if err != nil {
			m.logger.ErrorContext(ctx, "failed to sync settlement to financial account", "merchant_id", merchantID, "error", err)
			// 注意：此时本地账务已完成，跨项目失败可记录日志后补偿，此处不强制阻塞
		} else {
			m.logger.InfoContext(ctx, "settlement synced to financial account successfully", "merchant_id", merchantID)
		}
	}

	m.logger.InfoContext(ctx, "payment recorded in ledger", "entry_no", entry.EntryNo)
	return nil
}

// getAccountID 辅助方法。
func (m *SettlementManager) getAccountID(subjectCode, entityID string) uint64 {
	acc, err := m.ledgerService.CreateAccount(context.Background(), subjectCode, entityID)
	if err != nil {
		m.logger.Error("failed to get/create account", "subject", subjectCode, "entity", entityID, "error", err)
		return 0
	}
	return uint64(acc.ID)
}

// CreateSettlement 创建结算单。
func (m *SettlementManager) CreateSettlement(ctx context.Context, merchantID uint64, cycle string, startDate, endDate time.Time) (*domain.Settlement, error) {
	settlementNo := fmt.Sprintf("S%d%d", merchantID, time.Now().UnixNano())
	settlement := &domain.Settlement{
		SettlementNo: settlementNo,
		MerchantID:   merchantID,
		Cycle:        domain.SettlementCycle(cycle),
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       domain.SettlementStatusPending,
	}
	if err := m.repo.SaveSettlement(ctx, settlement); err != nil {
		return nil, err
	}
	return settlement, nil
}

// AddOrderToSettlement 添加订单到结算单.
func (m *SettlementManager) AddOrderToSettlement(ctx context.Context, settlementID uint64, orderID uint64, orderNo string, amount uint64) error {
	settlement, err := m.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != domain.SettlementStatusPending {
		return errors.New("settlement is not pending")
	}

	account, err := m.repo.GetMerchantAccount(ctx, settlement.MerchantID)
	if err != nil {
		return err
	}
	feeRate := 0.0
	if account != nil {
		feeRate = account.FeeRate
	}

	platformFee := uint64(float64(amount) * feeRate / 100)
	settlementAmount := amount - platformFee

	detail := &domain.SettlementDetail{
		SettlementID:     settlementID,
		OrderID:          orderID,
		OrderNo:          orderNo,
		OrderAmount:      amount,
		PlatformFee:      platformFee,
		SettlementAmount: settlementAmount,
	}

	if err := m.repo.SaveSettlementDetail(ctx, detail); err != nil {
		return err
	}

	settlement.OrderCount++
	settlement.TotalAmount += amount
	settlement.PlatformFee += platformFee
	settlement.SettlementAmount += settlementAmount

	return m.repo.SaveSettlement(ctx, settlement)
}

// ProcessSettlement 处理结算单（开始处理）。
func (m *SettlementManager) ProcessSettlement(ctx context.Context, id uint64) error {
	settlement, err := m.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != domain.SettlementStatusPending {
		return errors.New("settlement is not pending")
	}

	settlement.Status = domain.SettlementStatusProcessing
	return m.repo.SaveSettlement(ctx, settlement)
}

// CompleteSettlement 完成结算单。
func (m *SettlementManager) CompleteSettlement(ctx context.Context, id uint64) error {
	settlement, err := m.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != domain.SettlementStatusProcessing {
		return errors.New("settlement is not processing")
	}

	account, err := m.repo.GetMerchantAccount(ctx, settlement.MerchantID)
	if err != nil {
		return err
	}
	if account == nil {
		account = &domain.MerchantAccount{
			MerchantID: settlement.MerchantID,
			FeeRate:    0.0,
		}
	}

	account.Balance += settlement.SettlementAmount
	account.TotalIncome += settlement.SettlementAmount
	if err := m.repo.SaveMerchantAccount(ctx, account); err != nil {
		return err
	}

	now := time.Now()
	settlement.Status = domain.SettlementStatusCompleted
	settlement.SettledAt = &now
	return m.repo.SaveSettlement(ctx, settlement)
}
