package grpc

import (
	"context"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server gRPC 服务实现。
type Server struct {
	pb.UnimplementedPaymentServiceServer
	App *application.PaymentService
}

// NewServer 创建一个新的支付 gRPC 服务端实例。
func NewServer(app *application.PaymentService) *Server {
	return &Server{App: app}
}

// InitiatePayment 处理发起支付的 gRPC 请求。
func (s *Server) InitiatePayment(ctx context.Context, req *pb.InitiatePaymentRequest) (*pb.PaymentResponse, error) {
	start := time.Now()
	slog.Info("gRPC InitiatePayment received", "order_id", req.OrderId, "user_id", req.UserId, "amount", req.Amount, "method", req.PaymentMethod)

	payment, gatewayResp, err := s.App.InitiatePayment(ctx, req.OrderId, req.UserId, req.Amount, req.PaymentMethod)
	if err != nil {
		slog.Error("gRPC InitiatePayment failed", "order_id", req.OrderId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, err
	}

	slog.Info("gRPC InitiatePayment successful", "order_id", req.OrderId, "payment_no", payment.PaymentNo, "duration", time.Since(start))
	return &pb.PaymentResponse{
		PaymentUrl:    gatewayResp.PaymentURL,
		PrepayId:      gatewayResp.TransactionID,
		TransactionNo: payment.PaymentNo,
	}, nil
}

// HandlePaymentCallback 处理支付结果异步回调 (这里通常由 REST Handler 调用，但保留接口示例)
func (s *Server) HandlePaymentCallback(ctx context.Context, req *pb.HandlePaymentCallbackRequest) (*emptypb.Empty, error) {
	// ... 逻辑同之前，但需要调用正确的 App 方法（如果使用了 CallbackHandler）
	return &emptypb.Empty{}, nil
}

// GetPaymentStatus
func (s *Server) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.PaymentTransaction, error) {
	start := time.Now()
	slog.Debug("gRPC GetPaymentStatus received", "payment_transaction_id", req.PaymentTransactionId)

	// TODO: GetPaymentStatusRequest should include user_id for sharded lookup
	payment, err := s.App.GetPaymentStatus(ctx, 0, req.PaymentTransactionId)
	if err != nil {
		slog.Error("gRPC GetPaymentStatus failed", "id", req.PaymentTransactionId, "error", err, "duration", time.Since(start))
		return nil, err
	}

	slog.Debug("gRPC GetPaymentStatus successful", "id", req.PaymentTransactionId, "duration", time.Since(start))
	return convertPaymentToProto(payment), nil
}

// RequestRefund
func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.RefundTransaction, error) {
	start := time.Now()
	slog.Info("gRPC RequestRefund received", "payment_id", req.PaymentTransactionId, "user_id", req.UserId, "amount", req.RefundAmount)

	refund, err := s.App.RequestRefund(ctx, req.UserId, req.PaymentTransactionId, req.RefundAmount, req.Reason)
	if err != nil {
		slog.Error("gRPC RequestRefund failed", "id", req.PaymentTransactionId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, err
	}

	slog.Info("gRPC RequestRefund successful", "refund_id", refund.ID, "user_id", req.UserId, "duration", time.Since(start))
	return convertRefundToProto(refund), nil
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
		Id:                   uint64(p.ID), // 显式转换
		TransactionNo:        p.PaymentNo,
		OrderId:              p.OrderID,
		UserId:               p.UserID,
		PaymentMethod:        p.PaymentMethod,
		Amount:               p.Amount,
		Status:               pb.PaymentStatus(p.Status),
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
		Id:                   uint64(r.ID), // 显式转换
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
