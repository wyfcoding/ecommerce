package mysql

import (
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"

	"gorm.io/gorm"
)

// SettlementModel 结算单写模型（持久化专用）。
type SettlementModel struct {
	gorm.Model
	SettlementNo     string                  `gorm:"type:varchar(64);uniqueIndex;not null;comment:结算单号"`
	MerchantID       uint64                  `gorm:"index;not null;comment:商户ID"`
	Cycle            domain.SettlementCycle  `gorm:"type:varchar(32);not null;comment:结算周期"`
	StartDate        time.Time               `gorm:"not null;comment:开始日期"`
	EndDate          time.Time               `gorm:"not null;comment:结束日期"`
	OrderCount       int64                   `gorm:"not null;default:0;comment:订单数量"`
	TotalAmount      uint64                  `gorm:"not null;default:0;comment:总金额(分)"`
	PlatformFee      uint64                  `gorm:"not null;default:0;comment:平台手续费(分)"`
	CommissionAmount int64                   `gorm:"not null;default:0;comment:佣金金额(分)"`
	RebateAmount     int64                   `gorm:"not null;default:0;comment:返利金额(分)"`
	OtherFees        int64                   `gorm:"not null;default:0;comment:其他费用(分)"`
	SettlementAmount uint64                  `gorm:"not null;default:0;comment:结算金额(分)"`
	TotalFXGainLoss  int64                   `gorm:"not null;default:0;comment:总汇兑损益(分)"`
	Status           domain.SettlementStatus `gorm:"type:tinyint;not null;default:0;comment:状态"`
	SettledAt        *time.Time              `gorm:"comment:结算时间"`
	ApprovedBy       uint64                  `gorm:"index;comment:审核人"`
	ApprovedAt       *time.Time              `gorm:"comment:审核时间"`
	FailReason       string                  `gorm:"type:varchar(255);comment:失败原因"`
}

func (SettlementModel) TableName() string {
	return "settlements"
}

// SettlementDetailModel 结算明细写模型（持久化专用）。
type SettlementDetailModel struct {
	gorm.Model
	SettlementID     uint64  `gorm:"index;not null;comment:结算单ID"`
	OrderID          uint64  `gorm:"index;not null;comment:订单ID"`
	OrderNo          string  `gorm:"type:varchar(64);not null;comment:订单号"`
	OrderAmount      uint64  `gorm:"not null;comment:订单金额(分)"`
	PlatformFee      uint64  `gorm:"not null;comment:平台手续费(分)"`
	LogisticsFee     int64   `gorm:"not null;default:0;comment:物流费(分)"`
	ReturnFee        int64   `gorm:"not null;default:0;comment:退货费(分)"`
	OtherFee         int64   `gorm:"not null;default:0;comment:其他费用(分)"`
	SettlementAmount uint64  `gorm:"not null;comment:结算金额(分)"`
	SourceCurrency   string  `gorm:"type:varchar(3);comment:原始币种"`
	ExchangeRate     float64 `gorm:"type:decimal(18,6);comment:汇率"`
	FXGainLoss       int64   `gorm:"not null;default:0;comment:汇兑损益(分)"`
}

func (SettlementDetailModel) TableName() string {
	return "settlement_details"
}

// SettlementPaymentModel 结算支付记录写模型（持久化专用）。
type SettlementPaymentModel struct {
	gorm.Model
	SettlementID  uint64               `gorm:"not null;index;comment:结算ID"`
	MerchantID    uint64               `gorm:"not null;index;comment:商户ID"`
	Amount        int64                `gorm:"not null;comment:支付金额(分)"`
	Status        domain.PaymentStatus `gorm:"type:varchar(32);default:'pending';comment:状态"`
	TransactionID string               `gorm:"type:varchar(128);comment:交易流水号"`
	CompletedAt   *time.Time           `gorm:"comment:完成时间"`
}

func (SettlementPaymentModel) TableName() string {
	return "settlement_payments"
}

