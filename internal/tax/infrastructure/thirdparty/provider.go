// Package thirdparty 第三方税务服务集成
// 变更说明：集成Avalara、Vertex等第三方税务服务，提供统一的税务计算接口
package thirdparty

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wyfcoding/ecommerce/internal/tax/domain"
)

// ThirdPartyTaxProvider 第三方税务服务提供商接口
type ThirdPartyTaxProvider interface {
	// CalculateTax 计算税费
	CalculateTax(ctx context.Context, request *TaxCalculationRequest) (*TaxCalculationResponse, error)
	// ValidateAddress 验证地址
	ValidateAddress(ctx context.Context, address *Address) (*AddressValidationResult, error)
	// GetTaxRates 获取税率
	GetTaxRates(ctx context.Context, country, region string) ([]*TaxRateInfo, error)
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
	// ProviderName 提供商名称
	ProviderName() string
}

// TaxCalculationRequest 税务计算请求
type TaxCalculationRequest struct {
	TransactionDate time.Time           `json:"transaction_date"`
	CustomerCode    string              `json:"customer_code"`
	EntityUseCode   string              `json:"entity_use_code"` // 实体使用代码（免税原因）
	DocumentType    string              `json:"document_type"`   // SalesOrder, SalesInvoice等
	Addresses       map[string]*Address `json:"addresses"`       // ShipFrom, ShipTo, PointOfOrderOrigin等
	Lines           []*LineItem         `json:"lines"`
	Discount        float64             `json:"discount"`
	PurchaseOrderNo string              `json:"purchase_order_no"`
	ReferenceCode   string              `json:"reference_code"`
	Commit          bool                `json:"commit"` // 是否提交交易
}

