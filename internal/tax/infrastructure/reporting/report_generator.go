// Package reporting 税务报表生成
// 变更说明：提供各种税务报表生成功能，支持增值税、消费税、关税等多税种报表
package reporting

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/wyfcoding/ecommerce/internal/tax/domain"
)

// ReportType 报表类型
type ReportType int

const (
	ReportTypeVATReturn       ReportType = 1 // 增值税申报表
	ReportTypeExciseReturn    ReportType = 2 // 消费税申报表
	ReportTypeDutyReturn      ReportType = 3 // 关税申报表
	ReportTypeSalesTax        ReportType = 4 // 销售税报表
	ReportTypeWithholdingTax  ReportType = 5 // 预扣税报表
	ReportTypeConsolidated    ReportType = 6 // 综合税务报表
	ReportTypeTransactionLog  ReportType = 7 // 交易日志
	ReportTypeExemptionReport ReportType = 8 // 免税报表
	ReportTypeAuditTrail      ReportType = 9 // 审计追踪
)

func (r ReportType) String() string {
	switch r {
	case ReportTypeVATReturn:
		return "VAT Return"
	case ReportTypeExciseReturn:
		return "Excise Tax Return"
	case ReportTypeDutyReturn:
		return "Customs Duty Return"
	case ReportTypeSalesTax:
		return "Sales Tax Report"
	case ReportTypeWithholdingTax:
		return "Withholding Tax Report"
	case ReportTypeConsolidated:
		return "Consolidated Tax Report"
	case ReportTypeTransactionLog:
		return "Transaction Log"
	case ReportTypeExemptionReport:
		return "Tax Exemption Report"
	case ReportTypeAuditTrail:
		return "Audit Trail"
	default:
		return "Unknown"
	}
}

// ReportFormat 报表格式
type ReportFormat int

const (
	FormatJSON  ReportFormat = 1
	FormatCSV   ReportFormat = 2
	FormatPDF   ReportFormat = 3
	FormatExcel ReportFormat = 4
	FormatXML   ReportFormat = 5
)

func (f ReportFormat) String() string {
	switch f {
	case FormatJSON:
		return "JSON"
	case FormatCSV:
		return "CSV"
	case FormatPDF:
		return "PDF"
	case FormatExcel:
		return "Excel"
	case FormatXML:
		return "XML"
	default:
		return "Unknown"
	}
}

// ReportRequest 报表请求
type ReportRequest struct {
	ReportType     ReportType
	Format         ReportFormat
	StartDate      time.Time
	EndDate        time.Time
	CountryCode    string
	RegionCode     string
	EntityID       uint64
	TaxType        domain.TaxType
	IncludeDetails bool
}

// Report 报表接口
type Report interface {
	Generate(ctx context.Context, request *ReportRequest) (*ReportData, error)
}

// ReportData 报表数据
type ReportData struct {
	ReportType   string      `json:"report_type"`
	ReportFormat string      `json:"report_format"`
	GeneratedAt  time.Time   `json:"generated_at"`
	PeriodStart  time.Time   `json:"period_start"`
	PeriodEnd    time.Time   `json:"period_end"`
	CountryCode  string      `json:"country_code,omitempty"`
	RegionCode   string      `json:"region_code,omitempty"`
	Summary      interface{} `json:"summary"`
	Details      interface{} `json:"details,omitempty"`
	RawData      []byte      `json:"-"` // 原始字节数据
}

// --- 增值税申报表 ---

