package repository

import (
	"context"

	"ecommerce/internal/fraud_detection/model"
)

// FraudDetectionRepo defines the data storage interface for fraud detection data.
// The business layer depends on this interface, not on a concrete data implementation.
type FraudDetectionRepo interface {
	CreateFraudEvaluation(ctx context.Context, evaluation *model.FraudEvaluation) (*model.FraudEvaluation, error)
	GetFraudEvaluation(ctx context.Context, transactionID string) (*model.FraudEvaluation, error)
	UpdateFraudEvaluation(ctx context.Context, evaluation *model.FraudEvaluation) (*model.FraudEvaluation, error)
	CreateFraudReport(ctx context.Context, report *model.FraudReport) (*model.FraudReport, error)
}