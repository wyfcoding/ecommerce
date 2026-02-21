package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	accountv1 "github.com/wyfcoding/financialtrading/go-api/account/v1"
	"github.com/wyfcoding/pkg/messagequeue"
)

// SettlementCommandService 处理所有结算相关的写入操作（Commands）。
type SettlementCommandService struct {
	repo             domain.SettlementRepository
	ledgerService    *domain.LedgerService
	publisher        messagequeue.EventPublisher
	logger           *slog.Logger
	remoteAccountCli accountv1.AccountServiceClient
}

// NewSettlementCommandService 构造函数。
func NewSettlementCommandService(
	repo domain.SettlementRepository,
	ledgerService *domain.LedgerService,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *SettlementCommandService {
	return &SettlementCommandService{
		repo:          repo,
		ledgerService: ledgerService,
		publisher:     publisher,
		logger:        logger,
	}
}

func (s *SettlementCommandService) SetRemoteAccountClient(cli accountv1.AccountServiceClient) {
	s.remoteAccountCli = cli
}

// RecordPaymentSuccess 记录支付成功事件 (核心清分与记账逻辑)。
func (s *SettlementCommandService) RecordPaymentSuccess(ctx context.Context, orderID uint64, orderNo string, merchantID uint64, amount int64, _ int64) error {
	s.logger.InfoContext(ctx, "processing payment success for settlement", "order_no", orderNo, "amount", amount)

	// 1. 获取商户费率配置
	account, err := s.repo.GetMerchantAccount(ctx, merchantID)
	if err != nil {
		return err
	}
	feeRate := decimal.NewFromFloat(0.006)
	if account != nil {
		feeRate = account.FeeRate
	}

	// 2. 清分计算 (Clearing)
	amtDec := decimal.NewFromInt(amount).Div(decimal.NewFromInt(100))
	platformFeeDec := amtDec.Mul(feeRate)
	merchantReceivableDec := amtDec.Sub(platformFeeDec)

	// 3. 构造会计分录 (Accounting)
	entry := &domain.JournalEntry{
		TransactionID: orderNo,
		EventType:     "PAYMENT_SUCCESS",
		Description:   fmt.Sprintf("Payment for Order %s", orderNo),
		PostingDate:   time.Now(),
		Lines: []domain.EntryLine{
			{
				SubjectCode: "1001",
				AccountID:   s.getAccountID(ctx, "1001", "CHANNEL_GLOBAL"),
				Direction:   domain.Debit,
				Amount:      amtDec,
			},
			{
				SubjectCode: "2001",
				AccountID:   s.getAccountID(ctx, "2001", fmt.Sprintf("MERCH_%d", merchantID)),
				Direction:   domain.Credit,
				Amount:      merchantReceivableDec,
			},
			{
				SubjectCode: "6001",
				AccountID:   s.getAccountID(ctx, "6001", "PLATFORM_MAIN"),
				Direction:   domain.Credit,
				Amount:      platformFeeDec,
			},
		},
	}

	// 4. 调用账务核心记账 (由于 LedgerService.PostEntry 内部自带事务，且目前不支持外部注入，暂时保持)
	// 改进：为了保证原子性，如果后续需要可以将 PostEntry 重构为支持外部事务。
	if err := s.ledgerService.PostEntry(ctx, entry); err != nil {
		s.logger.ErrorContext(ctx, "failed to post ledger entry", "order_id", orderID, "error", err)
		return err
	}

	// 5. 跨项目同步 (Cross-Project Interaction) - 非强制一致性，失败记录日志
	if s.remoteAccountCli != nil {
		_, err := s.remoteAccountCli.Deposit(ctx, &accountv1.DepositRequest{
			UserId:   fmt.Sprintf("%d", merchantID),
			Amount:   merchantReceivableDec.String(),
			Currency: "USD",
		})
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to sync settlement to financial account", "merchant_id", merchantID, "error", err)
		} else {
			s.logger.InfoContext(ctx, "settlement synced to financial account successfully", "merchant_id", merchantID)
		}
	}

	s.logger.InfoContext(ctx, "payment recorded in ledger", "entry_no", entry.EntryNo)
	return nil
}

// getAccountID 辅助方法。
func (s *SettlementCommandService) getAccountID(ctx context.Context, subjectCode, entityID string) uint64 {
	acc, err := s.ledgerService.CreateAccount(ctx, subjectCode, entityID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get/create account", "subject", subjectCode, "entity", entityID, "error", err)
		return 0
	}
	return uint64(acc.ID)
}

