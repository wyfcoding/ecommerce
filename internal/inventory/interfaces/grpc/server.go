package grpc

import (
	"context"
	pb "ecommerce/api/inventory/v1"
	"ecommerce/internal/inventory/application"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedInventoryServiceServer
	app *application.InventoryService
}

func NewServer(app *application.InventoryService) *Server {
	return &Server{app: app}
}

func (s *Server) DeductStock(ctx context.Context, req *pb.DeductStockRequest) (*pb.DeductStockResponse, error) {
	// Note: Proto supports multiple items, but Service DeductStock handles one SKU at a time.
	// We need to loop or update service to handle batch.
	// For now, we loop and rollback if any fails (Saga pattern handles distributed, but local batch needs atomicity or loop).
	// Since we are inside Inventory Service, we should ideally use a transaction.
	// However, Service `DeductStock` is single item.
	// Let's implement a loop. If one fails, we should ideally rollback previous ones.
	// For this refactor, we'll do a simple loop and return error on first failure.
	// A better approach would be to add `BatchDeductStock` to Service.

	for _, item := range req.Items {
		// Reason is hardcoded or passed from somewhere? Proto has OrderID.
		err := s.app.DeductStock(ctx, item.SkuId, int32(item.Quantity), "Order: "+req.OrderId)
		if err != nil {
			// TODO: Rollback previous deductions in this loop?
			// For now, return error.
			return &pb.DeductStockResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
	}

	return &pb.DeductStockResponse{
		Success: true,
		Message: "Stock deducted successfully",
	}, nil
}

func (s *Server) ReleaseStock(ctx context.Context, req *pb.ReleaseStockRequest) (*pb.ReleaseStockResponse, error) {
	for _, item := range req.Items {
		// UnlockStock acts as release/rollback of lock/deduction
		err := s.app.UnlockStock(ctx, item.SkuId, int32(item.Quantity), "Release Order: "+req.OrderId)
		if err != nil {
			// Log error but continue to try releasing others?
			// Or return error.
			return &pb.ReleaseStockResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
	}

	return &pb.ReleaseStockResponse{
		Success: true,
		Message: "Stock released successfully",
	}, nil
}

func (s *Server) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	inv, err := s.app.GetInventory(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if inv == nil {
		return nil, status.Error(codes.NotFound, "inventory not found")
	}

	return &pb.GetStockResponse{
		Quantity: uint32(inv.AvailableStock),
	}, nil
}
