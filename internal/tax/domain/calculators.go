// Package domain 扩展税务计算引擎，支持增值税、消费税、关税专门计算
// 变更说明：添加专门的VAT、Excise、Duty计算引擎，支持复杂的计税规则和税率表
package domain

import (
	"context"
	"fmt"
	"math"
)

// --- 扩展税务类型 ---

// VATCalculationMethod 增值税计算方法
type VATCalculationMethod int

const (
	VATMethodStandard  VATCalculationMethod = 1 // 标准增值税（价外税）
	VATMethodInclusive VATCalculationMethod = 2 // 内含增值税
	VATMethodCascading VATCalculationMethod = 3 // 级联增值税
)

// ExciseTaxBase 消费税计税基础
type ExciseTaxBase int

const (
	ExciseBaseAdValorem ExciseTaxBase = 1 // 从价计征
	ExciseBaseSpecific  ExciseTaxBase = 2 // 从量计征
	ExciseBaseCompound  ExciseTaxBase = 3 // 复合计征
)

// DutyCalculationType 关税计算类型
type DutyCalculationType int

const (
	DutyTypeAdValorem   DutyCalculationType = 1 // 从价税
	DutyTypeSpecific    DutyCalculationType = 2 // 从量税
	DutyTypeCompound    DutyCalculationType = 3 // 复合税
	DutyTypeAlternative DutyCalculationType = 4 // 选择税
)

// --- VAT 增值税计算引擎 ---

// VATCalculator 增值税计算器
type VATCalculator struct {
	repo TaxRepository
}

func NewVATCalculator(repo TaxRepository) *VATCalculator {
	return &VATCalculator{repo: repo}
}

// VATCalculationInput VAT计算输入
type VATCalculationInput struct {
	CountryCode string
	RegionCode  string
	Category    string
	Amount      int64 // 金额（分）
	Method      VATCalculationMethod
	IncludeVAT  bool // 金额是否含税
}

// VATCalculationResult VAT计算结果
type VATCalculationResult struct {
	NetAmount   int64        // 不含税金额
	VATAmount   int64        // 增值税额
	GrossAmount int64        // 含税总额
	Rate        float64      // 适用税率
	Details     []*VATDetail // 税率明细（支持多档税率）
}

// VATDetail VAT明细
type VATDetail struct {
	RateName      string
	Rate          float64
	TaxableAmount int64
	VATAmount     int64
}

// Calculate 计算增值税
func (c *VATCalculator) Calculate(ctx context.Context, input *VATCalculationInput) (*VATCalculationResult, error) {
	// 查找适用的VAT规则
	rules, err := c.repo.FindActiveRules(ctx, input.CountryCode, input.RegionCode, input.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to find VAT rules: %w", err)
	}

	result := &VATCalculationResult{
		Details: make([]*VATDetail, 0),
	}

	// 筛选VAT规则
	var vatRules []*TaxRule
	for _, rule := range rules {
		if rule.TaxType == TaxTypeVAT {
			vatRules = append(vatRules, rule)
		}
	}

	if len(vatRules) == 0 {
		// 无VAT规则，返回原金额
		result.NetAmount = input.Amount
		result.GrossAmount = input.Amount
		return result, nil
	}

	// 计算VAT
	switch input.Method {
	case VATMethodStandard:
		result = c.calculateStandardVAT(input, vatRules)
	case VATMethodInclusive:
		result = c.calculateInclusiveVAT(input, vatRules)
	case VATMethodCascading:
		result = c.calculateCascadingVAT(input, vatRules)
	default:
		result = c.calculateStandardVAT(input, vatRules)
	}

	return result, nil
}

// calculateStandardVAT 标准增值税计算（价外税）
func (c *VATCalculator) calculateStandardVAT(input *VATCalculationInput, rules []*TaxRule) *VATCalculationResult {
	result := &VATCalculationResult{
		NetAmount: input.Amount,
		Details:   make([]*VATDetail, 0),
	}

	for _, rule := range rules {
		vatAmount := int64(float64(input.Amount) * rule.Rate)
		result.VATAmount += vatAmount
		result.Details = append(result.Details, &VATDetail{
			RateName:      rule.Name,
			Rate:          rule.Rate,
			TaxableAmount: input.Amount,
			VATAmount:     vatAmount,
		})
		result.Rate = rule.Rate // 使用最后一条规则的税率
	}

	result.GrossAmount = input.Amount + result.VATAmount
	return result
}