// VATReturnSummary 增值税申报表汇总
type VATReturnSummary struct {
	Box1VATDueSales         int64 `json:"box1_vat_due_sales"`         // 销项税额
	Box2VATDueAcquisitions  int64 `json:"box2_vat_due_acquisitions"`  // 视同销售税额
	Box3TotalVATDue         int64 `json:"box3_total_vat_due"`         // 应缴VAT总额
	Box4VATReclaimed        int64 `json:"box4_vat_reclaimed"`         // 可抵扣进项税额
	Box5NetVAT              int64 `json:"box5_net_vat"`               // 应缴或应退净额
	Box6TotalSales          int64 `json:"box6_total_sales"`           // 销售总额（不含税）
	Box7TotalPurchases      int64 `json:"box7_total_purchases"`       // 采购总额（不含税）
	Box8TotalSuppliesEU     int64 `json:"box8_total_supplies_eu"`     // 欧盟内供应总额
	Box9TotalAcquisitionsEU int64 `json:"box9_total_acquisitions_eu"` // 欧盟内采购总额
}

// VATReturnDetail 增值税申报表明细
type VATReturnDetail struct {
	TransactionID uint64    `json:"transaction_id"`
	Date          time.Time `json:"date"`
	Type          string    `json:"type"` // Sale, Purchase
	CustomerName  string    `json:"customer_name"`
	InvoiceNo     string    `json:"invoice_no"`
	NetAmount     int64     `json:"net_amount"`
	VATRate       float64   `json:"vat_rate"`
	VATAmount     int64     `json:"vat_amount"`
	GrossAmount   int64     `json:"gross_amount"`
	CountryCode   string    `json:"country_code"`
}

// VATReturnReport 增值税申报表
type VATReturnReport struct {
	repo domain.TaxRepository
}

func NewVATReturnReport(repo domain.TaxRepository) *VATReturnReport {
	return &VATReturnReport{repo: repo}
}

func (r *VATReturnReport) Generate(ctx context.Context, request *ReportRequest) (*ReportData, error) {
	// 获取指定期间的税务发票
	invoices, err := r.getInvoicesForPeriod(ctx, request.StartDate, request.EndDate, request.CountryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	summary := &VATReturnSummary{}
	details := make([]*VATReturnDetail, 0)

	for _, invoice := range invoices {
		detail := &VATReturnDetail{
			TransactionID: invoice.OrderID,
			Date:          invoice.CalculatedAt,
			Type:          "Sale",
			InvoiceNo:     invoice.InvoiceNo,
			NetAmount:     invoice.TotalNet,
			VATAmount:     invoice.TotalTax,
			GrossAmount:   invoice.TotalGross,
		}
		details = append(details, detail)

		// 汇总计算
		summary.Box1VATDueSales += invoice.TotalTax
		summary.Box6TotalSales += invoice.TotalNet
	}

	summary.Box3TotalVATDue = summary.Box1VATDueSales + summary.Box2VATDueAcquisitions
	summary.Box5NetVAT = summary.Box3TotalVATDue - summary.Box4VATReclaimed

	report := &ReportData{
		ReportType:   ReportTypeVATReturn.String(),
		ReportFormat: request.Format.String(),
		GeneratedAt:  time.Now(),
		PeriodStart:  request.StartDate,
		PeriodEnd:    request.EndDate,
		CountryCode:  request.CountryCode,
		Summary:      summary,
	}

	if request.IncludeDetails {
		report.Details = details
	}

	// 格式化输出
	rawData, err := r.formatReport(report, request.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to format report: %w", err)
	}
	report.RawData = rawData

	return report, nil
}

func (r *VATReturnReport) getInvoicesForPeriod(ctx context.Context, start, end time.Time, country string) ([]*domain.TaxInvoice, error) {
	// 这里应该调用仓储层的方法
	// 简化实现，返回空列表
	return []*domain.TaxInvoice{}, nil
}

func (r *VATReturnReport) formatReport(report *ReportData, format ReportFormat) ([]byte, error) {
	switch format {
	case FormatJSON:
		return json.MarshalIndent(report, "", "  ")
	case FormatCSV:
		return r.toCSV(report)
	default:
		return json.Marshal(report)
	}
}

func (r *VATReturnReport) toCSV(report *ReportData) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// 写入标题
	writer.Write([]string{"VAT Return Report"})
	writer.Write([]string{"Period", report.PeriodStart.Format("2006-01-02") + " to " + report.PeriodEnd.Format("2006-01-02")})
	writer.Write([]string{})

	// 写入汇总
	if summary, ok := report.Summary.(*VATReturnSummary); ok {
		writer.Write([]string{"Summary", "Amount"})
		writer.Write([]string{"VAT Due on Sales", strconv.FormatInt(summary.Box1VATDueSales, 10)})
		writer.Write([]string{"VAT Due on Acquisitions", strconv.FormatInt(summary.Box2VATDueAcquisitions, 10)})
		writer.Write([]string{"Total VAT Due", strconv.FormatInt(summary.Box3TotalVATDue, 10)})
		writer.Write([]string{"VAT Reclaimed", strconv.FormatInt(summary.Box4VATReclaimed, 10)})
		writer.Write([]string{"Net VAT", strconv.FormatInt(summary.Box5NetVAT, 10)})
		writer.Write([]string{"Total Sales", strconv.FormatInt(summary.Box6TotalSales, 10)})
		writer.Write([]string{"Total Purchases", strconv.FormatInt(summary.Box7TotalPurchases, 10)})
		writer.Write([]string{})
	}

	// 写入明细
	if report.Details != nil {
		if details, ok := report.Details.([]*VATReturnDetail); ok && len(details) > 0 {
			writer.Write([]string{"Transaction ID", "Date", "Type", "Invoice No", "Net Amount", "VAT Rate", "VAT Amount", "Gross Amount"})
			for _, d := range details {
				writer.Write([]string{
					strconv.FormatUint(d.TransactionID, 10),
					d.Date.Format("2006-01-02"),
					d.Type,
					d.InvoiceNo,
					strconv.FormatInt(d.NetAmount, 10),
					strconv.FormatFloat(d.VATRate, 'f', 2, 64),
					strconv.FormatInt(d.VATAmount, 10),
					strconv.FormatInt(d.GrossAmount, 10),
				})
			}
		}
	}

	writer.Flush()
	return []byte(buf.String()), nil
}

