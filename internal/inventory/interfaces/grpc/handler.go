package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/inventory/v1"          // 导入库存模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/inventory/application" // 导入库存模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/inventory/domain"      // 导入库存模块的领域层。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 Inventory 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedInventoryServiceServer                        // 嵌入生成的UnimplementedInventoryServiceServer，确保前向兼容性。
	app                                    *application.Inventory // 依赖Inventory应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Inventory gRPC 服务端实例。
func NewServer(app *application.Inventory) *Server {
	return &Server{app: app}
}

// CreateInventory 处理创建库存记录的gRPC请求。
func (s *Server) CreateInventory(ctx context.Context, req *pb.CreateInventoryRequest) (*pb.CreateInventoryResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateInventory received", "sku_id", req.SkuId, "product_id", req.ProductId, "warehouse_id", req.WarehouseId)

	inventory, err := s.app.CreateInventory(ctx, req.SkuId, req.ProductId, req.WarehouseId, req.TotalStock, req.WarningThreshold)
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

	inventory, err := s.app.GetInventory(ctx, req.SkuId)
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

	if err := s.app.AddStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
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

	if err := s.app.DeductStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
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

	if err := s.app.LockStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
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

	if err := s.app.UnlockStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
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

	if err := s.app.ConfirmDeduction(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
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

	inventories, total, err := s.app.ListInventories(ctx, page, pageSize)
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
		TotalCount:  uint64(total), // 总记录数。
	}, nil
}

// GetInventoryLogs 处理获取库存日志的gRPC请求。
func (s *Server) GetInventoryLogs(ctx context.Context, req *pb.GetInventoryLogsRequest) (*pb.GetInventoryLogsResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetInventoryLogs received", "inventory_id", req.InventoryId, "page_num", req.PageNum)

	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	logs, total, err := s.app.GetInventoryLogs(ctx, req.InventoryId, page, pageSize)
	if err != nil {
		slog.Error("gRPC GetInventoryLogs failed", "inventory_id", req.InventoryId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get inventory logs: %v", err))
	}

	pbLogs := make([]*pb.InventoryLog, len(logs))
	for i, log := range logs {
		pbLogs[i] = convertLogToProto(log)
	}

	slog.Debug("gRPC GetInventoryLogs successful", "inventory_id", req.InventoryId, "count", len(pbLogs), "duration", time.Since(start))
	return &pb.GetInventoryLogsResponse{
		Logs:       pbLogs,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// convertInventoryToProto 是一个辅助函数，将领域层的 Inventory 实体转换为 protobuf 的 Inventory 消息。
func convertInventoryToProto(inv *domain.Inventory) *pb.Inventory {
	if inv == nil {
		return nil
	}
	return &pb.Inventory{
		Id:               uint64(inv.ID),                 // 库存记录ID。
		SkuId:            inv.SkuID,                      // SKU ID。
		ProductId:        inv.ProductID,                  // 商品ID。
		WarehouseId:      inv.WarehouseID,                // 仓库ID。
		AvailableStock:   inv.AvailableStock,             // 可用库存。
		LockedStock:      inv.LockedStock,                // 锁定库存。
		TotalStock:       inv.TotalStock,                 // 总库存。
		Status:           int32(inv.Status),              // 状态。
		WarningThreshold: inv.WarningThreshold,           // 预警阈值。
		CreatedAt:        timestamppb.New(inv.CreatedAt), // 创建时间。
		UpdatedAt:        timestamppb.New(inv.UpdatedAt), // 更新时间。
	}
}

// convertLogToProto 是一个辅助函数，将领域层的 InventoryLog 实体转换为 protobuf 的 InventoryLog 消息。
func convertLogToProto(log *domain.InventoryLog) *pb.InventoryLog {
	if log == nil {
		return nil
	}
	return &pb.InventoryLog{
		Id:             uint64(log.ID),                 // 日志记录ID。
		InventoryId:    log.InventoryID,                // 库存ID。
		Action:         log.Action,                     // 操作类型。
		ChangeQuantity: log.ChangeQuantity,             // 变更数量。
		OldAvailable:   log.OldAvailable,               // 变更前可用库存。
		NewAvailable:   log.NewAvailable,               // 变更后可用库存。
		OldLocked:      log.OldLocked,                  // 变更前锁定库存。
		NewLocked:      log.NewLocked,                  // 变更后锁定库存。
		Reason:         log.Reason,                     // 原因。
		CreatedAt:      timestamppb.New(log.CreatedAt), // 创建时间。
	}
}