// calculateInclusiveVAT 内含增值税计算
func (c *VATCalculator) calculateInclusiveVAT(input *VATCalculationInput, rules []*TaxRule) *VATCalculationResult {
	result := &VATCalculationResult{
		GrossAmount: input.Amount,
		Details:     make([]*VATDetail, 0),
	}

	// 倒算不含税金额
	totalRate := 0.0
	for _, rule := range rules {
		totalRate += rule.Rate
	}

	// 价税分离：不含税金额 = 含税金额 / (1 + 税率)
	result.NetAmount = int64(float64(input.Amount) / (1 + totalRate))
	result.VATAmount = input.Amount - result.NetAmount

	// 分摊到各税率
	for _, rule := range rules {
		proportion := rule.Rate / totalRate
		vatForThisRate := int64(float64(result.VATAmount) * proportion)
		result.Details = append(result.Details, &VATDetail{
			RateName:      rule.Name,
			Rate:          rule.Rate,
			TaxableAmount: result.NetAmount,
			VATAmount:     vatForThisRate,
		})
		result.Rate = rule.Rate
	}

	return result
}

// calculateCascadingVAT 级联增值税计算
func (c *VATCalculator) calculateCascadingVAT(input *VATCalculationInput, rules []*TaxRule) *VATCalculationResult {
	result := &VATCalculationResult{
		NetAmount: input.Amount,
		Details:   make([]*VATDetail, 0),
	}

	cumulativeAmount := input.Amount
	for _, rule := range rules {
		vatAmount := int64(float64(cumulativeAmount) * rule.Rate)
		result.VATAmount += vatAmount
		result.Details = append(result.Details, &VATDetail{
			RateName:      rule.Name,
			Rate:          rule.Rate,
			TaxableAmount: cumulativeAmount,
			VATAmount:     vatAmount,
		})
		cumulativeAmount += vatAmount
		result.Rate = rule.Rate
	}

	result.GrossAmount = cumulativeAmount
	return result
}

// --- Excise 消费税计算引擎 ---

// ExciseCalculator 消费税计算器
type ExciseCalculator struct {
	repo TaxRepository
}

func NewExciseCalculator(repo TaxRepository) *ExciseCalculator {
	return &ExciseCalculator{repo: repo}
}

// ExciseCalculationInput 消费税计算输入
type ExciseCalculationInput struct {
	CountryCode string
	RegionCode  string
	Category    string        // 商品类别（烟草、酒精、成品油等）
	Amount      int64         // 商品金额（分）
	Quantity    float64       // 商品数量
	Unit        string        // 计量单位
	Base        ExciseTaxBase // 计税方式
}

// ExciseCalculationResult 消费税计算结果
type ExciseCalculationResult struct {
	AdValoremTax   int64 // 从价税额
	SpecificTax    int64 // 从量税额
	TotalExciseTax int64 // 总消费税额
	Details        []*ExciseDetail
}

// ExciseDetail 消费税明细
type ExciseDetail struct {
	TaxName          string
	TaxType          ExciseTaxBase
	Rate             float64 // 从价税率或从量税额
	BaseAmount       int64
	SpecificQuantity float64
	TaxAmount        int64
}

// Calculate 计算消费税
func (c *ExciseCalculator) Calculate(ctx context.Context, input *ExciseCalculationInput) (*ExciseCalculationResult, error) {
	rules, err := c.repo.FindActiveRules(ctx, input.CountryCode, input.RegionCode, input.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to find excise rules: %w", err)
	}

	result := &ExciseCalculationResult{
		Details: make([]*ExciseDetail, 0),
	}

	// 筛选Excise规则
	var exciseRules []*TaxRule
	for _, rule := range rules {
		if rule.TaxType == TaxTypeExcise {
			exciseRules = append(exciseRules, rule)
		}
	}

	if len(exciseRules) == 0 {
		return result, nil
	}

	switch input.Base {
	case ExciseBaseAdValorem:
		result = c.calculateAdValoremExcise(input, exciseRules)
	case ExciseBaseSpecific:
		result = c.calculateSpecificExcise(input, exciseRules)
	case ExciseBaseCompound:
		result = c.calculateCompoundExcise(input, exciseRules)
	default:
		// 根据规则自动判断
		result = c.calculateAutoExcise(input, exciseRules)
	}

	return result, nil
}

