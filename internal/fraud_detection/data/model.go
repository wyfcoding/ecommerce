package data

import (
	"time"

	"gorm.io/gorm"
)

// FraudEvaluation is the database model for a fraud evaluation result.
type FraudEvaluation struct {
	gorm.Model
	TransactionID string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	UserID        string    `gorm:"type:varchar(100);index"`
	Status        string    `gorm:"type:varchar(50);not null"` // e.g., PENDING, APPROVED, REJECTED, REVIEW
	Score         float64   `gorm:"not null"` // Fraud risk score
	Reasons       string    `gorm:"type:text"` // JSON string of reasons
	EvaluatedAt   time.Time `gorm:"not null"`
}

// TableName specifies the table name for the FraudEvaluation model.
func (FraudEvaluation) TableName() string {
	return "fraud_evaluations"
}

// FraudReport is the database model for a reported fraudulent activity.
type FraudReport struct {
	gorm.Model
	TransactionID string    `gorm:"type:varchar(100);index"`
	UserID        string    `gorm:"type:varchar(100);index"`
	ReportReason  string    `gorm:"type:text;not null"`
	Evidence      string    `gorm:"type:text"` // JSON string of evidence
	ReportedAt    time.Time `gorm:"not null"`
}

// TableName specifies the table name for the FraudReport model.
func (FraudReport) TableName() string {
	return "fraud_reports"
}
