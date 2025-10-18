package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	v1 "ecommerce/api/aftersales/v1"
	"ecommerce/internal/aftersales/model"
	"ecommerce/internal/aftersales/repository"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrReturnRequestNotFound = errors.New("return request not found")
	ErrRefundRequestNotFound = errors.New("refund request not found")
)

// AftersalesService 封装了售后相关的业务逻辑，实现了 aftersales.proto 中定义的 AftersalesServer 接口。
type AftersalesService struct {
	repo repository.AftersalesRepo
}

// NewAftersalesService 是 AftersalesService 的构造函数。
func NewAftersalesService(repo repository.AftersalesRepo) *AftersalesService {
	return &AftersalesService{repo: repo}
}

// CreateReturnRequest 实现了创建退货请求的 RPC 方法。
func (s *AftersalesService) CreateReturnRequest(ctx context.Context, req *v1.CreateReturnRequestRequest) (*v1.CreateReturnRequestResponse, error) {
	// TODO: 添加业务逻辑，例如验证订单，检查产品是否符合退货条件
	// 目前仅做基本转换和存储

	bizReq := &model.ReturnRequest{
		OrderID:   req.OrderId,
		UserID:    req.UserId,
		ProductID: req.ProductId,
		Quantity:  req.Quantity,
		Reason:    req.Reason,
		Status:    "PENDING", // 初始状态
	}

	createdReq, err := s.repo.CreateReturnRequest(ctx, bizReq)
	if err != nil {
		zap.S().Errorf("failed to create return request: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create return request")
	}

	return &v1.CreateReturnRequestResponse{
		Request: toProtoReturnRequest(createdReq),
	}, nil
}

// GetReturnRequest 实现了根据ID获取退货请求的 RPC 方法。
func (s *AftersalesService) GetReturnRequest(ctx context.Context, req *v1.GetReturnRequestRequest) (*v1.GetReturnRequestResponse, error) {
	id, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		zap.S().Warnf("GetReturnRequest: invalid ID format: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid return request ID format")
	}

	bizReq, err := s.repo.GetReturnRequest(ctx, uint(id))
	if err != nil {
		zap.S().Errorf("failed to get return request %d: %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to get return request")
	}
	if bizReq == nil {
		zap.S().Warnf("return request %d not found", id)
		return nil, status.Errorf(codes.NotFound, "return request not found")
	}

	return &v1.GetReturnRequestResponse{
		Request: toProtoReturnRequest(bizReq),
	}, nil
}

// UpdateReturnRequestStatus 实现了更新退货请求状态的 RPC 方法。
func (s *AftersalesService) UpdateReturnRequestStatus(ctx context.Context, req *v1.UpdateReturnRequestStatusRequest) (*v1.UpdateReturnRequestStatusResponse, error) {
	id, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		zap.S().Warnf("UpdateReturnRequestStatus: invalid ID format: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid return request ID format")
	}

	bizReq, err := s.repo.GetReturnRequest(ctx, uint(id))
	if err != nil {
		zap.S().Errorf("failed to get return request %d for status update: %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to get return request")
	}
	if bizReq == nil {
		zap.S().Warnf("return request %d not found for status update", id)
		return nil, status.Errorf(codes.NotFound, "return request not found")
	}

	// TODO: 添加状态流转的业务逻辑验证
	bizReq.Status = req.Status

	updatedReq, err := s.repo.UpdateReturnRequest(ctx, bizReq)
	if err != nil {
		zap.S().Errorf("failed to update return request %d status to %s: %v", id, req.Status, err)
		return nil, status.Errorf(codes.Internal, "failed to update return request status")
	}

	return &v1.UpdateReturnRequestStatusResponse{
		Request: toProtoReturnRequest(updatedReq),
	}, nil
}

// CreateRefundRequest 实现了创建退款请求的 RPC 方法。
func (s *AftersalesService) CreateRefundRequest(ctx context.Context, req *v1.CreateRefundRequestRequest) (*v1.CreateRefundRequestResponse, error) {
	// TODO: 添加业务逻辑，例如验证退货请求，检查支付状态
	// 目前仅做基本转换和存储

	bizReq := &model.RefundRequest{
		ReturnRequestID: req.ReturnRequestId,
		OrderID:         req.OrderId,
		UserID:          req.UserId,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          "PENDING", // 初始状态
	}

	createdReq, err := s.repo.CreateRefundRequest(ctx, bizReq)
	if err != nil {
		zap.S().Errorf("failed to create refund request: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create refund request")
	}

	return &v1.CreateRefundRequestResponse{
		Request: toProtoRefundRequest(createdReq),
	}, nil
}

