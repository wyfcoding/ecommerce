// Package domain 提供了税务计算的领域逻辑。
// 变更说明：补全通用税务计算引擎，支持增值税、关税及金融交易税 (Financial Transaction Tax) 的统一抽象。
package domain

import (
	"context"
)

// Additional Tax Types for Financial Trading
const (
	TaxTypeFTT   TaxType = 6 // 金融交易税 (Financial Transaction Tax)
	TaxTypeStamp TaxType = 7 // 印花税 (Stamp Duty)
)

// UniversalTaxCalculator 通用税务计算实现
type UniversalTaxCalculator struct {
	repo TaxRepository
}

func NewUniversalTaxCalculator(repo TaxRepository) *UniversalTaxCalculator {
	return &UniversalTaxCalculator{repo: repo}
}

// CalculateOrderTax 实现 TaxCalculator 接口，计算零售订单税费
func (c *UniversalTaxCalculator) CalculateOrderTax(ctx context.Context, country, region string, category string, amount int64) (*TaxCalculationResult, error) {
	rules, err := c.repo.FindActiveRules(ctx, country, region, category)
	if err != nil {
		return nil, err
	}

	result := &TaxCalculationResult{
		Items: make([]*TaxDetailItem, 0),
	}

	for _, rule := range rules {
		taxAmount := rule.CalculateTax(amount)
		if taxAmount > 0 {
			result.TotalTaxAmount += taxAmount
			result.Items = append(result.Items, &TaxDetailItem{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				TaxType:    rule.TaxType,
				BaseAmount: amount,
				Rate:       rule.Rate,
				Amount:     taxAmount,
			})
		}
	}

	return result, nil
}

// CalculateDuty 实现 TaxCalculator 接口，计算进口关税
func (c *UniversalTaxCalculator) CalculateDuty(ctx context.Context, originCountry, destCountry string, amount int64) (*TaxCalculationResult, error) {
	// 简化逻辑：查找目标国家针对特定原产地的关税规则
	rules, err := c.repo.FindActiveRules(ctx, destCountry, "ALL", "IMPORT_DUTY")
	if err != nil {
		return nil, err
	}

	result := &TaxCalculationResult{Items: make([]*TaxDetailItem, 0)}
	for _, rule := range rules {
		if rule.TaxType == TaxTypeDuty {
			tax := rule.CalculateTax(amount)
			result.TotalTaxAmount += tax
			result.Items = append(result.Items, &TaxDetailItem{
				RuleID:   rule.ID,
				RuleName: rule.Name,
				TaxType:  TaxTypeDuty,
				Amount:   tax,
			})
		}
	}
	return result, nil
}

// CalculateFinancialTax 计算金融交易税（跨项目协同能力）
func (c *UniversalTaxCalculator) CalculateFinancialTax(ctx context.Context, country string, assetType string, amount int64) (*TaxCalculationResult, error) {
	// 查找金融相关的税收规则 (FTT 或 Stamp Duty)
	rules, err := c.repo.FindActiveRules(ctx, country, "ALL", "FINANCIAL_"+assetType)
	if err != nil {
		return nil, err
	}

	result := &TaxCalculationResult{Items: make([]*TaxDetailItem, 0)}
	for _, rule := range rules {
		if rule.TaxType == TaxTypeFTT || rule.TaxType == TaxTypeStamp {
			tax := rule.CalculateTax(amount)
			result.TotalTaxAmount += tax
			result.Items = append(result.Items, &TaxDetailItem{
				RuleID:   rule.ID,
				RuleName: rule.Name,
				TaxType:  rule.TaxType,
				Amount:   tax,
			})
		}
	}
	return result, nil
}
