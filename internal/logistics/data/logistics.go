package data

import (
	"context"
	"ecommerce/internal/logistics/biz"
	"ecommerce/internal/logistics/data/model"
	"time"

	"gorm.io/gorm"
)

type logisticsRepo struct {
	data *Data
}

// NewLogisticsRepo creates a new LogisticsRepo.
func NewLogisticsRepo(data *Data) biz.LogisticsRepo {
	return &logisticsRepo{data: data}
}

// GetShippingRules retrieves applicable shipping rules.
func (r *logisticsRepo) GetShippingRules(ctx context.Context, origin, destination string) ([]*biz.ShippingRule, error) {
	var rules []*model.ShippingRule
	// Simplified query: In a real system, this would involve more complex matching
	// based on province, city, district, weight, volume, etc.
	if err := r.data.db.WithContext(ctx).Where("origin = ? AND destination = ?", origin, destination).Find(&rules).Error; err != nil {
		return nil, err
	}

	bizRules := make([]*biz.ShippingRule, len(rules))
	for i, r := range rules {
		bizRules[i] = &biz.ShippingRule{
			ID:          r.ID,
			Name:        r.Name,
			Origin:      r.Origin,
			Destination: r.Destination,
			MinWeight:   r.MinWeight,
			MaxWeight:   r.MaxWeight,
			BaseCost:    r.BaseCost,
			PerKgCost:   r.PerKgCost,
		}
	}
	return bizRules, nil
}
