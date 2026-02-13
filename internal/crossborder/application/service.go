// 变更说明：完善跨境电商应用层服务，增加完整的命令和查询服务
package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// CrossBorderCommandService 跨境电商命令服务
type CrossBorderCommandService struct {
	declRepo       domain.CrossBorderRepository
	hsCodeRepo     domain.HSCodeRepository
	orderRepo      domain.CrossBorderOrderRepository
	docRepo        domain.CustomsDocumentRepository
	eventRepo      domain.ClearanceEventRepository
	readRepo       domain.CrossBorderReadRepository
	taxService     domain.TaxCalculatorService
	customsGateway domain.CustomsGatewayService
	docStorage     domain.DocumentStorageService
	notification   domain.NotificationService
	publisher      messagequeue.EventPublisher
	topic          string
	logger         *slog.Logger
}

// NewCrossBorderCommandService 创建命令服务
func NewCrossBorderCommandService(
	declRepo domain.CrossBorderRepository,
	hsCodeRepo domain.HSCodeRepository,
	orderRepo domain.CrossBorderOrderRepository,
	docRepo domain.CustomsDocumentRepository,
	eventRepo domain.ClearanceEventRepository,
	readRepo domain.CrossBorderReadRepository,
	taxService domain.TaxCalculatorService,
	customsGateway domain.CustomsGatewayService,
	docStorage domain.DocumentStorageService,
	notification domain.NotificationService,
	publisher messagequeue.EventPublisher,
	topic string,
	logger *slog.Logger,
) *CrossBorderCommandService {
	return &CrossBorderCommandService{
		declRepo:       declRepo,
		hsCodeRepo:     hsCodeRepo,
		orderRepo:      orderRepo,
		docRepo:        docRepo,
		eventRepo:      eventRepo,
		readRepo:       readRepo,
		taxService:     taxService,
		customsGateway: customsGateway,
		docStorage:     docStorage,
		notification:   notification,
		publisher:      publisher,
		topic:          topic,
		logger:         logger,
	}
}

// CreateDeclarationCommand 创建报关单命令
type CreateDeclarationCommand struct {
	OrderID            string
	UserID             uint64
	MerchantID         uint64
	LogisticsNo        string
	DeclaredValue      decimal.Decimal
	Currency           string
	CustomsPort        string
	TradeMode          domain.TradeMode
	OriginCountry      string
	DestinationCountry string
	RecipientName      string
	RecipientPhone     string
	RecipientAddress   string
	RecipientIDNumber  string
	Items              []DeclarationItemCmd
}

// DeclarationItemCmd 报关明细命令
type DeclarationItemCmd struct {
	SKUID       string
	ProductName string
	HSCode      string
	Price       decimal.Decimal
	Quantity    int32
	Weight      decimal.Decimal
}

// CreateDeclaration 创建报关单
func (s *CrossBorderCommandService) CreateDeclaration(ctx context.Context, cmd CreateDeclarationCommand) (*domain.CustomsDeclaration, error) {
	s.logger.InfoContext(ctx, "creating declaration", "order_id", cmd.OrderID)

	decl := domain.NewDeclaration(cmd.OrderID, cmd.UserID, cmd.Currency, cmd.DeclaredValue)
	decl.MerchantID = cmd.MerchantID
	decl.LogisticsNo = cmd.LogisticsNo
	decl.CustomsPort = cmd.CustomsPort
	decl.TradeMode = cmd.TradeMode
	decl.OriginCountry = cmd.OriginCountry
	decl.DestinationCountry = cmd.DestinationCountry
	decl.RecipientName = cmd.RecipientName
	decl.RecipientPhone = cmd.RecipientPhone
	decl.RecipientAddress = cmd.RecipientAddress
	decl.RecipientIDNumber = cmd.RecipientIDNumber

	for _, item := range cmd.Items {
		decl.AddItem(item.SKUID, item.ProductName, item.HSCode, item.Price, item.Quantity, item.Weight)
	}

	hsCodes, err := s.getHSCodes(ctx, cmd.Items)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get hs codes", "error", err)
	}

	if len(hsCodes) > 0 {
		if err := decl.CalculateTax(hsCodes); err != nil {
			s.logger.WarnContext(ctx, "failed to calculate tax", "error", err)
		}
	}

	if err := s.declRepo.SaveDeclaration(ctx, decl); err != nil {
		return nil, fmt.Errorf("failed to save declaration: %w", err)
	}

	s.publishEvents(ctx, decl.GetDomainEvents())
	decl.ClearDomainEvents()

	s.logger.InfoContext(ctx, "declaration created", "declaration_id", decl.DeclarationID)
	return decl, nil
}