// calculateAdValoremExcise 从价计征消费税
func (c *ExciseCalculator) calculateAdValoremExcise(input *ExciseCalculationInput, rules []*TaxRule) *ExciseCalculationResult {
	result := &ExciseCalculationResult{Details: make([]*ExciseDetail, 0)}

	for _, rule := range rules {
		taxAmount := int64(float64(input.Amount) * rule.Rate)
		result.AdValoremTax += taxAmount
		result.TotalExciseTax += taxAmount
		result.Details = append(result.Details, &ExciseDetail{
			TaxName:    rule.Name,
			TaxType:    ExciseBaseAdValorem,
			Rate:       rule.Rate,
			BaseAmount: input.Amount,
			TaxAmount:  taxAmount,
		})
	}

	return result
}

// calculateSpecificExcise 从量计征消费税
func (c *ExciseCalculator) calculateSpecificExcise(input *ExciseCalculationInput, rules []*TaxRule) *ExciseCalculationResult {
	result := &ExciseCalculationResult{Details: make([]*ExciseDetail, 0)}

	for _, rule := range rules {
		// FixedAmount 表示单位税额（分）
		taxAmount := int64(float64(rule.FixedAmount) * input.Quantity)
		result.SpecificTax += taxAmount
		result.TotalExciseTax += taxAmount
		result.Details = append(result.Details, &ExciseDetail{
			TaxName:          rule.Name,
			TaxType:          ExciseBaseSpecific,
			Rate:             float64(rule.FixedAmount) / 100.0, // 转换为元
			SpecificQuantity: input.Quantity,
			TaxAmount:        taxAmount,
		})
	}

	return result
}

// calculateCompoundExcise 复合计征消费税
func (c *ExciseCalculator) calculateCompoundExcise(input *ExciseCalculationInput, rules []*TaxRule) *ExciseCalculationResult {
	result := &ExciseCalculationResult{Details: make([]*ExciseDetail, 0)}

	for _, rule := range rules {
		// 计算从价部分
		adValoremTax := int64(float64(input.Amount) * rule.Rate)
		result.AdValoremTax += adValoremTax
		result.Details = append(result.Details, &ExciseDetail{
			TaxName:    rule.Name + "(从价)",
			TaxType:    ExciseBaseAdValorem,
			Rate:       rule.Rate,
			BaseAmount: input.Amount,
			TaxAmount:  adValoremTax,
		})

		// 计算从量部分
		specificTax := int64(float64(rule.FixedAmount) * input.Quantity)
		result.SpecificTax += specificTax
		result.Details = append(result.Details, &ExciseDetail{
			TaxName:          rule.Name + "(从量)",
			TaxType:          ExciseBaseSpecific,
			Rate:             float64(rule.FixedAmount) / 100.0,
			SpecificQuantity: input.Quantity,
			TaxAmount:        specificTax,
		})

		result.TotalExciseTax += adValoremTax + specificTax
	}

	return result
}

// calculateAutoExcise 自动判断计税方式
func (c *ExciseCalculator) calculateAutoExcise(input *ExciseCalculationInput, rules []*TaxRule) *ExciseCalculationResult {
	// 简单策略：如果FixedAmount > 0且Quantity > 0，使用复合计征；如果只有Rate，使用从价；如果只有FixedAmount，使用从量
	hasAdValorem := false
	hasSpecific := false

	for _, rule := range rules {
		if rule.Rate > 0 {
			hasAdValorem = true
		}
		if rule.FixedAmount > 0 {
			hasSpecific = true
		}
	}

	if hasAdValorem && hasSpecific {
		return c.calculateCompoundExcise(input, rules)
	} else if hasSpecific {
		return c.calculateSpecificExcise(input, rules)
	} else {
		return c.calculateAdValoremExcise(input, rules)
	}
}

// --- Duty 关税计算引擎 ---

// DutyCalculator 关税计算器
type DutyCalculator struct {
	repo TaxRepository
}

func NewDutyCalculator(repo TaxRepository) *DutyCalculator {
	return &DutyCalculator{repo: repo}
}

// DutyCalculationInput 关税计算输入
type DutyCalculationInput struct {
	OriginCountry   string              // 原产国
	DestCountry     string              // 目的国
	HSCode          string              // 海关编码
	Category        string              // 商品类别
	CustomsValue    int64               // 完税价格（CIF价，分）
	Quantity        float64             // 数量
	Unit            string              // 计量单位
	Weight          float64             // 重量（kg）
	CalculationType DutyCalculationType // 计算类型
}

