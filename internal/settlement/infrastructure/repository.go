package infrastructure

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	"gorm.io/gorm"
)

type SettlementPO struct {
	ID               uint64           `gorm:"column:id;primaryKey;autoIncrement"`
	SettlementID     string           `gorm:"column:settlement_id;type:varchar(32);uniqueIndex;not null"`
	MerchantID       uint64           `gorm:"column:merchant_id;index;not null"`
	Cycle            string           `gorm:"column:cycle;type:varchar(20);not null"`
	PeriodStart      time.Time        `gorm:"column:period_start;not null"`
	PeriodEnd        time.Time        `gorm:"column:period_end;not null"`
	OrderCount       int64            `gorm:"column:order_count;default:0"`
	GrossAmount      decimal.Decimal  `gorm:"column:gross_amount;type:decimal(20,4);not null"`
	RefundAmount     decimal.Decimal  `gorm:"column:refund_amount;type:decimal(20,4);not null"`
	PlatformCommission decimal.Decimal `gorm:"column:platform_commission;type:decimal(20,4);not null"`
	PromotionFee     decimal.Decimal  `gorm:"column:promotion_fee;type:decimal(20,4);not null"`
	LogisticsFee     decimal.Decimal  `gorm:"column:logistics_fee;type:decimal(20,4);not null"`
	AdjustmentAmount decimal.Decimal  `gorm:"column:adjustment_amount;type:decimal(20,4);not null"`
	SettlementAmount decimal.Decimal  `gorm:"column:settlement_amount;type:decimal(20,4);not null"`
	Status           string           `gorm:"column:status;type:varchar(20);not null;default:'PENDING'"`
	BankAccountID    uint64           `gorm:"column:bank_account_id"`
	TransactionRef   string           `gorm:"column:transaction_ref;type:varchar(64)"`
	ApprovedBy       uint64           `gorm:"column:approved_by"`
	ApprovedAt       *time.Time       `gorm:"column:approved_at"`
	PaidAt           *time.Time       `gorm:"column:paid_at"`
	FailReason       string           `gorm:"column:fail_reason;type:varchar(255)"`
	Version          int64            `gorm:"column:version;default:0"`
	CreatedAt        time.Time        `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time        `gorm:"column:updated_at;autoUpdateTime"`
}

func (SettlementPO) TableName() string { return "merchant_settlements" }

type SettlementDetailPO struct {
	ID                 uint64          `gorm:"column:id;primaryKey;autoIncrement"`
	SettlementID       string          `gorm:"column:settlement_id;type:varchar(32);index;not null"`
	OrderID            uint64          `gorm:"column:order_id;not null"`
	OrderNo            string          `gorm:"column:order_no;type:varchar(32);not null"`
	OrderAmount        decimal.Decimal `gorm:"column:order_amount;type:decimal(20,4);not null"`
	RefundAmount       decimal.Decimal `gorm:"column:refund_amount;type:decimal(20,4);not null"`
	PlatformCommission decimal.Decimal `gorm:"column:platform_commission;type:decimal(20,4);not null"`
	PromotionFee       decimal.Decimal `gorm:"column:promotion_fee;type:decimal(20,4);not null"`
	LogisticsFee       decimal.Decimal `gorm:"column:logistics_fee;type:decimal(20,4);not null"`
	SettlementAmount   decimal.Decimal `gorm:"column:settlement_amount;type:decimal(20,4);not null"`
	CreatedAt          time.Time       `gorm:"column:created_at;autoCreateTime"`
}

func (SettlementDetailPO) TableName() string { return "merchant_settlement_details" }

type MerchantBankAccountPO struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	MerchantID  uint64    `gorm:"column:merchant_id;index;not null"`
	BankName    string    `gorm:"column:bank_name;type:varchar(64);not null"`
	BankCode    string    `gorm:"column:bank_code;type:varchar(20);not null"`
	AccountName string    `gorm:"column:account_name;type:varchar(64);not null"`
	AccountNo   string    `gorm:"column:account_no;type:varchar(32);not null"`
	BranchName  string    `gorm:"column:branch_name;type:varchar(128)"`
	IsDefault   bool      `gorm:"column:is_default;default:false"`
	Status      string    `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE'"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (MerchantBankAccountPO) TableName() string { return "merchant_bank_accounts" }

type MerchantSettlementConfigPO struct {
	ID                  uint64          `gorm:"column:id;primaryKey;autoIncrement"`
	MerchantID          uint64          `gorm:"column:merchant_id;uniqueIndex;not null"`
	Cycle               string          `gorm:"column:cycle;type:varchar(20);not null"`
	CommissionRate      decimal.Decimal `gorm:"column:commission_rate;type:decimal(10,4);not null"`
	MinSettlementAmount decimal.Decimal `gorm:"column:min_settlement_amount;type:decimal(20,4);not null"`
	AutoApprove         bool            `gorm:"column:auto_approve;default:false"`
	AutoPay             bool            `gorm:"column:auto_pay;default:false"`
	Status              string          `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE'"`
	CreatedAt           time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (MerchantSettlementConfigPO) TableName() string { return "merchant_settlement_configs" }

func toSettlementPO(s *domain.Settlement) *SettlementPO {
	return &SettlementPO{
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
		Version:            s.Version,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}

func toSettlement(po *SettlementPO) *domain.Settlement {
	return &domain.Settlement{
		ID:                 po.ID,
		SettlementID:       po.SettlementID,
		MerchantID:         po.MerchantID,
		Cycle:              domain.SettlementCycle(po.Cycle),
		PeriodStart:        po.PeriodStart,
		PeriodEnd:          po.PeriodEnd,
		OrderCount:         po.OrderCount,
		GrossAmount:        po.GrossAmount,
		RefundAmount:       po.RefundAmount,
		PlatformCommission: po.PlatformCommission,
		PromotionFee:       po.PromotionFee,
		LogisticsFee:       po.LogisticsFee,
		AdjustmentAmount:   po.AdjustmentAmount,
		SettlementAmount:   po.SettlementAmount,
		Status:             domain.SettlementStatus(po.Status),
		BankAccountID:      po.BankAccountID,
		TransactionRef:     po.TransactionRef,
		ApprovedBy:         po.ApprovedBy,
		ApprovedAt:         po.ApprovedAt,
		PaidAt:             po.PaidAt,
		FailReason:         po.FailReason,
		Version:            po.Version,
		CreatedAt:          po.CreatedAt,
		UpdatedAt:          po.UpdatedAt,
	}
}

func toSettlementDetailPO(d *domain.SettlementDetail) *SettlementDetailPO {
	return &SettlementDetailPO{
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

func toSettlementDetail(po *SettlementDetailPO) *domain.SettlementDetail {
	return &domain.SettlementDetail{
		ID:                 po.ID,
		SettlementID:       po.SettlementID,
		OrderID:            po.OrderID,
		OrderNo:            po.OrderNo,
		OrderAmount:        po.OrderAmount,
		RefundAmount:       po.RefundAmount,
		PlatformCommission: po.PlatformCommission,
		PromotionFee:       po.PromotionFee,
		LogisticsFee:       po.LogisticsFee,
		SettlementAmount:   po.SettlementAmount,
		CreatedAt:          po.CreatedAt,
	}
}

func toMerchantBankAccountPO(a *domain.MerchantBankAccount) *MerchantBankAccountPO {
	return &MerchantBankAccountPO{
		ID:          a.ID,
		MerchantID:  a.MerchantID,
		BankName:    a.BankName,
		BankCode:    a.BankCode,
		AccountName: a.AccountName,
		AccountNo:   a.AccountNo,
		BranchName:  a.BranchName,
		IsDefault:   a.IsDefault,
		Status:      string(a.Status),
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func toMerchantBankAccount(po *MerchantBankAccountPO) *domain.MerchantBankAccount {
	return &domain.MerchantBankAccount{
		ID:          po.ID,
		MerchantID:  po.MerchantID,
		BankName:    po.BankName,
		BankCode:    po.BankCode,
		AccountName: po.AccountName,
		AccountNo:   po.AccountNo,
		BranchName:  po.BranchName,
		IsDefault:   po.IsDefault,
		Status:      domain.AccountStatus(po.Status),
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

func toMerchantSettlementConfigPO(c *domain.MerchantSettlementConfig) *MerchantSettlementConfigPO {
	return &MerchantSettlementConfigPO{
		ID:                  c.ID,
		MerchantID:          c.MerchantID,
		Cycle:               string(c.Cycle),
		CommissionRate:      c.CommissionRate,
		MinSettlementAmount: c.MinSettlementAmount,
		AutoApprove:         c.AutoApprove,
		AutoPay:             c.AutoPay,
		Status:              string(c.Status),
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
}

func toMerchantSettlementConfig(po *MerchantSettlementConfigPO) *domain.MerchantSettlementConfig {
	return &domain.MerchantSettlementConfig{
		ID:                  po.ID,
		MerchantID:          po.MerchantID,
		Cycle:               domain.SettlementCycle(po.Cycle),
		CommissionRate:      po.CommissionRate,
		MinSettlementAmount: po.MinSettlementAmount,
		AutoApprove:         po.AutoApprove,
		AutoPay:             po.AutoPay,
		Status:              domain.ConfigStatus(po.Status),
		CreatedAt:           po.CreatedAt,
		UpdatedAt:           po.UpdatedAt,
	}
}

type GormSettlementRepository struct {
	db *gorm.DB
}

func NewGormSettlementRepository(db *gorm.DB) *GormSettlementRepository {
	return &GormSettlementRepository{db: db}
}

func (r *GormSettlementRepository) Save(ctx context.Context, s *domain.Settlement) error {
	po := toSettlementPO(s)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *GormSettlementRepository) Update(ctx context.Context, s *domain.Settlement) error {
	po := toSettlementPO(s)
	return r.db.WithContext(ctx).Model(&SettlementPO{}).Where("settlement_id = ? AND version = ?", s.SettlementID, s.Version).Updates(po).Error
}

func (r *GormSettlementRepository) GetByID(ctx context.Context, id string) (*domain.Settlement, error) {
	var po SettlementPO
	err := r.db.WithContext(ctx).Where("settlement_id = ?", id).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toSettlement(&po), nil
}

func (r *GormSettlementRepository) GetByIDForUpdate(ctx context.Context, id string) (*domain.Settlement, error) {
	var po SettlementPO
	err := r.db.WithContext(ctx).
		Set("gorm:query_option", "FOR UPDATE").
		Where("settlement_id = ?", id).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toSettlement(&po), nil
}

func (r *GormSettlementRepository) ListByMerchant(ctx context.Context, merchantID uint64, status domain.SettlementStatus, page, pageSize int) ([]*domain.Settlement, int64, error) {
	var pos []*SettlementPO
	var total int64

	query := r.db.WithContext(ctx).Model(&SettlementPO{}).Where("merchant_id = ?", merchantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	settlements := make([]*domain.Settlement, len(pos))
	for i, po := range pos {
		settlements[i] = toSettlement(po)
	}

	return settlements, total, nil
}

func (r *GormSettlementRepository) ListByPeriod(ctx context.Context, merchantID uint64, periodStart, periodEnd time.Time) ([]*domain.Settlement, error) {
	var pos []*SettlementPO
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("period_start >= ? AND period_end <= ?", periodStart, periodEnd).
		Find(&pos).Error
	if err != nil {
		return nil, err
	}

	settlements := make([]*domain.Settlement, len(pos))
	for i, po := range pos {
		settlements[i] = toSettlement(po)
	}
	return settlements, nil
}

func (r *GormSettlementRepository) GetByMerchantAndPeriod(ctx context.Context, merchantID uint64, periodStart, periodEnd time.Time) (*domain.Settlement, error) {
	var po SettlementPO
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("period_start = ? AND period_end = ?", periodStart, periodEnd).
		First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toSettlement(&po), nil
}

func (r *GormSettlementRepository) SaveDetail(ctx context.Context, d *domain.SettlementDetail) error {
	po := toSettlementDetailPO(d)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *GormSettlementRepository) SaveDetails(ctx context.Context, details []*domain.SettlementDetail) error {
	pos := make([]*SettlementDetailPO, len(details))
	for i, d := range details {
		pos[i] = toSettlementDetailPO(d)
	}
	return r.db.WithContext(ctx).CreateInBatches(pos, 100).Error
}

func (r *GormSettlementRepository) GetDetailsBySettlementID(ctx context.Context, settlementID string) ([]*domain.SettlementDetail, error) {
	var pos []*SettlementDetailPO
	err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&pos).Error
	if err != nil {
		return nil, err
	}

	details := make([]*domain.SettlementDetail, len(pos))
	for i, po := range pos {
		details[i] = toSettlementDetail(po)
	}
	return details, nil
}

type GormMerchantBankAccountRepository struct {
	db *gorm.DB
}

func NewGormMerchantBankAccountRepository(db *gorm.DB) *GormMerchantBankAccountRepository {
	return &GormMerchantBankAccountRepository{db: db}
}

func (r *GormMerchantBankAccountRepository) Save(ctx context.Context, a *domain.MerchantBankAccount) error {
	po := toMerchantBankAccountPO(a)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *GormMerchantBankAccountRepository) Update(ctx context.Context, a *domain.MerchantBankAccount) error {
	po := toMerchantBankAccountPO(a)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *GormMerchantBankAccountRepository) GetByID(ctx context.Context, id uint64) (*domain.MerchantBankAccount, error) {
	var po MerchantBankAccountPO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toMerchantBankAccount(&po), nil
}

func (r *GormMerchantBankAccountRepository) GetByMerchantID(ctx context.Context, merchantID uint64) ([]*domain.MerchantBankAccount, error) {
	var pos []*MerchantBankAccountPO
	err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Find(&pos).Error
	if err != nil {
		return nil, err
	}

	accounts := make([]*domain.MerchantBankAccount, len(pos))
	for i, po := range pos {
		accounts[i] = toMerchantBankAccount(po)
	}
	return accounts, nil
}

func (r *GormMerchantBankAccountRepository) GetDefaultByMerchantID(ctx context.Context, merchantID uint64) (*domain.MerchantBankAccount, error) {
	var po MerchantBankAccountPO
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND is_default = ? AND status = ?", merchantID, true, "ACTIVE").
		First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toMerchantBankAccount(&po), nil
}

func (r *GormMerchantBankAccountRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&MerchantBankAccountPO{}, id).Error
}

type GormMerchantSettlementConfigRepository struct {
	db *gorm.DB
}

func NewGormMerchantSettlementConfigRepository(db *gorm.DB) *GormMerchantSettlementConfigRepository {
	return &GormMerchantSettlementConfigRepository{db: db}
}

func (r *GormMerchantSettlementConfigRepository) Save(ctx context.Context, c *domain.MerchantSettlementConfig) error {
	po := toMerchantSettlementConfigPO(c)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *GormMerchantSettlementConfigRepository) Update(ctx context.Context, c *domain.MerchantSettlementConfig) error {
	po := toMerchantSettlementConfigPO(c)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *GormMerchantSettlementConfigRepository) GetByMerchantID(ctx context.Context, merchantID uint64) (*domain.MerchantSettlementConfig, error) {
	var po MerchantSettlementConfigPO
	err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toMerchantSettlementConfig(&po), nil
}
