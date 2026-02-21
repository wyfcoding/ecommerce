// Package mysql 供应链金融服务 MySQL 仓储实现
// 生成摘要：
// 1) 实现 SupplyChainFinance 领域的所有仓储接口
// 2) 使用 GORM 操作 MySQL，支持事务和复杂的关联查询
package mysql

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/supplychainfinance/domain"
	"github.com/wyfcoding/pkg/database"
	"gorm.io/gorm"
)

// FinanceApplicationModel 融资申请数据库模型
type FinanceApplicationModel struct {
	ID              string    `gorm:"primaryKey;type:varchar(64)"`
	ApplicantID     string    `gorm:"index;type:varchar(64);not null"`
	ApplicantName   string    `gorm:"type:varchar(128);not null"`
	ApplicantType   string    `gorm:"type:varchar(32);not null"`
	FinanceType     string    `gorm:"type:varchar(32);not null"`
	RequestedAmount float64   `gorm:"type:decimal(20,2);not null"`
	ApprovedAmount  float64   `gorm:"type:decimal(20,2);default:0"`
	Currency        string    `gorm:"type:varchar(10);not null"`
	Purpose         string    `gorm:"type:text"`
	TermDays        int       `gorm:"not null"`
	Status          string    `gorm:"index;type:varchar(32);not null"`
	CreditLineID    string    `gorm:"index;type:varchar(64)"`
	RiskLevel       string    `gorm:"type:varchar(32)"`
	RiskScore       float64   `gorm:"type:decimal(5,2)"`
	InterestRate    float64   `gorm:"type:decimal(10,4)"`
	FeeAmount       float64   `gorm:"type:decimal(20,2)"`
	SubmittedAt     time.Time `gorm:"index"`
	ApprovedAt      *time.Time
	RejectedAt      *time.Time
	RejectionReason string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (FinanceApplicationModel) TableName() string { return "scf_finance_applications" }

// CreditLineModel 授信额度数据库模型
type CreditLineModel struct {
	ID              string    `gorm:"primaryKey;type:varchar(64)"`
	OwnerID         string    `gorm:"index;type:varchar(64);not null"`
	OwnerName       string    `gorm:"type:varchar(128);not null"`
	OwnerType       string    `gorm:"type:varchar(32);not null"`
	TotalLimit      float64   `gorm:"type:decimal(20,2);not null"`
	UsedAmount      float64   `gorm:"type:decimal(20,2);default:0"`
	AvailableAmount float64   `gorm:"type:decimal(20,2);not null"`
	Currency        string    `gorm:"type:varchar(10);not null"`
	Status          string    `gorm:"index;type:varchar(32);not null"`
	InterestRate    float64   `gorm:"type:decimal(10,4)"`
	AnnualFee       float64   `gorm:"type:decimal(20,2)"`
	EffectiveFrom   time.Time `gorm:"not null"`
	EffectiveTo     time.Time `gorm:"not null"`
	ReviewFrequency string    `gorm:"type:varchar(32)"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (CreditLineModel) TableName() string { return "scf_credit_lines" }

// InvoiceFinancingModel 发票融资数据库模型
type InvoiceFinancingModel struct {
	ID                string    `gorm:"primaryKey;type:varchar(64)"`
	ApplicationID     string    `gorm:"index;type:varchar(64);not null"`
	BorrowerID        string    `gorm:"index;type:varchar(64);not null"`
	BorrowerName      string    `gorm:"type:varchar(128);not null"`
	InvoiceID         string    `gorm:"index;type:varchar(64);not null"`
	InvoiceAmount     float64   `gorm:"type:decimal(20,2);not null"`
	FinancingAmount   float64   `gorm:"type:decimal(20,2);not null"`
	AdvanceRate       float64   `gorm:"type:decimal(5,4);not null"`
	Currency          string    `gorm:"type:varchar(10);not null"`
	InterestRate      float64   `gorm:"type:decimal(10,4)"`
	OutstandingAmount float64   `gorm:"type:decimal(20,2);not null"`
	RepaidAmount      float64   `gorm:"type:decimal(20,2);default:0"`
	Status            string    `gorm:"index;type:varchar(32);not null"`
	MaturityDate      time.Time `gorm:"index"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
}

func (InvoiceFinancingModel) TableName() string { return "scf_invoice_financings" }

// 仓储实现基类
type baseRepository struct {
	db     *database.DB
	logger *slog.Logger
}

// FinanceApplicationRepositoryImpl 融资申请仓储实现
type FinanceApplicationRepositoryImpl struct {
	baseRepository
}

func NewFinanceApplicationRepository(db *database.DB, logger *slog.Logger) domain.FinanceApplicationRepository {
	return &FinanceApplicationRepositoryImpl{
		baseRepository{db: db, logger: logger.With("module", "scf_application_repo")},
	}
}

// Create 创建融资申请
func (r *FinanceApplicationRepositoryImpl) Create(ctx context.Context, app *domain.FinanceApplication) error {
	model := toApplicationModel(app)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create finance application", "error", err)
		return err
	}
	return nil
}

