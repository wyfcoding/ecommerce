package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/warehouse/v1"          // 导入仓库模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/warehouse/application" // 导入仓库模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"      // 导入仓库模块的领域。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 WarehouseService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedWarehouseServiceServer                               // 嵌入生成的UnimplementedWarehouseServiceServer。
	app                                    *application.WarehouseService // 依赖Warehouse应用服务 facade。
}

// NewServer 创建并返回一个新的 Warehouse gRPC 服务端实例。
func NewServer(app *application.WarehouseService) *Server {
	return &Server{app: app}
}

// CreateWarehouse 处理创建仓库的gRPC请求。
func (s *Server) CreateWarehouse(ctx context.Context, req *pb.CreateWarehouseRequest) (*pb.CreateWarehouseResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateWarehouse received", "code", req.Code, "name", req.Name)

	warehouse, err := s.app.CreateWarehouse(ctx, req.Code, req.Name)
	if err != nil {
		slog.Error("gRPC CreateWarehouse failed", "code", req.Code, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create warehouse: %v", err))
	}

	slog.Info("gRPC CreateWarehouse successful", "warehouse_id", warehouse.ID, "duration", time.Since(start))
	return &pb.CreateWarehouseResponse{
		Warehouse: convertWarehouseToProto(warehouse),
	}, nil
}

// ListWarehouses 处理列出仓库的gRPC请求。
func (s *Server) ListWarehouses(ctx context.Context, req *pb.ListWarehousesRequest) (*pb.ListWarehousesResponse, error) {
	start := time.Now()
	slog.Debug("gRPC ListWarehouses received", "page", req.Page)

	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	warehouses, total, err := s.app.ListWarehouses(ctx, page, pageSize)
	if err != nil {
		slog.Error("gRPC ListWarehouses failed", "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list warehouses: %v", err))
	}

	pbWarehouses := make([]*pb.Warehouse, len(warehouses))
	for i, w := range warehouses {
		pbWarehouses[i] = convertWarehouseToProto(w)
	}

	slog.Debug("gRPC ListWarehouses successful", "count", len(pbWarehouses), "duration", time.Since(start))
	return &pb.ListWarehousesResponse{
		Warehouses: pbWarehouses,
		TotalCount: total,
	}, nil
}