// DutyCalculationResult 关税计算结果
type DutyCalculationResult struct {
	CustomsDuty       int64 // 关税额
	Details           []*DutyDetail
	PreferentialRate  bool   // 是否使用优惠税率
	OriginCertificate string // 原产地证明编号
}

// DutyDetail 关税明细
type DutyDetail struct {
	TaxName        string
	DutyType       DutyCalculationType
	Rate           float64 // 税率
	SpecificAmount int64   // 单位税额
	BaseValue      int64   // 计税基础
	Quantity       float64
	DutyAmount     int64
}

// Calculate 计算关税
func (c *DutyCalculator) Calculate(ctx context.Context, input *DutyCalculationInput) (*DutyCalculationResult, error) {
	// 查找目的国针对原产国的关税规则
	rules, err := c.repo.FindActiveRules(ctx, input.DestCountry, input.OriginCountry, input.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to find duty rules: %w", err)
	}

	result := &DutyCalculationResult{
		Details: make([]*DutyDetail, 0),
	}

	// 筛选Duty规则
	var dutyRules []*TaxRule
	for _, rule := range rules {
		if rule.TaxType == TaxTypeDuty {
			dutyRules = append(dutyRules, rule)
		}
	}

	if len(dutyRules) == 0 {
		return result, nil
	}

	switch input.CalculationType {
	case DutyTypeAdValorem:
		result = c.calculateAdValoremDuty(input, dutyRules)
	case DutyTypeSpecific:
		result = c.calculateSpecificDuty(input, dutyRules)
	case DutyTypeCompound:
		result = c.calculateCompoundDuty(input, dutyRules)
	case DutyTypeAlternative:
		result = c.calculateAlternativeDuty(input, dutyRules)
	default:
		result = c.calculateAutoDuty(input, dutyRules)
	}

	return result, nil
}

// calculateAdValoremDuty 从价关税
func (c *DutyCalculator) calculateAdValoremDuty(input *DutyCalculationInput, rules []*TaxRule) *DutyCalculationResult {
	result := &DutyCalculationResult{Details: make([]*DutyDetail, 0)}

	for _, rule := range rules {
		dutyAmount := int64(float64(input.CustomsValue) * rule.Rate)
		result.CustomsDuty += dutyAmount
		result.Details = append(result.Details, &DutyDetail{
			TaxName:    rule.Name,
			DutyType:   DutyTypeAdValorem,
			Rate:       rule.Rate,
			BaseValue:  input.CustomsValue,
			DutyAmount: dutyAmount,
		})
	}

	return result
}

// calculateSpecificDuty 从量关税
func (c *DutyCalculator) calculateSpecificDuty(input *DutyCalculationInput, rules []*TaxRule) *DutyCalculationResult {
	result := &DutyCalculationResult{Details: make([]*DutyDetail, 0)}

	for _, rule := range rules {
		dutyAmount := int64(float64(rule.FixedAmount) * input.Quantity)
		result.CustomsDuty += dutyAmount
		result.Details = append(result.Details, &DutyDetail{
			TaxName:        rule.Name,
			DutyType:       DutyTypeSpecific,
			SpecificAmount: rule.FixedAmount,
			Quantity:       input.Quantity,
			DutyAmount:     dutyAmount,
		})
	}

	return result
}

// calculateCompoundDuty 复合关税
func (c *DutyCalculator) calculateCompoundDuty(input *DutyCalculationInput, rules []*TaxRule) *DutyCalculationResult {
	result := &DutyCalculationResult{Details: make([]*DutyDetail, 0)}

	for _, rule := range rules {
		adValoremDuty := int64(float64(input.CustomsValue) * rule.Rate)
		specificDuty := int64(float64(rule.FixedAmount) * input.Quantity)
		dutyAmount := adValoremDuty + specificDuty
		result.CustomsDuty += dutyAmount
		result.Details = append(result.Details, &DutyDetail{
			TaxName:        rule.Name,
			DutyType:       DutyTypeCompound,
			Rate:           rule.Rate,
			SpecificAmount: rule.FixedAmount,
			BaseValue:      input.CustomsValue,
			Quantity:       input.Quantity,
			DutyAmount:     dutyAmount,
		})
	}

	return result
}