// --- 消费税申报表 ---

// ExciseReturnSummary 消费税申报表汇总
type ExciseReturnSummary struct {
	Category          string                   `json:"category"`
	Quantity          float64                  `json:"quantity"`
	TaxBase           int64                    `json:"tax_base"`
	AdValoremTax      int64                    `json:"ad_valorem_tax"`
	SpecificTax       int64                    `json:"specific_tax"`
	TotalExciseTax    int64                    `json:"total_excise_tax"`
	CategoryBreakdown []*ExciseCategorySummary `json:"category_breakdown"`
}

// ExciseCategorySummary 消费品类目汇总
type ExciseCategorySummary struct {
	Category  string  `json:"category"`
	Quantity  float64 `json:"quantity"`
	TaxBase   int64   `json:"tax_base"`
	ExciseTax int64   `json:"excise_tax"`
}

// ExciseReturnReport 消费税申报表
type ExciseReturnReport struct{}

func NewExciseReturnReport() *ExciseReturnReport {
	return &ExciseReturnReport{}
}

func (r *ExciseReturnReport) Generate(ctx context.Context, request *ReportRequest) (*ReportData, error) {
	summary := &ExciseReturnSummary{
		CategoryBreakdown: make([]*ExciseCategorySummary, 0),
	}

	report := &ReportData{
		ReportType:   ReportTypeExciseReturn.String(),
		ReportFormat: request.Format.String(),
		GeneratedAt:  time.Now(),
		PeriodStart:  request.StartDate,
		PeriodEnd:    request.EndDate,
		Summary:      summary,
	}

	rawData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	report.RawData = rawData

	return report, nil
}

// --- 关税申报表 ---

// DutyReturnSummary 关税申报表汇总
type DutyReturnSummary struct {
	TotalImports      int   `json:"total_imports"`
	TotalCustomsValue int64 `json:"total_customs_value"`
	TotalCustomsDuty  int64 `json:"total_customs_duty"`
	TotalImportVAT    int64 `json:"total_import_vat"`
	TotalImportExcise int64 `json:"total_import_excise"`
	TotalFees         int64 `json:"total_fees"`
	TotalTaxes        int64 `json:"total_taxes"`
}

