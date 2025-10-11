package service

import (
	"context"

	v1 "ecommerce/api/payment/v1"
	"ecommerce/internal/payment/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PaymentService 是支付 gRPC 服务的实现。
type PaymentService struct {
	v1.UnimplementedPaymentServiceServer
	uc *biz.PaymentUsecase
}

// NewPaymentService 创建一个新的 PaymentService。
func NewPaymentService(uc *biz.PaymentUsecase) *PaymentService {
	return &PaymentService{uc: uc}
}

// bizPaymentToProto 将 biz.Payment 领域模型转换为 v1.Payment API 模型。
func bizPaymentToProto(p *biz.Payment) *v1.Payment {
	if p == nil {
		return nil
	}
	return &v1.Payment{
		PaymentId:     p.ID,
		OrderId:       p.OrderID,
		UserId:        p.UserID,
		Amount:        p.Amount,
		Method:        p.Method,
		Status:        p.Status,
		TransactionId: p.TransactionID,
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
		Currency:      &p.Currency,
		PaymentUrl:    p.PaymentURL,
	}
}

// CreatePayment 实现了 CreatePayment RPC。
func (s *PaymentService) CreatePayment(ctx context.Context, req *v1.CreatePaymentRequest) (*v1.CreatePaymentResponse, error) {
	// 基本参数校验
	if req.OrderId == "" || req.UserId == 0 || req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id, user_id, and a positive amount are required")
	}

	// 调用业务逻辑层创建支付
	payment, err := s.uc.CreatePayment(ctx, req.OrderId, req.UserId, req.Amount, req.GetCurrency(), req.Method)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create payment: %v", err)
	}

	// 返回包含支付信息的响应
	return &v1.CreatePaymentResponse{
		Payment: bizPaymentToProto(payment),
	}, nil
}

// ProcessCallback 实现了 ProcessCallback RPC。
func (s *PaymentService) ProcessCallback(ctx context.Context, req *v1.ProcessCallbackRequest) (*emptypb.Empty, error) {
	if req.PaymentId == "" || req.Data == nil {
		return nil, status.Error(codes.InvalidArgument, "payment_id and data are required")
	}

	err := s.uc.HandlePaymentCallback(ctx, req.PaymentId, req.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to handle payment callback: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// GetPaymentStatus 实现了 GetPaymentStatus RPC。
func (s *PaymentService) GetPaymentStatus(ctx context.Context, req *v1.GetPaymentStatusRequest) (*v1.Payment, error) {
	if req.PaymentId == "" {
		return nil, status.Error(codes.InvalidArgument, "payment_id is required")
	}

	payment, err := s.uc.repo.GetPaymentByID(ctx, req.PaymentId) // 直接通过 repo 获取
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "payment not found: %v", err)
	}

	return bizPaymentToProto(payment), nil
}

// CreateRefund 实现了 CreateRefund RPC。
func (s *PaymentService) CreateRefund(ctx context.Context, req *v1.CreateRefundRequest) (*v1.Refund, error) {
	// TODO: 实现退款业务逻辑
	return nil, status.Error(codes.Unimplemented, "rpc not implemented")
}