// Address 地址信息
type Address struct {
	AddressCode string  `json:"address_code"`
	Line1       string  `json:"line1"`
	Line2       string  `json:"line2"`
	Line3       string  `json:"line3"`
	City        string  `json:"city"`
	Region      string  `json:"region"`  // 州/省
	Country     string  `json:"country"` // ISO 3166-1 alpha-2
	PostalCode  string  `json:"postal_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// LineItem 行项目
type LineItem struct {
	LineNo             string   `json:"line_no"`
	ItemCode           string   `json:"item_code"`
	ItemDescription    string   `json:"item_description"`
	TaxCode            string   `json:"tax_code"` // 税务代码（Avalara/Vertex特定）
	Quantity           float64  `json:"quantity"`
	Amount             float64  `json:"amount"` // 行金额
	Discounted         bool     `json:"discounted"`
	OriginAddress      *Address `json:"origin_address"`
	DestinationAddress *Address `json:"destination_address"`
	HSCode             string   `json:"hs_code"` // 海关编码
	OriginCountry      string   `json:"origin_country"`
}

// TaxCalculationResponse 税务计算响应
type TaxCalculationResponse struct {
	DocumentCode       string        `json:"document_code"`
	TotalAmount        float64       `json:"total_amount"`
	TotalTax           float64       `json:"total_tax"`
	TotalTaxable       float64       `json:"total_taxable"`
	TotalExempt        float64       `json:"total_exempt"`
	TotalDiscount      float64       `json:"total_discount"`
	TaxLines           []*TaxLine    `json:"tax_lines"`
	TaxAddresses       []*TaxAddress `json:"tax_addresses"`
	Status             string        `json:"status"` // Saved, Posted, Committed
	ProviderName       string        `json:"provider_name"`
	ProviderResponseID string        `json:"provider_response_id"`
}

// TaxLine 税行明细
type TaxLine struct {
	LineNo           string       `json:"line_no"`
	TaxCode          string       `json:"tax_code"`
	TaxName          string       `json:"tax_name"`
	Rate             float64      `json:"rate"`
	Taxable          float64      `json:"taxable"`
	Tax              float64      `json:"tax"`
	Exempt           float64      `json:"exempt"`
	JurisdictionName string       `json:"jurisdiction_name"` // 税收管辖区
	JurisdictionType string       `json:"jurisdiction_type"` // Country, State, County, City
	TaxDetails       []*TaxDetail `json:"tax_details"`
}

// TaxDetail 税务明细
type TaxDetail struct {
	JurisdictionName string  `json:"jurisdiction_name"`
	TaxName          string  `json:"tax_name"`
	Rate             float64 `json:"rate"`
	Tax              float64 `json:"tax"`
	TaxType          string  `json:"tax_type"` // Sales, Use, VAT等
}

// TaxAddress 税务地址
type TaxAddress struct {
	AddressCode string  `json:"address_code"`
	Line1       string  `json:"line1"`
	Line2       string  `json:"line2"`
	Line3       string  `json:"line3"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Country     string  `json:"country"`
	PostalCode  string  `json:"postal_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	TaxRegionID int64   `json:"tax_region_id"`
	JurisCode   string  `json:"juris_code"`
}

// AddressValidationResult 地址验证结果
type AddressValidationResult struct {
	IsValid           bool     `json:"is_valid"`
	NormalizedAddress *Address `json:"normalized_address"`
	Messages          []string `json:"messages"`
	FiasCode          string   `json:"fias_code"` // 俄罗斯地址编码
	ResolutionQuality string   `json:"resolution_quality"`
}

// TaxRateInfo 税率信息
type TaxRateInfo struct {
	Country          string  `json:"country"`
	Region           string  `json:"region"`
	JurisdictionName string  `json:"jurisdiction_name"`
	TaxType          string  `json:"tax_type"`
	Rate             float64 `json:"rate"`
	EffectiveDate    string  `json:"effective_date"`
	EndDate          string  `json:"end_date"`
}

// ProviderConfig 第三方服务配置
type ProviderConfig struct {
	ProviderName string
	APIBaseURL   string
	APIKey       string
	AccountID    string
	CompanyCode  string
	LicenseKey   string
	Timeout      time.Duration
	MaxRetries   int
}

// ProviderFactory 提供商工厂
type ProviderFactory struct {
	configs map[string]*ProviderConfig
}

func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		configs: make(map[string]*ProviderConfig),
	}
}

// RegisterConfig 注册提供商配置
func (f *ProviderFactory) RegisterConfig(name string, config *ProviderConfig) {
	f.configs[name] = config
}

// CreateProvider 创建税务服务提供商
func (f *ProviderFactory) CreateProvider(name string) (ThirdPartyTaxProvider, error) {
	config, exists := f.configs[name]
	if !exists {
		return nil, fmt.Errorf("provider config not found: %s", name)
	}

	switch name {
	case "avalara":
		return NewAvalaraProvider(config), nil
	case "vertex":
		return NewVertexProvider(config), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}

// FallbackProvider 带降级机制的税务服务提供商
type FallbackProvider struct {
	primary   ThirdPartyTaxProvider
	fallbacks []ThirdPartyTaxProvider
	localCalc *domain.UniversalTaxCalculator
}

func NewFallbackProvider(primary ThirdPartyTaxProvider, fallbacks []ThirdPartyTaxProvider, localCalc *domain.UniversalTaxCalculator) *FallbackProvider {
	return &FallbackProvider{
		primary:   primary,
		fallbacks: fallbacks,
		localCalc: localCalc,
	}
}

// CalculateTax 带降级的税务计算
func (f *FallbackProvider) CalculateTax(ctx context.Context, request *TaxCalculationRequest) (*TaxCalculationResponse, error) {
	// 尝试主提供商
	if f.primary != nil {
		resp, err := f.primary.CalculateTax(ctx, request)
		if err == nil {
			return resp, nil
		}
	}

	// 尝试备用提供商
	for _, fallback := range f.fallbacks {
		resp, err := fallback.CalculateTax(ctx, request)
		if err == nil {
			return resp, nil
		}
	}

	// 使用本地计算作为最终降级
	return f.calculateWithLocalFallback(ctx, request)
}

// calculateWithLocalFallback 使用本地计算作为降级方案
func (f *FallbackProvider) calculateWithLocalFallback(ctx context.Context, request *TaxCalculationRequest) (*TaxCalculationResponse, error) {
	// 提取目的地地址
	var destAddress *Address
	if addr, ok := request.Addresses["ShipTo"]; ok {
		destAddress = addr
	} else if len(request.Addresses) > 0 {
		for _, addr := range request.Addresses {
			destAddress = addr
			break
		}
	}

	if destAddress == nil {
		return nil, fmt.Errorf("no destination address found")
	}

	// 计算总税额
	var totalTax int64
	taxLines := make([]*TaxLine, 0)

	for _, line := range request.Lines {
		amount := int64(line.Amount * 100) // 转换为分
		result, err := f.localCalc.CalculateOrderTax(ctx, destAddress.Country, destAddress.Region, line.ItemCode, amount)
		if err != nil {
			continue
		}

		totalTax += result.TotalTaxAmount

		for _, detail := range result.Items {
			taxLines = append(taxLines, &TaxLine{
				LineNo:           line.LineNo,
				TaxName:          detail.RuleName,
				Rate:             detail.Rate,
				Taxable:          float64(detail.BaseAmount) / 100,
				Tax:              float64(detail.Amount) / 100,
				JurisdictionName: destAddress.Region,
				JurisdictionType: "State",
			})
		}
	}

	return &TaxCalculationResponse{
		TotalAmount:        request.Lines[0].Amount,
		TotalTax:           float64(totalTax) / 100,
		TotalTaxable:       request.Lines[0].Amount,
		TaxLines:           taxLines,
		Status:             "Calculated",
		ProviderName:       "local_fallback",
		ProviderResponseID: "",
	}, nil
}

// ValidateAddress 地址验证
func (f *FallbackProvider) ValidateAddress(ctx context.Context, address *Address) (*AddressValidationResult, error) {
	if f.primary != nil {
		result, err := f.primary.ValidateAddress(ctx, address)
		if err == nil {
			return result, nil
		}
	}

	for _, fallback := range f.fallbacks {
		result, err := fallback.ValidateAddress(ctx, address)
		if err == nil {
			return result, nil
		}
	}

	// 本地简单验证：检查必填字段
	isValid := address.Line1 != "" && address.City != "" && address.Country != ""
	return &AddressValidationResult{
		IsValid:           isValid,
		NormalizedAddress: address,
		Messages:          []string{"Local validation only"},
	}, nil
}

// GetTaxRates 获取税率
func (f *FallbackProvider) GetTaxRates(ctx context.Context, country, region string) ([]*TaxRateInfo, error) {
	if f.primary != nil {
		rates, err := f.primary.GetTaxRates(ctx, country, region)
		if err == nil {
			return rates, nil
		}
	}

	for _, fallback := range f.fallbacks {
		rates, err := fallback.GetTaxRates(ctx, country, region)
		if err == nil {
			return rates, nil
		}
	}

	return nil, fmt.Errorf("all providers failed to get tax rates")
}

// HealthCheck 健康检查
func (f *FallbackProvider) HealthCheck(ctx context.Context) error {
	if f.primary != nil {
		if err := f.primary.HealthCheck(ctx); err == nil {
			return nil
		}
	}

	for _, fallback := range f.fallbacks {
		if err := fallback.HealthCheck(ctx); err == nil {
			return nil
		}
	}

	return fmt.Errorf("all providers are unhealthy")
}

// ProviderName 提供商名称
func (f *FallbackProvider) ProviderName() string {
	return "fallback_provider"
}

// HTTPClient HTTP客户端封装
type HTTPClient struct {
	client     *http.Client
	baseURL    string
	apiKey     string
	maxRetries int
}

func NewHTTPClient(baseURL, apiKey string, timeout time.Duration, maxRetries int) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:    baseURL,
		apiKey:     apiKey,
		maxRetries: maxRetries,
	}
}

// DoRequest 执行HTTP请求
func (c *HTTPClient) DoRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + path

	var bodyJSON []byte
	var err error
	if body != nil {
		bodyJSON, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		if len(bodyJSON) > 0 {
			req.Body = http.NoBody
			// 实际需要实现Body设置
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if result != nil {
				if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
					return fmt.Errorf("failed to decode response: %w", err)
				}
			}
			return nil
		}

		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
}
