// Package thirdparty Vertex税务服务实现
package thirdparty

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VertexProvider Vertex税务服务提供商
type VertexProvider struct {
	config *ProviderConfig
	client *http.Client
}

// NewVertexProvider 创建Vertex提供商
func NewVertexProvider(config *ProviderConfig) *VertexProvider {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &VertexProvider{
		config: config,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// CalculateTax 计算税费
func (p *VertexProvider) CalculateTax(ctx context.Context, request *TaxCalculationRequest) (*TaxCalculationResponse, error) {
	// 构建Vertex请求
	vertexReq := p.buildVertexRequest(request)

	url := fmt.Sprintf("%s/vertex-ws/rest/v1/CalculateTax60", p.config.APIBaseURL)

	bodyJSON, err := json.Marshal(vertexReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("Vertex API error: status=%d, error=%v", resp.StatusCode, errResp)
	}

	var vertexResp vertexCalculateTaxResponse
	if err := json.NewDecoder(resp.Body).Decode(&vertexResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertToTaxResponse(&vertexResp), nil
}

// ValidateAddress 验证地址
func (p *VertexProvider) ValidateAddress(ctx context.Context, address *Address) (*AddressValidationResult, error) {
	url := fmt.Sprintf("%s/vertex-ws/rest/v1/FindTaxAreas60", p.config.APIBaseURL)

	reqBody := vertexFindTaxAreasRequest{
		TaxAreaRequest: vertexTaxAreaRequest{
			PostalAddress: vertexPostalAddress{
				StreetAddress1: address.Line1,
				StreetAddress2: address.Line2,
				City:           address.City,
				MainDivision:   address.Region,
				PostalCode:     address.PostalCode,
				Country:        address.Country,
			},
		},
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Vertex API error: status=%d", resp.StatusCode)
	}

	var validateResp vertexFindTaxAreasResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertToValidationResult(&validateResp), nil
}

// GetTaxRates 获取税率
func (p *VertexProvider) GetTaxRates(ctx context.Context, country, region string) ([]*TaxRateInfo, error) {
	// Vertex没有直接的获取税率API，需要通过CalculateTax估算
	// 创建一个虚拟请求来获取税率
	testReq := &TaxCalculationRequest{
		TransactionDate: time.Now(),
		Addresses: map[string]*Address{
			"ShipFrom": {Country: country, Region: region},
			"ShipTo":   {Country: country, Region: region},
		},
		Lines: []*LineItem{
			{
				LineNo:   "1",
				Quantity: 1,
				Amount:   100.0,
			},
		},
	}

	resp, err := p.CalculateTax(ctx, testReq)
	if err != nil {
		return nil, err
	}

	// 从响应中提取税率信息
	rates := make([]*TaxRateInfo, 0)
	seenRates := make(map[string]bool)

	for _, line := range resp.TaxLines {
		for _, detail := range line.TaxDetails {
			key := fmt.Sprintf("%s-%f", detail.TaxType, detail.Rate)
			if !seenRates[key] {
				seenRates[key] = true
				rates = append(rates, &TaxRateInfo{
					Country:          country,
					Region:           region,
					JurisdictionName: detail.JurisdictionName,
					TaxType:          detail.TaxType,
					Rate:             detail.Rate,
				})
			}
		}
	}

	return rates, nil
}

// HealthCheck 健康检查
func (p *VertexProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/vertex-ws/rest/v1/Ping60", p.config.APIBaseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Basic "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status=%d", resp.StatusCode)
	}

	return nil
}

// ProviderName 提供商名称
func (p *VertexProvider) ProviderName() string {
	return "vertex"
}

// --- Vertex API 结构 ---

// vertexCalculateTaxRequest Vertex计算税务请求
type vertexCalculateTaxRequest struct {
	SaleTransaction []vertexSaleTransaction `json:"saleTransaction"`
}

type vertexSaleTransaction struct {
	TransactionId   string           `json:"transactionId"`
	TransactionDate string           `json:"transactionDate"`
	Seller          vertexSeller     `json:"seller"`
	Customer        vertexCustomer   `json:"customer"`
	LineItems       []vertexLineItem `json:"lineItem"`
	PurchaseOrderNo string           `json:"purchaseOrderNo,omitempty"`
	DocumentNumber  string           `json:"documentNumber,omitempty"`
}

type vertexSeller struct {
	CompanyCode string `json:"companyCode"`
}

type vertexCustomer struct {
	CustomerCode string            `json:"customerCode,omitempty"`
	Destination  vertexDestination `json:"destination"`
}

type vertexDestination struct {
	PostalAddress vertexPostalAddress `json:"postalAddress"`
}

type vertexPostalAddress struct {
	StreetAddress1 string `json:"streetAddress1"`
	StreetAddress2 string `json:"streetAddress2,omitempty"`
	City           string `json:"city"`
	MainDivision   string `json:"mainDivision"`
	PostalCode     string `json:"postalCode"`
	Country        string `json:"country"`
}

type vertexLineItem struct {
	LineItemId     string               `json:"lineItemId"`
	LineItemNumber string               `json:"lineItemNumber"`
	Product        vertexProduct        `json:"product"`
	UnitPrice      float64              `json:"unitPrice"`
	Quantity       vertexQuantity       `json:"quantity"`
	ExtendedPrice  float64              `json:"extendedPrice"`
	FlexibleFields vertexFlexibleFields `json:"flexibleFields,omitempty"`
}

type vertexProduct struct {
	ProductClass string `json:"productClass,omitempty"`
	Value        string `json:"value,omitempty"`
}

type vertexQuantity struct {
	Value float64 `json:"value"`
}

type vertexFlexibleFields struct {
	Field1 string `json:"field1,omitempty"`
	Field2 string `json:"field2,omitempty"`
}

// vertexCalculateTaxResponse Vertex计算税务响应
type vertexCalculateTaxResponse struct {
	SaleTransaction []vertexSaleTransactionResponse `json:"saleTransaction"`
}

type vertexSaleTransactionResponse struct {
	TransactionId   string                   `json:"transactionId"`
	TransactionDate string                   `json:"transactionDate"`
	TotalTax        float64                  `json:"totalTax"`
	Total           float64                  `json:"total"`
	LineItems       []vertexLineItemResponse `json:"lineItem"`
}

type vertexLineItemResponse struct {
	LineItemId     string      `json:"lineItemId"`
	LineItemNumber string      `json:"lineItemNumber"`
	TotalTax       float64     `json:"totalTax"`
	Taxes          []vertexTax `json:"taxes"`
}

type vertexTax struct {
	Jurisdiction  vertexJurisdiction `json:"jurisdiction"`
	CalculatedTax float64            `json:"calculatedTax"`
	EffectiveRate float64            `json:"effectiveRate"`
	Taxable       float64            `json:"taxable"`
	Exempt        float64            `json:"exempt"`
	TaxResult     string             `json:"taxResult"`
	TaxType       string             `json:"taxType"`
}

type vertexJurisdiction struct {
	JurisdictionId    int64  `json:"jurisdictionId"`
	Value             string `json:"value"`
	JurisdictionLevel string `json:"jurisdictionLevel"`
}

// vertexFindTaxAreasRequest 查找税收区域请求
type vertexFindTaxAreasRequest struct {
	TaxAreaRequest vertexTaxAreaRequest `json:"taxAreaRequest"`
}

type vertexTaxAreaRequest struct {
	PostalAddress vertexPostalAddress `json:"postalAddress"`
}

// vertexFindTaxAreasResponse 查找税收区域响应
type vertexFindTaxAreasResponse struct {
	TaxAreaResult []vertexTaxAreaResult `json:"taxAreaResult"`
}

type vertexTaxAreaResult struct {
	TaxAreaId           int64               `json:"taxAreaId"`
	Jurisdiction        vertexJurisdiction  `json:"jurisdiction"`
	PostalAddress       vertexPostalAddress `json:"postalAddress"`
	ConfidenceIndicator string              `json:"confidenceIndicator"`
}

// --- 转换方法 ---

func (p *VertexProvider) buildVertexRequest(req *TaxCalculationRequest) *vertexCalculateTaxRequest {
	// 获取地址
	var shipToAddr *Address
	if addr, ok := req.Addresses["ShipTo"]; ok {
		shipToAddr = addr
	}

	lineItems := make([]vertexLineItem, len(req.Lines))
	for i, line := range req.Lines {
		lineItems[i] = vertexLineItem{
			LineItemId:     line.LineNo,
			LineItemNumber: line.LineNo,
			Product: vertexProduct{
				ProductClass: line.TaxCode,
				Value:        line.ItemCode,
			},
			UnitPrice: line.Amount / line.Quantity,
			Quantity: vertexQuantity{
				Value: line.Quantity,
			},
			ExtendedPrice: line.Amount,
		}
	}

	saleTransaction := vertexSaleTransaction{
		TransactionId:   req.ReferenceCode,
		TransactionDate: req.TransactionDate.Format("2006-01-02"),
		Seller: vertexSeller{
			CompanyCode: p.config.CompanyCode,
		},
		PurchaseOrderNo: req.PurchaseOrderNo,
		DocumentNumber:  req.ReferenceCode,
	}

	if shipToAddr != nil {
		saleTransaction.Customer = vertexCustomer{
			CustomerCode: req.CustomerCode,
			Destination: vertexDestination{
				PostalAddress: vertexPostalAddress{
					StreetAddress1: shipToAddr.Line1,
					StreetAddress2: shipToAddr.Line2,
					City:           shipToAddr.City,
					MainDivision:   shipToAddr.Region,
					PostalCode:     shipToAddr.PostalCode,
					Country:        shipToAddr.Country,
				},
			},
		}
	}

	saleTransaction.LineItems = lineItems

	return &vertexCalculateTaxRequest{
		SaleTransaction: []vertexSaleTransaction{saleTransaction},
	}
}

func (p *VertexProvider) convertToTaxResponse(resp *vertexCalculateTaxResponse) *TaxCalculationResponse {
	if len(resp.SaleTransaction) == 0 {
		return &TaxCalculationResponse{
			ProviderName: p.ProviderName(),
		}
	}

	transaction := resp.SaleTransaction[0]

	taxLines := make([]*TaxLine, len(transaction.LineItems))
	for i, line := range transaction.LineItems {
		details := make([]*TaxDetail, len(line.Taxes))
		for j, tax := range line.Taxes {
			details[j] = &TaxDetail{
				JurisdictionName: tax.Jurisdiction.Value,
				TaxName:          tax.TaxType,
				Rate:             tax.EffectiveRate,
				Tax:              tax.CalculatedTax,
				TaxType:          tax.TaxType,
			}
		}

		totalTaxable := 0.0
		totalExempt := 0.0
		for _, tax := range line.Taxes {
			totalTaxable += tax.Taxable
			totalExempt += tax.Exempt
		}

		taxLines[i] = &TaxLine{
			LineNo:           line.LineItemNumber,
			Tax:              line.TotalTax,
			Taxable:          totalTaxable,
			Exempt:           totalExempt,
			JurisdictionName: "", // Vertex返回多个管辖区的税
			TaxDetails:       details,
		}
	}

	return &TaxCalculationResponse{
		DocumentCode:       transaction.TransactionId,
		TotalAmount:        transaction.Total,
		TotalTax:           transaction.TotalTax,
		TotalTaxable:       transaction.Total - transaction.TotalTax,
		TaxLines:           taxLines,
		Status:             "Calculated",
		ProviderName:       p.ProviderName(),
		ProviderResponseID: transaction.TransactionId,
	}
}

func (p *VertexProvider) convertToValidationResult(resp *vertexFindTaxAreasResponse) *AddressValidationResult {
	if len(resp.TaxAreaResult) == 0 {
		return &AddressValidationResult{
			IsValid:  false,
			Messages: []string{"Address validation failed - no tax area found"},
		}
	}

	result := resp.TaxAreaResult[0]
	isValid := result.ConfidenceIndicator == "EXACT" || result.ConfidenceIndicator == "GOOD"

	addr := &Address{
		Line1:      result.PostalAddress.StreetAddress1,
		Line2:      result.PostalAddress.StreetAddress2,
		City:       result.PostalAddress.City,
		Region:     result.PostalAddress.MainDivision,
		Country:    result.PostalAddress.Country,
		PostalCode: result.PostalAddress.PostalCode,
	}

	return &AddressValidationResult{
		IsValid:           isValid,
		NormalizedAddress: addr,
		Messages:          []string{fmt.Sprintf("Confidence: %s", result.ConfidenceIndicator)},
		ResolutionQuality: result.ConfidenceIndicator,
	}
}
