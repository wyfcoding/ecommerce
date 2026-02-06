package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dtm-labs/client/dtmgrpc"
	pb "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server gRPC 服务实现。
type Server struct {
	pb.UnimplementedPaymentServiceServer
	cmdService   *application.PaymentCommandService
	queryService *application.PaymentQuery
}

// NewServer 创建一个新的支付 gRPC 服务端实例。
func NewServer(cmd *application.PaymentCommandService, query *application.PaymentQuery) *Server {
	return &Server{
		cmdService:   cmd,
		queryService: query,
	}
}

// InitiatePayment 处理发起支付的 gRPC 请求。
func (s *Server) InitiatePayment(ctx context.Context, req *pb.InitiatePaymentRequest) (*pb.PaymentResponse, error) {
	start := time.Now()
	slog.Info("gRPC InitiatePayment received", "order_id", req.OrderId, "user_id", req.UserId, "amount", req.Amount, "method", req.PaymentMethod)

	cmd := &application.InitiatePaymentCommand{
		OrderID:       req.OrderId,
		UserID:        req.UserId,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		ClientIP:      req.ClientIp,
		DeviceID:      "unknown", // Header extraction usually happens at gateway or middleware
	}

	payment, gatewayResp, err := s.cmdService.InitiatePayment(ctx, cmd)
	if err != nil {
		slog.Error("gRPC InitiatePayment failed", "order_id", req.OrderId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, err.Error())
	}

	slog.Info("gRPC InitiatePayment successful", "order_id", req.OrderId, "payment_no", payment.PaymentNo, "duration", time.Since(start))
	return &pb.PaymentResponse{
		PaymentUrl:    gatewayResp.PaymentURL,
		PrepayId:      gatewayResp.TransactionID,
		TransactionNo: payment.PaymentNo,
	}, nil
}

// HandlePaymentCallback 处理支付结果异步回调
func (s *Server) HandlePaymentCallback(ctx context.Context, req *pb.HandlePaymentCallbackRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC HandlePaymentCallback received", "method", req.PaymentMethod)

	// 1. 提取核心参数
	paymentNo := req.CallbackData["out_trade_no"]
	if paymentNo == "" {
		paymentNo = req.CallbackData["payment_no"]
	}
	transactionID := req.CallbackData["trade_no"]
	if transactionID == "" {
		transactionID = req.CallbackData["transaction_id"]
	}
	statusVal := req.CallbackData["trade_status"]
	success := (statusVal == "TRADE_SUCCESS" || statusVal == "SUCCESS")

	if paymentNo == "" {
		slog.Error("gRPC HandlePaymentCallback failed: missing payment_no", "data", req.CallbackData)
		return nil, status.Error(codes.InvalidArgument, "missing out_trade_no in callback data")
	}

	// 2. 根据支付号反查用户 ID (用于分片路由)
	// We might need a global lookup or consistent hashing if sharded.
	// Assuming queryService handle this or we have a mapping.
	// For now, using a mock/temp userID if not provided, or querying via Repo.
	// The original code used App.GetUserIDByPaymentNo.
	// I'll add this to PaymentQuery.
	userID, err := s.queryService.GetUserIDByPaymentNo(ctx, paymentNo)
	if err != nil {
		slog.Error("gRPC HandlePaymentCallback failed: user not found", "payment_no", paymentNo, "error", err)
		return nil, status.Error(codes.NotFound, "payment record not found")
	}

	// 3. 调用应用层处理状态流转
	err = s.cmdService.HandlePaymentCallback(ctx, userID, paymentNo, success, transactionID, "", req.CallbackData)
	if err != nil {
		slog.Error("gRPC HandlePaymentCallback application error", "payment_no", paymentNo, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, err.Error())
	}

	slog.Info("gRPC HandlePaymentCallback processed successfully", "payment_no", paymentNo, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// GetPaymentStatus
func (s *Server) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.PaymentTransaction, error) {
	start := time.Now()
	slog.Debug("gRPC GetPaymentStatus received", "id", req.PaymentTransactionId)

	v := fmt.Sprintf("%v", req.PaymentTransactionId)

	// gRPC 请求中的 ID 可能是单号。尝试按单号查询。
	payment, err := s.queryService.GetPaymentStatus(ctx, 0, v)
	if err != nil {
		slog.Error("gRPC GetPaymentStatus failed", "id", req.PaymentTransactionId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if payment == nil {
		return nil, status.Error(codes.NotFound, "payment record not found")
	}

	slog.Debug("gRPC GetPaymentStatus successful", "id", req.PaymentTransactionId, "duration", time.Since(start))
	return convertPaymentToProto(payment), nil
}

// RequestRefund
func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.RefundTransaction, error) {
	start := time.Now()
	slog.Info("gRPC RequestRefund received", "payment_id", req.PaymentTransactionId, "user_id", req.UserId, "amount", req.RefundAmount)

	cmd := &application.RefundPaymentCommand{
		UserID:    req.UserId,
		PaymentID: req.PaymentTransactionId,
		Amount:    req.RefundAmount,
		Reason:    req.Reason,
	}

	refund, err := s.cmdService.RequestRefund(ctx, cmd)
	if err != nil {
		slog.Error("gRPC RequestRefund failed", "id", req.PaymentTransactionId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, err.Error())
	}

	slog.Info("gRPC RequestRefund successful", "refund_id", refund.ID, "user_id", req.UserId, "duration", time.Since(start))
	return convertRefundToProto(refund), nil
}

// SagaRefund Saga 正向: 退款
func (s *Server) SagaRefund(ctx context.Context, req *pb.SagaRefundRequest) (*pb.SagaRefundResponse, error) {
	barrier, err := dtmgrpc.BarrierFromGrpc(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "barrier error: %v", err)
	}

	refundNo, err := s.cmdService.SagaRefund(ctx, barrier, req.UserId, req.OrderId, req.RefundAmount, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "SagaRefund failed: %v", err)
	}

	return &pb.SagaRefundResponse{Success: true, RefundNo: refundNo}, nil
}

