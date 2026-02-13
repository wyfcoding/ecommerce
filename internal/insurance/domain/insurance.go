// 变更说明：完善保险服务领域模型，增加运费险、延保服务、理赔流程、保单管理等完整功能
package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PolicyType 保单类型
type PolicyType string

const (
	PolicyTypeShippingInsurance PolicyType = "SHIPPING_INSURANCE"
	PolicyTypeReturnInsurance   PolicyType = "RETURN_INSURANCE"
	PolicyTypePriceProtection   PolicyType = "PRICE_PROTECTION"
	PolicyTypeExtendedWarranty  PolicyType = "EXTENDED_WARRANTY"
	PolicyTypeQualityAssurance  PolicyType = "QUALITY_ASSURANCE"
	PolicyTypeDamageInsurance   PolicyType = "DAMAGE_INSURANCE"
)

// PolicyStatus 保单状态
type PolicyStatus string

const (
	PolicyStatusPending   PolicyStatus = "PENDING"
	PolicyStatusActive    PolicyStatus = "ACTIVE"
	PolicyStatusExpired   PolicyStatus = "EXPIRED"
	PolicyStatusCancelled PolicyStatus = "CANCELLED"
	PolicyStatusClaimed   PolicyStatus = "CLAIMED"
	PolicyStatusRefunded  PolicyStatus = "REFUNDED"
)

// ClaimStatus 理赔状态
type ClaimStatus string

const (
	ClaimStatusSubmitted   ClaimStatus = "SUBMITTED"
	ClaimStatusUnderReview ClaimStatus = "UNDER_REVIEW"
	ClaimStatusApproved    ClaimStatus = "APPROVED"
	ClaimStatusRejected    ClaimStatus = "REJECTED"
	ClaimStatusProcessing  ClaimStatus = "PROCESSING"
	ClaimStatusPaid        ClaimStatus = "PAID"
	ClaimStatusClosed      ClaimStatus = "CLOSED"
)

