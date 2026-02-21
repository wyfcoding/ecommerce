package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
)

type CreateSettlementCommand struct {
	MerchantID  uint64
	Cycle       domain.SettlementCycle
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type CreateSettlementHandler struct {
	settlementRepo domain.SettlementRepository
	configRepo     domain.MerchantSettlementConfigRepository
}

func NewCreateSettlementHandler(settlementRepo domain.SettlementRepository, configRepo domain.MerchantSettlementConfigRepository) *CreateSettlementHandler {
	return &CreateSettlementHandler{
		settlementRepo: settlementRepo,
		configRepo:     configRepo,
	}
}

func (h *CreateSettlementHandler) Handle(ctx context.Context, cmd *CreateSettlementCommand) (*domain.Settlement, error) {
	existing, err := h.settlementRepo.GetByMerchantAndPeriod(ctx, cmd.MerchantID, cmd.PeriodStart, cmd.PeriodEnd)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrDuplicateSettlement
	}

	settlementID := fmt.Sprintf("ST%s", uuid.New().String()[:16])
	settlement := domain.NewSettlement(settlementID, cmd.MerchantID, cmd.Cycle, cmd.PeriodStart, cmd.PeriodEnd)

	if err := h.settlementRepo.Save(ctx, settlement); err != nil {
		return nil, err
	}

	return settlement, nil
}

type CalculateSettlementCommand struct {
	SettlementID string
}

type CalculateSettlementHandler struct {
	settlementRepo domain.SettlementRepository
	calculator     domain.SettlementCalculatorService
	configRepo     domain.MerchantSettlementConfigRepository
}

func NewCalculateSettlementHandler(
	settlementRepo domain.SettlementRepository,
	calculator domain.SettlementCalculatorService,
	configRepo domain.MerchantSettlementConfigRepository,
) *CalculateSettlementHandler {
	return &CalculateSettlementHandler{
		settlementRepo: settlementRepo,
		calculator:     calculator,
		configRepo:     configRepo,
	}
}

func (h *CalculateSettlementHandler) Handle(ctx context.Context, cmd *CalculateSettlementCommand) error {
	settlement, err := h.settlementRepo.GetByIDForUpdate(ctx, cmd.SettlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return domain.ErrSettlementNotFound
	}

	if err := settlement.StartCalculation(); err != nil {
		return err
	}

	if err := h.settlementRepo.Update(ctx, settlement); err != nil {
		return err
	}

	result, err := h.calculator.CalculateSettlement(ctx, settlement.MerchantID, settlement.PeriodStart, settlement.PeriodEnd)
	if err != nil {
		settlement.FailPayment(fmt.Sprintf("calculation failed: %v", err))
		h.settlementRepo.Update(ctx, settlement)
		return err
	}

	for _, od := range result.OrderDetails {
		detail := domain.NewSettlementDetail(
			settlement.SettlementID,
			od.OrderID,
			od.OrderNo,
			decimal.NewFromInt(od.OrderAmount),
			decimal.NewFromInt(od.RefundAmount),
			decimal.NewFromInt(od.PlatformCommission),
			decimal.NewFromInt(od.PromotionFee),
			decimal.NewFromInt(od.LogisticsFee),
		)
		settlement.AddDetail(detail)
	}

	if err := settlement.CompleteCalculation(); err != nil {
		return err
	}

	if err := h.settlementRepo.Update(ctx, settlement); err != nil {
		return err
	}

	details := settlement.Details()
	if len(details) > 0 {
		if err := h.settlementRepo.SaveDetails(ctx, details); err != nil {
			return err
		}
	}

	config, _ := h.configRepo.GetByMerchantID(ctx, settlement.MerchantID)
	if config != nil && config.AutoApprove {
		approveCmd := &ApproveSettlementCommand{
			SettlementID: cmd.SettlementID,
			ApprovedBy:   0,
		}
		approveHandler := NewApproveSettlementHandler(h.settlementRepo)
		if err := approveHandler.Handle(ctx, approveCmd); err != nil {
			return err
		}
	}

	return nil
}

type ApproveSettlementCommand struct {
	SettlementID string
	ApprovedBy   uint64
}

type ApproveSettlementHandler struct {
	settlementRepo domain.SettlementRepository
}

func NewApproveSettlementHandler(settlementRepo domain.SettlementRepository) *ApproveSettlementHandler {
	return &ApproveSettlementHandler{settlementRepo: settlementRepo}
}