// SubmitDeclarationCommand 提交报关命令
type SubmitDeclarationCommand struct {
	DeclarationID string
}

// SubmitDeclaration 提交报关
func (s *CrossBorderCommandService) SubmitDeclaration(ctx context.Context, cmd SubmitDeclarationCommand) error {
	s.logger.InfoContext(ctx, "submitting declaration", "declaration_id", cmd.DeclarationID)

	decl, err := s.declRepo.GetDeclaration(ctx, cmd.DeclarationID)
	if err != nil {
		return err
	}
	if decl == nil {
		return errors.New("declaration not found")
	}

	if err := decl.Submit(); err != nil {
		return err
	}

	if err := s.declRepo.UpdateDeclaration(ctx, decl); err != nil {
		return err
	}

	s.publishEvents(ctx, decl.GetDomainEvents())
	decl.ClearDomainEvents()

	if s.notification != nil {
		_ = s.notification.NotifyDeclarationSubmitted(ctx, decl.UserID, decl.DeclarationID)
	}

	return nil
}

// SubmitCustomsDeclarationCommand 提交海关申报命令
type SubmitCustomsDeclarationCommand struct {
	DeclarationID string
	CustomsCode   string
	CustomsName   string
	DocumentIDs   []string
}

// SubmitCustomsDeclaration 提交海关申报
func (s *CrossBorderCommandService) SubmitCustomsDeclaration(ctx context.Context, cmd SubmitCustomsDeclarationCommand) (*domain.CustomsSubmitResult, error) {
	s.logger.InfoContext(ctx, "submitting customs declaration", "declaration_id", cmd.DeclarationID)

	decl, err := s.declRepo.GetDeclaration(ctx, cmd.DeclarationID)
	if err != nil {
		return nil, err
	}

	if err := decl.StartCustomsProcessing(cmd.CustomsCode, cmd.CustomsName); err != nil {
		return nil, err
	}

	var result *domain.CustomsSubmitResult
	if s.customsGateway != nil {
		result, err = s.customsGateway.SubmitDeclaration(ctx, decl)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to submit to customs", "error", err)
			_ = decl.CustomsReject(err.Error())
		} else {
			decl.CustomsDeclNo = result.CustomsDeclarationNo
		}
	} else {
		result = &domain.CustomsSubmitResult{
			CustomsDeclarationNo: fmt.Sprintf("CUS%d", time.Now().UnixNano()),
			Status:               "SUBMITTED",
			Message:              "Mock submission successful",
		}
		decl.CustomsDeclNo = result.CustomsDeclarationNo
	}

	if err := s.declRepo.UpdateDeclaration(ctx, decl); err != nil {
		return nil, err
	}

	for _, event := range decl.ClearanceEvents {
		_ = s.eventRepo.Save(ctx, &event)
	}

	return result, nil
}

// StartClearanceCommand 开始清关命令
type StartClearanceCommand struct {
	DeclarationID string
	CustomsPort   string
	AgentCode     string
	AgentName     string
}

// StartClearance 开始清关
func (s *CrossBorderCommandService) StartClearance(ctx context.Context, cmd StartClearanceCommand) (*domain.CustomsDeclaration, error) {
	s.logger.InfoContext(ctx, "starting clearance", "declaration_id", cmd.DeclarationID)

	decl, err := s.declRepo.GetDeclaration(ctx, cmd.DeclarationID)
	if err != nil {
		return nil, err
	}

	clearanceID := fmt.Sprintf("CLR%d", time.Now().UnixNano())
	if err := decl.StartClearance(clearanceID); err != nil {
		return nil, err
	}

	if err := s.declRepo.UpdateDeclaration(ctx, decl); err != nil {
		return nil, err
	}

	for _, event := range decl.ClearanceEvents {
		_ = s.eventRepo.Save(ctx, &event)
	}

	return decl, nil
}

// CompleteClearanceCommand 完成清关命令
type CompleteClearanceCommand struct {
	DeclarationID  string
	ReleaseOrderNo string
}

