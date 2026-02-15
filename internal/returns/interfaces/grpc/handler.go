package grpc

import (
	"context"
	"fmt"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/returns/v1"
	"github.com/wyfcoding/ecommerce/internal/returns/application"
	"github.com/wyfcoding/ecommerce/internal/returns/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedReturnsServiceServer
	svc    *application.ReturnService
	logger *slog.Logger
}

func NewServer(svc *application.ReturnService, logger *slog.Logger) pb.ReturnsServiceServer {
	return &Server{
		svc:    svc,
		logger: logger.With("component", "returns_grpc"),
	}
}

func (s *Server) CreateReturnRequest(ctx context.Context, req *pb.CreateReturnRequestRequest) (*pb.ReturnRequest, error) {
	items := make([]domain.ReturnItem, len(req.Items))
	for i, v := range req.Items {
		items[i] = domain.ReturnItem{
			SKUID:    v.SkuId,
			Quantity: v.Quantity,
			Reason:   v.Reason,
			Images:   v.Images,
		}
	}

	res, err := s.svc.CreateRequest(ctx, "temp_user_id", req.OrderId, items)
	if err != nil {
		return nil, err
	}
	return toReturnProto(res), nil
}

func (s *Server) UpdateReturnStatus(ctx context.Context, req *pb.UpdateReturnStatusRequest) (*pb.ReturnRequest, error) {
	// 示例简化：目前仅支持 Approve，如果是其他状态可在此扩展
	if req.Status == pb.ReturnStatus_RETURN_STATUS_APPROVED {
		res, err := s.svc.ApproveRequest(ctx, req.Id)
		if err != nil {
			return nil, err
		}
		return toReturnProto(res), nil
	}
	return nil, fmt.Errorf("unsupported status transition")
}

func (s *Server) LogQCResult(ctx context.Context, req *pb.LogQCResultRequest) (*pb.ReturnRequest, error) {
	res, err := s.svc.LogQC(ctx, req.Id, req.Passed, req.ConditionDetails)
	if err != nil {
		return nil, err
	}
	return toReturnProto(res), nil
}

func (s *Server) ListMyReturns(ctx context.Context, req *pb.ListMyReturnsRequest) (*pb.ListMyReturnsResponse, error) {
	list, total, err := s.svc.ListMyReturns(ctx, "temp_user_id", req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	protoList := make([]*pb.ReturnRequest, len(list))
	for i, r := range list {
		protoList[i] = toReturnProto(r)
	}

	return &pb.ListMyReturnsResponse{
		Returns: protoList,
		Total:   int32(total),
	}, nil
}

func (s *Server) GetReturnDetail(ctx context.Context, req *pb.GetReturnDetailRequest) (*pb.ReturnRequest, error) {
	res, err := s.svc.GetDetail(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toReturnProto(res), nil
}

func (s *Server) ProcessRefund(ctx context.Context, req *pb.ProcessRefundRequest) (*pb.ReturnRequest, error) {
	// TODO: 实现具体退款逻辑（调用 wallet）
	return nil, fmt.Errorf("refund processing not yet implemented")
}

func toReturnProto(req *domain.ReturnRequest) *pb.ReturnRequest {
	items := make([]*pb.Variation, len(req.Items))
	for i, v := range req.Items {
		items[i] = &pb.Variation{
			SkuId:    v.SKUID,
			Quantity: v.Quantity,
			Reason:   v.Reason,
			Images:   v.Images,
		}
	}

	return &pb.ReturnRequest{
		Id:             req.ID,
		OrderId:        req.OrderID,
		UserId:         req.UserID,
		Items:          items,
		Status:         req.Status,
		RmaNumber:      req.RMANumber,
		TrackingNumber: req.TrackingNumber,
		WarehouseNotes: req.WarehouseNotes,
		CreatedAt:      timestamppb.New(req.CreatedAt),
		UpdatedAt:      timestamppb.New(req.UpdatedAt),
	}
}
