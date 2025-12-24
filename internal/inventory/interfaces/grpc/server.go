package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/inventory/v1"         // 导入库存模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/inventory/application" // 导入库存模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/inventory/domain"      // 导入库存模块的领域层。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 InventoryService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedInventoryServiceServer                               // 嵌入生成的UnimplementedInventoryServiceServer，确保前向兼容性。
	app                                    *application.InventoryService // 依赖Inventory应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Inventory gRPC 服务端实例。
func NewServer(app *application.InventoryService) *Server {
	return &Server{app: app}
}

// CreateInventory 处理创建库存记录的gRPC请求。
// req: 包含SKU ID、商品ID、仓库ID、总库存和预警阈值的请求体。
// 返回created successfully的库存记录响应和可能发生的gRPC错误。
func (s *Server) CreateInventory(ctx context.Context, req *pb.CreateInventoryRequest) (*pb.CreateInventoryResponse, error) {
	inventory, err := s.app.CreateInventory(ctx, req.SkuId, req.ProductId, req.WarehouseId, req.TotalStock, req.WarningThreshold)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create inventory: %v", err))
	}

	return &pb.CreateInventoryResponse{
		Inventory: convertInventoryToProto(inventory),
	}, nil
}

// GetInventory 处理获取库存记录的gRPC请求。
// req: 包含SKU ID的请求体。
// 返回库存记录响应和可能发生的gRPC错误。
func (s *Server) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	inventory, err := s.app.GetInventory(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get inventory for sku %d: %v", req.SkuId, err))
	}
	// 应用服务层的 GetInventory 如果找不到记录，可能返回 (nil, nil) 或 (nil, ErrRecordNotFound)。
	// 这里显式检查 inventory == nil。
	if inventory == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("inventory not found for sku %d", req.SkuId))
	}

	return &pb.GetInventoryResponse{
		Inventory: convertInventoryToProto(inventory),
	}, nil
}

// AddStock 处理增加库存的gRPC请求。
// req: 包含SKU ID、增加数量和原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddStock(ctx context.Context, req *pb.AddStockRequest) (*emptypb.Empty, error) {
	if err := s.app.AddStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add stock: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// DeductStock 处理扣减库存的gRPC请求。
// req: 包含SKU ID、扣减数量和原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeductStock(ctx context.Context, req *pb.DeductStockRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to deduct stock: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// LockStock 处理锁定库存的gRPC请求。
// req: 包含SKU ID、锁定数量和原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) LockStock(ctx context.Context, req *pb.LockStockRequest) (*emptypb.Empty, error) {
	if err := s.app.LockStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to lock stock: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// UnlockStock 处理解锁库存的gRPC请求。
// req: 包含SKU ID、解锁数量和原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) UnlockStock(ctx context.Context, req *pb.UnlockStockRequest) (*emptypb.Empty, error) {
	if err := s.app.UnlockStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to unlock stock: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ConfirmDeduction 处理确认扣减库存的gRPC请求。
// req: 包含SKU ID、确认扣减数量和原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) ConfirmDeduction(ctx context.Context, req *pb.ConfirmDeductionRequest) (*emptypb.Empty, error) {
	if err := s.app.ConfirmDeduction(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to confirm deduction: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListInventories 处理列出库存记录的gRPC请求。
// req: 包含分页参数的请求体。
// 返回库存记录列表响应和可能发生的gRPC错误。
func (s *Server) ListInventories(ctx context.Context, req *pb.ListInventoriesRequest) (*pb.ListInventoriesResponse, error) {
	// 获取分页参数。
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取库存列表。
	inventories, total, err := s.app.ListInventories(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list inventories: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbInventories := make([]*pb.Inventory, len(inventories))
	for i, inv := range inventories {
		pbInventories[i] = convertInventoryToProto(inv)
	}

	return &pb.ListInventoriesResponse{
		Inventories: pbInventories,
		TotalCount:  uint64(total), // 总记录数。
	}, nil
}

// GetInventoryLogs 处理获取库存日志的gRPC请求。
// req: 包含库存ID和分页参数的请求体。
// 返回库存日志列表响应和可能发生的gRPC错误。
func (s *Server) GetInventoryLogs(ctx context.Context, req *pb.GetInventoryLogsRequest) (*pb.GetInventoryLogsResponse, error) {
	// 获取分页参数。
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取库存日志列表。
	logs, total, err := s.app.GetInventoryLogs(ctx, req.InventoryId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get inventory logs: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbLogs := make([]*pb.InventoryLog, len(logs))
	for i, log := range logs {
		pbLogs[i] = convertLogToProto(log)
	}

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
		// Proto中还包含 DeletedAt 等字段，但实体中没有或未映射。
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
		// Proto中还包含 UpdatedAt, DeletedAt 等字段，但实体中没有或未映射。
	}
}
