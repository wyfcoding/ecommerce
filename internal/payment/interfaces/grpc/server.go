package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/api/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedPaymentServer
	app *application.PaymentApplicationService
}

func NewServer(app *application.PaymentApplicationService) *Server {
	return &Server{app: app}
}

func (s *Server) InitiatePayment(ctx context.Context, req *pb.InitiatePaymentRequest) (*pb.PaymentResponse, error) {
	payment, err := s.app.InitiatePayment(ctx, req.OrderId, req.UserId, req.Amount, req.PaymentMethod)
	if err != nil {
		return nil, err
	}

	// Mock response for now
	return &pb.PaymentResponse{
		PaymentUrl:    "http://mock-payment-gateway.com/pay?id=" + payment.PaymentNo,
		TransactionNo: payment.PaymentNo,
	}, nil
}

func (s *Server) HandlePaymentCallback(ctx context.Context, req *pb.HandlePaymentCallbackRequest) (*emptypb.Empty, error) {
	// Extract data from callback_data
	paymentNo := req.CallbackData["payment_no"]
	success := req.CallbackData["status"] == "success"
	transactionID := req.CallbackData["transaction_id"]

	if err := s.app.HandlePaymentCallback(ctx, paymentNo, success, transactionID, ""); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.PaymentTransaction, error) {
	payment, err := s.app.GetPaymentStatus(ctx, req.PaymentTransactionId)
	if err != nil {
		return nil, err
	}
	return convertPaymentToProto(payment), nil
}

func (s *Server) RequestRefund(ctx context.Context, req *pb.RequestRefundRequest) (*pb.RefundTransaction, error) {
	refund, err := s.app.RequestRefund(ctx, req.PaymentTransactionId, req.RefundAmount, req.Reason)
	if err != nil {
		return nil, err
	}
	return convertRefundToProto(refund), nil
}

// Helper functions

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
		Status:               pb.RefundStatus(r.Status), // Map status correctly if needed
		Reason:               r.Reason,
		CreatedAt:            timestamppb.New(r.CreatedAt),
		UpdatedAt:            timestamppb.New(r.UpdatedAt),
	}
}