// Update 更新融资申请
func (r *FinanceApplicationRepositoryImpl) Update(ctx context.Context, app *domain.FinanceApplication) error {
	model := toApplicationModel(app)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update finance application", "error", err)
		return err
	}
	return nil
}

// FindByID 根据ID查找融资申请
func (r *FinanceApplicationRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.FinanceApplication, error) {
	var model FinanceApplicationModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toApplicationDomain(&model), nil
}

// FindByApplicantID 根据申请人ID查找
func (r *FinanceApplicationRepositoryImpl) FindByApplicantID(ctx context.Context, applicantID string) ([]*domain.FinanceApplication, error) {
	var models []FinanceApplicationModel
	if err := r.db.WithContext(ctx).Where("applicant_id = ?", applicantID).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, err
	}

	apps := make([]*domain.FinanceApplication, len(models))
	for i, m := range models {
		apps[i] = toApplicationDomain(&m)
	}
	return apps, nil
}

// FindByStatus 根据状态查找
func (r *FinanceApplicationRepositoryImpl) FindByStatus(ctx context.Context, status domain.FinanceStatus, limit, offset int) ([]*domain.FinanceApplication, int64, error) {
	var models []FinanceApplicationModel
	var total int64

	query := r.db.WithContext(ctx).Model(&FinanceApplicationModel{}).Where("status = ?", string(status))

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	apps := make([]*domain.FinanceApplication, len(models))
	for i, m := range models {
		apps[i] = toApplicationDomain(&m)
	}
	return apps, total, nil
}

// Delete 删除申请
func (r *FinanceApplicationRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&FinanceApplicationModel{}, "id = ?", id).Error
}

// CreditLineRepositoryImpl 授信额度仓储实现
type CreditLineRepositoryImpl struct {
	baseRepository
}

func NewCreditLineRepository(db *database.DB, logger *slog.Logger) domain.CreditLineRepository {
	return &CreditLineRepositoryImpl{
		baseRepository{db: db, logger: logger.With("module", "scf_creditline_repo")},
	}
}

func (r *CreditLineRepositoryImpl) Create(ctx context.Context, cl *domain.CreditLine) error {
	model := toCreditLineModel(cl)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *CreditLineRepositoryImpl) Update(ctx context.Context, cl *domain.CreditLine) error {
	model := toCreditLineModel(cl)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *CreditLineRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.CreditLine, error) {
	var model CreditLineModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toCreditLineDomain(&model), nil
}

func (r *CreditLineRepositoryImpl) FindByOwnerID(ctx context.Context, ownerID string) ([]*domain.CreditLine, error) {
	var models []CreditLineModel
	if err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&models).Error; err != nil {
		return nil, err
	}
	cls := make([]*domain.CreditLine, len(models))
	for i, m := range models {
		cls[i] = toCreditLineDomain(&m)
	}
	return cls, nil
}

func (r *CreditLineRepositoryImpl) FindByStatus(ctx context.Context, status domain.CreditLineStatus, limit, offset int) ([]*domain.CreditLine, int64, error) {
	var models []CreditLineModel
	var total int64
	query := r.db.WithContext(ctx).Model(&CreditLineModel{}).Where("status = ?", string(status))
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	cls := make([]*domain.CreditLine, len(models))
	for i, m := range models {
		cls[i] = toCreditLineDomain(&m)
	}
	return cls, total, nil
}

func (r *CreditLineRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&CreditLineModel{}, "id = ?", id).Error
}

// InvoiceFinancingRepositoryImpl 发票融资仓储实现
type InvoiceFinancingRepositoryImpl struct {
	baseRepository
}