// InsurancePolicy 保单聚合根
type InsurancePolicy struct {
	gorm.Model
	PolicyID       string          `gorm:"column:policy_id;type:varchar(32);uniqueIndex;not null" json:"policy_id"`
	OrderID        string          `gorm:"column:order_id;type:varchar(32);index;not null" json:"order_id"`
	OrderItemID    string          `gorm:"column:order_item_id;type:varchar(32);index" json:"order_item_id"`
	UserID         string          `gorm:"column:user_id;type:varchar(32);index;not null" json:"user_id"`
	MerchantID     string          `gorm:"column:merchant_id;type:varchar(32);index" json:"merchant_id"`
	ProductID      string          `gorm:"column:product_id;type:varchar(32)" json:"product_id"`
	ProductName    string          `gorm:"column:product_name;type:varchar(255)" json:"product_name"`
	Type           PolicyType      `gorm:"column:type;type:varchar(30);not null" json:"type"`
	Premium        decimal.Decimal `gorm:"column:premium;type:decimal(20,4);not null" json:"premium"`
	CoverageAmount decimal.Decimal `gorm:"column:coverage_amount;type:decimal(20,4);not null" json:"coverage_amount"`
	Deductible     decimal.Decimal `gorm:"column:deductible;type:decimal(20,4);default:0" json:"deductible"`
	Status         PolicyStatus    `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	StartTime      time.Time       `gorm:"column:start_time;not null" json:"start_time"`
	EndTime        time.Time       `gorm:"column:end_time;not null" json:"end_time"`
	ClaimCount     int             `gorm:"column:claim_count;default:0" json:"claim_count"`
	MaxClaims      int             `gorm:"column:max_claims;default:1" json:"max_claims"`
	TotalClaimed   decimal.Decimal `gorm:"column:total_claimed;type:decimal(20,4);default:0" json:"total_claimed"`
	Terms          string          `gorm:"column:terms;type:text" json:"terms"`
	CancelReason   string          `gorm:"column:cancel_reason;type:text" json:"cancel_reason"`
	CancelledAt    *time.Time      `gorm:"column:cancelled_at" json:"cancelled_at"`
	ActivatedAt    *time.Time      `gorm:"column:activated_at" json:"activated_at"`
}

// InsuranceClaim 理赔申请
type InsuranceClaim struct {
	gorm.Model
	ClaimID         string          `gorm:"column:claim_id;type:varchar(32);uniqueIndex;not null" json:"claim_id"`
	PolicyID        string          `gorm:"column:policy_id;type:varchar(32);index;not null" json:"policy_id"`
	UserID          string          `gorm:"column:user_id;type:varchar(32);index;not null" json:"user_id"`
	ClaimType       string          `gorm:"column:claim_type;type:varchar(30);not null" json:"claim_type"`
	Reason          string          `gorm:"column:reason;type:text;not null" json:"reason"`
	Description     string          `gorm:"column:description;type:text" json:"description"`
	AmountRequested decimal.Decimal `gorm:"column:amount_requested;type:decimal(20,4);not null" json:"amount_requested"`
	AmountApproved  decimal.Decimal `gorm:"column:amount_approved;type:decimal(20,4)" json:"amount_approved"`
	Status          ClaimStatus     `gorm:"column:status;type:varchar(20);not null;default:'SUBMITTED'" json:"status"`
	RejectReason    string          `gorm:"column:reject_reason;type:text" json:"reject_reason"`
	EvidenceURLs    string          `gorm:"column:evidence_urls;type:text" json:"evidence_urls"`
	ReviewedBy      string          `gorm:"column:reviewed_by;type:varchar(32)" json:"reviewed_by"`
	ReviewedAt      *time.Time      `gorm:"column:reviewed_at" json:"reviewed_at"`
	ApprovedAt      *time.Time      `gorm:"column:approved_at" json:"approved_at"`
	PaidAt          *time.Time      `gorm:"column:paid_at" json:"paid_at"`
	PaymentMethod   string          `gorm:"column:payment_method;type:varchar(30)" json:"payment_method"`
	TransactionID   string          `gorm:"column:transaction_id;type:varchar(64)" json:"transaction_id"`
}

// InsuranceProduct 保险产品定义
type InsuranceProduct struct {
	gorm.Model
	ProductID       string          `gorm:"column:product_id;type:varchar(32);uniqueIndex;not null" json:"product_id"`
	Name            string          `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Type            PolicyType      `gorm:"column:type;type:varchar(30);not null" json:"type"`
	Description     string          `gorm:"column:description;type:text" json:"description"`
	PremiumRate     decimal.Decimal `gorm:"column:premium_rate;type:decimal(10,6);not null" json:"premium_rate"`
	MinPremium      decimal.Decimal `gorm:"column:min_premium;type:decimal(20,4)" json:"min_premium"`
	MaxPremium      decimal.Decimal `gorm:"column:max_premium;type:decimal(20,4)" json:"max_premium"`
	CoverageRate    decimal.Decimal `gorm:"column:coverage_rate;type:decimal(10,6)" json:"coverage_rate"`
	MaxCoverage     decimal.Decimal `gorm:"column:max_coverage;type:decimal(20,4)" json:"max_coverage"`
	Deductible      decimal.Decimal `gorm:"column:deductible;type:decimal(20,4);default:0" json:"deductible"`
	DurationDays    int             `gorm:"column:duration_days;not null" json:"duration_days"`
	MaxClaims       int             `gorm:"column:max_claims;default:1" json:"max_claims"`
	Terms           string          `gorm:"column:terms;type:text" json:"terms"`
	Exclusions      string          `gorm:"column:exclusions;type:text" json:"exclusions"`
	Status          string          `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE'" json:"status"`
}

// ClaimDocument 理赔文档
type ClaimDocument struct {
	gorm.Model
	ClaimID    string    `gorm:"column:claim_id;type:varchar(32);index;not null" json:"claim_id"`
	DocumentType string  `gorm:"column:document_type;type:varchar(30);not null" json:"document_type"`
	FileName   string    `gorm:"column:file_name;type:varchar(255);not null" json:"file_name"`
	FileURL    string    `gorm:"column:file_url;type:varchar(512);not null" json:"file_url"`
	FileSize   int64     `gorm:"column:file_size" json:"file_size"`
	UploadedAt time.Time `gorm:"column:uploaded_at;not null" json:"uploaded_at"`
}

