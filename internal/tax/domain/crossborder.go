// Package domain 跨境税务规则配置
// 变更说明：添加跨境贸易税务规则、原产地规则、优惠税率协定支持
package domain

import (
	"context"
	"time"
)

// --- 跨境税务配置 ---

// CrossBorderTaxConfig 跨境税务配置聚合根
type CrossBorderTaxConfig struct {
	ID                     uint64                  `json:"id"`
	CreatedAt              time.Time               `json:"created_at"`
	UpdatedAt              time.Time               `json:"updated_at"`
	Name                   string                  `json:"name"`                    // 配置名称
	OriginCountry          string                  `json:"origin_country"`          // 原产国
	DestinationCountry     string                  `json:"destination_country"`     // 目的国
	TradeType              TradeType               `json:"trade_type"`              // 贸易类型
	TransactionMode        TransactionMode         `json:"transaction_mode"`        // 交易模式
	TaxCollectionMethod    TaxCollectionMethod     `json:"tax_collection_method"`   // 税款征收方式
	ImportTaxRules         []ImportTaxRule         `json:"import_tax_rules"`        // 进口税务规则
	ExportTaxRules         []ExportTaxRule         `json:"export_tax_rules"`        // 出口税务规则
	PreferentialAgreements []PreferentialAgreement `json:"preferential_agreements"` // 优惠协定
	DeMinimisThreshold     int64                   `json:"de_minimis_threshold"`    // 最低征税门槛（分）
	IsActive               bool                    `json:"is_active"`
	EffectiveFrom          time.Time               `json:"effective_from"`
	EffectiveTo            time.Time               `json:"effective_to"`
}

// TradeType 贸易类型
type TradeType int

const (
	TradeTypeB2B     TradeType = 1 // 企业对企业
	TradeTypeB2C     TradeType = 2 // 企业对消费者
	TradeTypeC2C     TradeType = 3 // 消费者对消费者
	TradeTypeImport  TradeType = 4 // 一般贸易进口
	TradeTypeExport  TradeType = 5 // 一般贸易出口
	TradeTypeTransit TradeType = 6 // 过境贸易
)

func (t TradeType) String() string {
	switch t {
	case TradeTypeB2B:
		return "B2B"
	case TradeTypeB2C:
		return "B2C"
	case TradeTypeC2C:
		return "C2C"
	case TradeTypeImport:
		return "Import"
	case TradeTypeExport:
		return "Export"
	case TradeTypeTransit:
		return "Transit"
	default:
		return "Unknown"
	}
}

// TransactionMode 交易模式
type TransactionMode int

const (
	TransactionModeDDP TransactionMode = 1 // 完税后交货
	TransactionModeDDU TransactionMode = 2 // 未完税交货
	TransactionModeDAP TransactionMode = 3 // 目的地交货
	TransactionModeEXW TransactionMode = 4 // 工厂交货
	TransactionModeFOB TransactionMode = 5 // 船上交货
	TransactionModeCIF TransactionMode = 6 // 成本加保险费运费
)

func (t TransactionMode) String() string {
	switch t {
	case TransactionModeDDP:
		return "DDP"
	case TransactionModeDDU:
		return "DDU"
	case TransactionModeDAP:
		return "DAP"
	case TransactionModeEXW:
		return "EXW"
	case TransactionModeFOB:
		return "FOB"
	case TransactionModeCIF:
		return "CIF"
	default:
		return "Unknown"
	}
}

// TaxCollectionMethod 税款征收方式
type TaxCollectionMethod int

const (
	CollectionMethodIOSS    TaxCollectionMethod = 1 // 进口一站式申报（欧盟）
	CollectionMethodOSS     TaxCollectionMethod = 2 // 一站式申报（欧盟）
	CollectionMethodCustoms TaxCollectionMethod = 3 // 海关征收
	CollectionMethodSeller  TaxCollectionMethod = 4 // 卖家代扣代缴
	CollectionMethodBuyer   TaxCollectionMethod = 5 // 买家自行申报
)