func NewInvoiceFinancingRepository(db *database.DB, logger *slog.Logger) domain.InvoiceFinancingRepository {
	return &InvoiceFinancingRepositoryImpl{
		baseRepository{db: db, logger: logger.With("module", "scf_invoice_repo")},
	}
}

func (r *InvoiceFinancingRepositoryImpl) Create(ctx context.Context, inf *domain.InvoiceFinancing) error {
	model := toInvoiceFinancingModel(inf)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *InvoiceFinancingRepositoryImpl) Update(ctx context.Context, inf *domain.InvoiceFinancing) error {
	model := toInvoiceFinancingModel(inf)
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *InvoiceFinancingRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.InvoiceFinancing, error) {
	var model InvoiceFinancingModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toInvoiceFinancingDomain(&model), nil
}

func (r *InvoiceFinancingRepositoryImpl) FindByBorrowerID(ctx context.Context, borrowerID string) ([]*domain.InvoiceFinancing, error) {
	var models []InvoiceFinancingModel
	if err := r.db.WithContext(ctx).Where("borrower_id = ?", borrowerID).Find(&models).Error; err != nil {
		return nil, err
	}
	infs := make([]*domain.InvoiceFinancing, len(models))
	for i, m := range models {
		infs[i] = toInvoiceFinancingDomain(&m)
	}
	return infs, nil
}

func (r *InvoiceFinancingRepositoryImpl) FindByInvoiceID(ctx context.Context, invoiceID string) (*domain.InvoiceFinancing, error) {
	var model InvoiceFinancingModel
	if err := r.db.WithContext(ctx).First(&model, "invoice_id = ?", invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toInvoiceFinancingDomain(&model), nil
}

func (r *InvoiceFinancingRepositoryImpl) FindByStatus(ctx context.Context, status domain.FinanceStatus, limit, offset int) ([]*domain.InvoiceFinancing, int64, error) {
	var models []InvoiceFinancingModel
	var total int64
	query := r.db.WithContext(ctx).Model(&InvoiceFinancingModel{}).Where("status = ?", string(status))
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	infs := make([]*domain.InvoiceFinancing, len(models))
	for i, m := range models {
		infs[i] = toInvoiceFinancingDomain(&m)
	}
	return infs, total, nil
}

func (r *InvoiceFinancingRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&InvoiceFinancingModel{}, "id = ?", id).Error
}

// 辅助转换函数