// UpdateStock 处理更新库存的gRPC请求（增加或减少）。
func (s *Server) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC UpdateStock received", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "quantity", req.Quantity)

	if err := s.app.UpdateStock(ctx, req.WarehouseId, req.SkuId, req.Quantity); err != nil {
		slog.Error("gRPC UpdateStock failed", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update stock: %v", err))
	}

	slog.Info("gRPC UpdateStock successful", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// GetStock 处理获取库存的gRPC请求。
func (s *Server) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetStock received", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId)

	stock, err := s.app.GetStock(ctx, req.WarehouseId, req.SkuId)
	if err != nil {
		slog.Error("gRPC GetStock failed", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get stock: %v", err))
	}

	slog.Debug("gRPC GetStock successful", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "duration", time.Since(start))
	return &pb.GetStockResponse{
		Stock: convertStockToProto(stock),
	}, nil
}

// CreateTransfer 处理创建库存调拨单的gRPC请求。
func (s *Server) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.CreateTransferResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateTransfer received", "from", req.FromWarehouseId, "to", req.ToWarehouseId, "sku_id", req.SkuId)

	transfer, err := s.app.CreateTransfer(ctx, req.FromWarehouseId, req.ToWarehouseId, req.SkuId, req.Quantity, req.CreatedBy)
	if err != nil {
		slog.Error("gRPC CreateTransfer failed", "from", req.FromWarehouseId, "to", req.ToWarehouseId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create transfer: %v", err))
	}

	slog.Info("gRPC CreateTransfer successful", "transfer_id", transfer.ID, "duration", time.Since(start))
	return &pb.CreateTransferResponse{
		Transfer: convertTransferToProto(transfer),
	}, nil
}

// CompleteTransfer 处理完成库存调拨的gRPC请求。
func (s *Server) CompleteTransfer(ctx context.Context, req *pb.CompleteTransferRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC CompleteTransfer received", "transfer_id", req.TransferId)

	if err := s.app.CompleteTransfer(ctx, req.TransferId); err != nil {
		slog.Error("gRPC CompleteTransfer failed", "transfer_id", req.TransferId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to complete transfer: %v", err))
	}

	slog.Info("gRPC CompleteTransfer successful", "transfer_id", req.TransferId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// DeductStock 扣减库存（Saga正向操作）。
func (s *Server) DeductStock(ctx context.Context, req *pb.DeductStockRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC DeductStock received", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "quantity", req.Quantity)

	if err := s.app.DeductStock(ctx, req.WarehouseId, req.SkuId, req.Quantity); err != nil {
		slog.Error("gRPC DeductStock failed", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Aborted, fmt.Sprintf("failed to deduct stock for saga: %v", err))
	}

	slog.Info("gRPC DeductStock successful", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// RevertStock 回滚库存（Saga补偿操作）。
func (s *Server) RevertStock(ctx context.Context, req *pb.RevertStockRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC RevertStock received", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "quantity", req.Quantity)

	if err := s.app.RevertStock(ctx, req.WarehouseId, req.SkuId, req.Quantity); err != nil {
		slog.Error("gRPC RevertStock failed", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to revert stock for saga: %v", err))
	}

	slog.Info("gRPC RevertStock successful", "warehouse_id", req.WarehouseId, "sku_id", req.SkuId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// convertWarehouseToProto 是一个辅助函数，将领域层的 Warehouse 实体转换为 protobuf 的 Warehouse 消息。
func convertWarehouseToProto(w *domain.Warehouse) *pb.Warehouse {
	if w == nil {
		return nil
	}
	return &pb.Warehouse{
		Id:            uint64(w.ID),
		Code:          w.Code,
		Name:          w.Name,
		WarehouseType: w.WarehouseType,
		Province:      w.Province,
		City:          w.City,
		District:      w.District,
		Address:       w.Address,
		Longitude:     w.Longitude,
		Latitude:      w.Latitude,
		ContactName:   w.ContactName,
		ContactPhone:  w.ContactPhone,
		Priority:      w.Priority,
		Status:        string(w.Status),
		Capacity:      w.Capacity,
		Description:   w.Description,
		CreatedAt:     timestamppb.New(w.CreatedAt),
		UpdatedAt:     timestamppb.New(w.UpdatedAt),
	}
}

// convertStockToProto 是一个辅助函数，将领域层的 WarehouseStock 实体转换为 protobuf 的 WarehouseStock 消息。
func convertStockToProto(s *domain.WarehouseStock) *pb.WarehouseStock {
	if s == nil {
		return nil
	}
	return &pb.WarehouseStock{
		Id:          uint64(s.ID),
		WarehouseId: s.WarehouseID,
		SkuId:       s.SkuID,
		Stock:       s.Stock,
		LockedStock: s.LockedStock,
		SafeStock:   s.SafeStock,
		MaxStock:    s.MaxStock,
		CreatedAt:   timestamppb.New(s.CreatedAt),
		UpdatedAt:   timestamppb.New(s.UpdatedAt),
	}
}

// convertTransferToProto 是一个辅助函数，将领域层的 StockTransfer 实体转换为 protobuf 的 StockTransfer 消息。
func convertTransferToProto(t *domain.StockTransfer) *pb.StockTransfer {
	if t == nil {
		return nil
	}
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
		Id:              uint64(t.ID),
		TransferNo:      t.TransferNo,
		FromWarehouseId: t.FromWarehouseID,
		ToWarehouseId:   t.ToWarehouseID,
		SkuId:           t.SkuID,
		Quantity:        t.Quantity,
		Status:          string(t.Status),
		Reason:          t.Reason,
		ApprovedBy:      t.ApprovedBy,
		ApprovedAt:      approvedAt,
		ShippedAt:       shippedAt,
		ReceivedAt:      receivedAt,
		CompletedAt:     completedAt,
		Remark:          t.Remark,
		CreatedBy:       t.CreatedBy,
		CreatedAt:       timestamppb.New(t.CreatedAt),
		UpdatedAt:       timestamppb.New(t.UpdatedAt),
	}
}