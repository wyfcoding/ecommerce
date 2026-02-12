// Package thirdparty Avalara税务服务实现
package thirdparty

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AvalaraProvider Avalara税务服务提供商
type AvalaraProvider struct {
	config *ProviderConfig
	client *http.Client
}

// NewAvalaraProvider 创建Avalara提供商
func NewAvalaraProvider(config *ProviderConfig) *AvalaraProvider {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &AvalaraProvider{
		config: config,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// CalculateTax 计算税费
func (p *AvalaraProvider) CalculateTax(ctx context.Context, request *TaxCalculationRequest) (*TaxCalculationResponse, error) {
	// 构建Avalara请求
	avalaraReq := p.buildAvalaraRequest(request)

	url := fmt.Sprintf("%s/api/v2/transactions/create", p.config.APIBaseURL)

	bodyJSON, err := json.Marshal(avalaraReq)
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

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("Avalara API error: status=%d, error=%v", resp.StatusCode, errResp)
	}

	var avalaraResp avalaraTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&avalaraResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertToTaxResponse(&avalaraResp), nil
}

// ValidateAddress 验证地址
func (p *AvalaraProvider) ValidateAddress(ctx context.Context, address *Address) (*AddressValidationResult, error) {
	url := fmt.Sprintf("%s/api/v2/addresses/resolve", p.config.APIBaseURL)

	params := fmt.Sprintf("?line1=%s&city=%s&region=%s&country=%s&postalCode=%s",
		address.Line1, address.City, address.Region, address.Country, address.PostalCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url+params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Avalara API error: status=%d", resp.StatusCode)
	}

	var validateResp avalaraAddressValidationResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertToValidationResult(&validateResp), nil
}

// GetTaxRates 获取税率
func (p *AvalaraProvider) GetTaxRates(ctx context.Context, country, region string) ([]*TaxRateInfo, error) {
	url := fmt.Sprintf("%s/api/v2/taxrates/bycountry/%s/region/%s", p.config.APIBaseURL, country, region)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Avalara API error: status=%d", resp.StatusCode)
	}

	var ratesResp avalaraRatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&ratesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return p.convertToTaxRates(&ratesResp), nil
}

// HealthCheck 健康检查
func (p *AvalaraProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v2/utilities/ping", p.config.APIBaseURL)

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
func (p *AvalaraProvider) ProviderName() string {
	return "avalara"
}

// --- Avalara API 结构 ---

// avalaraCreateTransactionRequest Avalara创建交易请求
type avalaraCreateTransactionRequest struct {
	Type            string                    `json:"type"`
	CompanyCode     string                    `json:"companyCode"`
	Date            string                    `json:"date"`
	CustomerCode    string                    `json:"customerCode"`
	PurchaseOrderNo string                    `json:"purchaseOrderNo,omitempty"`
	ReferenceCode   string                    `json:"referenceCode,omitempty"`
	EntityUseCode   string                    `json:"entityUseCode,omitempty"`
	Addresses       map[string]avalaraAddress `json:"addresses"`
	Lines           []avalaraLineItem         `json:"lines"`
	Commit          bool                      `json:"commit"`
}

type avalaraAddress struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	Line3      string `json:"line3,omitempty"`
	City       string `json:"city"`
	Region     string `json:"region"`
	Country    string `json:"country"`
	PostalCode string `json:"postalCode"`
}

type avalaraLineItem struct {
	Number             string  `json:"number"`
	ItemCode           string  `json:"itemCode,omitempty"`
	Description        string  `json:"description,omitempty"`
	TaxCode            string  `json:"taxCode,omitempty"`
	Quantity           float64 `json:"quantity"`
	Amount             float64 `json:"amount"`
	Discounted         bool    `json:"discounted,omitempty"`
	OriginAddress      string  `json:"originAddress,omitempty"`
	DestinationAddress string  `json:"destinationAddress,omitempty"`
}

// avalaraTransactionResponse Avalara交易响应
type avalaraTransactionResponse struct {
	ID            int64                    `json:"id"`
	Code          string                   `json:"code"`
	CompanyID     int64                    `json:"companyId"`
	Date          string                   `json:"date"`
	Status        string                   `json:"status"`
	Type          string                   `json:"type"`
	TotalAmount   float64                  `json:"totalAmount"`
	TotalTax      float64                  `json:"totalTax"`
	TotalTaxable  float64                  `json:"totalTaxable"`
	TotalExempt   float64                  `json:"totalExempt"`
	TotalDiscount float64                  `json:"totalDiscount"`
	Lines         []avalaraLineResponse    `json:"lines"`
	Summary       []avalaraSummary         `json:"summary"`
	Addresses     []avalaraAddressResponse `json:"addresses"`
}