// CompleteClearance 完成清关
func (s *CrossBorderCommandService) CompleteClearance(ctx context.Context, cmd CompleteClearanceCommand) error {
	s.logger.InfoContext(ctx, "completing clearance", "declaration_id", cmd.DeclarationID)

	decl, err := s.declRepo.GetDeclaration(ctx, cmd.DeclarationID)
	if err != nil {
		return err
	}

	if err := decl.CompleteClearance(cmd.ReleaseOrderNo); err != nil {
		return err
	}

	if err := s.declRepo.UpdateDeclaration(ctx, decl); err != nil {
		return err
	}

	for _, event := range decl.ClearanceEvents {
		_ = s.eventRepo.Save(ctx, &event)
	}

	s.publishEvents(ctx, decl.GetDomainEvents())
	decl.ClearDomainEvents()

	if s.readRepo != nil {
		_ = s.readRepo.SaveDeclaration(ctx, decl)
	}

	if s.notification != nil {
		_ = s.notification.NotifyDeclarationCleared(ctx, decl.UserID, decl.DeclarationID)
	}

	return nil
}

// CancelDeclarationCommand 取消报关命令
type CancelDeclarationCommand struct {
	DeclarationID string
	Reason        string
}

// CancelDeclaration 取消报关
func (s *CrossBorderCommandService) CancelDeclaration(ctx context.Context, cmd CancelDeclarationCommand) error {
	decl, err := s.declRepo.GetDeclaration(ctx, cmd.DeclarationID)
	if err != nil {
		return err
	}

	if err := decl.Cancel(cmd.Reason); err != nil {
		return err
	}

	return s.declRepo.UpdateDeclaration(ctx, decl)
}

// UploadDocumentCommand 上传证件命令
type UploadDocumentCommand struct {
	DeclarationID string
	DocumentType  domain.CustomsDocumentType
	DocumentName  string
	DocumentData  []byte
	DocumentURL   string
}

// UploadDocument 上传证件
func (s *CrossBorderCommandService) UploadDocument(ctx context.Context, cmd UploadDocumentCommand) (*domain.CustomsDocument, error) {
	decl, err := s.declRepo.GetDeclaration(ctx, cmd.DeclarationID)
	if err != nil {
		return nil, err
	}

	var docURL string
	if len(cmd.DocumentData) > 0 && s.docStorage != nil {
		docURL, err = s.docStorage.UploadDocument(ctx, cmd.DeclarationID, cmd.DocumentType, cmd.DocumentData)
		if err != nil {
			return nil, err
		}
	} else {
		docURL = cmd.DocumentURL
	}

	decl.AddDocument(cmd.DocumentType, cmd.DocumentName, docURL)

	if err := s.declRepo.UpdateDeclaration(ctx, decl); err != nil {
		return nil, err
	}

	for _, doc := range decl.Documents {
		if doc.DocumentURL == docURL {
			return &doc, nil
		}
	}

	return nil, nil
}

// CreateHSCodeCommand 创建HS编码命令
type CreateHSCodeCommand struct {
	Code                string
	Description         string
	DescriptionEn       string
	DutyRate            decimal.Decimal
	VATRate             decimal.Decimal
	ConsumptionTaxRate  decimal.Decimal
	Unit                string
	Restrictions        []string
}

// CreateHSCode 创建HS编码
func (s *CrossBorderCommandService) CreateHSCode(ctx context.Context, cmd CreateHSCodeCommand) (*domain.HSCode, error) {
	hsCode := &domain.HSCode{
		Code:               cmd.Code,
		Description:        cmd.Description,
		DescriptionEn:      cmd.DescriptionEn,
		DutyRate:           cmd.DutyRate,
		VATRate:            cmd.VATRate,
		ConsumptionTaxRate: cmd.ConsumptionTaxRate,
		Unit:               cmd.Unit,
		Active:             true,
	}

	if err := s.hsCodeRepo.Save(ctx, hsCode); err != nil {
		return nil, err
	}

	return hsCode, nil
}

// getHSCodes 获取HS编码
func (s *CrossBorderCommandService) getHSCodes(ctx context.Context, items []DeclarationItemCmd) (map[string]*domain.HSCode, error) {
	codes := make([]string, 0, len(items))
	for _, item := range items {
		if item.HSCode != "" {
			codes = append(codes, item.HSCode)
		}
	}
	if len(codes) == 0 {
		return nil, nil
	}
	return s.hsCodeRepo.GetByCodes(ctx, codes)
}

// publishEvents 发布事件
func (s *CrossBorderCommandService) publishEvents(ctx context.Context, events []domain.DomainEvent) {
	if s.publisher == nil {
		return
	}
	for _, event := range events {
		if err := s.publisher.Publish(ctx, s.topic, "", event); err != nil {
			s.logger.ErrorContext(ctx, "failed to publish event", "event", event.EventName(), "error", err)
		}
	}
}

