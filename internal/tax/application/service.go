// Package application 税务应用服务
// 变更说明：扩展TaxService，集成新的税务计算引擎、跨境税务规则和报表生成功能
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/tax/domain"
	"github.com/wyfcoding/ecommerce/internal/tax/infrastructure/reporting"
	"github.com/wyfcoding/ecommerce/internal/tax/infrastructure/thirdparty"
)

// TaxService 税务应用服务
type TaxService struct {
	repo                    domain.TaxRepository
	crossBorderRepo         domain.CrossBorderTaxRepository
	vatCalculator           *domain.VATCalculator
	exciseCalculator        *domain.ExciseCalculator
	dutyCalculator          *domain.DutyCalculator
	comprehensiveCalculator *domain.ComprehensiveTaxCalculator
	crossBorderEngine       *domain.CrossBorderTaxEngine
	reportGenerator         *reporting.ReportGenerator
	thirdPartyProvider      thirdparty.ThirdPartyTaxProvider
	logger                  *slog.Logger
}

// TaxServiceConfig 税务服务配置
type TaxServiceConfig struct {
	UseThirdPartyProvider bool
	ThirdPartyProvider    thirdparty.ThirdPartyTaxProvider
}

// NewTaxService 创建税务服务
func NewTaxService(
	repo domain.TaxRepository,
	crossBorderRepo domain.CrossBorderTaxRepository,
	config *TaxServiceConfig,
	logger *slog.Logger,
) *TaxService {
	svc := &TaxService{
		repo:                    repo,
		crossBorderRepo:         crossBorderRepo,
		vatCalculator:           domain.NewVATCalculator(repo),
		exciseCalculator:        domain.NewExciseCalculator(repo),
		dutyCalculator:          domain.NewDutyCalculator(repo),
		comprehensiveCalculator: domain.NewComprehensiveTaxCalculator(repo),
		crossBorderEngine:       domain.NewCrossBorderTaxEngine(crossBorderRepo),
		reportGenerator:         reporting.NewReportGenerator(repo),
		logger:                  logger,
	}

	if config != nil && config.UseThirdPartyProvider {
		svc.thirdPartyProvider = config.ThirdPartyProvider
	}

	return svc
}

// CalculateOrderTax 计算订单税费（基础方法）
func (s *TaxService) CalculateOrderTax(ctx context.Context, userID uint64, country, region, category string, amount int64) (*domain.TaxCalculationResult, error) {
	// 1. 检查免税
	exemption, err := s.repo.FindExemption(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check exemption", "error", err)
	}
	if exemption != nil {
		s.logger.InfoContext(ctx, "tax exemption applied", "user_id", userID, "reason", exemption.Reason)
		return &domain.TaxCalculationResult{
			TotalTaxAmount: 0,
			Currency:       "USD",
			Items:          []*domain.TaxDetailItem{},
		}, nil
	}

	// 2. 查找规则
	rules, err := s.repo.FindActiveRules(ctx, country, region, category)
	if err != nil {
		return nil, err
	}

	// 3. 计算
	result := &domain.TaxCalculationResult{
		Currency: "USD",
	}

	baseAmount := amount
	for _, rule := range rules {
		taxAmount := rule.CalculateTax(baseAmount)
		if taxAmount > 0 {
			item := &domain.TaxDetailItem{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				TaxType:    rule.TaxType,
				BaseAmount: baseAmount,
				Rate:       rule.Rate,
				Amount:     taxAmount,
			}
			result.Items = append(result.Items, item)
			result.TotalTaxAmount += taxAmount

			if rule.IsCompound {
				baseAmount += taxAmount
			}
		}
	}

	return result, nil
}

// CalculateVAT 计算增值税
func (s *TaxService) CalculateVAT(ctx context.Context, input *domain.VATCalculationInput) (*domain.VATCalculationResult, error) {
	return s.vatCalculator.Calculate(ctx, input)
}

// CalculateExcise 计算消费税
func (s *TaxService) CalculateExcise(ctx context.Context, input *domain.ExciseCalculationInput) (*domain.ExciseCalculationResult, error) {
	return s.exciseCalculator.Calculate(ctx, input)
}

// CalculateDuty 计算关税
func (s *TaxService) CalculateDuty(ctx context.Context, input *domain.DutyCalculationInput) (*domain.DutyCalculationResult, error) {
	return s.dutyCalculator.Calculate(ctx, input)
}

// CalculateComprehensiveTax 计算综合税费
func (s *TaxService) CalculateComprehensiveTax(ctx context.Context, input *domain.ComprehensiveTaxInput) (*domain.ComprehensiveTaxResult, error) {
	return s.comprehensiveCalculator.CalculateAll(ctx, input)
}

// CalculateCrossBorderTax 计算跨境税务
func (s *TaxService) CalculateCrossBorderTax(ctx context.Context, input *domain.CrossBorderTaxInput) (*domain.CrossBorderTaxResult, error) {
	return s.crossBorderEngine.Calculate(ctx, input)
}

// CalculateTaxWithThirdParty 使用第三方服务计算税费
func (s *TaxService) CalculateTaxWithThirdParty(ctx context.Context, request *thirdparty.TaxCalculationRequest) (*thirdparty.TaxCalculationResponse, error) {
	if s.thirdPartyProvider == nil {
		return nil, fmt.Errorf("third-party provider not configured")
	}
	return s.thirdPartyProvider.CalculateTax(ctx, request)
}