type avalaraLineResponse struct {
	ID            int64                   `json:"id"`
	TransactionID int64                   `json:"transactionId"`
	LineNumber    string                  `json:"lineNumber"`
	ItemCode      string                  `json:"itemCode,omitempty"`
	Description   string                  `json:"description,omitempty"`
	TaxCode       string                  `json:"taxCode,omitempty"`
	Quantity      float64                 `json:"quantity"`
	Amount        float64                 `json:"amount"`
	TaxableAmount float64                 `json:"taxableAmount"`
	ExemptAmount  float64                 `json:"exemptAmount"`
	Tax           float64                 `json:"tax"`
	Details       []avalaraDetailResponse `json:"details"`
}

type avalaraDetailResponse struct {
	ID                int64   `json:"id"`
	TransactionLineID int64   `json:"transactionLineId"`
	Country           string  `json:"country"`
	Region            string  `json:"region"`
	JurisType         string  `json:"jurisType"`
	JurisCode         string  `json:"jurisCode"`
	JurisName         string  `json:"jurisName"`
	TaxAuthorityType  int     `json:"taxAuthorityType"`
	TaxType           string  `json:"taxType"`
	RateType          string  `json:"rateType"`
	Rate              float64 `json:"rate"`
	Tax               float64 `json:"tax"`
	TaxableAmount     float64 `json:"taxableAmount"`
}

type avalaraSummary struct {
	Country          string  `json:"country"`
	Region           string  `json:"region"`
	JurisType        string  `json:"jurisType"`
	JurisCode        string  `json:"jurisCode"`
	JurisName        string  `json:"jurisName"`
	TaxAuthorityType int     `json:"taxAuthorityType"`
	TaxType          string  `json:"taxType"`
	RateType         string  `json:"rateType"`
	Rate             float64 `json:"rate"`
	Tax              float64 `json:"tax"`
	TaxableAmount    float64 `json:"taxableAmount"`
}

type avalaraAddressResponse struct {
	ID            int64   `json:"id"`
	TransactionID int64   `json:"transactionId"`
	BoundaryLevel string  `json:"boundaryLevel,omitempty"`
	Line1         string  `json:"line1"`
	Line2         string  `json:"line2,omitempty"`
	Line3         string  `json:"line3,omitempty"`
	City          string  `json:"city"`
	Region        string  `json:"region"`
	Country       string  `json:"country"`
	PostalCode    string  `json:"postalCode"`
	Latitude      float64 `json:"latitude,omitempty"`
	Longitude     float64 `json:"longitude,omitempty"`
	TaxRegionID   int64   `json:"taxRegionId,omitempty"`
	JurisCode     string  `json:"jurisCode,omitempty"`
}

// avalaraAddressValidationResponse 地址验证响应
type avalaraAddressValidationResponse struct {
	Address           *avalaraValidatedAddress `json:"address,omitempty"`
	Coordinates       *avalaraCoordinates      `json:"coordinates,omitempty"`
	ResolutionQuality string                   `json:"resolutionQuality,omitempty"`
	TaxAuthorities    []avalaraTaxAuthority    `json:"taxAuthorities,omitempty"`
	Messages          []avalaraMessage         `json:"messages,omitempty"`
}

type avalaraValidatedAddress struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	Line3      string `json:"line3,omitempty"`
	City       string `json:"city"`
	Region     string `json:"region"`
	Country    string `json:"country"`
	PostalCode string `json:"postalCode"`
}

type avalaraCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type avalaraTaxAuthority struct {
	JurisdictionName string `json:"jurisdictionName"`
	JurisdictionType string `json:"jurisdictionType"`
}

