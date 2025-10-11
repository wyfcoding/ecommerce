package service

import (
	"context"
	v1 "ecommerce/api/inventory/v1"
	"ecommerce/internal/inventory/biz"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InventoryService is the gRPC service implementation for inventory.
type InventoryService struct {
	v1.UnimplementedInventoryServiceServer
	uc *biz.InventoryUsecase
}

// NewInventoryService creates a new InventoryService.
func NewInventoryService(uc *biz.InventoryUsecase) *InventoryService {
	return &InventoryService{uc: uc}
}

// DeductStock implements the DeductStock RPC.
func (s *InventoryService) DeductStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.DeductStockResponse, error) {
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	// In a real system, this would involve a transaction across multiple items.
	// For simplicity, we'll deduct one by one.
	for _, item := range req.Items {
		err := s.uc.DeductStock(ctx, item.SkuId, item.Quantity)
		if err != nil {
			if errors.Is(err, biz.ErrInsufficientStock) {
				return &v1.DeductStockResponse{Success: false, Message: err.Error()}, status.Error(codes.FailedPrecondition, err.Error())
			}
			if errors.Is(err, biz.ErrStockNotFound) {
				return &v1.DeductStockResponse{Success: false, Message: err.Error()}, status.Error(codes.NotFound, err.Error())
			}
			return nil, status.Errorf(codes.Internal, "failed to deduct stock for SKU %d: %v", item.SkuId, err)
		}
	}

	return &v1.DeductStockResponse{Success: true, Message: "Stock deducted successfully"}, nil
}

// ReleaseStock implements the ReleaseStock RPC.
func (s *InventoryService) ReleaseStock(ctx context.Context, req *v1.ReleaseStockRequest) (*v1.ReleaseStockResponse, error) {
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	// In a real system, this would involve a transaction across multiple items.
	for _, item := range req.Items {
		err := s.uc.ReleaseStock(ctx, item.SkuId, item.Quantity)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to release stock for SKU %d: %v", item.SkuId, err)
		}
	}

	return &v1.ReleaseStockResponse{Success: true, Message: "Stock released successfully"}, nil
}

// GetStock implements the GetStock RPC.
func (s *InventoryService) GetStock(ctx context.Context, req *v1.GetStockRequest) (*v1.GetStockResponse, error) {
	if req.SkuId == 0 {
		return nil, status.Error(codes.InvalidArgument, "sku_id is required")
	}

	stock, err := s.uc.GetStock(ctx, req.SkuId)
	if err != nil {
		if errors.Is(err, biz.ErrStockNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to get stock for SKU %d: %v", req.SkuId, err)
	}

	return &v1.GetStockResponse{Quantity: stock.Quantity}, nil
}