// calculateAlternativeDuty 选择关税（从高或从低）
func (c *DutyCalculator) calculateAlternativeDuty(input *DutyCalculationInput, rules []*TaxRule) *DutyCalculationResult {
	result := &DutyCalculationResult{Details: make([]*DutyDetail, 0)}

	for _, rule := range rules {
		adValoremDuty := int64(float64(input.CustomsValue) * rule.Rate)
		specificDuty := int64(float64(rule.FixedAmount) * input.Quantity)

		// 选择较高者
		dutyAmount := int64(math.Max(float64(adValoremDuty), float64(specificDuty)))
		result.CustomsDuty += dutyAmount

		dutyType := DutyTypeAdValorem
		if specificDuty > adValoremDuty {
			dutyType = DutyTypeSpecific
		}

		result.Details = append(result.Details, &DutyDetail{
			TaxName:        rule.Name,
			DutyType:       dutyType,
			Rate:           rule.Rate,
			SpecificAmount: rule.FixedAmount,
			BaseValue:      input.CustomsValue,
			Quantity:       input.Quantity,
			DutyAmount:     dutyAmount,
		})
	}

	return result
}

// calculateAutoDuty 自动判断关税类型
func (c *DutyCalculator) calculateAutoDuty(input *DutyCalculationInput, rules []*TaxRule) *DutyCalculationResult {
	hasAdValorem := false
	hasSpecific := false

	for _, rule := range rules {
		if rule.Rate > 0 {
			hasAdValorem = true
		}
		if rule.FixedAmount > 0 {
			hasSpecific = true
		}
	}

	if hasAdValorem && hasSpecific {
		// 检查是否有明确的复合或选择标记，这里简化为复合税
		return c.calculateCompoundDuty(input, rules)
	} else if hasSpecific {
		return c.calculateSpecificDuty(input, rules)
	} else {
		return c.calculateAdValoremDuty(input, rules)
	}
}

// --- 统一税务计算服务 ---

// ComprehensiveTaxCalculator 综合税务计算器
type ComprehensiveTaxCalculator struct {
	vatCalculator    *VATCalculator
	exciseCalculator *ExciseCalculator
	dutyCalculator   *DutyCalculator
}

func NewComprehensiveTaxCalculator(repo TaxRepository) *ComprehensiveTaxCalculator {
	return &ComprehensiveTaxCalculator{
		vatCalculator:    NewVATCalculator(repo),
		exciseCalculator: NewExciseCalculator(repo),
		dutyCalculator:   NewDutyCalculator(repo),
	}
}

// ComprehensiveTaxInput 综合税务计算输入
type ComprehensiveTaxInput struct {
	VATInput    *VATCalculationInput
	ExciseInput *ExciseCalculationInput
	DutyInput   *DutyCalculationInput
}

// ComprehensiveTaxResult 综合税务计算结果
type ComprehensiveTaxResult struct {
	VATResult    *VATCalculationResult
	ExciseResult *ExciseCalculationResult
	DutyResult   *DutyCalculationResult
	TotalTax     int64
	NetAmount    int64
	GrossAmount  int64
}

// CalculateAll 计算所有税种
func (c *ComprehensiveTaxCalculator) CalculateAll(ctx context.Context, input *ComprehensiveTaxInput) (*ComprehensiveTaxResult, error) {
	result := &ComprehensiveTaxResult{}

	if input.VATInput != nil {
		vatResult, err := c.vatCalculator.Calculate(ctx, input.VATInput)
		if err != nil {
			return nil, fmt.Errorf("VAT calculation failed: %w", err)
		}
		result.VATResult = vatResult
		result.TotalTax += vatResult.VATAmount
	}

	if input.ExciseInput != nil {
		exciseResult, err := c.exciseCalculator.Calculate(ctx, input.ExciseInput)
		if err != nil {
			return nil, fmt.Errorf("Excise calculation failed: %w", err)
		}
		result.ExciseResult = exciseResult
		result.TotalTax += exciseResult.TotalExciseTax
	}

	if input.DutyInput != nil {
		dutyResult, err := c.dutyCalculator.Calculate(ctx, input.DutyInput)
		if err != nil {
			return nil, fmt.Errorf("Duty calculation failed: %w", err)
		}
		result.DutyResult = dutyResult
		result.TotalTax += dutyResult.CustomsDuty
	}

	// 计算净额和总额（基于VAT计算）
	if result.VATResult != nil {
		result.NetAmount = result.VATResult.NetAmount
		result.GrossAmount = result.VATResult.GrossAmount
	}

	return result, nil
}
