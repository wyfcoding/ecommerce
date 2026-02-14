package adapter

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wyfcoding/ecommerce/internal/invoice/domain"
)

type TaxPlatformAdapter struct {
	endpoint   string
	appKey     string
	appSecret  string
	merchantID string
	httpClient *http.Client
}

func NewTaxPlatformAdapter(endpoint, appKey, appSecret, merchantID string) *TaxPlatformAdapter {
	return &TaxPlatformAdapter{
		endpoint:   endpoint,
		appKey:     appKey,
		appSecret:  appSecret,
		merchantID: merchantID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (a *TaxPlatformAdapter) IssueInvoice(ctx context.Context, req *domain.IssueInvoiceRequest) (*domain.IssueInvoiceResult, error) {
	apiReq := map[string]any{
		"fpqqlsh":    req.OrderNo,
		"xsf_nsrsbh": a.merchantID,
		"ghfmc":      req.Title.Name,
		"ghf_nsrsbh": req.Title.TaxID,
		"ghfyhzh":    req.Title.Bank + " " + req.Title.Account,
		"ghfdzdh":    req.Title.Address + " " + req.Title.Phone,
		"kplx":       a.getInvoiceTypeCode(req.InvoiceType),
		"hzfs":       a.getMediumCode(req.InvoiceMedium),
		"items":      a.convertItems(req.Items),
		"hjje":       float64(req.Amount) / 100,
		"hjse":       float64(0) / 100,
		"bz":         req.Remark,
		"ghf_email":  req.Title.Email,
		"ghf_sj":     req.Title.ReceiverPhone,
	}

	resp, err := a.doRequest(ctx, "/api/v1/invoice/issue", apiReq)
	if err != nil {
		return nil, fmt.Errorf("issue invoice failed: %w", err)
	}

	return a.parseIssueResult(resp), nil
}

func (a *TaxPlatformAdapter) RedInvoice(ctx context.Context, req *domain.RedInvoiceRequest) (*domain.RedInvoiceResult, error) {
	apiReq := map[string]any{
		"yfp_dm": req.OriginalInvoiceCode,
		"yfp_hm": req.OriginalInvoiceNo,
		"kpyy":   req.Reason,
		"kprq":   time.Now().Format("2006-01-02"),
	}

	resp, err := a.doRequest(ctx, "/api/v1/invoice/red", apiReq)
	if err != nil {
		return nil, fmt.Errorf("red invoice failed: %w", err)
	}

	return a.parseRedResult(resp), nil
}

func (a *TaxPlatformAdapter) VerifyInvoice(ctx context.Context, req *domain.VerifyInvoiceRequest) (*domain.InvoiceVerification, error) {
	apiReq := map[string]any{
		"fp_dm": req.InvoiceCode,
		"fp_hm": req.InvoiceNo,
		"jym":   req.CheckCode,
		"kpje":  float64(req.Amount) / 100,
		"kprq":  req.IssueDate,
	}

	resp, err := a.doRequest(ctx, "/api/v1/invoice/verify", apiReq)
	if err != nil {
		return nil, fmt.Errorf("verify invoice failed: %w", err)
	}

	return a.parseVerifyResult(resp), nil
}

func (a *TaxPlatformAdapter) QueryInvoice(ctx context.Context, req *domain.QueryInvoiceRequest) (*domain.QueryInvoiceResult, error) {
	apiReq := map[string]any{
		"fp_dm": req.InvoiceCode,
		"fp_hm": req.InvoiceNo,
	}

	resp, err := a.doRequest(ctx, "/api/v1/invoice/query", apiReq)
	if err != nil {
		return nil, fmt.Errorf("query invoice failed: %w", err)
	}

	return a.parseQueryResult(resp), nil
}

func (a *TaxPlatformAdapter) DownloadInvoice(ctx context.Context, req *domain.DownloadInvoiceRequest) (*domain.DownloadInvoiceResult, error) {
	apiReq := map[string]any{
		"fp_dm":     req.InvoiceCode,
		"fp_hm":     req.InvoiceNo,
		"file_type": req.FileType,
	}

	resp, err := a.doRequest(ctx, "/api/v1/invoice/download", apiReq)
	if err != nil {
		return nil, fmt.Errorf("download invoice failed: %w", err)
	}

	return a.parseDownloadResult(resp), nil
}

func (a *TaxPlatformAdapter) doRequest(ctx context.Context, path string, body any) (map[string]any, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	signature := a.generateSignature(string(jsonBody), timestamp)

	url := fmt.Sprintf("%s%s", a.endpoint, path)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Key", a.appKey)
	req.Header.Set("X-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-Signature", signature)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s", string(respBody))
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (a *TaxPlatformAdapter) generateSignature(body string, timestamp int64) string {
	h := hmac.New(sha256.New, []byte(a.appSecret))
	h.Write(fmt.Appendf(nil, "%s%d", body, timestamp))
	return hex.EncodeToString(h.Sum(nil))
}

func (a *TaxPlatformAdapter) getInvoiceTypeCode(invType domain.InvoiceType) int {
	switch invType {
	case domain.InvoiceTypePersonalNormal:
		return 0
	case domain.InvoiceTypeCompanyNormal:
		return 0
	case domain.InvoiceTypeCompanySpecial:
		return 1
	default:
		return 0
	}
}

func (a *TaxPlatformAdapter) getMediumCode(medium domain.InvoiceMedium) int {
	switch medium {
	case domain.InvoiceMediumElectronic:
		return 2
	case domain.InvoiceMediumPaper:
		return 1
	default:
		return 2
	}
}

func (a *TaxPlatformAdapter) convertItems(items []domain.InvoiceItemRequest) []map[string]any {
	result := make([]map[string]any, len(items))
	for i, item := range items {
		result[i] = map[string]any{
			"xm":   item.ProductName,
			"ggxh": item.Spec,
			"dw":   item.Unit,
			"sl":   item.Quantity,
			"dj":   float64(item.Price) / 100,
			"je":   float64(item.Amount) / 100,
			"slv":  item.TaxRate,
			"se":   float64(item.TaxAmount) / 100,
		}
	}
	return result
}

func (a *TaxPlatformAdapter) parseIssueResult(resp map[string]any) *domain.IssueInvoiceResult {
	result := &domain.IssueInvoiceResult{}

	if data, ok := resp["data"].(map[string]any); ok {
		if code, ok := data["fp_dm"].(string); ok {
			result.InvoiceCode = code
		}
		if no, ok := data["fp_hm"].(string); ok {
			result.InvoiceNo = no
		}
		if checkCode, ok := data["jym"].(string); ok {
			result.CheckCode = checkCode
		}
		if pdfUrl, ok := data["pdf_url"].(string); ok {
			result.PDFUrl = pdfUrl
		}
		if xmlUrl, ok := data["xml_url"].(string); ok {
			result.XMLUrl = xmlUrl
		}
		if issuedAt, ok := data["kprq"].(string); ok {
			result.IssuedAt = issuedAt
		}
	}

	return result
}

func (a *TaxPlatformAdapter) parseRedResult(resp map[string]any) *domain.RedInvoiceResult {
	result := &domain.RedInvoiceResult{}

	if data, ok := resp["data"].(map[string]any); ok {
		if code, ok := data["fp_dm"].(string); ok {
			result.RedInvoiceCode = code
		}
		if no, ok := data["fp_hm"].(string); ok {
			result.RedInvoiceNo = no
		}
		if checkCode, ok := data["jym"].(string); ok {
			result.CheckCode = checkCode
		}
		if pdfUrl, ok := data["pdf_url"].(string); ok {
			result.PDFUrl = pdfUrl
		}
		if xmlUrl, ok := data["xml_url"].(string); ok {
			result.XMLUrl = xmlUrl
		}
	}

	return result
}

func (a *TaxPlatformAdapter) parseVerifyResult(resp map[string]any) *domain.InvoiceVerification {
	result := &domain.InvoiceVerification{}

	if data, ok := resp["data"].(map[string]any); ok {
		if code, ok := data["fp_dm"].(string); ok {
			result.InvoiceCode = code
		}
		if no, ok := data["fp_hm"].(string); ok {
			result.InvoiceNo = no
		}
		if valid, ok := data["valid"].(bool); ok {
			result.Valid = valid
		}
		if status, ok := data["fpzt"].(string); ok {
			result.InvoiceStatus = status
		}
		if sellerName, ok := data["xsfmc"].(string); ok {
			result.SellerName = sellerName
		}
		if sellerTaxID, ok := data["xsf_nsrsbh"].(string); ok {
			result.SellerTaxID = sellerTaxID
		}
		if buyerName, ok := data["ghfmc"].(string); ok {
			result.BuyerName = buyerName
		}
		if buyerTaxID, ok := data["ghf_nsrsbh"].(string); ok {
			result.BuyerTaxID = buyerTaxID
		}
		if amount, ok := data["kpje"].(float64); ok {
			result.Amount = int64(amount * 100)
		}
		if taxAmount, ok := data["kpse"].(float64); ok {
			result.TaxAmount = int64(taxAmount * 100)
		}
		if issueDate, ok := data["kprq"].(string); ok {
			result.IssueDate = issueDate
		}
		if invalidationMark, ok := data["zfbs"].(string); ok {
			result.InvalidationMark = invalidationMark
		}
	}

	result.VerifyTime = time.Now()
	return result
}

func (a *TaxPlatformAdapter) parseQueryResult(resp map[string]any) *domain.QueryInvoiceResult {
	result := &domain.QueryInvoiceResult{}

	if data, ok := resp["data"].(map[string]any); ok {
		if code, ok := data["fp_dm"].(string); ok {
			result.InvoiceCode = code
		}
		if no, ok := data["fp_hm"].(string); ok {
			result.InvoiceNo = no
		}
		if status, ok := data["fpzt"].(string); ok {
			result.Status = status
		}
		if pdfUrl, ok := data["pdf_url"].(string); ok {
			result.PDFUrl = pdfUrl
		}
		if xmlUrl, ok := data["xml_url"].(string); ok {
			result.XMLUrl = xmlUrl
		}
	}

	return result
}

func (a *TaxPlatformAdapter) parseDownloadResult(resp map[string]any) *domain.DownloadInvoiceResult {
	result := &domain.DownloadInvoiceResult{}

	if data, ok := resp["data"].(map[string]any); ok {
		if fileUrl, ok := data["file_url"].(string); ok {
			result.FileUrl = fileUrl
		}
		if fileData, ok := data["file_data"].([]byte); ok {
			result.FileData = fileData
		}
		if expireTime, ok := data["expire_time"].(string); ok {
			result.ExpireTime = expireTime
		}
	}

	return result
}

type MockInvoicePlatform struct{}

func NewMockInvoicePlatform() *MockInvoicePlatform {
	return &MockInvoicePlatform{}
}

func (m *MockInvoicePlatform) IssueInvoice(ctx context.Context, req *domain.IssueInvoiceRequest) (*domain.IssueInvoiceResult, error) {
	return &domain.IssueInvoiceResult{
		InvoiceCode: "044001900111",
		InvoiceNo:   fmt.Sprintf("%d", time.Now().UnixNano()%100000000),
		CheckCode:   fmt.Sprintf("%d", time.Now().UnixNano()%100000000000),
		PDFUrl:      "https://mock.example.com/invoice.pdf",
		XMLUrl:      "https://mock.example.com/invoice.xml",
		IssuedAt:    time.Now().Format("2006-01-02"),
	}, nil
}

func (m *MockInvoicePlatform) RedInvoice(ctx context.Context, req *domain.RedInvoiceRequest) (*domain.RedInvoiceResult, error) {
	return &domain.RedInvoiceResult{
		RedInvoiceCode: "044001900111",
		RedInvoiceNo:   fmt.Sprintf("%d", time.Now().UnixNano()%100000000),
		CheckCode:      fmt.Sprintf("%d", time.Now().UnixNano()%100000000000),
		PDFUrl:         "https://mock.example.com/red_invoice.pdf",
		XMLUrl:         "https://mock.example.com/red_invoice.xml",
	}, nil
}

func (m *MockInvoicePlatform) VerifyInvoice(ctx context.Context, req *domain.VerifyInvoiceRequest) (*domain.InvoiceVerification, error) {
	return &domain.InvoiceVerification{
		InvoiceCode:   req.InvoiceCode,
		InvoiceNo:     req.InvoiceNo,
		Valid:         true,
		VerifyTime:    time.Now(),
		InvoiceStatus: "有效",
		SellerName:    "测试商家",
		SellerTaxID:   "91110000000000000X",
		BuyerName:     "测试买家",
		BuyerTaxID:    "91110000000000001Y",
		Amount:        req.Amount,
		IssueDate:     req.IssueDate,
	}, nil
}

func (m *MockInvoicePlatform) QueryInvoice(ctx context.Context, req *domain.QueryInvoiceRequest) (*domain.QueryInvoiceResult, error) {
	return &domain.QueryInvoiceResult{
		InvoiceCode: req.InvoiceCode,
		InvoiceNo:   req.InvoiceNo,
		Status:      "有效",
		PDFUrl:      "https://mock.example.com/invoice.pdf",
		XMLUrl:      "https://mock.example.com/invoice.xml",
	}, nil
}

func (m *MockInvoicePlatform) DownloadInvoice(ctx context.Context, req *domain.DownloadInvoiceRequest) (*domain.DownloadInvoiceResult, error) {
	return &domain.DownloadInvoiceResult{
		FileUrl:    "https://mock.example.com/invoice." + req.FileType,
		ExpireTime: time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
	}, nil
}