// GetRefundRequest 实现了根据ID获取退款请求的 RPC 方法。
func (s *AftersalesService) GetRefundRequest(ctx context.Context, req *v1.GetRefundRequestRequest) (*v1.GetRefundRequestResponse, error) {
	id, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		zap.S().Warnf("GetRefundRequest: invalid ID format: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid refund request ID format")
	}

	bizReq, err := s.repo.GetRefundRequest(ctx, uint(id))
	if err != nil {
		zap.S().Errorf("failed to get refund request %d: %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to get refund request")
	}
	if bizReq == nil {
		zap.S().Warnf("refund request %d not found", id)
		return nil, status.Errorf(codes.NotFound, "refund request not found")
	}

	return &v1.GetRefundRequestResponse{
		Request: toProtoRefundRequest(bizReq),
	}, nil
}

// UpdateRefundRequestStatus 实现了更新退款请求状态的 RPC 方法。
func (s *AftersalesService) UpdateRefundRequestStatus(ctx context.Context, req *v1.UpdateRefundRequestStatusRequest) (*v1.UpdateRefundRequestStatusResponse, error) {
	id, err := strconv.ParseUint(req.Id, 10, 32)
	if err != nil {
		zap.S().Warnf("UpdateRefundRequestStatus: invalid ID format: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid refund request ID format")
	}

	bizReq, err := s.repo.GetRefundRequest(ctx, uint(id))
	if err != nil {
		zap.S().Errorf("failed to get refund request %d for status update: %v", id, err)
		return nil, status.Errorf(codes.Internal, "failed to get refund request")
	}
	if bizReq == nil {
		zap.S().Warnf("refund request %d not found for status update", id)
		return nil, status.Errorf(codes.NotFound, "refund request not found")
	}

	// TODO: 添加状态流转的业务逻辑验证
	bizReq.Status = req.Status

	updatedReq, err := s.repo.UpdateRefundRequest(ctx, bizReq)
	if err != nil {
		zap.S().Errorf("failed to update refund request %d status to %s: %v", id, req.Status, err)
		return nil, status.Errorf(codes.Internal, "failed to update refund request status")
	}

	return &v1.UpdateRefundRequestStatusResponse{
		Request: toProtoRefundRequest(updatedReq),
	}, nil
}

// --- 辅助函数 ---

// toProtoReturnRequest 将业务模型 model.ReturnRequest 转换为 proto 消息 v1.ReturnRequest。
func toProtoReturnRequest(bizReq *model.ReturnRequest) *v1.ReturnRequest {
	if bizReq == nil {
		return nil
	}
	return &v1.ReturnRequest{
		Id:        fmt.Sprintf("%d", bizReq.ID),
		OrderId:   bizReq.OrderID,
		UserId:    bizReq.UserID,
		ProductId: bizReq.ProductID,
		Quantity:  bizReq.Quantity,
		Reason:    bizReq.Reason,
		Status:    bizReq.Status,
		CreatedAt: timestamppb.New(bizReq.CreatedAt),
		UpdatedAt: timestamppb.New(bizReq.UpdatedAt),
	}
}

// toProtoRefundRequest 将业务模型 model.RefundRequest 转换为 proto 消息 v1.RefundRequest。
func toProtoRefundRequest(bizReq *model.RefundRequest) *v1.RefundRequest {
	if bizReq == nil {
		return nil
	}
	return &v1.RefundRequest{
		Id:              fmt.Sprintf("%d", bizReq.ID),
		ReturnRequestId: bizReq.ReturnRequestID,
		OrderId:         bizReq.OrderID,
		UserId:          bizReq.UserID,
		Amount:          bizReq.Amount,
		Currency:        bizReq.Currency,
		Status:          bizReq.Status,
		CreatedAt:       timestamppb.New(bizReq.CreatedAt),
		UpdatedAt:       timestamppb.New(bizReq.UpdatedAt),
	}
}
