package interfaces

import (
	"context"
	"strconv"

	pb "github.com/wyfcoding/ecommerce/go-api/crossborder/v1"
	"github.com/wyfcoding/ecommerce/internal/crossborder/application"
	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, err
	}

	id, err := h.app.CreateDeclaration(ctx, req.OrderId, userID, req.LogisticsNo, req.Currency, req.DeclaredValue, items)
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

	return &pb.GetDeclarationResponse{
		Declaration: &pb.CustomsDeclaration{
			DeclarationId: d.DeclarationID,
			OrderId:       d.OrderID,
			UserId:        strconv.FormatUint(d.UserID, 10),
			LogisticsNo:   d.LogisticsNo,
			DeclaredValue: d.DeclaredValue.InexactFloat64(),
			Currency:      d.Currency,
			DutyAmount:    d.DutyAmount.InexactFloat64(),
			TaxAmount:     d.TaxAmount.InexactFloat64(),
			Status:        pb.DeclarationStatus(d.Status),
			CreatedAt:     timestamppb.New(d.CreatedAt),
			UpdatedAt:     timestamppb.New(d.UpdatedAt),
		},
	}, nil
}

func (h *CrossBorderHandler) UpdateDeclarationStatus(ctx context.Context, req *pb.UpdateDeclarationStatusRequest) (*emptypb.Empty, error) {
	err := h.app.UpdateStatus(ctx, req.DeclarationId, req.Status.String(), req.RejectReason)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
