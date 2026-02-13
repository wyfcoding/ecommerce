package adapter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
)

type TaxCalculatorAdapter struct {
	taxClient TaxClient
	logger    *slog.Logger
}

type TaxClient interface {
	GetHSCodeInfo(ctx context.Context, hsCode string) (*HSCodeInfo, error)
	CalculateDuty(ctx context.Context, req *DutyCalculationRequest) (*DutyCalculationResult, error)
	CalculateVAT(ctx context.Context, req *VATCalculationRequest) (*VATCalculationResult, error)
}

type HSCodeInfo struct {
	HSCode          string
	Description     string
	DutyRate        decimal.Decimal
	VATRate         decimal.Decimal
	ConsumptionTax  decimal.Decimal
	RequiresLicense bool
}

type DutyCalculationRequest struct {
	HSCode             string
	Quantity           int32
	Value              decimal.Decimal
	OriginCountry      string
	DestinationCountry string
	TradeMode          string
}

type DutyCalculationResult struct {
	DutyAmount decimal.Decimal
	DutyRate   decimal.Decimal
	Currency   string
}

type VATCalculationRequest struct {
	HSCode     string
	Value      decimal.Decimal
	DutyAmount decimal.Decimal
	Quantity   int32
}

type VATCalculationResult struct {
	VATAmount decimal.Decimal
	VATRate   decimal.Decimal
	Currency  string
}

func NewTaxCalculatorAdapter(client TaxClient, logger *slog.Logger) domain.TaxCalculatorService {
	return &TaxCalculatorAdapter{
		taxClient: client,
		logger:    logger,
	}
}

func (a *TaxCalculatorAdapter) CalculateDuty(ctx context.Context, items []*domain.DeclarationItem, destinationCountry, originCountry string, tradeMode domain.TradeMode) (*domain.TaxResult, error) {
	a.logger.Info("calculating duty",
		"items_count", len(items),
		"destination", destinationCountry,
		"origin", originCountry,
		"trade_mode", tradeMode.String(),
	)

	result := &domain.TaxResult{
		DutyAmount:     decimal.Zero,
		VATAmount:      decimal.Zero,
		ConsumptionTax: decimal.Zero,
		TotalTax:       decimal.Zero,
		Currency:       "CNY",
		Breakdown:      make([]domain.TaxBreakdown, 0),
	}

	for _, item := range items {
		hsCodeInfo, err := a.taxClient.GetHSCodeInfo(ctx, item.HSCode)
		if err != nil {
			a.logger.Warn("failed to get HS code info, using default rates",
				"hs_code", item.HSCode,
				"error", err,
			)
			hsCodeInfo = &HSCodeInfo{
				HSCode:   item.HSCode,
				DutyRate: decimal.NewFromFloat(0.1),
				VATRate:  decimal.NewFromFloat(0.13),
			}
		}

		itemValue := decimal.NewFromFloat(item.Price * float64(item.Quantity))

		dutyReq := &DutyCalculationRequest{
			HSCode:             item.HSCode,
			Quantity:           item.Quantity,
			Value:              itemValue,
			OriginCountry:      originCountry,
			DestinationCountry: destinationCountry,
			TradeMode:          tradeMode.String(),
		}

		dutyResult, err := a.taxClient.CalculateDuty(ctx, dutyReq)
		if err != nil {
			a.logger.Error("duty calculation failed", "hs_code", item.HSCode, "error", err)
			dutyResult = &DutyCalculationResult{
				DutyAmount: itemValue.Mul(hsCodeInfo.DutyRate),
				DutyRate:   hsCodeInfo.DutyRate,
				Currency:   "CNY",
			}
		}

		vatReq := &VATCalculationRequest{
			HSCode:     item.HSCode,
			Value:      itemValue,
			DutyAmount: dutyResult.DutyAmount,
			Quantity:   item.Quantity,
		}

		vatResult, err := a.taxClient.CalculateVAT(ctx, vatReq)
		if err != nil {
			a.logger.Error("VAT calculation failed", "hs_code", item.HSCode, "error", err)
			vatResult = &VATCalculationResult{
				VATAmount: itemValue.Add(dutyResult.DutyAmount).Mul(hsCodeInfo.VATRate),
				VATRate:   hsCodeInfo.VATRate,
				Currency:  "CNY",
			}
		}

		result.Breakdown = append(result.Breakdown,
			domain.TaxBreakdown{
				TaxType:       "DUTY",
				TaxRate:       dutyResult.DutyRate,
				TaxableAmount: itemValue,
				TaxAmount:     dutyResult.DutyAmount,
			},
			domain.TaxBreakdown{
				TaxType:       "VAT",
				TaxRate:       vatResult.VATRate,
				TaxableAmount: itemValue.Add(dutyResult.DutyAmount),
				TaxAmount:     vatResult.VATAmount,
			},
		)

		result.DutyAmount = result.DutyAmount.Add(dutyResult.DutyAmount)
		result.VATAmount = result.VATAmount.Add(vatResult.VATAmount)
	}

	result.TotalTax = result.DutyAmount.Add(result.VATAmount).Add(result.ConsumptionTax)

	a.logger.Info("duty calculation completed",
		"total_duty", result.DutyAmount,
		"total_vat", result.VATAmount,
		"total_tax", result.TotalTax,
	)

	return result, nil
}

