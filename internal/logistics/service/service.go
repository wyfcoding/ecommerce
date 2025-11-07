package service

import (
	"context"
	v1 "ecommerce/api/logistics/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LogisticsService is the gRPC service implementation for logistics.
type LogisticsService struct {
	v1.UnimplementedLogisticsServiceServer
	uc *biz.LogisticsUsecase
}

// NewLogisticsService creates a new LogisticsService.
func NewLogisticsService(uc *biz.LogisticsUsecase) *LogisticsService {
	return &LogisticsService{uc: uc}
}

// CalculateShippingCost implements the CalculateShippingCost RPC.
func (s *LogisticsService) CalculateShippingCost(ctx context.Context, req *v1.CalculateShippingCostRequest) (*v1.CalculateShippingCostResponse, error) {
	if req.OriginAddress == nil || req.DestinationAddress == nil || len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "origin_address, destination_address, and items are required")
	}

	originAddress := &biz.AddressInfo{
		Province: req.OriginAddress.Province,
		City:     req.OriginAddress.City,
		District: req.OriginAddress.District,
	}
	destinationAddress := &biz.AddressInfo{
		Province: req.DestinationAddress.Province,
		City:     req.DestinationAddress.City,
		District: req.DestinationAddress.District,
	}

	bizItems := make([]*biz.ItemInfo, len(req.Items))
	for i, item := range req.Items {
		bizItems[i] = &biz.ItemInfo{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			// Assuming WeightKg is passed in the request or fetched from product service
			// For now, let's assume a dummy weight for demonstration
			WeightKg: 0.5, // Dummy weight
		}
	}

	shippingCost, err := s.uc.CalculateShippingCost(ctx, originAddress, destinationAddress, bizItems)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate shipping cost: %v", err)
	}

	return &v1.CalculateShippingCostResponse{
		ShippingCost: shippingCost,
		Currency:     "CNY", // Default currency
	}, nil
}
