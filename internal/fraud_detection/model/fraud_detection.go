package biz

import (
	"context"
	"errors"
	"time"
)

// ErrFraudEvaluationNotFound is a specific error for when a fraud evaluation is not found.
var ErrFraudEvaluationNotFound = errors.New("fraud evaluation not found")

// FraudEvaluation represents a fraud evaluation result in the business layer.
type FraudEvaluation struct {
	ID            uint
	TransactionID string
	UserID        string
	Status        string   // e.g., PENDING, APPROVED, REJECTED, REVIEW
	Score         float64  // Fraud risk score
	Reasons       []string // Reasons for the score/status
	EvaluatedAt   time.Time
}

// FraudReport represents a reported fraudulent activity in the business layer.
type FraudReport struct {
	ID            uint
	TransactionID string
	UserID        string
	ReportReason  string
	Evidence      map[string]string
	ReportedAt    time.Time
}

// FraudDetectionRepo defines the data storage interface for fraud detection data.
// The business layer depends on this interface, not on a concrete data implementation.
type FraudDetectionRepo interface {
	CreateFraudEvaluation(ctx context.Context, evaluation *FraudEvaluation) (*FraudEvaluation, error)
	GetFraudEvaluation(ctx context.Context, transactionID string) (*FraudEvaluation, error)
	UpdateFraudEvaluation(ctx context.Context, evaluation *FraudEvaluation) (*FraudEvaluation, error)
	CreateFraudReport(ctx context.Context, report *FraudReport) (*FraudReport, error)
}

// FraudDetectionUsecase is the use case for fraud detection operations.
// It orchestrates the business logic.
type FraudDetectionUsecase struct {
	repo FraudDetectionRepo
	// You can also inject other dependencies like a logger or an external fraud engine client
}

// NewFraudDetectionUsecase creates a new FraudDetectionUsecase.
func NewFraudDetectionUsecase(repo FraudDetectionRepo) *FraudDetectionUsecase {
	return &FraudDetectionUsecase{repo: repo}
}

// EvaluateTransaction evaluates a transaction for fraud risk.
func (uc *FraudDetectionUsecase) EvaluateTransaction(ctx context.Context, transactionID, userID string, amount float64, currency, paymentMethodType, ipAddress, userAgent string, additionalData map[string]string) (*FraudEvaluation, error) {
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

	evaluation := &FraudEvaluation{
		TransactionID: transactionID,
		UserID:        userID,
		Status:        status,
		Score:         score,
		Reasons:       reasons,
		EvaluatedAt:   time.Now(),
	}

	return uc.repo.CreateFraudEvaluation(ctx, evaluation)
}

// GetEvaluationStatus retrieves the status of a fraud evaluation.
func (uc *FraudDetectionUsecase) GetEvaluationStatus(ctx context.Context, transactionID string) (*FraudEvaluation, error) {
	return uc.repo.GetFraudEvaluation(ctx, transactionID)
}

// ReportFraud reports a suspected fraudulent activity.
func (uc *FraudDetectionUsecase) ReportFraud(ctx context.Context, transactionID, userID, reportReason string, evidence map[string]string) (*FraudReport, error) {
	report := &FraudReport{
		TransactionID: transactionID,
		UserID:        userID,
		ReportReason:  reportReason,
		Evidence:      evidence,
		ReportedAt:    time.Now(),
	}
	return uc.repo.CreateFraudReport(ctx, report)
}