func (a *TaxCalculatorAdapter) CalculateTax(ctx context.Context, items []*domain.DeclarationItem, customsPort string, tradeMode domain.TradeMode) (*domain.TaxResult, error) {
	return a.CalculateDuty(ctx, items, "CN", "", tradeMode)
}

type CustomsGatewayAdapter struct {
	customsClient CustomsClient
	logger        *slog.Logger
}

type CustomsClient interface {
	SubmitDeclaration(ctx context.Context, req *CustomsSubmitRequest) (*CustomsSubmitResponse, error)
	QueryStatus(ctx context.Context, customsDeclNo string) (*CustomsStatusResponse, error)
	QueryResult(ctx context.Context, customsDeclNo string) (*CustomsResultResponse, error)
}

type CustomsSubmitRequest struct {
	DeclarationID string
	OrderID       string
	CustomsCode   string
	TradeMode     string
	Items         []CustomsItem
	TotalValue    decimal.Decimal
	Currency      string
}

type CustomsItem struct {
	HSCode      string
	ProductName string
	Quantity    int32
	Price       float64
	Currency    string
}

type CustomsSubmitResponse struct {
	CustomsDeclarationNo string
	Status               string
	Message              string
}

type CustomsStatusResponse struct {
	CustomsDeclarationNo string
	Status               string
	StatusMessage        string
}

type CustomsResultResponse struct {
	CustomsDeclarationNo string
	Status               string
	Result               string
	Issues               []string
}

func NewCustomsGatewayAdapter(client CustomsClient, logger *slog.Logger) domain.CustomsGatewayService {
	return &CustomsGatewayAdapter{
		customsClient: client,
		logger:        logger,
	}
}

func (a *CustomsGatewayAdapter) SubmitDeclaration(ctx context.Context, decl *domain.CustomsDeclaration) (*domain.CustomsSubmitResult, error) {
	a.logger.Info("submitting declaration to customs",
		"declaration_id", decl.ID,
		"customs_code", decl.CustomsCode,
	)

	items := make([]CustomsItem, len(decl.Items))
	for i, item := range decl.Items {
		items[i] = CustomsItem{
			HSCode:      item.HSCode,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Currency:    "CNY",
		}
	}

	req := &CustomsSubmitRequest{
		DeclarationID: decl.ID,
		OrderID:       decl.OrderID,
		CustomsCode:   decl.CustomsCode,
		TradeMode:     decl.TradeMode.String(),
		Items:         items,
		TotalValue:    decl.TotalValue(),
		Currency:      "CNY",
	}

	resp, err := a.customsClient.SubmitDeclaration(ctx, req)
	if err != nil {
		a.logger.Error("customs submission failed", "error", err)
		return nil, fmt.Errorf("customs submit: %w", err)
	}

	a.logger.Info("declaration submitted to customs",
		"customs_decl_no", resp.CustomsDeclarationNo,
		"status", resp.Status,
	)

	return &domain.CustomsSubmitResult{
		CustomsDeclarationNo: resp.CustomsDeclarationNo,
		Status:               resp.Status,
		Message:              resp.Message,
	}, nil
}

func (a *CustomsGatewayAdapter) QueryStatus(ctx context.Context, customsDeclNo string) (*domain.CustomsStatusResult, error) {
	a.logger.Info("querying customs status", "customs_decl_no", customsDeclNo)

	resp, err := a.customsClient.QueryStatus(ctx, customsDeclNo)
	if err != nil {
		return nil, fmt.Errorf("query status: %w", err)
	}

	return &domain.CustomsStatusResult{
		CustomsDeclarationNo: resp.CustomsDeclarationNo,
		Status:               resp.Status,
		StatusMessage:        resp.StatusMessage,
	}, nil
}