// CrossBorderQueryService 跨境电商查询服务
type CrossBorderQueryService struct {
	declRepo   domain.CrossBorderRepository
	hsCodeRepo domain.HSCodeRepository
	orderRepo  domain.CrossBorderOrderRepository
	docRepo    domain.CustomsDocumentRepository
	eventRepo  domain.ClearanceEventRepository
	readRepo   domain.CrossBorderReadRepository
	logger     *slog.Logger
}

// NewCrossBorderQueryService 创建查询服务
func NewCrossBorderQueryService(
	declRepo domain.CrossBorderRepository,
	hsCodeRepo domain.HSCodeRepository,
	orderRepo domain.CrossBorderOrderRepository,
	docRepo domain.CustomsDocumentRepository,
	eventRepo domain.ClearanceEventRepository,
	readRepo domain.CrossBorderReadRepository,
	logger *slog.Logger,
) *CrossBorderQueryService {
	return &CrossBorderQueryService{
		declRepo:   declRepo,
		hsCodeRepo: hsCodeRepo,
		orderRepo:  orderRepo,
		docRepo:    docRepo,
		eventRepo:  eventRepo,
		readRepo:   readRepo,
		logger:     logger,
	}
}

// DeclarationDTO 报关单DTO
type DeclarationDTO struct {
	DeclarationID      string              `json:"declaration_id"`
	OrderID            string              `json:"order_id"`
	UserID             uint64              `json:"user_id"`
	MerchantID         uint64              `json:"merchant_id"`
	LogisticsNo        string              `json:"logistics_no"`
	DeclaredValue      string              `json:"declared_value"`
	Currency           string              `json:"currency"`
	DutyAmount         string              `json:"duty_amount"`
	TaxAmount          string              `json:"tax_amount"`
	Status             string              `json:"status"`
	RejectReason       string              `json:"reject_reason"`
	CustomsPort        string              `json:"customs_port"`
	TradeMode          string              `json:"trade_mode"`
	CustomsDeclNo      string              `json:"customs_decl_no"`
	ClearanceID        string              `json:"clearance_id"`
	ClearanceStatus    string              `json:"clearance_status"`
	Items              []DeclarationItemDTO `json:"items"`
	Documents          []CustomsDocumentDTO `json:"documents"`
	ClearanceEvents    []ClearanceEventDTO  `json:"clearance_events"`
	CreatedAt          time.Time           `json:"created_at"`
	SubmittedAt        *time.Time          `json:"submitted_at"`
	ClearedAt          *time.Time          `json:"cleared_at"`
}

// DeclarationItemDTO 报关明细DTO
type DeclarationItemDTO struct {
	SKUID       string `json:"sku_id"`
	ProductName string `json:"product_name"`
	HSCode      string `json:"hs_code"`
	Price       string `json:"price"`
	Quantity    int32  `json:"quantity"`
	Weight      string `json:"weight"`
	DutyRate    string `json:"duty_rate"`
	TaxRate     string `json:"tax_rate"`
	DutyAmount  string `json:"duty_amount"`
	TaxAmount   string `json:"tax_amount"`
}

