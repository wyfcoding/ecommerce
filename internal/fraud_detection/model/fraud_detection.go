package model

import "time"

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