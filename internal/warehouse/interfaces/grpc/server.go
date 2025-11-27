package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/api/warehouse/v1"
	"github.com/wyfcoding/ecommerce/internal/warehouse/application"
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedWarehouseServiceServer
	app *application.WarehouseService
}

func NewServer(app *application.WarehouseService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateWarehouse(ctx context.Context, req *pb.CreateWarehouseRequest) (*pb.CreateWarehouseResponse, error) {
	warehouse, err := s.app.CreateWarehouse(ctx, req.Code, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateWarehouseResponse{
		Warehouse: convertWarehouseToProto(warehouse),
	}, nil
}

func (s *Server) ListWarehouses(ctx context.Context, req *pb.ListWarehousesRequest) (*pb.ListWarehousesResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	warehouses, total, err := s.app.ListWarehouses(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbWarehouses := make([]*pb.Warehouse, len(warehouses))
	for i, w := range warehouses {
		pbWarehouses[i] = convertWarehouseToProto(w)
	}

	return &pb.ListWarehousesResponse{
		Warehouses: pbWarehouses,
		TotalCount: total,
	}, nil
}

func (s *Server) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*emptypb.Empty, error) {
	if err := s.app.UpdateStock(ctx, req.WarehouseId, req.SkuId, req.Quantity); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	stock, err := s.app.GetStock(ctx, req.WarehouseId, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetStockResponse{
		Stock: convertStockToProto(stock),
	}, nil
}

func (s *Server) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.CreateTransferResponse, error) {
	transfer, err := s.app.CreateTransfer(ctx, req.FromWarehouseId, req.ToWarehouseId, req.SkuId, req.Quantity, req.CreatedBy)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateTransferResponse{
		Transfer: convertTransferToProto(transfer),
	}, nil
}

func (s *Server) CompleteTransfer(ctx context.Context, req *pb.CompleteTransferRequest) (*emptypb.Empty, error) {
	if err := s.app.CompleteTransfer(ctx, req.TransferId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func convertWarehouseToProto(w *entity.Warehouse) *pb.Warehouse {
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

func convertStockToProto(s *entity.WarehouseStock) *pb.WarehouseStock {
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

func convertTransferToProto(t *entity.StockTransfer) *pb.StockTransfer {
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