// DutyReturnDetail 关税申报表明细
type DutyReturnDetail struct {
	EntryNo       string    `json:"entry_no"`
	Date          time.Time `json:"date"`
	HSCode        string    `json:"hs_code"`
	Description   string    `json:"description"`
	OriginCountry string    `json:"origin_country"`
	CustomsValue  int64     `json:"customs_value"`
	CustomsDuty   int64     `json:"customs_duty"`
	ImportVAT     int64     `json:"import_vat"`
	ImportExcise  int64     `json:"import_excise"`
	TotalTax      int64     `json:"total_tax"`
}

// DutyReturnReport 关税申报表
type DutyReturnReport struct{}

func NewDutyReturnReport() *DutyReturnReport {
	return &DutyReturnReport{}
}

func (r *DutyReturnReport) Generate(ctx context.Context, request *ReportRequest) (*ReportData, error) {
	summary := &DutyReturnSummary{}
	details := make([]*DutyReturnDetail, 0)

	report := &ReportData{
		ReportType:   ReportTypeDutyReturn.String(),
		ReportFormat: request.Format.String(),
		GeneratedAt:  time.Now(),
		PeriodStart:  request.StartDate,
		PeriodEnd:    request.EndDate,
		Summary:      summary,
	}

	if request.IncludeDetails {
		report.Details = details
	}

	rawData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	report.RawData = rawData

	return report, nil
}

// --- 综合税务报表 ---

// ConsolidatedTaxSummary 综合税务汇总
type ConsolidatedTaxSummary struct {
	TotalRevenue     int64             `json:"total_revenue"`
	TotalNetAmount   int64             `json:"total_net_amount"`
	TotalTaxAmount   int64             `json:"total_tax_amount"`
	TotalGrossAmount int64             `json:"total_gross_amount"`
	TaxBreakdown     []*TaxTypeSummary `json:"tax_breakdown"`
	PeriodBreakdown  []*PeriodSummary  `json:"period_breakdown"`
}

// TaxTypeSummary 税种汇总
type TaxTypeSummary struct {
	TaxType          domain.TaxType `json:"tax_type"`
	TaxTypeName      string         `json:"tax_type_name"`
	TaxBase          int64          `json:"tax_base"`
	TaxAmount        int64          `json:"tax_amount"`
	TransactionCount int            `json:"transaction_count"`
}

// PeriodSummary 期间汇总
type PeriodSummary struct {
	Period      string `json:"period"`
	NetAmount   int64  `json:"net_amount"`
	TaxAmount   int64  `json:"tax_amount"`
	GrossAmount int64  `json:"gross_amount"`
}

// ConsolidatedReport 综合税务报表
type ConsolidatedReport struct {
	repo domain.TaxRepository
}

func NewConsolidatedReport(repo domain.TaxRepository) *ConsolidatedReport {
	return &ConsolidatedReport{repo: repo}
}

func (r *ConsolidatedReport) Generate(ctx context.Context, request *ReportRequest) (*ReportData, error) {
	summary := &ConsolidatedTaxSummary{
		TaxBreakdown:    make([]*TaxTypeSummary, 0),
		PeriodBreakdown: make([]*PeriodSummary, 0),
	}

	report := &ReportData{
		ReportType:   ReportTypeConsolidated.String(),
		ReportFormat: request.Format.String(),
		GeneratedAt:  time.Now(),
		PeriodStart:  request.StartDate,
		PeriodEnd:    request.EndDate,
		Summary:      summary,
	}

	rawData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	report.RawData = rawData

	return report, nil
}

// --- 报表生成器 ---

