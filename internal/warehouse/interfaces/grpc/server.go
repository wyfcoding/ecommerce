package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。

	pb "github.com/wyfcoding/ecommerce/go-api/warehouse/v1"           // 导入仓库模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/warehouse/application"   // 导入仓库模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/entity" // 导入仓库模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 WarehouseService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedWarehouseServiceServer                               // 嵌入生成的UnimplementedWarehouseServiceServer，确保前向兼容性。
	app                                    *application.WarehouseService // 依赖Warehouse应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Warehouse gRPC 服务端实例。
func NewServer(app *application.WarehouseService) *Server {
	return &Server{app: app}
}

// CreateWarehouse 处理创建仓库的gRPC请求。
// req: 包含仓库代码和名称的请求体。
// 返回created successfully的仓库响应和可能发生的gRPC错误。
func (s *Server) CreateWarehouse(ctx context.Context, req *pb.CreateWarehouseRequest) (*pb.CreateWarehouseResponse, error) {
	warehouse, err := s.app.CreateWarehouse(ctx, req.Code, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create warehouse: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateWarehouseResponse{
		Warehouse: convertWarehouseToProto(warehouse),
	}, nil
}

// ListWarehouses 处理列出仓库的gRPC请求。
// req: 包含分页参数的请求体。
// 返回仓库列表响应和可能发生的gRPC错误。
func (s *Server) ListWarehouses(ctx context.Context, req *pb.ListWarehousesRequest) (*pb.ListWarehousesResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取仓库列表。
	warehouses, total, err := s.app.ListWarehouses(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list warehouses: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbWarehouses := make([]*pb.Warehouse, len(warehouses))
	for i, w := range warehouses {
		pbWarehouses[i] = convertWarehouseToProto(w)
	}

	return &pb.ListWarehousesResponse{
		Warehouses: pbWarehouses,
		TotalCount: total, // 总记录数。
	}, nil
}

// UpdateStock 处理更新库存的gRPC请求（增加或减少）。
// req: 包含仓库ID、SKU ID和数量的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*emptypb.Empty, error) {
	if err := s.app.UpdateStock(ctx, req.WarehouseId, req.SkuId, req.Quantity); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update stock: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// GetStock 处理获取库存的gRPC请求。
// req: 包含仓库ID和SKU ID的请求体。
// 返回库存信息响应和可能发生的gRPC错误。
func (s *Server) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	stock, err := s.app.GetStock(ctx, req.WarehouseId, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get stock: %v", err))
	}

	return &pb.GetStockResponse{
		Stock: convertStockToProto(stock),
	}, nil
}

// CreateTransfer 处理创建库存调拨单的gRPC请求。
// req: 包含调出/调入仓库ID、SKU ID、数量和创建人ID的请求体。
// 返回created successfully的调拨单响应和可能发生的gRPC错误。
func (s *Server) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.CreateTransferResponse, error) {
	transfer, err := s.app.CreateTransfer(ctx, req.FromWarehouseId, req.ToWarehouseId, req.SkuId, req.Quantity, req.CreatedBy)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create transfer: %v", err))
	}

	return &pb.CreateTransferResponse{
		Transfer: convertTransferToProto(transfer),
	}, nil
}

