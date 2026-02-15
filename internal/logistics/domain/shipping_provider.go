// 生成摘要：
// - 从 internal/shipping 服务合并而来，补充物流服务商管理、报价查询、运费计算记录等能力
// - 与 shipping_template.go 互补：shipping_template 负责运费模板规则，shipping_provider 负责物流服务商管理
// - 新增 ShippingProvider（物流服务商）、ShippingQuote（报价查询结果）、
//   ShippingFeeCalculation（运费计算记录）等领域对象
// - 新增 ShippingProviderRepository 仓储接口

package domain

import (
	"context"
	"time"
)

// ShippingProvider 物流服务商实体
// 管理各物流服务商的接入信息、支持能力和覆盖区域
type ShippingProvider struct {
	ID               uint      `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Code             string    `json:"code"`              // 服务商编码（如 SF / YT / ZT）
	Name             string    `json:"name"`              // 服务商名称
	Logo             string    `json:"logo"`              // Logo URL
	Description      string    `json:"description"`       // 描述
	Enabled          bool      `json:"enabled"`           // 是否启用
	Priority         int       `json:"priority"`          // 优先级（越小越高）
	APIEndpoint      string    `json:"api_endpoint"`      // API 接入地址
	APIKey           string    `json:"api_key"`           // API Key
	SecretKey        string    `json:"secret_key"`        // Secret Key
	SupportCOD       bool      `json:"support_cod"`       // 是否支持货到付款
	SupportInsurance bool      `json:"support_insurance"` // 是否支持保价
	SupportPickup    bool      `json:"support_pickup"`    // 是否支持上门取件
	TrackingURL      string    `json:"tracking_url"`      // 物流跟踪 URL 模板
	Coverage         []string  `json:"coverage"`          // 覆盖区域编码列表
	RateTable        string    `json:"rate_table"`        // 费率表（JSON 格式）
}

// ShippingQuote 物流报价结果
// 表示某一物流服务商对特定运输请求的报价
type ShippingQuote struct {
	ProviderID        uint   `json:"provider_id"`
	ProviderCode      string `json:"provider_code"`
	ProviderName      string `json:"provider_name"`
	ServiceType       string `json:"service_type"`       // 服务类型（标准/加急/次日达）
	ServiceName       string `json:"service_name"`       // 服务名称
	EstimatedDays     int    `json:"estimated_days"`     // 预计送达天数
	BaseFee           int64  `json:"base_fee"`           // 基础费用（分）
	TotalFee          int64  `json:"total_fee"`          // 总费用（分）
	InsuranceFee      int64  `json:"insurance_fee"`      // 保价费用（分）
	CODFee            int64  `json:"cod_fee"`            // 货到付款手续费（分）
	Discount          int64  `json:"discount"`           // 折扣金额（分）
	IsRecommended     bool   `json:"is_recommended"`     // 是否推荐
	IsAvailable       bool   `json:"is_available"`       // 是否可用
	UnavailableReason string `json:"unavailable_reason"` // 不可用原因
}

// ShippingFeeCalculation 运费计算记录
// 持久化每次运费计算的详细结果，用于对账和审计
type ShippingFeeCalculation struct {
	ID                  uint      `json:"id"`
	CreatedAt           time.Time `json:"created_at"`
	OrderID             uint64    `json:"order_id"`
	MerchantID          uint64    `json:"merchant_id"`
	TemplateID          uint      `json:"template_id"`
	DestinationCode     string    `json:"destination_code"`
	Weight              int64     `json:"weight"`                // 重量（克）
	Volume              int64     `json:"volume"`                // 体积（立方厘米）
	Quantity            int32     `json:"quantity"`              // 数量
	Subtotal            int64     `json:"subtotal"`              // 商品小计（分）
	BaseFee             int64     `json:"base_fee"`              // 基础运费（分）
	AdditionalFee       int64     `json:"additional_fee"`        // 续费（分）
	DiscountFee         int64     `json:"discount_fee"`          // 优惠减免（分）
	InsuranceFee        int64     `json:"insurance_fee"`         // 保价费用（分）
	CODFee              int64     `json:"cod_fee"`               // 货到付款手续费（分）
	TotalFee            int64     `json:"total_fee"`             // 最终运费（分）
	FreeShippingApplied bool      `json:"free_shipping_applied"` // 是否命中包邮规则
	FreeShippingRuleID  uint      `json:"free_shipping_rule_id"` // 命中的包邮规则 ID
}

// NewShippingProvider 创建新的物流服务商
func NewShippingProvider(code, name string) *ShippingProvider {
	return &ShippingProvider{
		Code:             code,
		Name:             name,
		Enabled:          true,
		Coverage:         make([]string, 0),
		SupportCOD:       false,
		SupportInsurance: false,
		SupportPickup:    false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// SetAPI 设置服务商 API 接入信息
func (p *ShippingProvider) SetAPI(endpoint, apiKey, secretKey string) {
	p.APIEndpoint = endpoint
	p.APIKey = apiKey
	p.SecretKey = secretKey
	p.UpdatedAt = time.Now()
}

// SetCoverage 设置服务商覆盖区域
func (p *ShippingProvider) SetCoverage(regions []string) {
	p.Coverage = regions
	p.UpdatedAt = time.Now()
}

// Enable 启用服务商
func (p *ShippingProvider) Enable() {
	p.Enabled = true
	p.UpdatedAt = time.Now()
}

// Disable 禁用服务商
func (p *ShippingProvider) Disable() {
	p.Enabled = false
	p.UpdatedAt = time.Now()
}

// ShippingProviderRepository 物流服务商仓储接口
type ShippingProviderRepository interface {
	// Save 保存服务商
	Save(ctx context.Context, provider *ShippingProvider) error
	// FindByID 根据 ID 查找服务商
	FindByID(ctx context.Context, id uint) (*ShippingProvider, error)
	// FindByCode 根据编码查找服务商
	FindByCode(ctx context.Context, code string) (*ShippingProvider, error)
	// FindEnabled 查找所有启用的服务商
	FindEnabled(ctx context.Context) ([]*ShippingProvider, error)
	// Update 更新服务商
	Update(ctx context.Context, provider *ShippingProvider) error
	// Delete 删除服务商
	Delete(ctx context.Context, id uint) error

	// SaveCalculation 保存运费计算记录
	SaveCalculation(ctx context.Context, calc *ShippingFeeCalculation) error
	// FindCalculationByOrderID 根据订单 ID 查找运费计算记录
	FindCalculationByOrderID(ctx context.Context, orderID uint64) (*ShippingFeeCalculation, error)
}
