package grpc

import (
	"context"

	v1 "github.com/wyfcoding/ecommerce/go-api/tax/v1"
	"github.com/wyfcoding/ecommerce/internal/tax/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	v1.UnimplementedTaxServiceServer
	app *application.TaxService
}

func NewServer(app *application.TaxService) *Server {
	return &Server{app: app}
}

func (s *Server) CalculateTax(ctx context.Context, req *v1.CalculateTaxRequest) (*v1.CalculateTaxResponse, error) {
	if req.Amount < 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be non-negative")
	}

	result, err := s.app.CalculateOrderTax(
		ctx,
		req.UserId,
		req.DestinationCountry,
		"",
		req.ProductCategory,
		req.Amount,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate tax: %v", err)
	}

	taxRate := 0.0
	if req.Amount > 0 {
		taxRate = float64(result.TotalTaxAmount) / float64(req.Amount)
	}

	taxType := "UNKNOWN"
	if len(result.Items) > 0 {
		taxType = result.Items[0].TaxType.String()
	}

	return &v1.CalculateTaxResponse{
		TotalTax:     result.TotalTaxAmount,
		TaxRate:      taxRate,
		Jurisdiction: req.DestinationCountry,
		TaxType:      taxType,
	}, nil
}

func (s *Server) ValidateTaxID(_ context.Context, req *v1.ValidateTaxIDRequest) (*v1.ValidateTaxIDResponse, error) {
	if req.CountryCode == "" || req.TaxId == "" {
		return nil, status.Error(codes.InvalidArgument, "country_code and tax_id are required")
	}

	// 当前服务仅提供基础格式校验，外部税号校验网关后续接入。
	isValid := len(req.TaxId) >= 6
	return &v1.ValidateTaxIDResponse{
		IsValid:      isValid,
		BusinessName: "",
	}, nil
}

func (s *Server) GetTaxSummary(_ context.Context, req *v1.GetTaxSummaryRequest) (*v1.GetTaxSummaryResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// 当前版本先返回空汇总，后续由对账/报表任务异步聚合。
	return &v1.GetTaxSummaryResponse{
		TotalTaxPaid: 0,
		Breakdown:    []*v1.TaxBreakdown{},
	}, nil
}