func toApplicationModel(d *domain.FinanceApplication) *FinanceApplicationModel {
	reqAmt, _ := d.RequestedAmount.Float64()
	appAmt, _ := d.ApprovedAmount.Float64()
	riskScore, _ := d.RiskScore.Float64()
	intRate, _ := d.InterestRate.Float64()
	feeAmt, _ := d.FeeAmount.Float64()

	return &FinanceApplicationModel{
		ID:              d.ID,
		ApplicantID:     d.ApplicantID,
		ApplicantName:   d.ApplicantName,
		ApplicantType:   d.ApplicantType,
		FinanceType:     string(d.FinanceType),
		RequestedAmount: reqAmt,
		ApprovedAmount:  appAmt,
		Currency:        d.Currency,
		Purpose:         d.Purpose,
		TermDays:        d.TermDays,
		Status:          string(d.Status),
		CreditLineID:    d.CreditLineID,
		RiskLevel:       string(d.RiskLevel),
		RiskScore:       riskScore,
		InterestRate:    intRate,
		FeeAmount:       feeAmt,
		SubmittedAt:     d.SubmittedAt,
		ApprovedAt:      d.ApprovedAt,
		RejectedAt:      d.RejectedAt,
		RejectionReason: d.RejectionReason,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func toApplicationDomain(m *FinanceApplicationModel) *domain.FinanceApplication {
	return &domain.FinanceApplication{
		ID:              m.ID,
		ApplicantID:     m.ApplicantID,
		ApplicantName:   m.ApplicantName,
		ApplicantType:   m.ApplicantType,
		FinanceType:     domain.FinanceType(m.FinanceType),
		RequestedAmount: decimal.NewFromFloat(m.RequestedAmount),
		ApprovedAmount:  decimal.NewFromFloat(m.ApprovedAmount),
		Currency:        m.Currency,
		Purpose:         m.Purpose,
		TermDays:        m.TermDays,
		Status:          domain.FinanceStatus(m.Status),
		CreditLineID:    m.CreditLineID,
		RiskLevel:       domain.RiskLevel(m.RiskLevel),
		RiskScore:       decimal.NewFromFloat(m.RiskScore),
		InterestRate:    decimal.NewFromFloat(m.InterestRate),
		FeeAmount:       decimal.NewFromFloat(m.FeeAmount),
		SubmittedAt:     m.SubmittedAt,
		ApprovedAt:      m.ApprovedAt,
		RejectedAt:      m.RejectedAt,
		RejectionReason: m.RejectionReason,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func toCreditLineModel(d *domain.CreditLine) *CreditLineModel {
	total, _ := d.TotalLimit.Float64()
	used, _ := d.UsedAmount.Float64()
	avail, _ := d.AvailableAmount.Float64()
	rate, _ := d.InterestRate.Float64()
	fee, _ := d.AnnualFee.Float64()

	return &CreditLineModel{
		ID:              d.ID,
		OwnerID:         d.OwnerID,
		OwnerName:       d.OwnerName,
		OwnerType:       d.OwnerType,
		TotalLimit:      total,
		UsedAmount:      used,
		AvailableAmount: avail,
		Currency:        d.Currency,
		Status:          string(d.Status),
		InterestRate:    rate,
		AnnualFee:       fee,
		EffectiveFrom:   d.EffectiveFrom,
		EffectiveTo:     d.EffectiveTo,
		ReviewFrequency: d.ReviewFrequency,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

func toCreditLineDomain(m *CreditLineModel) *domain.CreditLine {
	return &domain.CreditLine{
		ID:              m.ID,
		OwnerID:         m.OwnerID,
		OwnerName:       m.OwnerName,
		OwnerType:       m.OwnerType,
		TotalLimit:      decimal.NewFromFloat(m.TotalLimit),
		UsedAmount:      decimal.NewFromFloat(m.UsedAmount),
		AvailableAmount: decimal.NewFromFloat(m.AvailableAmount),
		Currency:        m.Currency,
		Status:          domain.CreditLineStatus(m.Status),
		InterestRate:    decimal.NewFromFloat(m.InterestRate),
		AnnualFee:       decimal.NewFromFloat(m.AnnualFee),
		EffectiveFrom:   m.EffectiveFrom,
		EffectiveTo:     m.EffectiveTo,
		ReviewFrequency: m.ReviewFrequency,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func toInvoiceFinancingModel(d *domain.InvoiceFinancing) *InvoiceFinancingModel {
	invAmt, _ := d.InvoiceAmount.Float64()
	finAmt, _ := d.FinancingAmount.Float64()
	advRate, _ := d.AdvanceRate.Float64()
	intRate, _ := d.InterestRate.Float64()
	outAmt, _ := d.OutstandingAmount.Float64()
	repAmt, _ := d.RepaidAmount.Float64()

	return &InvoiceFinancingModel{
		ID:                d.ID,
		ApplicationID:     d.ApplicationID,
		BorrowerID:        d.BorrowerID,
		BorrowerName:      d.BorrowerName,
		InvoiceID:         d.InvoiceID,
		InvoiceAmount:     invAmt,
		FinancingAmount:   finAmt,
		AdvanceRate:       advRate,
		Currency:          d.Currency,
		InterestRate:      intRate,
		OutstandingAmount: outAmt,
		RepaidAmount:      repAmt,
		Status:            string(d.Status),
		MaturityDate:      d.MaturityDate,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

func toInvoiceFinancingDomain(m *InvoiceFinancingModel) *domain.InvoiceFinancing {
	return &domain.InvoiceFinancing{
		ID:                m.ID,
		ApplicationID:     m.ApplicationID,
		BorrowerID:        m.BorrowerID,
		BorrowerName:      m.BorrowerName,
		InvoiceID:         m.InvoiceID,
		InvoiceAmount:     decimal.NewFromFloat(m.InvoiceAmount),
		FinancingAmount:   decimal.NewFromFloat(m.FinancingAmount),
		AdvanceRate:       decimal.NewFromFloat(m.AdvanceRate),
		Currency:          m.Currency,
		InterestRate:      decimal.NewFromFloat(m.InterestRate),
		OutstandingAmount: decimal.NewFromFloat(m.OutstandingAmount),
		RepaidAmount:      decimal.NewFromFloat(m.RepaidAmount),
		Status:            domain.FinanceStatus(m.Status),
		MaturityDate:      m.MaturityDate,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}
