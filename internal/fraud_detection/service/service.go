package service

import (
	"context"

	"ecommerce/internal/fraud_detection/biz"
	// Assuming the generated pb.go file will be in this path
	// v1 "ecommerce/api/fraud_detection/v1"
)

// FraudDetectionService is a gRPC service that implements the FraudDetectionServer interface.
// It holds a reference to the business logic layer.
type FraudDetectionService struct {
	// v1.UnimplementedFraudDetectionServer

	uc *biz.FraudDetectionUsecase
}

// NewFraudDetectionService creates a new FraudDetectionService.
func NewFraudDetectionService(uc *biz.FraudDetectionUsecase) *FraudDetectionService {
	return &FraudDetectionService{uc: uc}
}

// Note: The actual RPC methods like EvaluateTransaction, GetEvaluationStatus, etc., will be implemented here.
// These methods will call the corresponding business logic in the 'biz' layer.

/*
Example Implementation (once gRPC code is generated):

func (s *FraudDetectionService) EvaluateTransaction(ctx context.Context, req *v1.EvaluateTransactionRequest) (*v1.EvaluateTransactionResponse, error) {
    // 1. Call business logic
    evaluation, err := s.uc.EvaluateTransaction(ctx, req.TransactionId, req.UserId, req.Amount, req.Currency, req.PaymentMethodType, req.IpAddress, req.UserAgent, req.AdditionalData)
    if err != nil {
        return nil, err
    }

    // 2. Convert biz model to API model and return
    return &v1.EvaluateTransactionResponse{Evaluation: evaluation}, nil
}

*/
