package service

import (
	"context"
	"fmt"
	"strings"

	v1 "ecommerce/api/inventory/v1"
	"ecommerce/internal/inventory/model"
	"ecommerce/internal/inventory/repository"
	"ecommerce/pkg/errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InventoryService 库存服务实现
type InventoryService struct {
	v1.UnimplementedInventoryServiceServer
	repo repository.InventoryRepository
}

// NewInventoryService 创建库存服务实例
func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

// DeductStock 扣减库存（订单支付后）
func (s *InventoryService) DeductStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.DeductStockResponse, error) {
	zap.S().Infow("DeductStock request", "items_count", len(req.Items))

	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	// 扣减每个SKU的库存
	for _, item := range req.Items {
		if item.SkuId == "" {
			return nil, status.Error(codes.InvalidArgument, "sku_id is required")
		}
		if item.Quantity <= 0 {
			return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
		}

		// 默认仓库ID为1（实际应根据业务逻辑选择仓库）
		warehouseID := uint(1)
		if item.WarehouseId > 0 {
			warehouseID = uint(item.WarehouseId)
		}

		// 调用repository扣减库存
		err := s.repo.AdjustStock(
			ctx,
			item.SkuId,
			warehouseID,
			-int(item.Quantity), // 负数表示扣减
			model.MovementTypeOutbound,
			req.OrderNo,
			"Order payment completed",
		)

		if err != nil {
			zap.S().Errorw("Failed to deduct stock",
				"sku_id", item.SkuId,
				"quantity", item.Quantity,
				"error", err,
			)

			if strings.Contains(err.Error(), "物理库存不足") || strings.Contains(err.Error(), "可用库存不足") {
				return &v1.DeductStockResponse{
					Success: false,
					Message: fmt.Sprintf("Insufficient stock for SKU %s", item.SkuId),
				}, nil
			}

			return nil, status.Errorf(codes.Internal, "failed to deduct stock for SKU %s: %v", item.SkuId, err)
		}
	}

	zap.S().Infow("Stock deducted successfully", "order_no", req.OrderNo)
	return &v1.DeductStockResponse{Success: true, Message: "Stock deducted successfully"}, nil
}

// ReleaseStock 释放预留库存（订单取消时）
func (s *InventoryService) ReleaseStock(ctx context.Context, req *v1.ReleaseStockRequest) (*v1.ReleaseStockResponse, error) {
	zap.S().Infow("ReleaseStock request", "items_count", len(req.Items))

	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	for _, item := range req.Items {
		if item.SkuId == "" {
			return nil, status.Error(codes.InvalidArgument, "sku_id is required")
		}
		if item.Quantity <= 0 {
			return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
		}

		warehouseID := uint(1)
		if item.WarehouseId > 0 {
			warehouseID = uint(item.WarehouseId)
		}

		// 释放预留库存
		err := s.repo.ReleaseStock(ctx, item.SkuId, warehouseID, int(item.Quantity))
		if err != nil {
			zap.S().Errorw("Failed to release stock",
				"sku_id", item.SkuId,
				"quantity", item.Quantity,
				"error", err,
			)
			return nil, status.Errorf(codes.Internal, "failed to release stock for SKU %s: %v", item.SkuId, err)
		}
	}

	zap.S().Infow("Stock released successfully", "order_no", req.OrderNo)
	return &v1.ReleaseStockResponse{Success: true, Message: "Stock released successfully"}, nil
}

// ReserveStock 预留库存（下单时）
func (s *InventoryService) ReserveStock(ctx context.Context, req *v1.ReserveStockRequest) (*v1.ReserveStockResponse, error) {
	zap.S().Infow("ReserveStock request", "items_count", len(req.Items))

	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	for _, item := range req.Items {
		if item.SkuId == "" {
			return nil, status.Error(codes.InvalidArgument, "sku_id is required")
		}
		if item.Quantity <= 0 {
			return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
		}

		warehouseID := uint(1)
		if item.WarehouseId > 0 {
			warehouseID = uint(item.WarehouseId)
		}

		// 预留库存
		err := s.repo.ReserveStock(ctx, item.SkuId, warehouseID, int(item.Quantity))
		if err != nil {
			zap.S().Errorw("Failed to reserve stock",
				"sku_id", item.SkuId,
				"quantity", item.Quantity,
				"error", err,
			)

			if strings.Contains(err.Error(), "可用库存不足") {
				return &v1.ReserveStockResponse{
					Success: false,
					Message: fmt.Sprintf("Insufficient available stock for SKU %s", item.SkuId),
				}, nil
			}

			return nil, status.Errorf(codes.Internal, "failed to reserve stock for SKU %s: %v", item.SkuId, err)
		}
	}

	zap.S().Infow("Stock reserved successfully", "order_no", req.OrderNo)
	return &v1.ReserveStockResponse{Success: true, Message: "Stock reserved successfully"}, nil
}

// GetStock 查询库存
func (s *InventoryService) GetStock(ctx context.Context, req *v1.GetStockRequest) (*v1.GetStockResponse, error) {
	if req.SkuId == "" {
		return nil, status.Error(codes.InvalidArgument, "sku_id is required")
	}

	warehouseID := uint(1)
	if req.WarehouseId > 0 {
		warehouseID = uint(req.WarehouseId)
	}

	inventory, err := s.repo.GetInventory(ctx, req.SkuId, warehouseID)
	if err != nil {
		zap.S().Errorw("Failed to get stock", "sku_id", req.SkuId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get stock for SKU %s: %v", req.SkuId, err)
	}

	if inventory == nil {
		return nil, errors.ErrStockNotFound.ToGRPCError()
	}

	return &v1.GetStockResponse{
		SkuId:             req.SkuId,
		WarehouseId:       uint64(inventory.WarehouseID),
		QuantityOnHand:    int32(inventory.QuantityOnHand),
		QuantityReserved:  int32(inventory.QuantityReserved),
		QuantityAvailable: int32(inventory.QuantityAvailable),
	}, nil
}
