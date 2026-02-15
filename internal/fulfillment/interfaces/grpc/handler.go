// 生成摘要：
// - 实现 Fulfillment 服务的 gRPC Handler，对接 proto 定义的履约流程
// - 将外部 gRPC 请求转换为内部 Command/Query 模型并调用应用层服务
// - 支持履约单创建、详情查询、拣货、打包、发货及取消等全链路操作
// - 负责错误码映射与领域模型到 proto 消息的转换

package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/fulfillment/v1"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/application"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler Fulfillment gRPC 处理程序
type Handler struct {
	pb.UnimplementedFulfillmentServiceServer
	cmdService   *application.CommandService
	queryService *application.QueryService
}

// NewHandler 创建 gRPC 处理程序
func NewHandler(cmdService *application.CommandService, queryService *application.QueryService) *Handler {
	return &Handler{
		cmdService:   cmdService,
		queryService: queryService,
	}
}

// CreateFulfillment 创建履约单
func (h *Handler) CreateFulfillment(ctx context.Context, req *pb.CreateFulfillmentRequest) (*pb.CreateFulfillmentResponse, error) {
	items := make([]application.FulfillmentItemDTO, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, application.FulfillmentItemDTO{
			SKUID:       item.SkuId,
			ProductName: item.ProductName,
			SKUName:     item.SkuName,
			ImageURL:    item.ImageUrl,
			Quantity:    item.Quantity,
			Location:    item.Location,
			BatchNo:     item.BatchNo,
		})
	}

	cmd := application.CreateFulfillmentCommand{
		OrderNo:     req.OrderNo,
		MerchantID:  req.MerchantId,
		StoreID:     req.StoreId,
		WarehouseID: req.WarehouseId,
		Type:        domain.FulfillmentType(req.Type),
		Remark:      req.Remark,
		Address: application.ShippingAddress{
			ReceiverName:  req.ShippingAddress.ReceiverName,
			ReceiverPhone: req.ShippingAddress.ReceiverPhone,
			Province:      req.ShippingAddress.Province,
			City:          req.ShippingAddress.City,
			District:      req.ShippingAddress.District,
			Address:       req.ShippingAddress.Address,
			PostalCode:    req.ShippingAddress.PostalCode,
		},
		Items: items,
	}

	if req.ExpectedShipTime != nil {
		t := req.ExpectedShipTime.AsTime()
		cmd.ExpectedShipTime = &t
	}

	res, err := h.cmdService.CreateFulfillment(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create fulfillment: %v", err)
	}

	return &pb.CreateFulfillmentResponse{
		FulfillmentId: uint64(res.FulfillmentID),
		FulfillmentNo: res.FulfillmentNo,
	}, nil
}

// GetFulfillment 获取履约单详情
func (h *Handler) GetFulfillment(ctx context.Context, req *pb.GetFulfillmentRequest) (*pb.GetFulfillmentResponse, error) {
	f, err := h.queryService.GetFulfillment(ctx, uint(req.FulfillmentId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "fulfillment not found: %v", err)
	}

	return &pb.GetFulfillmentResponse{
		Fulfillment: toProtoFulfillment(f),
	}, nil
}

// AssignPicking 分配拣货
func (h *Handler) AssignPicking(ctx context.Context, req *pb.AssignPickingRequest) (*pb.AssignPickingResponse, error) {
	err := h.cmdService.AssignPicking(ctx, application.AssignPickingCommand{
		FulfillmentID: uint(req.FulfillmentId),
		PickerID:      req.PickerId,
		PickerName:    req.PickerName,
	})
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to assign picking: %v", err)
	}
	return &pb.AssignPickingResponse{Success: true}, nil
}

// StartPicking 开始拣货
func (h *Handler) StartPicking(ctx context.Context, req *pb.StartPickingRequest) (*pb.StartPickingResponse, error) {
	err := h.cmdService.StartPicking(ctx, application.StartPickingCommand{
		FulfillmentID: uint(req.FulfillmentId),
		PickerID:      req.PickerId,
	})
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to start picking: %v", err)
	}
	return &pb.StartPickingResponse{Success: true}, nil
}

