package service

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/fraud_detection/model"
	"ecommerce/internal/fraud_detection/repository"
)

// ErrFraudEvaluationNotFound is a specific error for when a fraud evaluation is not found.
var ErrFraudEvaluationNotFound = errors.New("fraud evaluation not found")

// FraudDetectionService is the use case for fraud detection operations.
// It orchestrates the business logic.
type FraudDetectionService struct {
	repo repository.FraudDetectionRepo
	// You can also inject other dependencies like a logger or an external fraud engine client
}

// NewFraudDetectionService creates a new FraudDetectionService.
func NewFraudDetectionService(repo repository.FraudDetectionRepo) *FraudDetectionService {
	return &FraudDetectionService{repo: repo}
}

// EvaluateTransaction evaluates a transaction for fraud risk.
func (s *FraudDetectionService) EvaluateTransaction(ctx context.Context, transactionID, userID string, amount float64, currency, paymentMethodType, ipAddress, userAgent string, additionalData map[string]string) (*model.FraudEvaluation, error) {
	// TODO: Implement actual fraud detection logic here.
	// This would involve calling external fraud engines, applying rules, ML models, etc.
	// For now, we'll return a dummy evaluation.

	score := 0.1         // Dummy score
	status := "APPROVED" // Dummy status
	reasons := []string{}

	if amount > 1000 && paymentMethodType == "credit_card" {
		score = 0.8
		status = "REVIEW"
		reasons = append(reasons, "High value credit card transaction")
	}

	evaluation := &model.FraudEvaluation{
		TransactionID: transactionID,
		UserID:        userID,
		Status:        status,
		Score:         score,
		Reasons:       reasons,
		EvaluatedAt:   time.Now(),
	}

	return s.repo.CreateFraudEvaluation(ctx, evaluation)
}

// GetEvaluationStatus retrieves the status of a fraud evaluation.
func (s *FraudDetectionService) GetEvaluationStatus(ctx context.Context, transactionID string) (*model.FraudEvaluation, error) {
	return s.repo.GetFraudEvaluation(ctx, transactionID)
}

// ReportFraud reports a suspected fraudulent activity.
func (s *FraudDetectionService) ReportFraud(ctx context.Context, transactionID, userID, reportReason string, evidence map[string]string) (*model.FraudReport, error) {
	report := &model.FraudReport{
		TransactionID: transactionID,
		UserID:        userID,
		ReportReason:  reportReason,
		Evidence:      evidence,
		ReportedAt:    time.Now(),
	}
	return s.repo.CreateFraudReport(ctx, report)
}
