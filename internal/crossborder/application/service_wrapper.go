package application

import (
	"context"
	"log"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
)

type CrossBorderService struct {
	cmdSvc   *CrossBorderCommandService
	querySvc *CrossBorderQueryService
}

func NewCrossBorderService(
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
	logger interface{},
) *CrossBorderService {
	lg := getLogger(logger)

	cmdSvc := NewCrossBorderCommandService(
		declRepo, hsCodeRepo, orderRepo, docRepo, eventRepo, readRepo,
		taxService, customsGateway, docStorage, notification, nil, "", lg,
	)

	querySvc := NewCrossBorderQueryService(
		declRepo, hsCodeRepo, orderRepo, docRepo, eventRepo, readRepo, lg,
	)

	return &CrossBorderService{
		cmdSvc:   cmdSvc,
		querySvc: querySvc,
	}
}

func (s *CrossBorderService) CreateDeclaration(ctx context.Context, orderID string, userID uint64, logisticsNo, currency string, declaredValue float64, items []struct {
	SKUID  string
	HSCode string
	Price  float64
	Qty    int32
}) (string, error) {
	cmd := CreateDeclarationCommand{
		OrderID:       orderID,
		UserID:        userID,
		LogisticsNo:   logisticsNo,
		Currency:      currency,
		DeclaredValue: decimal.NewFromFloat(declaredValue),
	}

	for _, it := range items {
		cmd.Items = append(cmd.Items, DeclarationItemCmd{
			SKUID:    it.SKUID,
			HSCode:   it.HSCode,
			Price:    decimal.NewFromFloat(it.Price),
			Quantity: it.Qty,
		})
	}

	decl, err := s.cmdSvc.CreateDeclaration(ctx, cmd)
	if err != nil {
		return "", err
	}
	return decl.DeclarationID, nil
}

func (s *CrossBorderService) CalculateDuty(ctx context.Context, items []struct {
	HSCode string
	Price  float64
	Qty    int32
}, destinationCountry string) (float64, float64, error) {
	hsCodes := make(map[string]*domain.HSCode)
	for _, item := range items {
		hsCodes[item.HSCode] = &domain.HSCode{
			Code:     item.HSCode,
			DutyRate: decimal.NewFromFloat(0.1),
			VATRate:  decimal.NewFromFloat(0.13),
		}
	}

	decl := &domain.CustomsDeclaration{
		Items: make([]domain.DeclarationItem, len(items)),
	}
	for i, it := range items {
		decl.Items[i] = domain.DeclarationItem{
			HSCode:   it.HSCode,
			Price:    decimal.NewFromFloat(it.Price),
			Quantity: it.Qty,
		}
	}

	decl.CalculateTax(hsCodes)

	return decl.DutyAmount.InexactFloat64(), decl.TaxAmount.InexactFloat64(), nil
}

func (s *CrossBorderService) UpdateStatus(ctx context.Context, declarationID, status, rejectReason string) error {
	if status == "CANCELLED" {
		return s.cmdSvc.CancelDeclaration(ctx, CancelDeclarationCommand{
			DeclarationID: declarationID,
			Reason:        rejectReason,
		})
	}
	return nil
}

func (s *CrossBorderService) GetDeclaration(ctx context.Context, declarationID string) (*DeclarationDTO, error) {
	return s.querySvc.GetDeclaration(ctx, declarationID)
}

func (s *CrossBorderService) ListDeclarations(ctx context.Context, page, pageSize int, status domain.DeclarationStatus, userID uint64) ([]*DeclarationDTO, int64, error) {
	return s.querySvc.ListDeclarations(ctx, page, pageSize, status, userID)
}

type logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func getLogger(l interface{}) *log.Logger {
	if lg, ok := l.(*log.Logger); ok {
		return lg
	}
	return log.Default()
}
