package service

import (
	"context"
	"errors"
	"time"

	v1 "ecommerce/api/bi/v1"
	"ecommerce/internal/bi/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BiService is the gRPC service implementation for BI reports.
type BiService struct {
	v1.UnimplementedBiServiceServer
	uc *biz.BiUsecase
}

// NewBiService creates a new BiService.
func NewBiService(uc *biz.BiUsecase) *BiService {
	return &BiService{uc: uc}
}

// GetSalesOverview implements the GetSalesOverview RPC.
func (s *BiService) GetSalesOverview(ctx context.Context, req *v1.GetSalesOverviewRequest) (*v1.SalesOverviewResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format: %v", err)
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid end_date format: %v", err)
		}
		endDate = &t
	}

	overview, err := s.uc.GetSalesOverview(ctx, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get sales overview: %v", err)
	}

	return &v1.SalesOverviewResponse{
		Overview: &v1.SalesOverview{
			TotalSalesAmount: overview.TotalSalesAmount,
			TotalOrders:      overview.TotalOrders,
			TotalUsers:       overview.TotalUsers,
			ConversionRate:   overview.ConversionRate,
		},
	},
			nil
}

// GetTopSellingProducts implements the GetTopSellingProducts RPC.
func (s *BiService) GetTopSellingProducts(ctx context.Context, req *v1.GetTopSellingProductsRequest) (*v1.TopSellingProductsResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10 // Default limit
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid start_date format: %v", err)
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid end_date format: %v", err)
		}
		endDate = &t
	}

	products, err := s.uc.GetTopSellingProducts(ctx, req.Limit, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get top selling products: %v", err)
	}

	protoProducts := make([]*v1.ProductSalesData, len(products))
	for i, p := range products {
		protoProducts[i] = &v1.ProductSalesData{
			ProductId:    p.ProductID,
			ProductName:  p.ProductName,
			SalesQuantity: p.SalesQuantity,
			SalesAmount:  p.SalesAmount,
		}
	}

	return &v1.TopSellingProductsResponse{Products: protoProducts}, nil
}