// CustomsDocumentDTO 证件DTO
type CustomsDocumentDTO struct {
	DocumentID   string    `json:"document_id"`
	DocumentType string    `json:"document_type"`
	DocumentName string    `json:"document_name"`
	DocumentURL  string    `json:"document_url"`
	Status       string    `json:"status"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// ClearanceEventDTO 清关事件DTO
type ClearanceEventDTO struct {
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// HSCodeDTO HS编码DTO
type HSCodeDTO struct {
	Code                string   `json:"code"`
	Description         string   `json:"description"`
	DescriptionEn       string   `json:"description_en"`
	DutyRate            string   `json:"duty_rate"`
	VATRate             string   `json:"vat_rate"`
	ConsumptionTaxRate  string   `json:"consumption_tax_rate"`
	Unit                string   `json:"unit"`
	Restrictions        []string `json:"restrictions"`
	RequiredDocuments   []string `json:"required_documents"`
	Active              bool     `json:"active"`
}

// GetDeclaration 获取报关单
func (s *CrossBorderQueryService) GetDeclaration(ctx context.Context, declarationID string) (*DeclarationDTO, error) {
	decl, err := s.declRepo.GetDeclaration(ctx, declarationID)
	if err != nil {
		return nil, err
	}
	if decl == nil {
		return nil, errors.New("declaration not found")
	}
	return s.toDeclarationDTO(decl), nil
}

// GetDeclarationByOrder 根据订单获取报关单
func (s *CrossBorderQueryService) GetDeclarationByOrder(ctx context.Context, orderID string) (*DeclarationDTO, error) {
	decl, err := s.declRepo.GetDeclarationByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if decl == nil {
		return nil, errors.New("declaration not found")
	}
	return s.toDeclarationDTO(decl), nil
}

// ListDeclarations 获取报关单列表
func (s *CrossBorderQueryService) ListDeclarations(ctx context.Context, page, pageSize int, status domain.DeclarationStatus, userID uint64) ([]*DeclarationDTO, int64, error) {
	decls, total, err := s.declRepo.ListDeclarations(ctx, page, pageSize, status, userID)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*DeclarationDTO, len(decls))
	for i, decl := range decls {
		dtos[i] = s.toDeclarationDTO(decl)
	}

	return dtos, total, nil
}

// GetHSCode 获取HS编码
func (s *CrossBorderQueryService) GetHSCode(ctx context.Context, code string) (*HSCodeDTO, error) {
	hsCode, err := s.hsCodeRepo.Get(ctx, code)
	if err != nil {
		return nil, err
	}
	if hsCode == nil {
		return nil, errors.New("hs code not found")
	}
	return s.toHSCodeDTO(hsCode), nil
}

// SearchHSCodes 搜索HS编码
func (s *CrossBorderQueryService) SearchHSCodes(ctx context.Context, keyword string, page, pageSize int) ([]*HSCodeDTO, int64, error) {
	hsCodes, total, err := s.hsCodeRepo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*HSCodeDTO, len(hsCodes))
	for i, hsCode := range hsCodes {
		dtos[i] = s.toHSCodeDTO(hsCode)
	}

	return dtos, total, nil
}

// toDeclarationDTO 转换为DTO
func (s *CrossBorderQueryService) toDeclarationDTO(decl *domain.CustomsDeclaration) *DeclarationDTO {
	dto := &DeclarationDTO{
		DeclarationID:   decl.DeclarationID,
		OrderID:         decl.OrderID,
		UserID:          decl.UserID,
		MerchantID:      decl.MerchantID,
		LogisticsNo:     decl.LogisticsNo,
		DeclaredValue:   decl.DeclaredValue.String(),
		Currency:        decl.Currency,
		DutyAmount:      decl.DutyAmount.String(),
		TaxAmount:       decl.TaxAmount.String(),
		Status:          decl.Status.String(),
		RejectReason:    decl.RejectReason,
		CustomsPort:     decl.CustomsPort,
		TradeMode:       decl.TradeMode.String(),
		CustomsDeclNo:   decl.CustomsDeclNo,
		ClearanceID:     decl.ClearanceID,
		ClearanceStatus: decl.ClearanceStatus.String(),
		CreatedAt:       decl.CreatedAt,
		SubmittedAt:     decl.SubmittedAt,
		ClearedAt:       decl.ClearedAt,
	}

	for _, item := range decl.Items {
		dto.Items = append(dto.Items, DeclarationItemDTO{
			SKUID:       item.SKUID,
			ProductName: item.ProductName,
			HSCode:      item.HSCode,
			Price:       item.Price.String(),
			Quantity:    item.Quantity,
			Weight:      item.Weight.String(),
			DutyRate:    item.DutyRate.String(),
			TaxRate:     item.TaxRate.String(),
			DutyAmount:  item.DutyAmount.String(),
			TaxAmount:   item.TaxAmount.String(),
		})
	}

	for _, doc := range decl.Documents {
		dto.Documents = append(dto.Documents, CustomsDocumentDTO{
			DocumentID:   doc.DocumentID,
			DocumentType: fmt.Sprintf("%d", doc.DocumentType),
			DocumentName: doc.DocumentName,
			DocumentURL:  doc.DocumentURL,
			Status:       doc.Status,
			UploadedAt:   doc.UploadedAt,
		})
	}

	for _, event := range decl.ClearanceEvents {
		dto.ClearanceEvents = append(dto.ClearanceEvents, ClearanceEventDTO{
			EventType:   event.EventType,
			Description: event.Description,
			Location:    event.Location,
			OccurredAt:  event.OccurredAt,
		})
	}

	return dto
}

// toHSCodeDTO 转换为DTO
func (s *CrossBorderQueryService) toHSCodeDTO(hsCode *domain.HSCode) *HSCodeDTO {
	return &HSCodeDTO{
		Code:               hsCode.Code,
		Description:        hsCode.Description,
		DescriptionEn:      hsCode.DescriptionEn,
		DutyRate:           hsCode.DutyRate.String(),
		VATRate:            hsCode.VATRate.String(),
		ConsumptionTaxRate: hsCode.ConsumptionTaxRate.String(),
		Unit:               hsCode.Unit,
		Active:             hsCode.Active,
	}
}
