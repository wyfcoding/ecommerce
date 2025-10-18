package repository

import (
	"context"

	"ecommerce/internal/risk_security/model"
)

// RiskSecurityRepo defines the interface for risk and security data access.
type RiskSecurityRepo interface {
	SaveFraudCheckResult(ctx context.Context, result *model.FraudCheckResult) (*model.FraudCheckResult, error)
}