type avalaraMessage struct {
	Summary  string `json:"summary"`
	Details  string `json:"details"`
	RefersTo string `json:"refersTo"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
}

// avalaraRatesResponse 税率响应
type avalaraRatesResponse struct {
	TotalRate float64       `json:"totalRate"`
	Rates     []avalaraRate `json:"rates"`
}

type avalaraRate struct {
	Rate         float64              `json:"rate"`
	Name         string               `json:"name"`
	Type         string               `json:"type"`
	Jurisdiction *avalaraJurisdiction `json:"jurisdiction,omitempty"`
}

type avalaraJurisdiction struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	Name    string `json:"name"`
}

// --- 转换方法 ---

func (p *AvalaraProvider) buildAvalaraRequest(req *TaxCalculationRequest) *avalaraCreateTransactionRequest {
	addresses := make(map[string]avalaraAddress)
	for code, addr := range req.Addresses {
		addresses[code] = avalaraAddress{
			Line1:      addr.Line1,
			Line2:      addr.Line2,
			Line3:      addr.Line3,
			City:       addr.City,
			Region:     addr.Region,
			Country:    addr.Country,
			PostalCode: addr.PostalCode,
		}
	}

	lines := make([]avalaraLineItem, len(req.Lines))
	for i, line := range req.Lines {
		lines[i] = avalaraLineItem{
			Number:      line.LineNo,
			ItemCode:    line.ItemCode,
			Description: line.ItemDescription,
			TaxCode:     line.TaxCode,
			Quantity:    line.Quantity,
			Amount:      line.Amount,
			Discounted:  line.Discounted,
		}
		if line.OriginAddress != nil {
			lines[i].OriginAddress = "ShipFrom"
		}
		if line.DestinationAddress != nil {
			lines[i].DestinationAddress = "ShipTo"
		}
	}

	return &avalaraCreateTransactionRequest{
		Type:            req.DocumentType,
		CompanyCode:     p.config.CompanyCode,
		Date:            req.TransactionDate.Format("2006-01-02"),
		CustomerCode:    req.CustomerCode,
		PurchaseOrderNo: req.PurchaseOrderNo,
		ReferenceCode:   req.ReferenceCode,
		EntityUseCode:   req.EntityUseCode,
		Addresses:       addresses,
		Lines:           lines,
		Commit:          req.Commit,
	}
}

func (p *AvalaraProvider) convertToTaxResponse(resp *avalaraTransactionResponse) *TaxCalculationResponse {
	taxLines := make([]*TaxLine, len(resp.Lines))
	for i, line := range resp.Lines {
		details := make([]*TaxDetail, len(line.Details))
		for j, detail := range line.Details {
			details[j] = &TaxDetail{
				JurisdictionName: detail.JurisName,
				TaxName:          detail.TaxType,
				Rate:             detail.Rate,
				Tax:              detail.Tax,
				TaxType:          detail.TaxType,
			}
		}

		taxLines[i] = &TaxLine{
			LineNo:           line.LineNumber,
			TaxCode:          line.TaxCode,
			TaxName:          line.TaxCode,
			Rate:             0, // 计算平均税率
			Taxable:          line.TaxableAmount,
			Tax:              line.Tax,
			Exempt:           line.ExemptAmount,
			JurisdictionName: "", // 从details中提取
			TaxDetails:       details,
		}
	}

	taxAddresses := make([]*TaxAddress, len(resp.Addresses))
	for i, addr := range resp.Addresses {
		taxAddresses[i] = &TaxAddress{
			AddressCode: "Location" + string(rune('0'+i)),
			Line1:       addr.Line1,
			Line2:       addr.Line2,
			Line3:       addr.Line3,
			City:        addr.City,
			Region:      addr.Region,
			Country:     addr.Country,
			PostalCode:  addr.PostalCode,
			Latitude:    addr.Latitude,
			Longitude:   addr.Longitude,
			TaxRegionID: addr.TaxRegionID,
			JurisCode:   addr.JurisCode,
		}
	}

	return &TaxCalculationResponse{
		DocumentCode:       resp.Code,
		TotalAmount:        resp.TotalAmount,
		TotalTax:           resp.TotalTax,
		TotalTaxable:       resp.TotalTaxable,
		TotalExempt:        resp.TotalExempt,
		TotalDiscount:      resp.TotalDiscount,
		TaxLines:           taxLines,
		TaxAddresses:       taxAddresses,
		Status:             resp.Status,
		ProviderName:       p.ProviderName(),
		ProviderResponseID: fmt.Sprintf("%d", resp.ID),
	}
}

func (p *AvalaraProvider) convertToValidationResult(resp *avalaraAddressValidationResponse) *AddressValidationResult {
	if resp.Address == nil {
		return &AddressValidationResult{
			IsValid:  false,
			Messages: []string{"Address validation failed"},
		}
	}

	addr := &Address{
		Line1:      resp.Address.Line1,
		Line2:      resp.Address.Line2,
		Line3:      resp.Address.Line3,
		City:       resp.Address.City,
		Region:     resp.Address.Region,
		Country:    resp.Address.Country,
		PostalCode: resp.Address.PostalCode,
	}

	if resp.Coordinates != nil {
		addr.Latitude = resp.Coordinates.Latitude
		addr.Longitude = resp.Coordinates.Longitude
	}

	messages := make([]string, len(resp.Messages))
	for i, msg := range resp.Messages {
		messages[i] = fmt.Sprintf("%s: %s", msg.Severity, msg.Summary)
	}

	return &AddressValidationResult{
		IsValid:           resp.ResolutionQuality != "",
		NormalizedAddress: addr,
		Messages:          messages,
		ResolutionQuality: resp.ResolutionQuality,
	}
}

func (p *AvalaraProvider) convertToTaxRates(resp *avalaraRatesResponse) []*TaxRateInfo {
	rates := make([]*TaxRateInfo, len(resp.Rates))
	for i, rate := range resp.Rates {
		rates[i] = &TaxRateInfo{
			TaxType:          rate.Type,
			Rate:             rate.Rate,
			JurisdictionName: rate.Name,
		}
		if rate.Jurisdiction != nil {
			rates[i].Country = rate.Jurisdiction.Country
			rates[i].Region = rate.Jurisdiction.Region
		}
	}
	return rates
}
