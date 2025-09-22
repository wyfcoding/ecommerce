package data

import (
	"context"
	"ecommerce/internal/risk_security/biz"
	"ecommerce/internal/risk_security/data/model"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type riskSecurityRepo struct {
	data *Data
}

// NewRiskSecurityRepo creates a new RiskSecurityRepo.
func NewRiskSecurityRepo(data *Data) biz.RiskSecurityRepo {
	return &riskSecurityRepo{data: data}
}

// SaveFraudCheckResult saves a fraud check result to the database.
func (r *riskSecurityRepo) SaveFraudCheckResult(ctx context.Context, result *biz.FraudCheckResult) (*biz.FraudCheckResult, error) {
	additionalDataBytes, err := json.Marshal(result.AdditionalData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal additional data: %w", err)
	}

	po := &model.FraudCheckResult{
		UserID:        result.UserID,
		IPAddress:     result.IPAddress,
		DeviceInfo:    result.DeviceInfo,
		OrderID:       result.OrderID,
		Amount:        result.Amount,
		IsFraud:       result.IsFraud,
		RiskScore:     result.RiskScore,
		Decision:      result.Decision,
		Message:       result.Message,
		AdditionalData: string(additionalDataBytes),
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	result.ID = po.ID
	return result, nil
}
