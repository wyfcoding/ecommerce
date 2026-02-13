package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/crossborder/v1"
	"github.com/wyfcoding/ecommerce/internal/crossborder/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedCrossBorderServiceServer
	app *application.CrossBorderService
}

func NewServer(app *application.CrossBorderService) *Server {
	return &Server{app: app}
}

func (s *Server) CalculateDuty(ctx context.Context, req *pb.CalculateDutyRequest) (*pb.CalculateDutyResponse, error) {
	var items []struct {
		HSCode string
		Price  float64
		Qty    int32
	}
	for _, item := range req.Items {
		items = append(items, struct {
			HSCode string
			Price  float64
			Qty    int32
		}{
			HSCode: item.HsCode,
			Price:  item.Price,
			Qty:    item.Quantity,
		})
	}

	duty, tax, err := s.app.CalculateDuty(ctx, items, req.DestinationCountry)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate duty: %v", err)
	}

	return &pb.CalculateDutyResponse{
		DutyAmount: duty,
		TaxAmount:  tax,
		TotalCost:  duty + tax,
		Currency:   req.Currency,
	}, nil
}

func (s *Server) CreateDeclaration(ctx context.Context, req *pb.CreateDeclarationRequest) (*pb.CreateDeclarationResponse, error) {
	var items []struct {
		SKUID  string
		HSCode string
		Price  float64
		Qty    int32
	}
	for _, item := range req.Items {
		items = append(items, struct {
			SKUID  string
			HSCode string
			Price  float64
			Qty    int32
		}{
			SKUID:  item.SkuId,
			HSCode: item.HsCode,
			Price:  item.Price,
			Qty:    item.Quantity,
		})
	}

	id, err := s.app.CreateDeclaration(ctx, req.OrderId, req.UserId, req.LogisticsNo, req.Currency, req.DeclaredValue, items)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create declaration: %v", err)
	}

	return &pb.CreateDeclarationResponse{
		DeclarationId: id,
	}, nil
}

func (s *Server) GetDeclaration(ctx context.Context, req *pb.GetDeclarationRequest) (*pb.GetDeclarationResponse, error) {
	decl, err := s.app.GetDeclaration(ctx, req.DeclarationId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "declaration not found: %v", err)
	}

	return &pb.GetDeclarationResponse{
		DeclarationId: decl.DeclarationID,
		OrderId:       decl.OrderID,
		Status:        decl.Status.String(),
		TotalTax:      decl.DutyAmount.Add(decl.TaxAmount).InexactFloat64(),
		CreatedAt:     decl.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *Server) UpdateDeclarationStatus(ctx context.Context, req *pb.UpdateDeclarationStatusRequest) (*pb.UpdateDeclarationStatusResponse, error) {
	err := s.app.UpdateStatus(ctx, req.DeclarationId, req.Status, req.RejectReason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update status: %v", err)
	}

	return &pb.UpdateDeclarationStatusResponse{Success: true}, nil
}