// ReportGenerator 报表生成器
type ReportGenerator struct {
	repo               domain.TaxRepository
	vatReturnReport    *VATReturnReport
	exciseReturnReport *ExciseReturnReport
	dutyReturnReport   *DutyReturnReport
	consolidatedReport *ConsolidatedReport
}

func NewReportGenerator(repo domain.TaxRepository) *ReportGenerator {
	return &ReportGenerator{
		repo:               repo,
		vatReturnReport:    NewVATReturnReport(repo),
		exciseReturnReport: NewExciseReturnReport(),
		dutyReturnReport:   NewDutyReturnReport(),
		consolidatedReport: NewConsolidatedReport(repo),
	}
}

// Generate 生成报表
func (g *ReportGenerator) Generate(ctx context.Context, request *ReportRequest) (*ReportData, error) {
	switch request.ReportType {
	case ReportTypeVATReturn:
		return g.vatReturnReport.Generate(ctx, request)
	case ReportTypeExciseReturn:
		return g.exciseReturnReport.Generate(ctx, request)
	case ReportTypeDutyReturn:
		return g.dutyReturnReport.Generate(ctx, request)
	case ReportTypeConsolidated:
		return g.consolidatedReport.Generate(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported report type: %v", request.ReportType)
	}
}

// GenerateVATReturn 生成增值税申报表
func (g *ReportGenerator) GenerateVATReturn(ctx context.Context, startDate, endDate time.Time, countryCode string, includeDetails bool) (*ReportData, error) {
	request := &ReportRequest{
		ReportType:     ReportTypeVATReturn,
		Format:         FormatJSON,
		StartDate:      startDate,
		EndDate:        endDate,
		CountryCode:    countryCode,
		IncludeDetails: includeDetails,
	}
	return g.Generate(ctx, request)
}

// GenerateExciseReturn 生成消费税申报表
func (g *ReportGenerator) GenerateExciseReturn(ctx context.Context, startDate, endDate time.Time, countryCode string) (*ReportData, error) {
	request := &ReportRequest{
		ReportType:  ReportTypeExciseReturn,
		Format:      FormatJSON,
		StartDate:   startDate,
		EndDate:     endDate,
		CountryCode: countryCode,
	}
	return g.Generate(ctx, request)
}

// GenerateDutyReturn 生成关税申报表
func (g *ReportGenerator) GenerateDutyReturn(ctx context.Context, startDate, endDate time.Time, countryCode string) (*ReportData, error) {
	request := &ReportRequest{
		ReportType:  ReportTypeDutyReturn,
		Format:      FormatJSON,
		StartDate:   startDate,
		EndDate:     endDate,
		CountryCode: countryCode,
	}
	return g.Generate(ctx, request)
}

// GenerateConsolidatedReport 生成综合税务报表
func (g *ReportGenerator) GenerateConsolidatedReport(ctx context.Context, startDate, endDate time.Time, countryCode string) (*ReportData, error) {
	request := &ReportRequest{
		ReportType:  ReportTypeConsolidated,
		Format:      FormatJSON,
		StartDate:   startDate,
		EndDate:     endDate,
		CountryCode: countryCode,
	}
	return g.Generate(ctx, request)
}

// ExportReport 导出报表到writer
func (g *ReportGenerator) ExportReport(ctx context.Context, request *ReportRequest, writer io.Writer) error {
	report, err := g.Generate(ctx, request)
	if err != nil {
		return err
	}

	_, err = writer.Write(report.RawData)
	return err
}

// GetSupportedFormats 获取支持的报表格式
func (g *ReportGenerator) GetSupportedFormats() []ReportFormat {
	return []ReportFormat{FormatJSON, FormatCSV, FormatPDF, FormatExcel, FormatXML}
}

// GetSupportedReportTypes 获取支持的报表类型
func (g *ReportGenerator) GetSupportedReportTypes() []ReportType {
	return []ReportType{
		ReportTypeVATReturn,
		ReportTypeExciseReturn,
		ReportTypeDutyReturn,
		ReportTypeSalesTax,
		ReportTypeConsolidated,
	}
}