func (t TaxCollectionMethod) String() string {
	switch t {
	case CollectionMethodIOSS:
		return "IOSS"
	case CollectionMethodOSS:
		return "OSS"
	case CollectionMethodCustoms:
		return "Customs"
	case CollectionMethodSeller:
		return "Seller"
	case CollectionMethodBuyer:
		return "Buyer"
	default:
		return "Unknown"
	}
}

// ImportTaxRule 进口税务规则
type ImportTaxRule struct {
	ID                    uint64   `json:"id"`
	TaxType               TaxType  `json:"tax_type"`                // 税种
	Rate                  float64  `json:"rate"`                    // 税率
	FixedAmount           int64    `json:"fixed_amount"`            // 固定税额
	ThresholdAmount       int64    `json:"threshold_amount"`        // 起征点
	ExemptCategories      []string `json:"exempt_categories"`       // 免税类别
	ReducedRateCategories []string `json:"reduced_rate_categories"` // 减税类别
}

// ExportTaxRule 出口税务规则
type ExportTaxRule struct {
	ID           uint64   `json:"id"`
	TaxType      TaxType  `json:"tax_type"`
	RefundRate   float64  `json:"refund_rate"`   // 退税率
	Exempt       bool     `json:"exempt"`        // 是否免税出口
	RequiredDocs []string `json:"required_docs"` // 所需单证
}

// PreferentialAgreement 优惠贸易协定
type PreferentialAgreement struct {
	ID                 uint64    `json:"id"`
	Name               string    `json:"name"`           // 协定名称
	AgreementCode      string    `json:"agreement_code"` // 协定代码
	OriginCountry      string    `json:"origin_country"`
	DestinationCountry string    `json:"destination_country"`
	PreferentialRate   float64   `json:"preferential_rate"` // 优惠税率
	RulesOfOrigin      []ROORule `json:"rules_of_origin"`   // 原产地规则
	ValidFrom          time.Time `json:"valid_from"`
	ValidTo            time.Time `json:"valid_to"`
}

// ROORule 原产地规则 (Rule of Origin)
type ROORule struct {
	ID              uint64  `json:"id"`
	RuleType        ROOType `json:"rule_type"` // 规则类型
	Description     string  `json:"description"`
	MinLocalContent float64 `json:"min_local_content"` // 最低本地成分比例
	HSCodePrefix    string  `json:"hs_code_prefix"`    // HS编码前缀
	RequiredProcess string  `json:"required_process"`  // 必要加工工序
}

// ROOType 原产地规则类型
type ROOType int

const (
	ROOTypeWO  ROOType = 1 // 完全获得 (Wholly Obtained)
	ROOTypePE  ROOType = 2 // 完全生产 (Produced Exclusively)
	ROOTypeRVC ROOType = 3 // 区域价值成分 (Regional Value Content)
	ROOTypeCTH ROOType = 4 // 税则归类改变 (Change in Tariff Heading)
	ROOTypeSP  ROOType = 5 // 特定加工工序 (Specific Process)
)

// --- 跨境税务计算 ---

// CrossBorderTaxInput 跨境税务计算输入
type CrossBorderTaxInput struct {
	OriginCountry      string
	DestinationCountry string
	TradeType          TradeType
	TransactionMode    TransactionMode
	HSCode             string
	Category           string
	ProductValue       int64 // 商品价值（CIF，分）
	ShippingCost       int64 // 运费（分）
	InsuranceCost      int64 // 保险费（分）
	Quantity           float64
	Weight             float64
	HasOriginCert      bool   // 是否有原产地证明
	OriginCertNo       string // 原产地证明编号
}

// CrossBorderTaxResult 跨境税务计算结果
type CrossBorderTaxResult struct {
	CustomsValue        int64 // 完税价格
	CustomsDuty         int64 // 关税
	ImportVAT           int64 // 进口增值税
	ImportExcise        int64 // 进口消费税
	OtherTaxes          int64 // 其他税费
	TotalImportTax      int64 // 进口总税费
	TaxCollectionMethod TaxCollectionMethod
	PreferentialApplied bool                   // 是否应用优惠税率
	AppliedAgreement    *PreferentialAgreement // 应用的协定
	RequiredDocuments   []string               // 所需单证
}