// InsuranceStatistics 保险统计
type InsuranceStatistics struct {
	TotalPolicies      int64           `json:"total_policies"`
	ActivePolicies     int64           `json:"active_policies"`
	TotalPremium       decimal.Decimal `json:"total_premium"`
	TotalClaims        int64           `json:"total_claims"`
	PendingClaims      int64           `json:"pending_claims"`
	ApprovedClaims     int64           `json:"approved_claims"`
	TotalClaimAmount   decimal.Decimal `json:"total_claim_amount"`
	ClaimApprovalRate  decimal.Decimal `json:"claim_approval_rate"`
	AvgProcessingTime  int64           `json:"avg_processing_time"`
}

func (InsurancePolicy) TableName() string   { return "insurance_policies" }
func (InsuranceClaim) TableName() string    { return "insurance_claims" }
func (InsuranceProduct) TableName() string  { return "insurance_products" }
func (ClaimDocument) TableName() string     { return "claim_documents" }

// NewInsurancePolicy 创建新保单
func NewInsurancePolicy(orderID, orderItemID, userID, merchantID, productID, productName string, policyType PolicyType, premium, coverageAmount decimal.Decimal, startTime, endTime time.Time) *InsurancePolicy {
	return &InsurancePolicy{
		PolicyID:       generatePolicyID(),
		OrderID:        orderID,
		OrderItemID:    orderItemID,
		UserID:         userID,
		MerchantID:     merchantID,
		ProductID:      productID,
		ProductName:    productName,
		Type:           policyType,
		Premium:        premium,
		CoverageAmount: coverageAmount,
		Status:         PolicyStatusPending,
		StartTime:      startTime,
		EndTime:        endTime,
		MaxClaims:      1,
		TotalClaimed:   decimal.Zero,
	}
}

// Activate 激活保单
func (p *InsurancePolicy) Activate() error {
	if p.Status != PolicyStatusPending {
		return errors.New("policy is not pending")
	}
	p.Status = PolicyStatusActive
	now := time.Now()
	p.ActivatedAt = &now
	return nil
}

// Cancel 取消保单
func (p *InsurancePolicy) Cancel(reason string) error {
	if p.Status != PolicyStatusPending && p.Status != PolicyStatusActive {
		return errors.New("policy cannot be cancelled")
	}
	p.Status = PolicyStatusCancelled
	p.CancelReason = reason
	now := time.Now()
	p.CancelledAt = &now
	return nil
}

// IsExpired 检查是否过期
func (p *InsurancePolicy) IsExpired() bool {
	return time.Now().After(p.EndTime)
}

// CanClaim 检查是否可以理赔
func (p *InsurancePolicy) CanClaim() bool {
	if p.Status != PolicyStatusActive {
		return false
	}
	if p.IsExpired() {
		return false
	}
	if p.ClaimCount >= p.MaxClaims {
		return false
	}
	return true
}

// NewInsuranceClaim 创建理赔申请
func NewInsuranceClaim(policyID, userID, claimType, reason string, amountRequested decimal.Decimal, evidenceURLs string) *InsuranceClaim {
	return &InsuranceClaim{
		ClaimID:         generateClaimID(),
		PolicyID:        policyID,
		UserID:          userID,
		ClaimType:       claimType,
		Reason:          reason,
		AmountRequested: amountRequested,
		EvidenceURLs:    evidenceURLs,
		Status:          ClaimStatusSubmitted,
	}
}

// Submit 提交理赔
func (c *InsuranceClaim) Submit() error {
	if c.Status != ClaimStatusSubmitted {
		return errors.New("claim is already submitted")
	}
	c.Status = ClaimStatusUnderReview
	return nil
}

// Approve 批准理赔
func (c *InsuranceClaim) Approve(amount decimal.Decimal, reviewerID string) error {
	if c.Status != ClaimStatusUnderReview {
		return errors.New("claim is not under review")
	}
	c.Status = ClaimStatusApproved
	c.AmountApproved = amount
	c.ReviewedBy = reviewerID
	now := time.Now()
	c.ReviewedAt = &now
	c.ApprovedAt = &now
	return nil
}

// Reject 拒绝理赔
func (c *InsuranceClaim) Reject(reason, reviewerID string) error {
	if c.Status != ClaimStatusUnderReview {
		return errors.New("claim is not under review")
	}
	c.Status = ClaimStatusRejected
	c.RejectReason = reason
	c.ReviewedBy = reviewerID
	now := time.Now()
	c.ReviewedAt = &now
	return nil
}

