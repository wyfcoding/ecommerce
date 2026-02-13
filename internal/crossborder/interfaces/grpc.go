package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/crossborder/v1"
	"github.com/wyfcoding/ecommerce/internal/crossborder/application"
	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
)

type CrossBorderHandler struct {
	pb.UnimplementedCrossBorderServiceServer
	app  *application.CrossBorderService
	repo domain.CrossBorderRepository
}

func NewCrossBorderHandler(app *application.CrossBorderService, repo domain.CrossBorderRepository) *CrossBorderHandler {
	return &CrossBorderHandler{app: app, repo: repo}
}

func (h *CrossBorderHandler) CalculateDuty(ctx context.Context, req *pb.CalculateDutyRequest) (*pb.CalculateDutyResponse, error) {
	items := make([]struct {
		HSCode string
		Price  float64
		Qty    int32
	}, len(req.Items))
	
	for i, it := range req.Items {
		items[i] = struct {
			HSCode string
			Price  float64
			Qty    int32
		}{it.HsCode, it.Price, it.Quantity}
	}

	duty, tax, err := h.app.CalculateDuty(ctx, items, req.DestinationCountry)
	if err != nil {
		return nil, err
	}

	return &pb.CalculateDutyResponse{
		DutyAmount: duty,
		TaxAmount:  tax,
		TotalCost:  duty + tax,
		Currency:   req.Currency,
	}, nil
}

func (h *CrossBorderHandler) CreateDeclaration(ctx context.Context, req *pb.CreateDeclarationRequest) (*pb.CreateDeclarationResponse, error) {
	items := make([]struct {
		SKUID  string
		HSCode string
		Price  float64
		Qty    int32
	}, len(req.Items))
	
	for i, it := range req.Items {
		items[i] = struct {
			SKUID  string
			HSCode string
			Price  float64
			Qty    int32
		}{it.SkuId, it.HsCode, it.Price, it.Quantity}
	}

	id, err := h.app.CreateDeclaration(ctx, req.OrderId, req.UserId, req.LogisticsNo, req.Currency, req.DeclaredValue, items)
	if err != nil {
		return nil, err
	}
	return &pb.CreateDeclarationResponse{DeclarationId: id}, nil
}

func (h *CrossBorderHandler) GetDeclaration(ctx context.Context, req *pb.GetDeclarationRequest) (*pb.GetDeclarationResponse, error) {
	d, err := h.repo.GetDeclaration(ctx, req.DeclarationId)
	if err != nil {
		return nil, err
	}

	totalTax := d.DutyAmount.Add(d.TaxAmount).InexactFloat64()

	return &pb.GetDeclarationResponse{
		DeclarationId: d.DeclarationID,
		OrderId:       d.OrderID,
		Status:        d.Status.String(),
		TotalTax:      totalTax,
		CreatedAt:     d.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (h *CrossBorderHandler) UpdateDeclarationStatus(ctx context.Context, req *pb.UpdateDeclarationStatusRequest) (*pb.UpdateDeclarationStatusResponse, error) {
	err := h.app.UpdateStatus(ctx, req.DeclarationId, req.Status, req.RejectReason)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateDeclarationStatusResponse{Success: true}, nil
}
