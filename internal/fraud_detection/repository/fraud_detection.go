package data

import (
	"context"
	"ecommerce/internal/fraud_detection/biz"
	"encoding/json"

	"gorm.io/gorm"
)

// fraudDetectionRepo is the data layer implementation for FraudDetectionRepo.
type fraudDetectionRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.FraudEvaluation model to a biz.FraudEvaluation entity.
func (e *FraudEvaluation) toBiz() *biz.FraudEvaluation {
	if e == nil {
		return nil
	}
	var reasons []string
	_ = json.Unmarshal([]byte(e.Reasons), &reasons) // Ignore error for simplicity

	return &biz.FraudEvaluation{
		ID:            e.ID,
		TransactionID: e.TransactionID,
		UserID:        e.UserID,
		Status:        e.Status,
		Score:         e.Score,
		Reasons:       reasons,
		EvaluatedAt:   e.EvaluatedAt,
	}
}

// fromBiz converts a biz.FraudEvaluation entity to a data.FraudEvaluation model.
func fromBizFraudEvaluation(b *biz.FraudEvaluation) *FraudEvaluation {
	if b == nil {
		return nil
	}
	reasonsJSON, _ := json.Marshal(b.Reasons) // Ignore error for simplicity

	return &FraudEvaluation{
		TransactionID: b.TransactionID,
		UserID:        b.UserID,
		Status:        b.Status,
		Score:         b.Score,
		Reasons:       string(reasonsJSON),
		EvaluatedAt:   b.EvaluatedAt,
	}
}

// toBiz converts a data.FraudReport model to a biz.FraudReport entity.
func (r *FraudReport) toBiz() *biz.FraudReport {
	if r == nil {
		return nil
	}
	var evidence map[string]string
	_ = json.Unmarshal([]byte(r.Evidence), &evidence) // Ignore error for simplicity

	return &biz.FraudReport{
		ID:            r.ID,
		TransactionID: r.TransactionID,
		UserID:        r.UserID,
		ReportReason:  r.ReportReason,
		Evidence:      evidence,
		ReportedAt:    r.ReportedAt,
	}
}

// fromBiz converts a biz.FraudReport entity to a data.FraudReport model.
func fromBizFraudReport(b *biz.FraudReport) *FraudReport {
	if b == nil {
		return nil
	}
	evidenceJSON, _ := json.Marshal(b.Evidence) // Ignore error for simplicity

	return &FraudReport{
		TransactionID: b.TransactionID,
		UserID:        b.UserID,
		ReportReason:  b.ReportReason,
		Evidence:      string(evidenceJSON),
		ReportedAt:    b.ReportedAt,
	}
}

// CreateFraudEvaluation creates a new fraud evaluation record in the database.
func (r *fraudDetectionRepo) CreateFraudEvaluation(ctx context.Context, b *biz.FraudEvaluation) (*biz.FraudEvaluation, error) {
	evaluation := fromBizFraudEvaluation(b)
	if err := r.data.db.WithContext(ctx).Create(evaluation).Error; err != nil {
		return nil, err
	}
	return evaluation.toBiz(), nil
}

// GetFraudEvaluation retrieves a fraud evaluation record from the database by transaction ID.
func (r *fraudDetectionRepo) GetFraudEvaluation(ctx context.Context, transactionID string) (*biz.FraudEvaluation, error) {
	var evaluation FraudEvaluation
	if err := r.data.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&evaluation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrFraudEvaluationNotFound
		}
		return nil, err
	}
	return evaluation.toBiz(), nil
}

// UpdateFraudEvaluation updates an existing fraud evaluation record in the database.
func (r *fraudDetectionRepo) UpdateFraudEvaluation(ctx context.Context, b *biz.FraudEvaluation) (*biz.FraudEvaluation, error) {
	evaluation := fromBizFraudEvaluation(b)
	if err := r.data.db.WithContext(ctx).Save(evaluation).Error; err != nil {
		return nil, err
	}
	return evaluation.toBiz(), nil
}

// CreateFraudReport creates a new fraud report record in the database.
func (r *fraudDetectionRepo) CreateFraudReport(ctx context.Context, b *biz.FraudReport) (*biz.FraudReport, error) {
	report := fromBizFraudReport(b)
	if err := r.data.db.WithContext(ctx).Create(report).Error; err != nil {
		return nil, err
	}
	return report.toBiz(), nil
}
