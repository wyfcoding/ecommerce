package domain

import (
	"context"
	"time"
)

type SettlementRepository interface {
	Save(ctx context.Context, settlement *Settlement) error
	Update(ctx context.Context, settlement *Settlement) error
	GetByID(ctx context.Context, id string) (*Settlement, error)
	GetByIDForUpdate(ctx context.Context, id string) (*Settlement, error)
	ListByMerchant(ctx context.Context, merchantID uint64, status SettlementStatus, page, pageSize int) ([]*Settlement, int64, error)
	ListByPeriod(ctx context.Context, merchantID uint64, periodStart, periodEnd time.Time) ([]*Settlement, error)
	GetByMerchantAndPeriod(ctx context.Context, merchantID uint64, periodStart, periodEnd time.Time) (*Settlement, error)
	SaveDetail(ctx context.Context, detail *SettlementDetail) error
	SaveDetails(ctx context.Context, details []*SettlementDetail) error
	GetDetailsBySettlementID(ctx context.Context, settlementID string) ([]*SettlementDetail, error)
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
	OrderCount       int64
	GrossAmount      int64
	RefundAmount     int64
	PlatformCommission int64
	PromotionFee     int64
	LogisticsFee     int64
	OrderDetails     []*OrderSettlementDetail
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
