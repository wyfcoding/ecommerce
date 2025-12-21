package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server gRPC 服务实现。
type Server struct {
	pb.UnimplementedPaymentServer
	App *application.PaymentService
}

// NewServer 创建一个新的支付 gRPC 服务端实例。
func NewServer(app *application.PaymentService) *Server {
	return &Server{App: app}
}

// InitiatePayment 处理发起支付的 gRPC 请求。
func (s *Server) InitiatePayment(ctx context.Context, req *pb.InitiatePaymentRequest) (*pb.PaymentResponse, error) {
	payment, gatewayResp, err := s.App.InitiatePayment(ctx, req.OrderId, req.UserId, req.Amount, req.PaymentMethod)
	if err != nil {
		return nil, err
	}

	return &pb.PaymentResponse{
		PaymentUrl:    gatewayResp.PaymentURL,
		PrepayId:      gatewayResp.TransactionID,
		TransactionNo: payment.PaymentNo,
	}, nil
}

// HandlePaymentCallback 处理支付结果异步回调的 gRPC 封装请求（通常由网关或中转服务发起）。
func (s *Server) HandlePaymentCallback(ctx context.Context, req *pb.HandlePaymentCallbackRequest) (*emptypb.Empty, error) {
	// 尝试从 callback_data 中提取关键信息
	paymentNo := req.CallbackData["payment_no"]
	if paymentNo == "" {
		paymentNo = req.CallbackData["out_trade_no"]
	}

	success := req.CallbackData["status"] == "success" ||
		req.CallbackData["trade_status"] == "TRADE_SUCCESS" ||
		req.CallbackData["result_code"] == "SUCCESS"

	transactionID := req.CallbackData["trade_no"]
	if transactionID == "" {
		transactionID = req.CallbackData["transaction_id"]
	}

	if err := s.App.HandlePaymentCallback(ctx, paymentNo, success, transactionID, "", req.CallbackData); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetPaymentStatus 处理根据ID查询支付单状态的 gRPC 请求。
func (s *Server) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.PaymentTransaction, error) {
	payment, err := s.App.GetPaymentStatus(ctx, req.PaymentTransactionId)
	if err != nil {
		return nil, err
	}
	return convertPaymentToProto(payment), nil
}

// RequestRefund 处理针对支付单发起退款申请的 gRPC 请求。
func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.RefundTransaction, error) {
	refund, err := s.App.RequestRefund(ctx, req.PaymentTransactionId, req.RefundAmount, req.Reason)
	if err != nil {
		return nil, err
	}
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
		Id:                   p.ID,
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
		Id:                   r.ID,
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
