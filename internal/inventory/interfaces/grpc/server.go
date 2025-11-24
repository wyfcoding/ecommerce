package grpc

import (
	"context"
	pb "ecommerce/api/inventory/v1"
	"ecommerce/internal/inventory/application"
	"ecommerce/internal/inventory/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedInventoryServiceServer
	app *application.InventoryService
}

func NewServer(app *application.InventoryService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateInventory(ctx context.Context, req *pb.CreateInventoryRequest) (*pb.CreateInventoryResponse, error) {
	inventory, err := s.app.CreateInventory(ctx, req.SkuId, req.ProductId, req.WarehouseId, req.TotalStock, req.WarningThreshold)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateInventoryResponse{
		Inventory: convertInventoryToProto(inventory),
	}, nil
}

func (s *Server) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	inventory, err := s.app.GetInventory(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if inventory == nil {
		return nil, status.Error(codes.NotFound, "inventory not found")
	}

	return &pb.GetInventoryResponse{
		Inventory: convertInventoryToProto(inventory),
	}, nil
}

func (s *Server) AddStock(ctx context.Context, req *pb.AddStockRequest) (*emptypb.Empty, error) {
	if err := s.app.AddStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) DeductStock(ctx context.Context, req *pb.DeductStockRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) LockStock(ctx context.Context, req *pb.LockStockRequest) (*emptypb.Empty, error) {
	if err := s.app.LockStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) UnlockStock(ctx context.Context, req *pb.UnlockStockRequest) (*emptypb.Empty, error) {
	if err := s.app.UnlockStock(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ConfirmDeduction(ctx context.Context, req *pb.ConfirmDeductionRequest) (*emptypb.Empty, error) {
	if err := s.app.ConfirmDeduction(ctx, req.SkuId, req.Quantity, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListInventories(ctx context.Context, req *pb.ListInventoriesRequest) (*pb.ListInventoriesResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	inventories, total, err := s.app.ListInventories(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbInventories := make([]*pb.Inventory, len(inventories))
	for i, inv := range inventories {
		pbInventories[i] = convertInventoryToProto(inv)
	}

	return &pb.ListInventoriesResponse{
		Inventories: pbInventories,
		TotalCount:  uint64(total),
	}, nil
}

func (s *Server) GetInventoryLogs(ctx context.Context, req *pb.GetInventoryLogsRequest) (*pb.GetInventoryLogsResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	logs, total, err := s.app.GetInventoryLogs(ctx, req.InventoryId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbLogs := make([]*pb.InventoryLog, len(logs))
	for i, log := range logs {
		pbLogs[i] = convertLogToProto(log)
	}

	return &pb.GetInventoryLogsResponse{
		Logs:       pbLogs,
		TotalCount: uint64(total),
	}, nil
}

func convertInventoryToProto(inv *entity.Inventory) *pb.Inventory {
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

func convertLogToProto(log *entity.InventoryLog) *pb.InventoryLog {
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
