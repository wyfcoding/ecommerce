package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/inventory/v1"
	"github.com/wyfcoding/ecommerce/internal/inventory/application"
	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/optimization"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 Inventory 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedInventoryServiceServer
	cmdService   *application.InventoryCommandService
	queryService *application.InventoryQueryService
}

// NewServer 创建并返回一个新的 Inventory gRPC 服务端实例。
func NewServer(cmd *application.InventoryCommandService, query *application.InventoryQueryService) *Server {
	return &Server{cmdService: cmd, queryService: query}
}

// CreateInventory 处理创建库存记录的gRPC请求。
func (s *Server) CreateInventory(ctx context.Context, req *pb.CreateInventoryRequest) (*pb.CreateInventoryResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateInventory received", "sku_id", req.SkuId, "product_id", req.ProductId, "warehouse_id", req.WarehouseId)

	inventory, err := s.cmdService.CreateInventory(ctx, req.SkuId, req.ProductId, req.WarehouseId, req.TotalStock, req.WarningThreshold)
	if err != nil {
		slog.Error("gRPC CreateInventory failed", "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create inventory: %v", err))
	}

	slog.Info("gRPC CreateInventory successful", "sku_id", req.SkuId, "inventory_id", inventory.ID, "duration", time.Since(start))
	return &pb.CreateInventoryResponse{
		Inventory: convertInventoryToProto(inventory),
	}, nil
}

// GetInventory 处理获取库存记录的gRPC请求。
func (s *Server) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetInventory received", "sku_id", req.SkuId)

	inventory, err := s.queryService.GetInventory(ctx, req.SkuId)
	if err != nil {
		slog.Error("gRPC GetInventory failed", "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get inventory for sku %d: %v", req.SkuId, err))
	}
	if inventory == nil {
		slog.Debug("gRPC GetInventory successful (not found)", "sku_id", req.SkuId, "duration", time.Since(start))
		return nil, status.Error(codes.NotFound, fmt.Sprintf("inventory not found for sku %d", req.SkuId))
	}

	slog.Debug("gRPC GetInventory successful", "sku_id", req.SkuId, "duration", time.Since(start))
	return &pb.GetInventoryResponse{
		Inventory: convertInventoryToProto(inventory),
	}, nil
}

// AddStock 处理增加库存的gRPC请求。
func (s *Server) AddStock(ctx context.Context, req *pb.AddStockRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC AddStock received", "sku_id", req.SkuId, "quantity", req.Quantity, "reason", req.Reason)

	if err := s.cmdService.AddStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		slog.Error("gRPC AddStock failed", "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add stock: %v", err))
	}

	slog.Info("gRPC AddStock successful", "sku_id", req.SkuId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// DeductStock 处理扣减库存的gRPC请求。
func (s *Server) DeductStock(ctx context.Context, req *pb.DeductStockRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC DeductStock received", "sku_id", req.SkuId, "quantity", req.Quantity, "reason", req.Reason)

	if err := s.cmdService.DeductStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		slog.Error("gRPC DeductStock failed", "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to deduct stock: %v", err))
	}

	slog.Info("gRPC DeductStock successful", "sku_id", req.SkuId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// LockStock 处理锁定库存的gRPC请求。
func (s *Server) LockStock(ctx context.Context, req *pb.LockStockRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC LockStock received", "sku_id", req.SkuId, "quantity", req.Quantity, "reason", req.Reason)

	if err := s.cmdService.LockStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		slog.Error("gRPC LockStock failed", "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to lock stock: %v", err))
	}

	slog.Info("gRPC LockStock successful", "sku_id", req.SkuId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// UnlockStock 处理解锁库存的gRPC请求。
func (s *Server) UnlockStock(ctx context.Context, req *pb.UnlockStockRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC UnlockStock received", "sku_id", req.SkuId, "quantity", req.Quantity, "reason", req.Reason)

	if err := s.cmdService.UnlockStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		slog.Error("gRPC UnlockStock failed", "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to unlock stock: %v", err))
	}

	slog.Info("gRPC UnlockStock successful", "sku_id", req.SkuId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// ConfirmDeduction 处理确认扣减库存的gRPC请求。
func (s *Server) ConfirmDeduction(ctx context.Context, req *pb.ConfirmDeductionRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC ConfirmDeduction received", "sku_id", req.SkuId, "quantity", req.Quantity, "reason", req.Reason)

	if err := s.cmdService.ConfirmDeduction(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		slog.Error("gRPC ConfirmDeduction failed", "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to confirm deduction: %v", err))
	}

	slog.Info("gRPC ConfirmDeduction successful", "sku_id", req.SkuId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// ListInventories 处理列出库存记录的gRPC请求。
func (s *Server) ListInventories(ctx context.Context, req *pb.ListInventoriesRequest) (*pb.ListInventoriesResponse, error) {
	start := time.Now()
	slog.Debug("gRPC ListInventories received", "page_num", req.PageNum, "page_size", req.PageSize)

	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	inventories, total, err := s.queryService.ListInventories(ctx, page, pageSize)
	if err != nil {
		slog.Error("gRPC ListInventories failed", "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list inventories: %v", err))
	}

	pbInventories := make([]*pb.Inventory, len(inventories))
	for i, inv := range inventories {
		pbInventories[i] = convertInventoryToProto(inv)
	}

	slog.Debug("gRPC ListInventories successful", "count", len(pbInventories), "duration", time.Since(start))
	return &pb.ListInventoriesResponse{
		Inventories: pbInventories,
		TotalCount:  uint64(total),
	}, nil
}

// GetInventoryLogs 处理获取库存日志的gRPC请求。
func (s *Server) GetInventoryLogs(ctx context.Context, req *pb.GetInventoryLogsRequest) (*pb.GetInventoryLogsResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetInventoryLogs received", "inventory_id", req.InventoryId, "sku_id", req.SkuId, "page_num", req.PageNum)

	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	logs, total, err := s.queryService.GetInventoryLogs(ctx, req.SkuId, req.InventoryId, page, pageSize)
	if err != nil {
		slog.Error("gRPC GetInventoryLogs failed", "inventory_id", req.InventoryId, "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get inventory logs: %v", err))
	}

	pbLogs := make([]*pb.InventoryLog, len(logs))
	for i, log := range logs {
		pbLogs[i] = convertLogToProto(log)
	}

	slog.Debug("gRPC GetInventoryLogs successful", "inventory_id", req.InventoryId, "count", len(pbLogs), "duration", time.Since(start))
	return &pb.GetInventoryLogsResponse{
		Logs:       pbLogs,
		TotalCount: uint64(total),
	}, nil
}