func (a *CustomsGatewayAdapter) QueryResult(ctx context.Context, customsDeclNo string) (*domain.CustomsResultDetail, error) {
	a.logger.Info("querying customs result", "customs_decl_no", customsDeclNo)

	resp, err := a.customsClient.QueryResult(ctx, customsDeclNo)
	if err != nil {
		return nil, fmt.Errorf("query result: %w", err)
	}

	return &domain.CustomsResultDetail{
		CustomsDeclarationNo: resp.CustomsDeclarationNo,
		Status:               resp.Status,
		Result:               resp.Result,
		Issues:               resp.Issues,
	}, nil
}

type DocumentStorageAdapter struct {
	storageClient StorageClient
	logger        *slog.Logger
}

type StorageClient interface {
	Upload(ctx context.Context, bucket, key string, data []byte) (string, error)
	GetURL(ctx context.Context, bucket, key string) (string, error)
	Delete(ctx context.Context, bucket, key string) error
}

func NewDocumentStorageAdapter(client StorageClient, logger *slog.Logger) domain.DocumentStorageService {
	return &DocumentStorageAdapter{
		storageClient: client,
		logger:        logger,
	}
}

func (a *DocumentStorageAdapter) UploadDocument(ctx context.Context, declarationID string, docType domain.CustomsDocumentType, data []byte) (string, error) {
	a.logger.Info("uploading document",
		"declaration_id", declarationID,
		"doc_type", int(docType),
	)

	key := fmt.Sprintf("customs/%s/%d/%d", declarationID, docType, len(data))
	url, err := a.storageClient.Upload(ctx, "customs-documents", key, data)
	if err != nil {
		a.logger.Error("document upload failed", "error", err)
		return "", fmt.Errorf("upload document: %w", err)
	}

	a.logger.Info("document uploaded", "url", url)
	return url, nil
}

func (a *DocumentStorageAdapter) GetDocumentURL(ctx context.Context, documentID string) (string, error) {
	url, err := a.storageClient.GetURL(ctx, "customs-documents", documentID)
	if err != nil {
		return "", fmt.Errorf("get document url: %w", err)
	}
	return url, nil
}

func (a *DocumentStorageAdapter) DeleteDocument(ctx context.Context, documentID string) error {
	a.logger.Info("deleting document", "document_id", documentID)

	if err := a.storageClient.Delete(ctx, "customs-documents", documentID); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	return nil
}

type NotificationAdapter struct {
	notifyClient NotifyClient
	logger       *slog.Logger
}

type NotifyClient interface {
	Send(ctx context.Context, userID uint64, eventType, title, content string, data map[string]string) error
}

func NewNotificationAdapter(client NotifyClient, logger *slog.Logger) domain.NotificationService {
	return &NotificationAdapter{
		notifyClient: client,
		logger:       logger,
	}
}

func (a *NotificationAdapter) NotifyDeclarationSubmitted(ctx context.Context, userID uint64, declarationID string) error {
	a.logger.Info("sending declaration submitted notification", "user_id", userID, "declaration_id", declarationID)

	return a.notifyClient.Send(ctx, userID, "DECLARATION_SUBMITTED",
		"报关单已提交",
		fmt.Sprintf("您的报关单 %s 已提交海关审核", declarationID),
		map[string]string{"declaration_id": declarationID},
	)
}

func (a *NotificationAdapter) NotifyDeclarationCleared(ctx context.Context, userID uint64, declarationID string) error {
	a.logger.Info("sending declaration cleared notification", "user_id", userID, "declaration_id", declarationID)

	return a.notifyClient.Send(ctx, userID, "DECLARATION_CLEARED",
		"报关单已清关",
		fmt.Sprintf("您的报关单 %s 已完成清关", declarationID),
		map[string]string{"declaration_id": declarationID},
	)
}

func (a *NotificationAdapter) NotifyDeclarationRejected(ctx context.Context, userID uint64, declarationID, reason string) error {
	a.logger.Info("sending declaration rejected notification", "user_id", userID, "declaration_id", declarationID)

	return a.notifyClient.Send(ctx, userID, "DECLARATION_REJECTED",
		"报关单被拒绝",
		fmt.Sprintf("您的报关单 %s 被拒绝，原因：%s", declarationID, reason),
		map[string]string{
			"declaration_id": declarationID,
			"reason":         reason,
		},
	)
}