// CompleteTransfer 处理完成库存调拨的gRPC请求。
// req: 包含调拨单ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) CompleteTransfer(ctx context.Context, req *pb.CompleteTransferRequest) (*emptypb.Empty, error) {
	if err := s.app.CompleteTransfer(ctx, req.TransferId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to complete transfer: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// DeductStock 扣减库存（Saga正向操作）。
// req: 包含仓库ID、SKU ID和数量的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeductStock(ctx context.Context, req *pb.DeductStockRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductStock(ctx, req.WarehouseId, req.SkuId, req.Quantity); err != nil {
		// DTM 分布式事务管理器期望在业务逻辑失败时返回 codes.Aborted 状态码，
		// 以触发Saga事务的回滚操作。
		return nil, status.Error(codes.Aborted, fmt.Sprintf("failed to deduct stock for saga: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RevertStock 回滚库存（Saga补偿操作）。
// req: 包含仓库ID、SKU ID和数量的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RevertStock(ctx context.Context, req *pb.RevertStockRequest) (*emptypb.Empty, error) {
	if err := s.app.RevertStock(ctx, req.WarehouseId, req.SkuId, req.Quantity); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to revert stock for saga: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// convertWarehouseToProto 是一个辅助函数，将领域层的 Warehouse 实体转换为 protobuf 的 Warehouse 消息。
func convertWarehouseToProto(w *entity.Warehouse) *pb.Warehouse {
	if w == nil {
		return nil
	}
	return &pb.Warehouse{
		Id:            uint64(w.ID),                 // ID。
		Code:          w.Code,                       // 代码。
		Name:          w.Name,                       // 名称。
		WarehouseType: w.WarehouseType,              // 类型。
		Province:      w.Province,                   // 省份。
		City:          w.City,                       // 城市。
		District:      w.District,                   // 区。
		Address:       w.Address,                    // 地址。
		Longitude:     w.Longitude,                  // 经度。
		Latitude:      w.Latitude,                   // 纬度。
		ContactName:   w.ContactName,                // 联系人。
		ContactPhone:  w.ContactPhone,               // 联系电话。
		Priority:      w.Priority,                   // 优先级。
		Status:        string(w.Status),             // 状态。
		Capacity:      w.Capacity,                   // 容量。
		Description:   w.Description,                // 描述。
		CreatedAt:     timestamppb.New(w.CreatedAt), // 创建时间。
		UpdatedAt:     timestamppb.New(w.UpdatedAt), // 更新时间。
	}
}

// convertStockToProto 是一个辅助函数，将领域层的 WarehouseStock 实体转换为 protobuf 的 WarehouseStock 消息。
func convertStockToProto(s *entity.WarehouseStock) *pb.WarehouseStock {
	if s == nil {
		return nil
	}
	return &pb.WarehouseStock{
		Id:          uint64(s.ID),                 // ID。
		WarehouseId: s.WarehouseID,                // 仓库ID。
		SkuId:       s.SkuID,                      // SKU ID。
		Stock:       s.Stock,                      // 库存数量。
		LockedStock: s.LockedStock,                // 锁定库存。
		SafeStock:   s.SafeStock,                  // 安全库存。
		MaxStock:    s.MaxStock,                   // 最大库存。
		CreatedAt:   timestamppb.New(s.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(s.UpdatedAt), // 更新时间。
	}
}

// convertTransferToProto 是一个辅助函数，将领域层的 StockTransfer 实体转换为 protobuf 的 StockTransfer 消息。
func convertTransferToProto(t *entity.StockTransfer) *pb.StockTransfer {
	if t == nil {
		return nil
	}
	// 转换可选的时间字段。
	var approvedAt, shippedAt, receivedAt, completedAt *timestamppb.Timestamp
	if t.ApprovedAt != nil {
		approvedAt = timestamppb.New(*t.ApprovedAt)
	}
	if t.ShippedAt != nil {
		shippedAt = timestamppb.New(*t.ShippedAt)
	}
	if t.ReceivedAt != nil {
		receivedAt = timestamppb.New(*t.ReceivedAt)
	}
	if t.CompletedAt != nil {
		completedAt = timestamppb.New(*t.CompletedAt)
	}

	return &pb.StockTransfer{
		Id:              uint64(t.ID),                 // ID。
		TransferNo:      t.TransferNo,                 // 调拨单号。
		FromWarehouseId: t.FromWarehouseID,            // 调出仓库ID。
		ToWarehouseId:   t.ToWarehouseID,              // 调入仓库ID。
		SkuId:           t.SkuID,                      // SKU ID。
		Quantity:        t.Quantity,                   // 数量。
		Status:          string(t.Status),             // 状态。
		Reason:          t.Reason,                     // 原因。
		ApprovedBy:      t.ApprovedBy,                 // 审核人ID。
		ApprovedAt:      approvedAt,                   // 审核时间。
		ShippedAt:       shippedAt,                    // 发货时间。
		ReceivedAt:      receivedAt,                   // 收货时间。
		CompletedAt:     completedAt,                  // 完成时间。
		Remark:          t.Remark,                     // 备注。
		CreatedBy:       t.CreatedBy,                  // 创建人ID。
		CreatedAt:       timestamppb.New(t.CreatedAt), // 创建时间。
		UpdatedAt:       timestamppb.New(t.UpdatedAt), // 更新时间。
	}
}