// SagaCancelRefund Saga 补偿: 记录失败
func (s *Server) SagaCancelRefund(ctx context.Context, req *pb.SagaRefundRequest) (*pb.SagaRefundResponse, error) {
	barrier, err := dtmgrpc.BarrierFromGrpc(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "barrier error: %v", err)
	}

	if err := s.cmdService.SagaCancelRefund(ctx, barrier, req.UserId, req.OrderId); err != nil {
		return nil, status.Errorf(codes.Internal, "SagaCancelRefund failed: %v", err)
	}

	return &pb.SagaRefundResponse{Success: true}, nil
}

// 辅助函数：将领域层的 Payment 实体转换为 Proto 消息对象。
func convertPaymentToProto(p *domain.Payment) *pb.PaymentTransaction {
	if p == nil {
		return nil
	}
	var paidAt *timestamppb.Timestamp
	if p.PaidAt != nil {
		paidAt = timestamppb.New(*p.PaidAt)
	}

	return &pb.PaymentTransaction{
		Id:                   uint64(p.ID),
		TransactionNo:        p.PaymentNo,
		OrderId:              p.OrderID,
		UserId:               p.UserID,
		PaymentMethod:        p.PaymentMethod,
		Amount:               p.Amount,
		Status:               p.Status,
		GatewayTransactionId: p.TransactionID,
		CreatedAt:            timestamppb.New(p.CreatedAt),
		UpdatedAt:            timestamppb.New(p.UpdatedAt),
		PaidAt:               paidAt,
	}
}

// 辅助函数：将领域层的 Refund 实体转换为 Proto 消息对象。
func convertRefundToProto(r *domain.Refund) *pb.RefundTransaction {
	if r == nil {
		return nil
	}
	return &pb.RefundTransaction{
		Id:                   uint64(r.ID),
		RefundNo:             r.RefundNo,
		PaymentTransactionId: r.PaymentID,
		OrderId:              r.OrderID,
		UserId:               r.UserID,
		RefundAmount:         r.RefundAmount,
		Status:               pb.RefundStatus(r.Status),
		Reason:               r.Reason,
		CreatedAt:            timestamppb.New(r.CreatedAt),
		UpdatedAt:            timestamppb.New(r.UpdatedAt),
	}
}