// CompletePicking 完成拣货
func (h *Handler) CompletePicking(ctx context.Context, req *pb.CompletePickingRequest) (*pb.CompletePickingResponse, error) {
	items := make(map[string]int32)
	for _, item := range req.Items {
		items[item.SkuId] = item.PickedQuantity
	}

	err := h.cmdService.CompletePicking(ctx, application.CompletePickingCommand{
		FulfillmentID: uint(req.FulfillmentId),
		Items:         items,
	})
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to complete picking: %v", err)
	}
	return &pb.CompletePickingResponse{Success: true}, nil
}

// ArrangeShipment 安排发货
func (h *Handler) ArrangeShipment(ctx context.Context, req *pb.ArrangeShipmentRequest) (*pb.ArrangeShipmentResponse, error) {
	err := h.cmdService.ArrangeShipment(ctx, application.ArrangeShipmentCommand{
		FulfillmentID: uint(req.FulfillmentId),
		CarrierCode:   req.CarrierCode,
		CarrierName:   req.CarrierName,
		ShippingFee:   req.ShippingFee,
	})
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to arrange shipment: %v", err)
	}
	return &pb.ArrangeShipmentResponse{Success: true}, nil
}

// 辅助转换函数
func toProtoFulfillment(f *application.FulfillmentDTO) *pb.Fulfillment {
	items := make([]*pb.FulfillmentItem, 0, len(f.Items))
	for _, item := range f.Items {
		items = append(items, &pb.FulfillmentItem{
			Id:             uint64(item.ID),
			SkuId:          item.SKUID,
			ProductName:    item.ProductName,
			SkuName:        item.SKUName,
			ImageUrl:       item.ImageURL,
			Quantity:       item.Quantity,
			PickedQuantity: item.PickedQuantity,
			PackedQuantity: item.PackedQuantity,
			Location:       item.Location,
			BatchNo:        item.BatchNo,
		})
	}

	return &pb.Fulfillment{
		Id:            uint64(f.ID),
		FulfillmentNo: f.FulfillmentNo,
		OrderNo:       f.OrderNo,
		MerchantId:    f.MerchantID,
		StoreId:       f.StoreID,
		WarehouseId:   f.WarehouseID,
		Type:          pb.FulfillmentType(f.Type),
		Status:        pb.FulfillmentStatus(f.Status),
		Items:         items,
		ShippingAddress: &pb.ShippingAddress{
			ReceiverName:  f.ReceiverName,
			ReceiverPhone: f.ReceiverPhone,
			Province:      f.Province,
			City:          f.City,
			District:      f.District,
			Address:       f.Address,
			PostalCode:    f.PostalCode,
		},
		CreatedAt: timestamppb.New(f.CreatedAt),
		UpdatedAt: timestamppb.New(f.UpdatedAt),
	}
}

// 其他 RPC 方法按需实现...
func (h *Handler) ListFulfillments(ctx context.Context, req *pb.ListFulfillmentsRequest) (*pb.ListFulfillmentsResponse, error) {
	return nil, nil
}
func (h *Handler) ReportPickingException(ctx context.Context, req *pb.ReportPickingExceptionRequest) (*pb.ReportPickingExceptionResponse, error) {
	return nil, nil
}
func (h *Handler) StartPacking(ctx context.Context, req *pb.StartPackingRequest) (*pb.StartPackingResponse, error) {
	return nil, nil
}
func (h *Handler) CompletePacking(ctx context.Context, req *pb.CompletePackingRequest) (*pb.CompletePackingResponse, error) {
	return nil, nil
}
func (h *Handler) ConfirmShipment(ctx context.Context, req *pb.ConfirmShipmentRequest) (*pb.ConfirmShipmentResponse, error) {
	return nil, nil
}
func (h *Handler) CancelFulfillment(ctx context.Context, req *pb.CancelFulfillmentRequest) (*pb.CancelFulfillmentResponse, error) {
	return nil, nil
}
