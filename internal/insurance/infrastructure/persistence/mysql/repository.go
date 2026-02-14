package mysql

import (
	"context"
	"errors"
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("policy not found: %s", policyID)
		}
		return nil, err
	}
	return &policy, nil
}

func (r *InsuranceRepository) GetPolicyByOrder(ctx context.Context, orderID string) (*domain.InsurancePolicy, error) {
	var policy domain.InsurancePolicy
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at desc").First(&policy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *InsuranceRepository) ListPolicies(ctx context.Context, userID string, status domain.PolicyStatus, offset, limit int) ([]*domain.InsurancePolicy, int64, error) {
	var policies []*domain.InsurancePolicy
	var total int64
	query := r.db.WithContext(ctx)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Model(&domain.InsurancePolicy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&policies).Error; err != nil {
		return nil, 0, err
	}
	return policies, total, nil
}

func (r *InsuranceRepository) SaveClaim(ctx context.Context, claim *domain.InsuranceClaim) error {
	return r.db.WithContext(ctx).Save(claim).Error
}

func (r *InsuranceRepository) GetClaim(ctx context.Context, claimID string) (*domain.InsuranceClaim, error) {
	var claim domain.InsuranceClaim
	if err := r.db.WithContext(ctx).Where("claim_id = ?", claimID).First(&claim).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("claim not found: %s", claimID)
		}
		return nil, err
	}
	return &claim, nil
}

func (r *InsuranceRepository) ListClaims(ctx context.Context, policyID string, status domain.ClaimStatus, offset, limit int) ([]*domain.InsuranceClaim, int64, error) {
	var claims []*domain.InsuranceClaim
	var total int64
	query := r.db.WithContext(ctx).Model(&domain.InsuranceClaim{})
	if policyID != "" {
		query = query.Where("policy_id = ?", policyID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&claims).Error; err != nil {
		return nil, 0, err
	}
	return claims, total, nil
}

func (r *InsuranceRepository) ListPendingClaims(ctx context.Context, limit int) ([]*domain.InsuranceClaim, error) {
	if limit <= 0 {
		limit = 100
	}
	var claims []*domain.InsuranceClaim
	if err := r.db.WithContext(ctx).
		Where("status IN ?", []domain.ClaimStatus{domain.ClaimStatusSubmitted, domain.ClaimStatusUnderReview}).
		Order("created_at asc").
		Limit(limit).
		Find(&claims).Error; err != nil {
		return nil, err
	}
	return claims, nil
}

func (r *InsuranceRepository) SaveInsuranceProduct(ctx context.Context, product *domain.InsuranceProduct) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *InsuranceRepository) GetInsuranceProduct(ctx context.Context, productID string) (*domain.InsuranceProduct, error) {
	var product domain.InsuranceProduct
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *InsuranceRepository) ListInsuranceProducts(ctx context.Context, policyType domain.PolicyType) ([]*domain.InsuranceProduct, error) {
	var products []*domain.InsuranceProduct
	query := r.db.WithContext(ctx).Model(&domain.InsuranceProduct{})
	if policyType != "" {
		query = query.Where("type = ?", policyType)
	}
	if err := query.Where("status = ?", "ACTIVE").Order("created_at desc").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// Ensure models are available for migration
func GetModels() []any {
	return []any{
		&domain.InsurancePolicy{},
		&domain.InsuranceClaim{},
		&domain.InsuranceProduct{},
		&domain.ClaimDocument{},
	}
}
