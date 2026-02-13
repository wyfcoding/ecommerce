package mysql

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/insurance/domain"
	"gorm.io/gorm"
)

type InsuranceRepository struct {
	db *gorm.DB
}

func NewInsuranceRepository(db *gorm.DB) domain.Repository {
	return &InsuranceRepository{db: db}
}

func (r *InsuranceRepository) SavePolicy(ctx context.Context, policy *domain.InsurancePolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

func (r *InsuranceRepository) GetPolicy(ctx context.Context, policyID string) (*domain.InsurancePolicy, error) {
	var policy domain.InsurancePolicy
	if err := r.db.WithContext(ctx).Where("policy_id = ?", policyID).First(&policy).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("policy not found: %s", policyID)
		}
		return nil, err
	}
	return &policy, nil
}

func (r *InsuranceRepository) ListPolicies(ctx context.Context, userID, orderID string, offset, limit int) ([]*domain.InsurancePolicy, error) {
	var policies []*domain.InsurancePolicy
	query := r.db.WithContext(ctx)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if orderID != "" {
		query = query.Where("order_id = ?", orderID)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&policies).Error; err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *InsuranceRepository) SaveClaim(ctx context.Context, claim *domain.InsuranceClaim) error {
	return r.db.WithContext(ctx).Save(claim).Error
}

func (r *InsuranceRepository) GetClaim(ctx context.Context, claimID string) (*domain.InsuranceClaim, error) {
	var claim domain.InsuranceClaim
	if err := r.db.WithContext(ctx).Where("claim_id = ?", claimID).First(&claim).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("claim not found: %s", claimID)
		}
		return nil, err
	}
	return &claim, nil
}

// Ensure models are available for migration
func GetModels() []interface{} {
	return []interface{}{
		&domain.InsurancePolicy{},
		&domain.InsuranceClaim{},
	}
}