// MerchantAccountModel 商户账户写模型（持久化专用）。
type MerchantAccountModel struct {
	gorm.Model
	MerchantID    uint64  `gorm:"uniqueIndex;not null;comment:商户ID"`
	Balance       uint64  `gorm:"not null;default:0;comment:余额(分)"`
	FrozenBalance uint64  `gorm:"not null;default:0;comment:冻结金额(分)"`
	TotalIncome   uint64  `gorm:"not null;default:0;comment:总收入(分)"`
	TotalWithdraw uint64  `gorm:"not null;default:0;comment:总提现(分)"`
	FeeRate       float64 `gorm:"type:decimal(5,2);not null;default:0;comment:费率(%)"`
}

func (MerchantAccountModel) TableName() string {
	return "merchant_accounts"
}

// SubjectModel 科目模型（持久化专用）。
type SubjectModel struct {
	Code        string             `gorm:"primaryKey;type:varchar(32);comment:科目代码"`
	Name        string             `gorm:"type:varchar(64);not null;comment:科目名称"`
	Type        domain.AccountType `gorm:"type:varchar(32);not null;comment:科目类型"`
	Description string             `gorm:"type:varchar(255);comment:描述"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (SubjectModel) TableName() string {
	return "subjects"
}

// AccountModel 账户模型（持久化专用）。
type AccountModel struct {
	gorm.Model
	SubjectCode string `gorm:"index;type:varchar(32);not null;comment:关联科目"`
	EntityID    string `gorm:"uniqueIndex:idx_sub_ent;type:varchar(64);not null;comment:关联实体ID"`
	Balance     int64  `gorm:"not null;default:0;comment:余额(分)"`
	Currency    string `gorm:"type:varchar(3);default:'CNY';comment:币种"`
	Version     int64  `gorm:"default:0;comment:乐观锁版本"`
}

func (AccountModel) TableName() string {
	return "accounts"
}

// JournalEntryModel 记账凭证模型（持久化专用）。
type JournalEntryModel struct {
	gorm.Model
	EntryNo       string           `gorm:"uniqueIndex;type:varchar(64);not null;comment:凭证号"`
	TransactionID string           `gorm:"index;type:varchar(64);not null;comment:关联业务流水号"`
	EventType     string           `gorm:"type:varchar(32);not null;comment:事件类型"`
	PostingDate   time.Time        `gorm:"index;not null;comment:入账日期"`
	Description   string           `gorm:"type:varchar(255);comment:摘要"`
	Lines         []EntryLineModel `gorm:"foreignKey:EntryID"`
}

func (JournalEntryModel) TableName() string {
	return "journal_entries"
}

// EntryLineModel 凭证分录模型（持久化专用）。
type EntryLineModel struct {
	gorm.Model
	EntryID     uint64                `gorm:"index;not null;comment:关联凭证ID"`
	AccountID   uint64                `gorm:"index;not null;comment:关联账户ID"`
	SubjectCode string                `gorm:"type:varchar(32);not null;comment:冗余科目代码"`
	Direction   domain.EntryDirection `gorm:"type:tinyint;not null;comment:借贷方向(1:借, -1:贷)"`
	Amount      int64                 `gorm:"not null;comment:发生金额(分)"`
}

func (EntryLineModel) TableName() string {
	return "entry_lines"
}

// LedgerModel 账本模型
type LedgerModel struct {
	gorm.Model
	MerchantID   uint64 `gorm:"index;not null"`
	SettlementID uint64 `gorm:"index;not null"`
	TotalAmount  int64  `gorm:"not null"`
	Currency     string `gorm:"type:varchar(3);not null"`
	Status       string `gorm:"type:varchar(32);not null"`
}

func (LedgerModel) TableName() string {
	return "ledgers"
}

func toLedgerModel(ledger *domain.Ledger) *LedgerModel {
	if ledger == nil {
		return nil
	}
	return &LedgerModel{
		Model: gorm.Model{
			ID:        uint(ledger.ID),
			CreatedAt: ledger.CreatedAt,
			UpdatedAt: ledger.UpdatedAt,
		},
		MerchantID:   ledger.MerchantID,
		SettlementID: ledger.SettlementID,
		TotalAmount:  ledger.TotalAmount.Mul(decimal.NewFromInt(100)).IntPart(),
		Currency:     ledger.Currency,
		Status:       ledger.Status,
	}
}

func toLedger(model *LedgerModel) *domain.Ledger {
	if model == nil {
		return nil
	}
	return &domain.Ledger{
		ID:           uint64(model.ID),
		MerchantID:   model.MerchantID,
		SettlementID: model.SettlementID,
		TotalAmount:  decimal.NewFromInt(model.TotalAmount).Div(decimal.NewFromInt(100)),
		Currency:     model.Currency,
		Status:       model.Status,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func toSettlementModel(settlement *domain.Settlement) *SettlementModel {
	if settlement == nil {
		return nil
	}
	return &SettlementModel{
		Model: gorm.Model{
			ID:        uint(settlement.ID),
			CreatedAt: settlement.CreatedAt,
			UpdatedAt: settlement.UpdatedAt,
		},
		SettlementNo:     settlement.SettlementNo,
		MerchantID:       settlement.MerchantID,
		Cycle:            settlement.Cycle,
		StartDate:        settlement.StartDate,
		EndDate:          settlement.EndDate,
		OrderCount:       settlement.OrderCount,
		TotalAmount:      uint64(settlement.GrossAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		PlatformFee:      uint64(settlement.PlatformCommission.Mul(decimal.NewFromInt(100)).IntPart()),
		CommissionAmount: 0, // 假设映射
		RebateAmount:     settlement.RebateAmount.Mul(decimal.NewFromInt(100)).IntPart(),
		OtherFees:        settlement.OtherFees.Mul(decimal.NewFromInt(100)).IntPart(),
		SettlementAmount: uint64(settlement.SettlementAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		TotalFXGainLoss:  settlement.TotalFXGainLoss.Mul(decimal.NewFromInt(100)).IntPart(),
		Status:           settlement.Status,
		SettledAt:        settlement.SettledAt,
		ApprovedBy:       settlement.ApprovedBy,
		ApprovedAt:       settlement.ApprovedAt,
		FailReason:       settlement.FailReason,
	}
}

func toSettlement(model *SettlementModel) *domain.Settlement {
	if model == nil {
		return nil
	}
	return &domain.Settlement{
		ID:                 uint64(model.ID),
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
		SettlementNo:       model.SettlementNo,
		MerchantID:         model.MerchantID,
		Cycle:              model.Cycle,
		StartDate:          model.StartDate,
		EndDate:            model.EndDate,
		OrderCount:         model.OrderCount,
		GrossAmount:        decimal.NewFromUint64(model.TotalAmount).Div(decimal.NewFromInt(100)),
		PlatformCommission: decimal.NewFromUint64(model.PlatformFee).Div(decimal.NewFromInt(100)),
		CommissionAmount:   decimal.NewFromInt(model.CommissionAmount).Div(decimal.NewFromInt(100)),
		RebateAmount:       decimal.NewFromInt(model.RebateAmount).Div(decimal.NewFromInt(100)),
		OtherFees:          decimal.NewFromInt(model.OtherFees).Div(decimal.NewFromInt(100)),
		SettlementAmount:   decimal.NewFromUint64(model.SettlementAmount).Div(decimal.NewFromInt(100)),
		TotalFXGainLoss:    decimal.NewFromInt(model.TotalFXGainLoss).Div(decimal.NewFromInt(100)),
		Status:             model.Status,
		SettledAt:          model.SettledAt,
		ApprovedBy:         model.ApprovedBy,
		ApprovedAt:         model.ApprovedAt,
		FailReason:         model.FailReason,
	}
}

func toSettlementDetailModel(detail *domain.SettlementDetail) *SettlementDetailModel {
	if detail == nil {
		return nil
	}
	return &SettlementDetailModel{
		Model: gorm.Model{
			ID:        uint(detail.ID),
			CreatedAt: detail.CreatedAt,
			UpdatedAt: detail.UpdatedAt,
		},
		SettlementID:     toUint64(detail.SettlementID),
		OrderID:          detail.OrderID,
		OrderNo:          detail.OrderNo,
		OrderAmount:      uint64(detail.OrderAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		PlatformFee:      uint64(detail.PlatformCommission.Mul(decimal.NewFromInt(100)).IntPart()),
		LogisticsFee:     detail.LogisticsFee.Mul(decimal.NewFromInt(100)).IntPart(),
		ReturnFee:        detail.RefundAmount.Mul(decimal.NewFromInt(100)).IntPart(),
		OtherFee:         detail.OtherFees.Mul(decimal.NewFromInt(100)).IntPart(),
		SettlementAmount: uint64(detail.SettlementAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		SourceCurrency:   detail.SourceCurrency,
		ExchangeRate:     detail.ExchangeRate.InexactFloat64(),
		FXGainLoss:       detail.FXGainLoss.Mul(decimal.NewFromInt(100)).IntPart(),
	}
}

// toUint64 辅助函数
func toUint64(s string) uint64 {
	u, _ := strconv.ParseUint(s, 10, 64)
	return u
}

func toSettlementDetail(model *SettlementDetailModel) *domain.SettlementDetail {
	if model == nil {
		return nil
	}
	return &domain.SettlementDetail{
		ID:                 uint64(model.ID),
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
		SettlementID:       fmt.Sprintf("%d", model.SettlementID),
		OrderID:            model.OrderID,
		OrderNo:            model.OrderNo,
		OrderAmount:        decimal.NewFromUint64(model.OrderAmount).Div(decimal.NewFromInt(100)),
		PlatformCommission: decimal.NewFromUint64(model.PlatformFee).Div(decimal.NewFromInt(100)),
		LogisticsFee:       decimal.NewFromInt(model.LogisticsFee).Div(decimal.NewFromInt(100)),
		RefundAmount:       decimal.NewFromInt(model.ReturnFee).Div(decimal.NewFromInt(100)),
		OtherFees:          decimal.NewFromInt(model.OtherFee).Div(decimal.NewFromInt(100)),
		SettlementAmount:   decimal.NewFromUint64(model.SettlementAmount).Div(decimal.NewFromInt(100)),
		SourceCurrency:     model.SourceCurrency,
		ExchangeRate:       decimal.NewFromFloat(model.ExchangeRate),
		FXGainLoss:         decimal.NewFromInt(model.FXGainLoss).Div(decimal.NewFromInt(100)),
	}
}

func toMerchantAccountModel(account *domain.MerchantAccount) *MerchantAccountModel {
	if account == nil {
		return nil
	}
	return &MerchantAccountModel{
		Model: gorm.Model{
			ID:        uint(account.ID),
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		},
		MerchantID:    account.MerchantID,
		Balance:       uint64(account.Balance.Mul(decimal.NewFromInt(100)).IntPart()),
		FrozenBalance: uint64(account.FrozenBalance.Mul(decimal.NewFromInt(100)).IntPart()),
		TotalIncome:   uint64(account.TotalIncome.Mul(decimal.NewFromInt(100)).IntPart()),
		TotalWithdraw: uint64(account.TotalWithdraw.Mul(decimal.NewFromInt(100)).IntPart()),
		FeeRate:       account.FeeRate.InexactFloat64(),
	}
}

func toMerchantAccount(model *MerchantAccountModel) *domain.MerchantAccount {
	if model == nil {
		return nil
	}
	return &domain.MerchantAccount{
		ID:            uint64(model.ID),
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		MerchantID:    model.MerchantID,
		Balance:       decimal.NewFromUint64(model.Balance).Div(decimal.NewFromInt(100)),
		FrozenBalance: decimal.NewFromUint64(model.FrozenBalance).Div(decimal.NewFromInt(100)),
		TotalIncome:   decimal.NewFromUint64(model.TotalIncome).Div(decimal.NewFromInt(100)),
		TotalWithdraw: decimal.NewFromUint64(model.TotalWithdraw).Div(decimal.NewFromInt(100)),
		FeeRate:       decimal.NewFromFloat(model.FeeRate),
	}
}

func toSubjectModel(subject *domain.Subject) *SubjectModel {
	if subject == nil {
		return nil
	}
	return &SubjectModel{
		Code:        subject.Code,
		Name:        subject.Name,
		Type:        subject.Type,
		Description: subject.Description,
		CreatedAt:   subject.CreatedAt,
		UpdatedAt:   subject.UpdatedAt,
	}
}

func toSubject(model *SubjectModel) *domain.Subject {
	if model == nil {
		return nil
	}
	return &domain.Subject{
		Code:        model.Code,
		Name:        model.Name,
		Type:        model.Type,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func toAccountModel(account *domain.Account) *AccountModel {
	if account == nil {
		return nil
	}
	return &AccountModel{
		Model: gorm.Model{
			ID:        uint(account.ID),
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		},
		SubjectCode: account.SubjectCode,
		EntityID:    account.EntityID,
		Balance:     account.Balance.Mul(decimal.NewFromInt(100)).IntPart(),
		Currency:    account.Currency,
		Version:     account.Version,
	}
}

func toAccount(model *AccountModel) *domain.Account {
	if model == nil {
		return nil
	}
	return &domain.Account{
		ID:          uint64(model.ID),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		SubjectCode: model.SubjectCode,
		EntityID:    model.EntityID,
		Balance:     decimal.NewFromInt(model.Balance).Div(decimal.NewFromInt(100)),
		Currency:    model.Currency,
		Version:     model.Version,
	}
}

func toJournalEntryModel(entry *domain.JournalEntry) *JournalEntryModel {
	if entry == nil {
		return nil
	}
	model := &JournalEntryModel{
		Model: gorm.Model{
			ID:        uint(entry.ID),
			CreatedAt: entry.CreatedAt,
			UpdatedAt: entry.UpdatedAt,
		},
		EntryNo:       entry.EntryNo,
		TransactionID: entry.TransactionID,
		EventType:     entry.EventType,
		PostingDate:   entry.PostingDate,
		Description:   entry.Description,
	}
	if len(entry.Lines) > 0 {
		model.Lines = make([]EntryLineModel, 0, len(entry.Lines))
		for _, line := range entry.Lines {
			model.Lines = append(model.Lines, toEntryLineModel(&line))
		}
	}
	return model
}

func toJournalEntry(model *JournalEntryModel) *domain.JournalEntry {
	if model == nil {
		return nil
	}
	entry := &domain.JournalEntry{
		ID:            uint64(model.ID),
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		EntryNo:       model.EntryNo,
		TransactionID: model.TransactionID,
		EventType:     model.EventType,
		PostingDate:   model.PostingDate,
		Description:   model.Description,
	}
	if len(model.Lines) > 0 {
		entry.Lines = make([]domain.EntryLine, 0, len(model.Lines))
		for _, line := range model.Lines {
			entry.Lines = append(entry.Lines, toEntryLine(&line))
		}
	}
	return entry
}

func toEntryLineModel(line *domain.EntryLine) EntryLineModel {
	if line == nil {
		return EntryLineModel{}
	}
	return EntryLineModel{
		Model: gorm.Model{
			ID:        uint(line.ID),
			CreatedAt: line.CreatedAt,
			UpdatedAt: line.UpdatedAt,
		},
		EntryID:     line.EntryID,
		AccountID:   line.AccountID,
		SubjectCode: line.SubjectCode,
		Direction:   line.Direction,
		Amount:      line.Amount.Mul(decimal.NewFromInt(100)).IntPart(),
	}
}

func toEntryLine(model *EntryLineModel) domain.EntryLine {
	if model == nil {
		return domain.EntryLine{}
	}
	return domain.EntryLine{
		ID:          uint64(model.ID),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		EntryID:     model.EntryID,
		AccountID:   model.AccountID,
		SubjectCode: model.SubjectCode,
		Direction:   model.Direction,
		Amount:      decimal.NewFromInt(model.Amount).Div(decimal.NewFromInt(100)),
	}
}
