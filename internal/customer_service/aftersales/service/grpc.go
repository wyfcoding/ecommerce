package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "ecommerce/api/aftersales/v1"
	"ecommerce/internal/aftersales/model"
)

// AftersalesGRPCService 是 aftersales.v1.AftersalesServiceServer 接口的实现。
// 它将 gRPC 请求转换为业务逻辑调用，并处理响应。
type AftersalesGRPCService struct {
	v1.UnimplementedAftersalesServiceServer
	bizService AftersalesService // 业务逻辑服务接口
	logger     *zap.Logger
}

// NewAftersalesGRPCService 创建一个新的 AftersalesGRPCService 实例。
// 接收业务逻辑服务接口和 zap.Logger 实例。
func NewAftersalesGRPCService(bizService AftersalesService, logger *zap.Logger) *AftersalesGRPCService {
	return &AftersalesGRPCService{
		bizService: bizService,
		logger:     logger,
	}
}

// CreateReturnRequest 实现了 aftersales.v1.AftersalesServiceServer 接口的 CreateReturnRequest 方法。
// 它处理创建退货请求的 gRPC 调用，将 proto 请求转换为业务模型，并调用业务逻辑服务。
func (s *AftersalesGRPCService) CreateReturnRequest(ctx context.Context, req *v1.CreateReturnRequestRequest) (*v1.CreateReturnRequestResponse, error) {
	s.logger.Info("Received CreateReturnRequest RPC", zap.Uint64("order_id", req.OrderId), zap.Uint64("user_id", req.UserId))

	// 转换 proto 请求到业务模型
	items := make([]model.AftersalesItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = model.AftersalesItem{
			OrderItemID: uint(item.OrderItemId),
			ProductID:   uint(item.ProductId),
			ProductSKU:  item.ProductSku,
			Quantity:    int(item.Quantity),
		}
	}

	bizApp, err := s.bizService.CreateApplication(ctx, uint(req.UserId), uint(req.OrderId), model.TypeReturn, req.Reason, items)
	if err != nil {
		s.logger.Error("Failed to create return request via business service", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to create return request: %v", err)
	}

	return &v1.CreateReturnRequestResponse{
		Request: toProtoAftersalesApplication(bizApp),
	}, nil
}

// GetReturnRequest 实现了 aftersales.v1.AftersalesServiceServer 接口的 GetReturnRequest 方法。
// 它处理获取退货请求详情的 gRPC 调用，并调用业务逻辑服务。
func (s *AftersalesGRPCService) GetReturnRequest(ctx context.Context, req *v1.GetReturnRequestRequest) (*v1.GetReturnRequestResponse, error) {
	s.logger.Info("Received GetReturnRequest RPC", zap.Uint64("request_id", req.RequestId))

	appID := uint(req.RequestId)
	// 假设这里没有用户ID，或者从认证信息中获取
	// 为了简化，这里假设是管理员调用，或者不进行用户ID校验
	bizApp, err := s.bizService.GetApplication(ctx, appID, nil, true) // 假设是管理员调用
	if err != nil {
		s.logger.Error("Failed to get return request via business service", zap.Error(err))
		if errors.Is(err, ErrApplicationNotFound) {
			return nil, status.Errorf(codes.NotFound, "Return request not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get return request: %v", err)
	}

	return &v1.GetReturnRequestResponse{
		Request: toProtoAftersalesApplication(bizApp),
	}, nil
}

// UpdateReturnRequestStatus 实现了 aftersales.v1.AftersalesServiceServer 接口的 UpdateReturnRequestStatus 方法。
// 它处理更新退货请求状态的 gRPC 调用，并根据请求的状态调用业务逻辑服务的不同方法。
func (s *AftersalesGRPCService) UpdateReturnRequestStatus(ctx context.Context, req *v1.UpdateReturnRequestStatusRequest) (*v1.UpdateReturnRequestStatusResponse, error) {
	s.logger.Info("Received UpdateReturnRequestStatus RPC", zap.Uint64("request_id", req.RequestId), zap.String("status", req.Status))

	appID := uint(req.RequestId)
	var updatedApp *model.AftersalesApplication
	var err error

	switch model.ApplicationStatus(req.Status) {
	case model.StatusApproved:
		updatedApp, err = s.bizService.ApproveApplication(ctx, appID, req.AdminRemarks)
	case model.StatusRejected:
		updatedApp, err = s.bizService.RejectApplication(ctx, appID, req.AdminRemarks)
	case model.StatusGoodsReceived:
		// 假设这里是管理员手动标记已收货，然后进入处理中
		app, getErr := s.bizService.GetApplication(ctx, appID, nil, true)
		if getErr != nil {
			s.logger.Error("Failed to get application for goods received status update", zap.Error(getErr))
			return nil, status.Errorf(codes.Internal, "Failed to get application: %v", getErr)
		}
		if app == nil {
			return nil, status.Errorf(codes.NotFound, "Application not found")
		}
		if app.Status != model.StatusApproved {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid status transition to GOODS_RECEIVED, expected APPROVED")
		}
		app.Status = model.StatusGoodsReceived
		app.AdminRemarks = req.AdminRemarks
		app.UpdatedAt = time.Now()
		if updateErr := s.bizService.repo.UpdateApplication(ctx, app); updateErr != nil { // 直接调用 repo 更新，跳过业务逻辑验证
			s.logger.Error("Failed to update application status to GOODS_RECEIVED", zap.Error(updateErr))
			return nil, status.Errorf(codes.Internal, "Failed to update application status")
		}
		updatedApp = app
	case model.StatusCompleted:
		updatedApp, err = s.bizService.CompleteApplication(ctx, appID, req.AdminRemarks)
	case model.StatusCancelled:
		updatedApp, err = s.bizService.CancelApplication(ctx, appID, req.AdminRemarks)
	default:
		s.logger.Warn("UpdateReturnRequestStatus: invalid status provided", zap.String("status", req.Status))
		return nil, status.Errorf(codes.InvalidArgument, "Invalid status provided")
	}

	if err != nil {
		s.logger.Error("Failed to update return request status via business service", zap.Error(err))
		if errors.Is(err, ErrApplicationNotFound) {
			return nil, status.Errorf(codes.NotFound, "Return request not found")
		}
		if errors.Is(err, ErrInvalidApplicationStatus) {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid application status transition")
		}
		return nil, status.Errorf(codes.Internal, "Failed to update return request status: %v", err)
	}

	return &v1.UpdateReturnRequestStatusResponse{
		Request: toProtoAftersalesApplication(updatedApp),
	}, nil
}

// CreateRefundRequest 实现了 aftersales.v1.AftersalesServiceServer 接口的 CreateRefundRequest 方法。
// 它处理创建退款请求的 gRPC 调用，并调用业务逻辑服务。
func (s *AftersalesGRPCService) CreateRefundRequest(ctx context.Context, req *v1.CreateRefundRequestRequest) (*v1.CreateRefundRequestResponse, error) {
	s.logger.Info("Received CreateRefundRequest RPC", zap.Uint64("return_request_id", req.ReturnRequestId))

	// 假设这里直接调用 ProcessReturnedGoods 来触发退款流程
	// 实际中可能需要更复杂的逻辑来创建独立的退款请求
	bizApp, err := s.bizService.ProcessReturnedGoods(ctx, uint(req.ReturnRequestId), req.Amount)
	if err != nil {
		s.logger.Error("Failed to create refund request via business service (ProcessReturnedGoods)", zap.Error(err))
		if errors.Is(err, ErrApplicationNotFound) {
			return nil, status.Errorf(codes.NotFound, "Return request not found for refund")
		}
		if errors.Is(err, ErrInvalidApplicationStatus) {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid application status for refund processing")
		}
		return nil, status.Errorf(codes.Internal, "Failed to create refund request: %v", err)
	}

	// 伪造一个 RefundRequest 响应，因为业务层直接处理了退款
	return &v1.CreateRefundRequestResponse{
		Request: &v1.RefundRequest{
			Id:              fmt.Sprintf("%d", bizApp.ID), // 使用售后申请ID作为退款ID
			ReturnRequestId: req.ReturnRequestId,
			OrderId:         bizApp.OrderID,
			UserId:          bizApp.UserID,
			Amount:          bizApp.RefundAmount,
			Currency:        req.Currency, // 假设货币从请求中来
			Status:          string(bizApp.Status),
			CreatedAt:       timestamppb.New(bizApp.CreatedAt),
			UpdatedAt:       timestamppb.New(bizApp.UpdatedAt),
		},
	}, nil
}

// GetRefundRequest 实现了 aftersales.v1.AftersalesServiceServer 接口的 GetRefundRequest 方法。
// 它处理获取退款请求详情的 gRPC 调用，并调用业务逻辑服务。
func (s *AftersalesGRPCService) GetRefundRequest(ctx context.Context, req *v1.GetRefundRequestRequest) (*v1.GetRefundRequestResponse, error) {
	s.logger.Info("Received GetRefundRequest RPC", zap.Uint64("refund_request_id", req.RefundRequestId))

	// 假设退款请求ID就是售后申请ID
	appID := uint(req.RefundRequestId)
	bizApp, err := s.bizService.GetApplication(ctx, appID, nil, true) // 假设是管理员调用
	if err != nil {
		s.logger.Error("Failed to get refund request via business service", zap.Error(err))
		if errors.Is(err, ErrApplicationNotFound) {
			return nil, status.Errorf(codes.NotFound, "Refund request not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get refund request: %v", err)
	}

	return &v1.GetRefundRequestResponse{
		Request: &v1.RefundRequest{
			Id:              fmt.Sprintf("%d", bizApp.ID),
			ReturnRequestId: bizApp.ID,
			OrderId:         bizApp.OrderID,
			UserId:          bizApp.UserID,
			Amount:          bizApp.RefundAmount,
			Currency:        "CNY", // 假设货币
			Status:          string(bizApp.Status),
			CreatedAt:       timestamppb.New(bizApp.CreatedAt),
			UpdatedAt:       timestamppb.New(bizApp.UpdatedAt),
		},
	}, nil
}

// UpdateRefundRequestStatus 实现了 aftersales.v1.AftersalesServiceServer 接口的 UpdateRefundRequestStatus 方法。
// 它处理更新退款请求状态的 gRPC 调用，并调用业务逻辑服务。
func (s *AftersalesGRPCService) UpdateRefundRequestStatus(ctx context.Context, req *v1.UpdateRefundRequestStatusRequest) (*v1.UpdateRefundRequestStatusResponse, error) {
	s.logger.Info("Received UpdateRefundRequestStatus RPC", zap.Uint64("refund_request_id", req.RefundRequestId), zap.String("status", req.Status))

	// 假设退款请求ID就是售后申请ID
	appID := uint(req.RefundRequestId)
	bizApp, err := s.bizService.GetApplication(ctx, appID, nil, true) // 假设是管理员调用
	if err != nil {
		s.logger.Error("Failed to get application for refund status update", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to get application: %v", err)
	}
	if bizApp == nil {
		return nil, status.Errorf(codes.NotFound, "Application not found")
	}

	// 假设这里直接更新售后申请的状态来反映退款状态
	// 实际中退款状态可能独立于售后申请状态
	bizApp.Status = model.ApplicationStatus(req.Status) // 直接映射状态
	bizApp.UpdatedAt = time.Now()

	if err := s.bizService.repo.UpdateApplication(ctx, bizApp); err != nil {
		s.logger.Error("Failed to update application status for refund request", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to update refund request status: %v", err)
	}

	return &v1.UpdateRefundRequestStatusResponse{
		Request: &v1.RefundRequest{
			Id:              fmt.Sprintf("%d", bizApp.ID),
			ReturnRequestId: bizApp.ID,
			OrderId:         bizApp.OrderID,
			UserId:          bizApp.UserID,
			Amount:          bizApp.RefundAmount,
			Currency:        req.Currency,
			Status:          string(bizApp.Status),
			CreatedAt:       timestamppb.New(bizApp.CreatedAt),
			UpdatedAt:       timestamppb.New(bizApp.UpdatedAt),
		},
	}, nil
}

// toProtoAftersalesApplication 将业务模型 model.AftersalesApplication 转换为 proto 消息 v1.AftersalesApplication。
func toProtoAftersalesApplication(bizApp *model.AftersalesApplication) *v1.AftersalesApplication {
	if bizApp == nil {
		return nil
	}
	protoItems := make([]*v1.AftersalesItem, len(bizApp.Items))
	for i, item := range bizApp.Items {
		protoItems[i] = &v1.AftersalesItem{
			Id:          uint64(item.ID),
			ApplicationId: uint64(item.ApplicationID),
			OrderItemId: uint64(item.OrderItemID),
			ProductId:   uint64(item.ProductID),
			ProductSku:  item.ProductSKU,
			Quantity:    int32(item.Quantity),
		}
	}

	return &v1.AftersalesApplication{
		Id:            uint64(bizApp.ID),
		ApplicationSn: bizApp.ApplicationSN,
		UserId:        uint64(bizApp.UserID),
		OrderId:       uint64(bizApp.OrderID),
		OrderSn:       bizApp.OrderSN,
		Type:          string(bizApp.Type),
		Status:        string(bizApp.Status),
		Reason:        bizApp.Reason,
		UserRemarks:   bizApp.UserRemarks,
		AdminRemarks:  bizApp.AdminRemarks,
		RefundAmount:  bizApp.RefundAmount,
		CreatedAt:     timestamppb.New(bizApp.CreatedAt),
		UpdatedAt:     timestamppb.New(bizApp.UpdatedAt),
		Items:         protoItems,
	}
}
