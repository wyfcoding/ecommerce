package grpc

import (
	"context"
	pb "ecommerce/api/logistics/v1"
	"ecommerce/internal/logistics/application"
)

type Server struct {
	pb.UnimplementedLogisticsServiceServer
	app *application.LogisticsService
}

func NewServer(app *application.LogisticsService) *Server {
	return &Server{app: app}
}

func (s *Server) CalculateShippingCost(ctx context.Context, req *pb.CalculateShippingCostRequest) (*pb.CalculateShippingCostResponse, error) {
	// Service doesn't have CalculateShippingCost method yet.
	// This is a feature gap.
	// For now, we return a mock or simple calculation based on item count/weight if available.
	// Proto has Items and Address.

	// Mock calculation: 1000 cents (10 CNY) base + 100 cents per item
	var totalQuantity uint32
	for _, item := range req.Items {
		totalQuantity += item.Quantity
	}

	cost := 1000 + uint64(totalQuantity)*100

	return &pb.CalculateShippingCostResponse{
		ShippingCost: cost,
		Currency:     "CNY",
	}, nil
}