// ValidateAddress 验证地址（使用第三方服务）
func (s *TaxService) ValidateAddress(ctx context.Context, address *thirdparty.Address) (*thirdparty.AddressValidationResult, error) {
	if s.thirdPartyProvider == nil {
		return nil, fmt.Errorf("third-party provider not configured")
	}
	return s.thirdPartyProvider.ValidateAddress(ctx, address)
}

// GetTaxRates 获取税率（使用第三方服务）
func (s *TaxService) GetTaxRates(ctx context.Context, country, region string) ([]*thirdparty.TaxRateInfo, error) {
	if s.thirdPartyProvider == nil {
		return nil, fmt.Errorf("third-party provider not configured")
	}
	return s.thirdPartyProvider.GetTaxRates(ctx, country, region)
}

// RecordInvoice 记录税务发票
func (s *TaxService) RecordInvoice(ctx context.Context, orderID uint64, result *domain.TaxCalculationResult) error {
	detailsJSON, _ := json.Marshal(result.Items)

	invoice := &domain.TaxInvoice{
		OrderID:      orderID,
		InvoiceNo:    fmt.Sprintf("INV-%d-%d", orderID, time.Now().Unix()),
		TotalNet:     0,
		TotalTax:     result.TotalTaxAmount,
		TotalGross:   0,
		CalculatedAt: time.Now(),
		TaxDetails:   string(detailsJSON),
	}

	return s.repo.SaveInvoice(ctx, invoice)
}

// --- 报表服务 ---

// GenerateVATReturn 生成增值税申报表
func (s *TaxService) GenerateVATReturn(ctx context.Context, startDate, endDate time.Time, countryCode string, includeDetails bool) (*reporting.ReportData, error) {
	return s.reportGenerator.GenerateVATReturn(ctx, startDate, endDate, countryCode, includeDetails)
}

// GenerateExciseReturn 生成消费税申报表
func (s *TaxService) GenerateExciseReturn(ctx context.Context, startDate, endDate time.Time, countryCode string) (*reporting.ReportData, error) {
	return s.reportGenerator.GenerateExciseReturn(ctx, startDate, endDate, countryCode)
}

// GenerateDutyReturn 生成关税申报表
func (s *TaxService) GenerateDutyReturn(ctx context.Context, startDate, endDate time.Time, countryCode string) (*reporting.ReportData, error) {
	return s.reportGenerator.GenerateDutyReturn(ctx, startDate, endDate, countryCode)
}

// GenerateConsolidatedReport 生成综合税务报表
func (s *TaxService) GenerateConsolidatedReport(ctx context.Context, startDate, endDate time.Time, countryCode string) (*reporting.ReportData, error) {
	return s.reportGenerator.GenerateConsolidatedReport(ctx, startDate, endDate, countryCode)
}

// GenerateCustomReport 生成自定义报表
func (s *TaxService) GenerateCustomReport(ctx context.Context, request *reporting.ReportRequest) (*reporting.ReportData, error) {
	return s.reportGenerator.Generate(ctx, request)
}

// ExportReport 导出报表
func (s *TaxService) ExportReport(ctx context.Context, request *reporting.ReportRequest, writer io.Writer) error {
	return s.reportGenerator.ExportReport(ctx, request, writer)
}

// GetSupportedReportFormats 获取支持的报表格式
func (s *TaxService) GetSupportedReportFormats() []reporting.ReportFormat {
	return s.reportGenerator.GetSupportedFormats()
}

// GetSupportedReportTypes 获取支持的报表类型
func (s *TaxService) GetSupportedReportTypes() []reporting.ReportType {
	return s.reportGenerator.GetSupportedReportTypes()
}

// --- 跨境税务配置管理 ---

// SaveCrossBorderConfig 保存跨境税务配置
func (s *TaxService) SaveCrossBorderConfig(ctx context.Context, config *domain.CrossBorderTaxConfig) error {
	return s.crossBorderRepo.SaveConfig(ctx, config)
}

// GetCrossBorderConfig 获取跨境税务配置
func (s *TaxService) GetCrossBorderConfig(ctx context.Context, origin, destination string, tradeType domain.TradeType) (*domain.CrossBorderTaxConfig, error) {
	return s.crossBorderRepo.FindConfig(ctx, origin, destination, tradeType)
}

// GetPreferentialAgreements 获取优惠贸易协定
func (s *TaxService) GetPreferentialAgreements(ctx context.Context, origin, destination string) ([]*domain.PreferentialAgreement, error) {
	return s.crossBorderRepo.ListAgreements(ctx, origin, destination)
}

// --- 免税管理 ---

// ApplyTaxExemption 申请税务减免
func (s *TaxService) ApplyTaxExemption(ctx context.Context, exemption *domain.TaxExemption) error {
	return s.repo.SaveExemption(ctx, exemption)
}

// GetTaxExemption 获取用户税务减免
func (s *TaxService) GetTaxExemption(ctx context.Context, userID uint64) (*domain.TaxExemption, error) {
	return s.repo.FindExemption(ctx, userID)
}

// --- 税务规则管理 ---

// CreateTaxRule 创建税务规则
func (s *TaxService) CreateTaxRule(ctx context.Context, rule *domain.TaxRule) error {
	return s.repo.SaveRule(ctx, rule)
}

// GetTaxRules 获取税务规则
func (s *TaxService) GetTaxRules(ctx context.Context, country, region, category string) ([]*domain.TaxRule, error) {
	return s.repo.FindActiveRules(ctx, country, region, category)
}

// --- 发票查询 ---

// GetInvoiceByOrder 根据订单获取发票
func (s *TaxService) GetInvoiceByOrder(ctx context.Context, orderID uint64) (*domain.TaxInvoice, error) {
	return s.repo.FindByOrder(ctx, orderID)
}