// CrossBorderTaxEngine 跨境税务计算引擎
type CrossBorderTaxEngine struct {
	repo CrossBorderTaxRepository
}

func NewCrossBorderTaxEngine(repo CrossBorderTaxRepository) *CrossBorderTaxEngine {
	return &CrossBorderTaxEngine{repo: repo}
}

// Calculate 计算跨境税务
func (e *CrossBorderTaxEngine) Calculate(ctx context.Context, input *CrossBorderTaxInput) (*CrossBorderTaxResult, error) {
	// 查找适用的跨境税务配置
	config, err := e.repo.FindConfig(ctx, input.OriginCountry, input.DestinationCountry, input.TradeType)
	if err != nil {
		return nil, err
	}

	result := &CrossBorderTaxResult{
		TaxCollectionMethod: config.TaxCollectionMethod,
	}

	// 计算完税价格
	result.CustomsValue = input.ProductValue + input.ShippingCost + input.InsuranceCost

	// 检查最低征税门槛
	if result.CustomsValue <= config.DeMinimisThreshold {
		// 低于门槛，免税
		return result, nil
	}

	// 检查是否有适用的优惠协定
	var appliedAgreement *PreferentialAgreement
	if input.HasOriginCert {
		appliedAgreement, err = e.findApplicableAgreement(ctx, config, input)
		if err != nil {
			return nil, err
		}
	}

	// 计算进口税费
	for _, rule := range config.ImportTaxRules {
		switch rule.TaxType {
		case TaxTypeDuty:
			result.CustomsDuty = e.calculateCustomsDuty(result.CustomsValue, rule, appliedAgreement)
		case TaxTypeVAT:
			// 进口VAT = (完税价格 + 关税 + 消费税) * 增值税率
			taxBase := result.CustomsValue + result.CustomsDuty + result.ImportExcise
			result.ImportVAT = int64(float64(taxBase) * rule.Rate)
		case TaxTypeExcise:
			result.ImportExcise = int64(float64(result.CustomsValue) * rule.Rate)
		}
	}

	result.TotalImportTax = result.CustomsDuty + result.ImportVAT + result.ImportExcise + result.OtherTaxes
	result.PreferentialApplied = appliedAgreement != nil
	result.AppliedAgreement = appliedAgreement

	// 收集所需单证
	result.RequiredDocuments = e.collectRequiredDocuments(config, input)

	return result, nil
}

// findApplicableAgreement 查找适用的优惠协定
func (e *CrossBorderTaxEngine) findApplicableAgreement(ctx context.Context, config *CrossBorderTaxConfig, input *CrossBorderTaxInput) (*PreferentialAgreement, error) {
	for _, agreement := range config.PreferentialAgreements {
		// 检查协定有效期
		now := time.Now()
		if now.Before(agreement.ValidFrom) || now.After(agreement.ValidTo) {
			continue
		}

		// 检查原产地规则
		if e.checkROORules(agreement.RulesOfOrigin, input) {
			return &agreement, nil
		}
	}
	return nil, nil
}

// checkROORules 检查原产地规则
func (e *CrossBorderTaxEngine) checkROORules(rules []ROORule, input *CrossBorderTaxInput) bool {
	// 简化实现：实际应该验证各种原产地规则
	for _, rule := range rules {
		switch rule.RuleType {
		case ROOTypeWO:
			// 完全获得：通常适用于农产品、矿产品等
			return true
		case ROOTypeRVC:
			// 区域价值成分检查
			// 实际应该计算产品的本地成分比例
			return true
		case ROOTypeCTH:
			// 检查HS编码是否发生归类改变
			if input.HSCode != "" && rule.HSCodePrefix != "" {
				return true
			}
		}
	}
	return true // 默认通过
}