func (h *ApproveSettlementHandler) Handle(ctx context.Context, cmd *ApproveSettlementCommand) error {
	settlement, err := h.settlementRepo.GetByIDForUpdate(ctx, cmd.SettlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return domain.ErrSettlementNotFound
	}

	if err := settlement.Approve(cmd.ApprovedBy); err != nil {
		return err
	}

	return h.settlementRepo.Update(ctx, settlement)
}

type RejectSettlementCommand struct {
	SettlementID string
	Reason       string
}

type RejectSettlementHandler struct {
	settlementRepo domain.SettlementRepository
}

func NewRejectSettlementHandler(settlementRepo domain.SettlementRepository) *RejectSettlementHandler {
	return &RejectSettlementHandler{settlementRepo: settlementRepo}
}

func (h *RejectSettlementHandler) Handle(ctx context.Context, cmd *RejectSettlementCommand) error {
	settlement, err := h.settlementRepo.GetByIDForUpdate(ctx, cmd.SettlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return domain.ErrSettlementNotFound
	}

	if err := settlement.Reject(cmd.Reason); err != nil {
		return err
	}

	return h.settlementRepo.Update(ctx, settlement)
}

type PaySettlementCommand struct {
	SettlementID  string
	BankAccountID uint64
}

type PaySettlementHandler struct {
	settlementRepo  domain.SettlementRepository
	bankAccountRepo domain.MerchantBankAccountRepository
	paymentGateway  domain.PaymentGateway
}

func NewPaySettlementHandler(
	settlementRepo domain.SettlementRepository,
	bankAccountRepo domain.MerchantBankAccountRepository,
	paymentGateway domain.PaymentGateway,
) *PaySettlementHandler {
	return &PaySettlementHandler{
		settlementRepo:  settlementRepo,
		bankAccountRepo: bankAccountRepo,
		paymentGateway:  paymentGateway,
	}
}

func (h *PaySettlementHandler) Handle(ctx context.Context, cmd *PaySettlementCommand) error {
	settlement, err := h.settlementRepo.GetByIDForUpdate(ctx, cmd.SettlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return domain.ErrSettlementNotFound
	}

	bankAccount, err := h.bankAccountRepo.GetByID(ctx, cmd.BankAccountID)
	if err != nil {
		return err
	}
	if bankAccount == nil {
		return domain.ErrBankAccountNotFound
	}

	if err := settlement.StartPayment(cmd.BankAccountID); err != nil {
		return err
	}

	if err := h.settlementRepo.Update(ctx, settlement); err != nil {
		return err
	}

	result, err := h.paymentGateway.Transfer(ctx, &domain.TransferRequest{
		MerchantID:    settlement.MerchantID,
		BankAccountID: cmd.BankAccountID,
		Amount:        settlement.SettlementAmount.IntPart(),
		SettlementID:  settlement.SettlementID,
		Description:   fmt.Sprintf("商户结算 %s", settlement.SettlementID),
	})
	if err != nil {
		settlement.FailPayment(err.Error())
		h.settlementRepo.Update(ctx, settlement)
		return err
	}

	if result.Status == "SUCCESS" {
		return settlement.CompletePayment(result.TransactionID)
	} else {
		return settlement.FailPayment(result.FailedReason)
	}
}

type AdjustSettlementCommand struct {
	SettlementID     string
	AdjustmentAmount decimal.Decimal
	Reason           string
}

type AdjustSettlementHandler struct {
	settlementRepo domain.SettlementRepository
}

func NewAdjustSettlementHandler(settlementRepo domain.SettlementRepository) *AdjustSettlementHandler {
	return &AdjustSettlementHandler{settlementRepo: settlementRepo}
}

func (h *AdjustSettlementHandler) Handle(ctx context.Context, cmd *AdjustSettlementCommand) error {
	settlement, err := h.settlementRepo.GetByIDForUpdate(ctx, cmd.SettlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return domain.ErrSettlementNotFound
	}

	settlement.SetAdjustment(cmd.AdjustmentAmount, cmd.Reason)
	return h.settlementRepo.Update(ctx, settlement)
}

type CancelSettlementCommand struct {
	SettlementID string
	Reason       string
}

type CancelSettlementHandler struct {
	settlementRepo domain.SettlementRepository
}

func NewCancelSettlementHandler(settlementRepo domain.SettlementRepository) *CancelSettlementHandler {
	return &CancelSettlementHandler{settlementRepo: settlementRepo}
}

func (h *CancelSettlementHandler) Handle(ctx context.Context, cmd *CancelSettlementCommand) error {
	settlement, err := h.settlementRepo.GetByIDForUpdate(ctx, cmd.SettlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return domain.ErrSettlementNotFound
	}

	return settlement.Cancel(cmd.Reason)
}
