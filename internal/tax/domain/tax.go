// 变更说明：新增税务管理功能，支持全球税率映射、增值税（VAT）、关税（Duty）计算及税务减免。
// 假设：税率根据收货地址（国家+地区）进行匹配，支持百分比及固定值计税。
package domain

import (
	"context"
	"time"
)

// --- 税务类型 ---

// TaxType 税务类型
type TaxType int

const (
	TaxTypeVAT       TaxType = 1 // 增值税 (Value Added Tax)
	TaxTypeSales     TaxType = 2 // 销售税 (Sales Tax)
	TaxTypeDuty      TaxType = 3 // 关税 (Customs Duty)
	TaxTypeExcise    TaxType = 4 // 消费税 (Excise Tax)
	TaxTypeSurcharge TaxType = 5 // 附加税
)

// --- 税务规则 ---

// TaxRule 税务规则聚合根
type TaxRule struct {
	ID          uint64    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	CountryCode string    `json:"country_code"` // 国家代码（ISO 3166-1 alpha-2）
	RegionCode  string    `json:"region_code"`  // 地区代码（省/州）
	TaxType     TaxType   `json:"tax_type"`
	Rate        float64   `json:"rate"`         // 税率（百分比，如 0.13 表示 13%）
	FixedAmount int64     `json:"fixed_amount"` // 固定税额（分）
	Category    string    `json:"category"`     // 商品分类适用（ALL/FOOD/LUXURY等）
	Priority    int       `json:"priority"`     // 优先级
	IsCompound  bool      `json:"is_compound"`  // 是否复合计税（在已有税额基础上再计税）
	IsActive    bool      `json:"is_active"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

// --- 税务计算结果 ---

// TaxCalculationResult 税务计算结果
type TaxCalculationResult struct {
	TotalTaxAmount int64            `json:"total_tax_amount"` // 总税额
	Items          []*TaxDetailItem `json:"items"`            // 税务明细
	Currency       string           `json:"currency"`
}

// TaxDetailItem 税务明细项
type TaxDetailItem struct {
	RuleID     uint64  `json:"rule_id"`
	RuleName   string  `json:"rule_name"`
	TaxType    TaxType `json:"tax_type"`
	BaseAmount int64   `json:"base_amount"` // 计税基数
	Rate       float64 `json:"rate"`
	Amount     int64   `json:"amount"` // 本项税额
}

// --- 税务减免 ---

// TaxExemption 税务减免
type TaxExemption struct {
	ID             uint64    `json:"id"`
	UserID         uint64    `json:"user_id"`
	Reason         string    `json:"reason"`          // 减免原因：慈善、外交、出口、政府项目等
	CertificateID  string    `json:"certificate_id"`  // 免税证明编号
	CertificateImg string    `json:"certificate_img"` // 证明照片
	ValidFrom      time.Time `json:"valid_from"`
	ValidTo        time.Time `json:"valid_to"`
	Status         int       `json:"status"` // 1:有效 2:已过期 3:已注销
}

// --- 税务发票/凭证 ---

// TaxInvoice 税务凭证
type TaxInvoice struct {
	ID           uint64    `json:"id"`
	OrderID      uint64    `json:"order_id"`
	InvoiceNo    string    `json:"invoice_no"`  // 税务发票号
	TaxID        string    `json:"tax_id"`      // 纳税人识别号
	TotalNet     int64     `json:"total_net"`   // 总净额
	TotalTax     int64     `json:"total_tax"`   // 总税额
	TotalGross   int64     `json:"total_gross"` // 总毛额
	CalculatedAt time.Time `json:"calculated_at"`
	TaxDetails   string    `json:"tax_details"` // JSON存储明细
}

// --- 领域服务接口 ---

// TaxCalculator 税务计算领域服务
type TaxCalculator interface {
	// CalculateOrderTax 计算订单总税费
	CalculateOrderTax(ctx context.Context, country, region string, category string, amount int64) (*TaxCalculationResult, error)
	// CalculateDuty 计算进口关税
	CalculateDuty(ctx context.Context, originCountry, destCountry string, amount int64) (*TaxCalculationResult, error)
}

// --- 仓储接口 ---

// TaxRepository 税务仓储接口
type TaxRepository interface {
	FindActiveRules(ctx context.Context, country, region string, category string) ([]*TaxRule, error)
	SaveRule(ctx context.Context, rule *TaxRule) error

	FindExemption(ctx context.Context, userID uint64) (*TaxExemption, error)
	SaveExemption(ctx context.Context, exemption *TaxExemption) error

	SaveInvoice(ctx context.Context, invoice *TaxInvoice) error
	FindByOrder(ctx context.Context, orderID uint64) (*TaxInvoice, error)
}

// --- 业务逻辑示例 ---

// CalculateTaxByRule 根据单条规则计算税额
func (r *TaxRule) CalculateTax(amount int64) int64 {
	if !r.IsActive {
		return 0
	}
	now := time.Now()
	if now.Before(r.StartTime) || now.After(r.EndTime) {
		return 0
	}

	var tax int64
	if r.Rate > 0 {
		tax = int64(float64(amount) * r.Rate)
	}
	if r.FixedAmount > 0 {
		tax += r.FixedAmount
	}
	return tax
}

// FormatTaxType 格式化税务类型
func (t TaxType) String() string {
	switch t {
	case TaxTypeVAT:
		return "VAT"
	case TaxTypeSales:
		return "Sales Tax"
	case TaxTypeDuty:
		return "Customs Duty"
	case TaxTypeExcise:
		return "Excise Tax"
	case TaxTypeSurcharge:
		return "Surcharge"
	default:
		return "Unknown"
	}
}