// calculateCustomsDuty 计算关税
func (e *CrossBorderTaxEngine) calculateCustomsDuty(customsValue int64, rule ImportTaxRule, agreement *PreferentialAgreement) int64 {
	rate := rule.Rate
	if agreement != nil && agreement.PreferentialRate > 0 {
		rate = agreement.PreferentialRate
	}
	return int64(float64(customsValue) * rate)
}

// collectRequiredDocuments 收集所需单证
func (e *CrossBorderTaxEngine) collectRequiredDocuments(config *CrossBorderTaxConfig, input *CrossBorderTaxInput) []string {
	docs := []string{
		"Commercial Invoice",
		"Packing List",
		"Bill of Lading/Airway Bill",
	}

	// 根据商品类别添加特定单证
	switch input.Category {
	case "FOOD", "AGRICULTURE":
		docs = append(docs, "Health Certificate", "Phytosanitary Certificate")
	case "CHEMICAL", "HAZARDOUS":
		docs = append(docs, "MSDS", "Dangerous Goods Declaration")
	case "ELECTRONICS":
		docs = append(docs, "FCC Certificate", "CE Certificate")
	}

	// 如果有优惠协定，需要原产地证明
	if input.HasOriginCert {
		docs = append(docs, "Certificate of Origin")
	}

	return docs
}

// --- 仓储接口 ---

// CrossBorderTaxRepository 跨境税务仓储接口
type CrossBorderTaxRepository interface {
	FindConfig(ctx context.Context, origin, destination string, tradeType TradeType) (*CrossBorderTaxConfig, error)
	SaveConfig(ctx context.Context, config *CrossBorderTaxConfig) error
	FindAgreement(ctx context.Context, agreementCode string) (*PreferentialAgreement, error)
	ListAgreements(ctx context.Context, origin, destination string) ([]*PreferentialAgreement, error)
}

// --- 欧盟特定配置 ---

// EUOSSConfig 欧盟OSS/IOSS配置
type EUOSSConfig struct {
	ID          uint64    `json:"id"`
	MerchantID  uint64    `json:"merchant_id"`
	OSSNumber   string    `json:"oss_number"`   // OSS登记号
	IOSSNumber  string    `json:"ioss_number"`  // IOSS登记号
	MemberState string    `json:"member_state"` // 成员国
	ValidFrom   time.Time `json:"valid_from"`
	ValidTo     time.Time `json:"valid_to"`
	IsActive    bool      `json:"is_active"`
}

// OSSReturn  OSS申报
type OSSReturn struct {
	ID           uint64            `json:"id"`
	Period       string            `json:"period"` // 申报期（如 2024-Q1）
	MemberState  string            `json:"member_state"`
	SalesAmount  int64             `json:"sales_amount"` // 销售额
	VATPayable   int64             `json:"vat_payable"`  // 应缴VAT
	Status       OSSReturnStatus   `json:"status"`
	SubmittedAt  *time.Time        `json:"submitted_at"`
	Transactions []*OSSTransaction `json:"transactions"`
}

// OSSReturnStatus 申报状态
type OSSReturnStatus int

const (
	OSSStatusDraft     OSSReturnStatus = 1
	OSSStatusSubmitted OSSReturnStatus = 2
	OSSStatusPaid      OSSReturnStatus = 3
	OSSStatusCompleted OSSReturnStatus = 4
)

// OSSTransaction OSS交易明细
type OSSTransaction struct {
	ID          uint64  `json:"id"`
	OrderID     uint64  `json:"order_id"`
	CountryCode string  `json:"country_code"` // 消费国
	Amount      int64   `json:"amount"`
	VATRate     float64 `json:"vat_rate"`
	VATAmount   int64   `json:"vat_amount"`
}

// EUOSSRepository 欧盟OSS仓储接口
type EUOSSRepository interface {
	FindOSSConfig(ctx context.Context, merchantID uint64) (*EUOSSConfig, error)
	SaveOSSReturn(ctx context.Context, returnRecord *OSSReturn) error
	GetOSSReturns(ctx context.Context, merchantID uint64, period string) ([]*OSSReturn, error)
}