// ProcessPayment 处理支付
func (c *InsuranceClaim) ProcessPayment(transactionID, paymentMethod string) error {
	if c.Status != ClaimStatusApproved {
		return errors.New("claim is not approved")
	}
	c.Status = ClaimStatusProcessing
	c.TransactionID = transactionID
	c.PaymentMethod = paymentMethod
	return nil
}

// MarkPaid 标记已支付
func (c *InsuranceClaim) MarkPaid() error {
	if c.Status != ClaimStatusProcessing {
		return errors.New("claim is not processing")
	}
	c.Status = ClaimStatusPaid
	now := time.Now()
	c.PaidAt = &now
	return nil
}

// Close 关闭理赔
func (c *InsuranceClaim) Close() {
	c.Status = ClaimStatusClosed
}

// NewInsuranceProduct 创建保险产品
func NewInsuranceProduct(name string, policyType PolicyType, premiumRate, coverageRate decimal.Decimal, durationDays int) *InsuranceProduct {
	return &InsuranceProduct{
		ProductID:    generateInsuranceProductID(),
		Name:         name,
		Type:         policyType,
		PremiumRate:  premiumRate,
		CoverageRate: coverageRate,
		DurationDays: durationDays,
		Status:       "ACTIVE",
	}
}

// CalculatePremium 计算保费
func (p *InsuranceProduct) CalculatePremium(orderValue decimal.Decimal) decimal.Decimal {
	premium := orderValue.Mul(p.PremiumRate).Div(decimal.NewFromInt(100))
	if premium.LessThan(p.MinPremium) {
		return p.MinPremium
	}
	if premium.GreaterThan(p.MaxPremium) && !p.MaxPremium.IsZero() {
		return p.MaxPremium
	}
	return premium
}

// CalculateCoverage 计算保额
func (p *InsuranceProduct) CalculateCoverage(orderValue decimal.Decimal) decimal.Decimal {
	coverage := orderValue.Mul(p.CoverageRate).Div(decimal.NewFromInt(100))
	if coverage.GreaterThan(p.MaxCoverage) && !p.MaxCoverage.IsZero() {
		return p.MaxCoverage
	}
	return coverage
}

// 辅助函数
func generatePolicyID() string {
	return fmt.Sprintf("POL%d", time.Now().UnixNano())
}

func generateClaimID() string {
	return fmt.Sprintf("CLM%d", time.Now().UnixNano())
}

func generateInsuranceProductID() string {
	return fmt.Sprintf("IP%d", time.Now().UnixNano())
}

// Repository 仓储接口
type Repository interface {
	SavePolicy(ctx context.Context, policy *InsurancePolicy) error
	GetPolicy(ctx context.Context, policyID string) (*InsurancePolicy, error)
	GetPolicyByOrder(ctx context.Context, orderID string) (*InsurancePolicy, error)
	ListPolicies(ctx context.Context, userID string, status PolicyStatus, offset, limit int) ([]*InsurancePolicy, int64, error)
	
	SaveClaim(ctx context.Context, claim *InsuranceClaim) error
	GetClaim(ctx context.Context, claimID string) (*InsuranceClaim, error)
	ListClaims(ctx context.Context, policyID string, status ClaimStatus, offset, limit int) ([]*InsuranceClaim, int64, error)
	ListPendingClaims(ctx context.Context, limit int) ([]*InsuranceClaim, error)
	
	SaveInsuranceProduct(ctx context.Context, product *InsuranceProduct) error
	GetInsuranceProduct(ctx context.Context, productID string) (*InsuranceProduct, error)
	ListInsuranceProducts(ctx context.Context, policyType PolicyType) ([]*InsuranceProduct, error)
}

// 错误定义
var (
	ErrPolicyNotFound      = errors.New("policy not found")
	ErrClaimNotFound       = errors.New("claim not found")
	ErrPolicyNotActive     = errors.New("policy is not active")
	ErrPolicyExpired       = errors.New("policy has expired")
	ErrClaimLimitExceeded  = errors.New("claim limit exceeded")
	ErrInvalidClaimAmount  = errors.New("invalid claim amount")
	ErrClaimAlreadyProcessed = errors.New("claim already processed")
)