// CreateSettlement 创建结算单。
func (s *SettlementCommandService) CreateSettlement(ctx context.Context, merchantID uint64, cycle string, startDate, endDate time.Time) (*domain.Settlement, error) {
	settlementNo := fmt.Sprintf("S%d%d", merchantID, time.Now().UnixNano())
	settlement := &domain.Settlement{
		SettlementNo: settlementNo,
		MerchantID:   merchantID,
		Cycle:        domain.SettlementCycle(cycle),
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       domain.SettlementStatusPending,
	}

	err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveSettlementInTx(ctx, tx, settlement); err != nil {
			return err
		}

		event := &domain.SettlementCreatedEvent{
			SettlementNo: settlement.SettlementNo,
			MerchantID:   settlement.MerchantID,
			TotalAmount:  uint64(settlement.GrossAmount.Mul(decimal.NewFromInt(100)).IntPart()),
			Timestamp:    time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, "settlement.created", settlement.SettlementNo, event)
	})
	if err != nil {
		return nil, err
	}

	return settlement, nil
}

// AddOrderToSettlement 添加订单到结算单.
func (s *SettlementCommandService) AddOrderToSettlement(ctx context.Context, settlementID uint64, orderID uint64, orderNo string, amount uint64) error {
	return s.repo.WithTx(ctx, func(tx any) error {
		settlement, err := s.repo.GetSettlement(ctx, settlementID)
		if err != nil {
			return err
		}
		if settlement == nil {
			return errors.New("settlement not found")
		}

		if settlement.Status != domain.SettlementStatusPending {
			return errors.New("settlement is not pending")
		}

		account, err := s.repo.GetMerchantAccount(ctx, settlement.MerchantID)
		if err != nil {
			return err
		}
		feeRate := decimal.Zero
		if account != nil {
			feeRate = account.FeeRate
		}

		amtDec := decimal.NewFromUint64(amount).Div(decimal.NewFromInt(100))
		platformFeeDec := amtDec.Mul(feeRate)
		settlementAmountDec := amtDec.Sub(platformFeeDec)

		detail := &domain.SettlementDetail{
			SettlementID:     fmt.Sprintf("%d", settlementID),
			OrderID:          orderID,
			OrderNo:          orderNo,
			OrderAmount:      amtDec,
			PlatformFee:      platformFeeDec,
			SettlementAmount: settlementAmountDec,
		}

		if err := s.repo.SaveSettlementDetailInTx(ctx, tx, detail); err != nil {
			return err
		}

		settlement.OrderCount++
		settlement.GrossAmount = settlement.GrossAmount.Add(amtDec)
		settlement.PlatformCommission = settlement.PlatformCommission.Add(platformFeeDec)
		settlement.SettlementAmount = settlement.SettlementAmount.Add(settlementAmountDec)

		return s.repo.SaveSettlementInTx(ctx, tx, settlement)
	})
}

// ProcessSettlement 处理结算单（开始处理）。
func (s *SettlementCommandService) ProcessSettlement(ctx context.Context, id uint64) error {
	return s.repo.WithTx(ctx, func(tx any) error {
		settlement, err := s.repo.GetSettlement(ctx, id)
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
		if err := s.repo.SaveSettlementInTx(ctx, tx, settlement); err != nil {
			return err
		}

		event := &domain.SettlementProcessedEvent{
			SettlementNo: settlement.SettlementNo,
			MerchantID:   settlement.MerchantID,
			Amount:       uint64(settlement.SettlementAmount.Mul(decimal.NewFromInt(100)).IntPart()),
			Timestamp:    time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, "settlement.processed", settlement.SettlementNo, event)
	})
}

// CompleteSettlement 完成结算单。
func (s *SettlementCommandService) CompleteSettlement(ctx context.Context, id uint64) error {
	return s.repo.WithTx(ctx, func(tx any) error {
		settlement, err := s.repo.GetSettlement(ctx, id)
		if err != nil {
			return err
		}
		if settlement == nil {
			return errors.New("settlement not found")
		}

		if settlement.Status != domain.SettlementStatusProcessing {
			return errors.New("settlement is not processing")
		}

		account, err := s.repo.GetMerchantAccount(ctx, settlement.MerchantID)
		if err != nil {
			return err
		}
		if account == nil {
			account = &domain.MerchantAccount{
				MerchantID: settlement.MerchantID,
				FeeRate:    decimal.Zero,
			}
		}

		account.Balance = account.Balance.Add(settlement.SettlementAmount)
		account.TotalIncome = account.TotalIncome.Add(settlement.SettlementAmount)
		if err := s.repo.SaveMerchantAccountInTx(ctx, tx, account); err != nil {
			return err
		}

		now := time.Now()
		settlement.Status = domain.SettlementStatusCompleted
		settlement.SettledAt = &now
		if err := s.repo.SaveSettlementInTx(ctx, tx, settlement); err != nil {
			return err
		}

		event := &domain.SettlementCompletedEvent{
			SettlementNo: settlement.SettlementNo,
			MerchantID:   settlement.MerchantID,
			Amount:       uint64(settlement.SettlementAmount.Mul(decimal.NewFromInt(100)).IntPart()),
			Timestamp:    time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, "settlement.completed", settlement.SettlementNo, event)
	})
}
