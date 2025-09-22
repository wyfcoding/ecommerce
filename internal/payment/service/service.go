package service

import (
	"context"
	"errors"
	v1 "ecommerce/api/payment/v1"
	"ecommerce/internal/payment/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaymentService is the gRPC service implementation for payment.
type PaymentService struct {
	v1.UnimplementedPaymentServiceServer
	uc *biz.PaymentUsecase
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(uc *biz.PaymentUsecase) *PaymentService {
	return &PaymentService{uc: uc}
}

// CreatePayment implements the CreatePayment RPC.
func (s *PaymentService) CreatePayment(ctx context.Context, req *v1.CreatePaymentRequest) (*v1.CreatePaymentResponse, error) {
	if req.OrderId == 0 || req.UserId == 0 || req.Amount == 0 || req.Currency == "" || req.PaymentMethod == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id, user_id, amount, currency, and payment_method are required")
	}

	tx, redirectURL, qrCodeURL, err := s.uc.CreatePayment(ctx, req.OrderId, req.UserId, req.Amount, req.Currency, req.PaymentMethod, req.ReturnUrl)
	if err != nil {
		if errors.Is(err, biz.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, biz.ErrInvalidAmount) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to create payment: %v", err)
	}

	return &v1.CreatePaymentResponse{
		PaymentId:   tx.PaymentID,
		RedirectUrl: redirectURL,
		QrCodeUrl:   qrCodeURL,
		Status:      tx.Status,
	}, nil
}

// HandlePaymentCallback implements the HandlePaymentCallback RPC.
func (s *PaymentService) HandlePaymentCallback(ctx context.Context, req *v1.HandlePaymentCallbackRequest) (*v1.HandlePaymentCallbackResponse, error) {
	if req.PaymentMethod == "" || req.CallbackData == nil {
		return nil, status.Error(codes.InvalidArgument, "payment_method and callback_data are required")
	}

	tx, err := s.uc.HandlePaymentCallback(ctx, req.PaymentMethod, req.CallbackData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to handle payment callback: %v", err)
	}

	return &v1.HandlePaymentCallbackResponse{
		Status:  tx.Status,
		Message: "Callback processed successfully",
	}, nil
}
