package application

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/domain"
)

type SettlementDTO struct {
	ID               uint64          `json:"id"`
	SettlementID     string          `json:"settlement_id"`
	MerchantID       uint64          `json:"merchant_id"`
	Cycle            string          `json:"cycle"`
	PeriodStart      time.Time       `json:"period_start"`
	PeriodEnd        time.Time       `json:"period_end"`
	OrderCount       int64           `json:"order_count"`
	GrossAmount      decimal.Decimal `json:"gross_amount"`
	RefundAmount     decimal.Decimal `json:"refund_amount"`
	PlatformCommission decimal.Decimal `json:"platform_commission"`
	PromotionFee     decimal.Decimal `json:"promotion_fee"`
	LogisticsFee     decimal.Decimal `json:"logistics_fee"`
	AdjustmentAmount decimal.Decimal `json:"adjustment_amount"`
	SettlementAmount decimal.Decimal `json:"settlement_amount"`
	Status           string          `json:"status"`
	BankAccountID    uint64          `json:"bank_account_id"`
	TransactionRef   string          `json:"transaction_ref"`
	ApprovedBy       uint64          `json:"approved_by"`
	ApprovedAt       *time.Time      `json:"approved_at"`
	PaidAt           *time.Time      `json:"paid_at"`
	FailReason       string          `json:"fail_reason"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type SettlementDetailDTO struct {
	ID                 uint64          `json:"id"`
	SettlementID       string          `json:"settlement_id"`
	OrderID            uint64          `json:"order_id"`
	OrderNo            string          `json:"order_no"`
	OrderAmount        decimal.Decimal `json:"order_amount"`
	RefundAmount       decimal.Decimal `json:"refund_amount"`
	PlatformCommission decimal.Decimal `json:"platform_commission"`
	PromotionFee       decimal.Decimal `json:"promotion_fee"`
	LogisticsFee       decimal.Decimal `json:"logistics_fee"`
	SettlementAmount   decimal.Decimal `json:"settlement_amount"`
	CreatedAt          time.Time       `json:"created_at"`
}

type BankAccountDTO struct {
	ID          uint64 `json:"id"`
	MerchantID  uint64 `json:"merchant_id"`
	BankName    string `json:"bank_name"`
	BankCode    string `json:"bank_code"`
	AccountName string `json:"account_name"`
	AccountNo   string `json:"account_no"`
	BranchName  string `json:"branch_name"`
	IsDefault   bool   `json:"is_default"`
	Status      string `json:"status"`
}

type SettlementConfigDTO struct {
	ID                  uint64          `json:"id"`
	MerchantID          uint64          `json:"merchant_id"`
	Cycle               string          `json:"cycle"`
	CommissionRate      decimal.Decimal `json:"commission_rate"`
	MinSettlementAmount decimal.Decimal `json:"min_settlement_amount"`
	AutoApprove         bool            `json:"auto_approve"`
	AutoPay             bool            `json:"auto_pay"`
	Status              string          `json:"status"`
}

type GetSettlementQuery struct {
	SettlementID string
}

type GetSettlementHandler struct {
	settlementRepo domain.SettlementRepository
}

func NewGetSettlementHandler(settlementRepo domain.SettlementRepository) *GetSettlementHandler {
	return &GetSettlementHandler{settlementRepo: settlementRepo}
}

func (h *GetSettlementHandler) Handle(ctx context.Context, query *GetSettlementQuery) (*SettlementDTO, error) {
	settlement, err := h.settlementRepo.GetByID(ctx, query.SettlementID)
	if err != nil {
		return nil, err
	}
	if settlement == nil {
		return nil, domain.ErrSettlementNotFound
	}
	return toSettlementDTO(settlement), nil
}

type ListSettlementsQuery struct {
	MerchantID uint64
	Status     domain.SettlementStatus
	Page       int
	PageSize   int
}

type ListSettlementsResult struct {
	Settlements []*SettlementDTO `json:"settlements"`
	Total       int64            `json:"total"`
	Page        int              `json:"page"`
	PageSize    int              `json:"page_size"`
}

type ListSettlementsHandler struct {
	settlementRepo domain.SettlementRepository
}

func NewListSettlementsHandler(settlementRepo domain.SettlementRepository) *ListSettlementsHandler {
	return &ListSettlementsHandler{settlementRepo: settlementRepo}
}

func (h *ListSettlementsHandler) Handle(ctx context.Context, query *ListSettlementsQuery) (*ListSettlementsResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	settlements, total, err := h.settlementRepo.ListByMerchant(ctx, query.MerchantID, query.Status, query.Page, query.PageSize)
	if err != nil {
		return nil, err
	}

	dtos := make([]*SettlementDTO, len(settlements))
	for i, s := range settlements {
		dtos[i] = toSettlementDTO(s)
	}

	return &ListSettlementsResult{
		Settlements: dtos,
		Total:       total,
		Page:        query.Page,
		PageSize:    query.PageSize,
	}, nil
}

type GetSettlementDetailsQuery struct {
	SettlementID string
}

type GetSettlementDetailsHandler struct {
	settlementRepo domain.SettlementRepository
}

func NewGetSettlementDetailsHandler(settlementRepo domain.SettlementRepository) *GetSettlementDetailsHandler {
	return &GetSettlementDetailsHandler{settlementRepo: settlementRepo}
}

func (h *GetSettlementDetailsHandler) Handle(ctx context.Context, query *GetSettlementDetailsQuery) ([]*SettlementDetailDTO, error) {
	details, err := h.settlementRepo.GetDetailsBySettlementID(ctx, query.SettlementID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*SettlementDetailDTO, len(details))
	for i, d := range details {
		dtos[i] = toSettlementDetailDTO(d)
	}

	return dtos, nil
}

type ListBankAccountsQuery struct {
	MerchantID uint64
}

type ListBankAccountsHandler struct {
	bankAccountRepo domain.MerchantBankAccountRepository
}

func NewListBankAccountsHandler(bankAccountRepo domain.MerchantBankAccountRepository) *ListBankAccountsHandler {
	return &ListBankAccountsHandler{bankAccountRepo: bankAccountRepo}
}

func (h *ListBankAccountsHandler) Handle(ctx context.Context, query *ListBankAccountsQuery) ([]*BankAccountDTO, error) {
	accounts, err := h.bankAccountRepo.GetByMerchantID(ctx, query.MerchantID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*BankAccountDTO, len(accounts))
	for i, a := range accounts {
		dtos[i] = toBankAccountDTO(a)
	}

	return dtos, nil
}

type GetSettlementConfigQuery struct {
	MerchantID uint64
}

type GetSettlementConfigHandler struct {
	configRepo domain.MerchantSettlementConfigRepository
}

func NewGetSettlementConfigHandler(configRepo domain.MerchantSettlementConfigRepository) *GetSettlementConfigHandler {
	return &GetSettlementConfigHandler{configRepo: configRepo}
}

func (h *GetSettlementConfigHandler) Handle(ctx context.Context, query *GetSettlementConfigQuery) (*SettlementConfigDTO, error) {
	config, err := h.configRepo.GetByMerchantID(ctx, query.MerchantID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, domain.ErrConfigNotFound
	}
	return toSettlementConfigDTO(config), nil
}

func toSettlementDTO(s *domain.Settlement) *SettlementDTO {
	return &SettlementDTO{
		ID:                 s.ID,
		SettlementID:       s.SettlementID,
		MerchantID:         s.MerchantID,
		Cycle:              string(s.Cycle),
		PeriodStart:        s.PeriodStart,
		PeriodEnd:          s.PeriodEnd,
		OrderCount:         s.OrderCount,
		GrossAmount:        s.GrossAmount,
		RefundAmount:       s.RefundAmount,
		PlatformCommission: s.PlatformCommission,
		PromotionFee:       s.PromotionFee,
		LogisticsFee:       s.LogisticsFee,
		AdjustmentAmount:   s.AdjustmentAmount,
		SettlementAmount:   s.SettlementAmount,
		Status:             string(s.Status),
		BankAccountID:      s.BankAccountID,
		TransactionRef:     s.TransactionRef,
		ApprovedBy:         s.ApprovedBy,
		ApprovedAt:         s.ApprovedAt,
		PaidAt:             s.PaidAt,
		FailReason:         s.FailReason,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}

func toSettlementDetailDTO(d *domain.SettlementDetail) *SettlementDetailDTO {
	return &SettlementDetailDTO{
		ID:                 d.ID,
		SettlementID:       d.SettlementID,
		OrderID:            d.OrderID,
		OrderNo:            d.OrderNo,
		OrderAmount:        d.OrderAmount,
		RefundAmount:       d.RefundAmount,
		PlatformCommission: d.PlatformCommission,
		PromotionFee:       d.PromotionFee,
		LogisticsFee:       d.LogisticsFee,
		SettlementAmount:   d.SettlementAmount,
		CreatedAt:          d.CreatedAt,
	}
}

func toBankAccountDTO(a *domain.MerchantBankAccount) *BankAccountDTO {
	return &BankAccountDTO{
		ID:          a.ID,
		MerchantID:  a.MerchantID,
		BankName:    a.BankName,
		BankCode:    a.BankCode,
		AccountName: a.AccountName,
		AccountNo:   a.AccountNo,
		BranchName:  a.BranchName,
		IsDefault:   a.IsDefault,
		Status:      string(a.Status),
	}
}

func toSettlementConfigDTO(c *domain.MerchantSettlementConfig) *SettlementConfigDTO {
	return &SettlementConfigDTO{
		ID:                  c.ID,
		MerchantID:          c.MerchantID,
		Cycle:               string(c.Cycle),
		CommissionRate:      c.CommissionRate,
		MinSettlementAmount: c.MinSettlementAmount,
		AutoApprove:         c.AutoApprove,
		AutoPay:             c.AutoPay,
		Status:              string(c.Status),
	}
}