// AllocateOrderStock 处理订单库存分配请求。
func (s *Server) AllocateOrderStock(ctx context.Context, req *pb.AllocateOrderStockRequest) (*pb.AllocateOrderStockResponse, error) {
	// 转换请求参数为算法输入 DTO
	algoItems := make([]algorithm.OrderItem, len(req.Items))
	for i, it := range req.Items {
		algoItems[i] = algorithm.OrderItem{
			SkuID:    it.SkuId,
			Quantity: it.Quantity,
		}
	}

	// 调用应用服务层（已有的分配逻辑）
	results, err := s.cmdService.AllocateStock(ctx, req.UserLat, req.UserLon, algoItems)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to allocate stock: %v", err))
	}

	// 映射结果
	pbAllocations := make([]*pb.WarehouseAllocation, len(results))
	for i, res := range results {
		pbItems := make([]*pb.OrderItemShort, len(res.Items))
		for j, it := range res.Items {
			pbItems[j] = &pb.OrderItemShort{
				SkuId:    it.SkuID,
				Quantity: it.Quantity,
			}
		}
		pbAllocations[i] = &pb.WarehouseAllocation{
			WarehouseId:   res.WarehouseID,
			Items:         pbItems,
			Distance:      res.Distance,
			EstimatedCost: res.TotalCost,
		}
	}

	return &pb.AllocateOrderStockResponse{
		OrderId:     req.OrderId,
		Allocations: pbAllocations,
	}, nil
}

// convertInventoryToProto 是一个辅助函数，将领域层的 Inventory 实体转换为 protobuf 的 Inventory 消息。
func convertInventoryToProto(inv *domain.Inventory) *pb.Inventory {
	if inv == nil {
		return nil
	}
	return &pb.Inventory{
		Id:               uint64(inv.ID),
		SkuId:            inv.SkuID,
		ProductId:        inv.ProductID,
		WarehouseId:      inv.WarehouseID,
		AvailableStock:   inv.AvailableStock,
		LockedStock:      inv.LockedStock,
		TotalStock:       inv.TotalStock,
		Status:           int32(inv.Status),
		WarningThreshold: inv.WarningThreshold,
		CreatedAt:        timestamppb.New(inv.CreatedAt),
		UpdatedAt:        timestamppb.New(inv.UpdatedAt),
	}
}

// convertLogToProto 是一个辅助函数，将领域层的 InventoryLog 实体转换为 protobuf 的 InventoryLog 消息。
func convertLogToProto(log *domain.InventoryLog) *pb.InventoryLog {
	if log == nil {
		return nil
	}
	return &pb.InventoryLog{
		Id:             uint64(log.ID),
		InventoryId:    log.InventoryID,
		Action:         log.Action,
		ChangeQuantity: log.ChangeQuantity,
		OldAvailable:   log.OldAvailable,
		NewAvailable:   log.NewAvailable,
		OldLocked:      log.OldLocked,
		NewLocked:      log.NewLocked,
		Reason:         log.Reason,
		CreatedAt:      timestamppb.New(log.CreatedAt),
	}
}
