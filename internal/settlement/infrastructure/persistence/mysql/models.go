package mysql

import (
	"time"

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
	Status           domain.SettlementStatus `gorm:"type:tinyint;not null;default:0;comment:状态"`
	SettledAt        *time.Time              `gorm:"comment:结算时间"`
	ApprovedBy       string                  `gorm:"type:varchar(64);comment:审核人"`
	ApprovedAt       *time.Time              `gorm:"comment:审核时间"`
	FailReason       string                  `gorm:"type:varchar(255);comment:失败原因"`
}

func (SettlementModel) TableName() string {
	return "settlements"
}

// SettlementDetailModel 结算明细写模型（持久化专用）。
type SettlementDetailModel struct {
	gorm.Model
	SettlementID     uint64 `gorm:"index;not null;comment:结算单ID"`
	OrderID          uint64 `gorm:"index;not null;comment:订单ID"`
	OrderNo          string `gorm:"type:varchar(64);not null;comment:订单号"`
	OrderAmount      uint64 `gorm:"not null;comment:订单金额(分)"`
	PlatformFee      uint64 `gorm:"not null;comment:平台手续费(分)"`
	LogisticsFee     int64  `gorm:"not null;default:0;comment:物流费(分)"`
	ReturnFee        int64  `gorm:"not null;default:0;comment:退货费(分)"`
	OtherFee         int64  `gorm:"not null;default:0;comment:其他费用(分)"`
	SettlementAmount uint64 `gorm:"not null;comment:结算金额(分)"`
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
	EntryID     uint64           `gorm:"index;not null;comment:关联凭证ID"`
	AccountID   uint64           `gorm:"index;not null;comment:关联账户ID"`
	SubjectCode string           `gorm:"type:varchar(32);not null;comment:冗余科目代码"`
	Direction   domain.Direction `gorm:"type:tinyint;not null;comment:借贷方向(1:借, -1:贷)"`
	Amount      int64            `gorm:"not null;comment:发生金额(分)"`
}

func (EntryLineModel) TableName() string {
	return "entry_lines"
}

func toSettlementModel(settlement *domain.Settlement) *SettlementModel {
	if settlement == nil {
		return nil
	}
	return &SettlementModel{
		Model: gorm.Model{
			ID:        settlement.ID,
			CreatedAt: settlement.CreatedAt,
			UpdatedAt: settlement.UpdatedAt,
		},
		SettlementNo:     settlement.SettlementNo,
		MerchantID:       settlement.MerchantID,
		Cycle:            settlement.Cycle,
		StartDate:        settlement.StartDate,
		EndDate:          settlement.EndDate,
		OrderCount:       settlement.OrderCount,
		TotalAmount:      settlement.TotalAmount,
		PlatformFee:      settlement.PlatformFee,
		CommissionAmount: settlement.CommissionAmount,
		RebateAmount:     settlement.RebateAmount,
		OtherFees:        settlement.OtherFees,
		SettlementAmount: settlement.SettlementAmount,
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
		ID:               model.ID,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		SettlementNo:     model.SettlementNo,
		MerchantID:       model.MerchantID,
		Cycle:            model.Cycle,
		StartDate:        model.StartDate,
		EndDate:          model.EndDate,
		OrderCount:       model.OrderCount,
		TotalAmount:      model.TotalAmount,
		PlatformFee:      model.PlatformFee,
		CommissionAmount: model.CommissionAmount,
		RebateAmount:     model.RebateAmount,
		OtherFees:        model.OtherFees,
		SettlementAmount: model.SettlementAmount,
		Status:           model.Status,
		SettledAt:        model.SettledAt,
		ApprovedBy:       model.ApprovedBy,
		ApprovedAt:       model.ApprovedAt,
		FailReason:       model.FailReason,
	}
}

func toSettlementDetailModel(detail *domain.SettlementDetail) *SettlementDetailModel {
	if detail == nil {
		return nil
	}
	return &SettlementDetailModel{
		Model: gorm.Model{
			ID:        detail.ID,
			CreatedAt: detail.CreatedAt,
			UpdatedAt: detail.UpdatedAt,
		},
		SettlementID:     detail.SettlementID,
		OrderID:          detail.OrderID,
		OrderNo:          detail.OrderNo,
		OrderAmount:      detail.OrderAmount,
		PlatformFee:      detail.PlatformFee,
		LogisticsFee:     detail.LogisticsFee,
		ReturnFee:        detail.ReturnFee,
		OtherFee:         detail.OtherFee,
		SettlementAmount: detail.SettlementAmount,
	}
}

func toSettlementDetail(model *SettlementDetailModel) *domain.SettlementDetail {
	if model == nil {
		return nil
	}
	return &domain.SettlementDetail{
		ID:               model.ID,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		SettlementID:     model.SettlementID,
		OrderID:          model.OrderID,
		OrderNo:          model.OrderNo,
		OrderAmount:      model.OrderAmount,
		PlatformFee:      model.PlatformFee,
		LogisticsFee:     model.LogisticsFee,
		ReturnFee:        model.ReturnFee,
		OtherFee:         model.OtherFee,
		SettlementAmount: model.SettlementAmount,
	}
}

func toMerchantAccountModel(account *domain.MerchantAccount) *MerchantAccountModel {
	if account == nil {
		return nil
	}
	return &MerchantAccountModel{
		Model: gorm.Model{
			ID:        account.ID,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		},
		MerchantID:    account.MerchantID,
		Balance:       account.Balance,
		FrozenBalance: account.FrozenBalance,
		TotalIncome:   account.TotalIncome,
		TotalWithdraw: account.TotalWithdraw,
		FeeRate:       account.FeeRate,
	}
}

func toMerchantAccount(model *MerchantAccountModel) *domain.MerchantAccount {
	if model == nil {
		return nil
	}
	return &domain.MerchantAccount{
		ID:            model.ID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		MerchantID:    model.MerchantID,
		Balance:       model.Balance,
		FrozenBalance: model.FrozenBalance,
		TotalIncome:   model.TotalIncome,
		TotalWithdraw: model.TotalWithdraw,
		FeeRate:       model.FeeRate,
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
			ID:        account.ID,
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		},
		SubjectCode: account.SubjectCode,
		EntityID:    account.EntityID,
		Balance:     account.Balance,
		Currency:    account.Currency,
		Version:     account.Version,
	}
}

func toAccount(model *AccountModel) *domain.Account {
	if model == nil {
		return nil
	}
	return &domain.Account{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		SubjectCode: model.SubjectCode,
		EntityID:    model.EntityID,
		Balance:     model.Balance,
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
			ID:        entry.ID,
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
		ID:            model.ID,
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
			ID:        line.ID,
			CreatedAt: line.CreatedAt,
			UpdatedAt: line.UpdatedAt,
		},
		EntryID:     line.EntryID,
		AccountID:   line.AccountID,
		SubjectCode: line.SubjectCode,
		Direction:   line.Direction,
		Amount:      line.Amount,
	}
}

func toEntryLine(model *EntryLineModel) domain.EntryLine {
	if model == nil {
		return domain.EntryLine{}
	}
	return domain.EntryLine{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		EntryID:     model.EntryID,
		AccountID:   model.AccountID,
		SubjectCode: model.SubjectCode,
		Direction:   model.Direction,
		Amount:      model.Amount,
	}
}
