package domain

import (
	"context"
	"time"
)

// SettlementRepository 是结算模块的核心仓储接口。
type SettlementRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- 结算单管理 (Settlement methods) ---
	Save(ctx context.Context, settlement *Settlement) error
	Update(ctx context.Context, settlement *Settlement) error
	GetByID(ctx context.Context, id string) (*Settlement, error)
	GetByIDForUpdate(ctx context.Context, id string) (*Settlement, error)
	GetByMerchantAndPeriod(ctx context.Context, merchantID uint64, start, end time.Time) (*Settlement, error)
	ListSettlements(ctx context.Context, merchantID uint64, status *SettlementStatus, offset, limit int) ([]*Settlement, int64, error)
	ListByMerchant(ctx context.Context, merchantID uint64, status *SettlementStatus, offset, limit int) ([]*Settlement, int64, error)
	SaveSettlement(ctx context.Context, settlement *Settlement) error
	SaveSettlementInTx(ctx context.Context, tx any, settlement *Settlement) error
	GetSettlement(ctx context.Context, id uint64) (*Settlement, error)
	GetSettlementByNo(ctx context.Context, no string) (*Settlement, error)

	// --- 结算明细管理 (SettlementDetail methods) ---
	SaveDetails(ctx context.Context, details []*SettlementDetail) error
	SaveSettlementDetail(ctx context.Context, detail *SettlementDetail) error
	SaveSettlementDetailInTx(ctx context.Context, tx any, detail *SettlementDetail) error
	ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*SettlementDetail, error)
	GetDetailsBySettlementID(ctx context.Context, settlementID string) ([]*SettlementDetail, error)

	// --- 商户账户管理 (MerchantAccount methods) ---
	GetMerchantAccount(ctx context.Context, merchantID uint64) (*MerchantAccount, error)
	SaveMerchantAccount(ctx context.Context, account *MerchantAccount) error
	SaveMerchantAccountInTx(ctx context.Context, tx any, account *MerchantAccount) error
}

// LedgerRepository 账本仓储接口
type LedgerRepository interface {
	Save(ctx context.Context, ledger *Ledger) error
	GetByID(ctx context.Context, id uint64) (*Ledger, error)
	GetBySettlementID(ctx context.Context, settlementID uint64) (*Ledger, error)
	ListByMerchant(ctx context.Context, merchantID uint64, offset, limit int) ([]*Ledger, int64, error)
}

type MerchantBankAccountRepository interface {
	Save(ctx context.Context, account *MerchantBankAccount) error
	Update(ctx context.Context, account *MerchantBankAccount) error
	GetByID(ctx context.Context, id uint64) (*MerchantBankAccount, error)
	GetByMerchantID(ctx context.Context, merchantID uint64) ([]*MerchantBankAccount, error)
	GetDefaultByMerchantID(ctx context.Context, merchantID uint64) (*MerchantBankAccount, error)
	Delete(ctx context.Context, id uint64) error
}

type MerchantSettlementConfigRepository interface {
	Save(ctx context.Context, config *MerchantSettlementConfig) error
	Update(ctx context.Context, config *MerchantSettlementConfig) error
	GetByMerchantID(ctx context.Context, merchantID uint64) (*MerchantSettlementConfig, error)
}

type SettlementCalculatorService interface {
	CalculateSettlement(ctx context.Context, merchantID uint64, periodStart, periodEnd time.Time) (*SettlementCalculationResult, error)
}

type SettlementCalculationResult struct {
	OrderCount         int64
	GrossAmount        int64
	RefundAmount       int64
	PlatformCommission int64
	PromotionFee       int64
	LogisticsFee       int64
	OrderDetails       []*OrderSettlementDetail
}

type OrderSettlementDetail struct {
	OrderID            uint64
	OrderNo            string
	OrderAmount        int64
	RefundAmount       int64
	PlatformCommission int64
	PromotionFee       int64
	LogisticsFee       int64
}

type PaymentGateway interface {
	Transfer(ctx context.Context, req *TransferRequest) (*TransferResult, error)
	QueryTransfer(ctx context.Context, transactionID string) (*TransferResult, error)
}

type TransferRequest struct {
	MerchantID    uint64
	BankAccountID uint64
	Amount        int64
	SettlementID  string
	Description   string
}

type TransferResult struct {
	TransactionID string
	Status        string
	FailedReason  string
}
