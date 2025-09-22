package biz

import (
	"context"
	"fmt"
	"time"
)

// ShippingRule represents a shipping rule in the business logic layer.
type ShippingRule struct {
	ID          uint
	Name        string
	Origin      string
	Destination string
	MinWeight   float64
	MaxWeight   float64
	BaseCost    uint64
	PerKgCost   uint64
}

// ItemInfo represents item information for shipping cost calculation.
type ItemInfo struct {
	ProductID uint64
	Quantity  uint32
	WeightKg  float64 // Weight per item in kilograms
}

// AddressInfo represents address information for shipping cost calculation.
type AddressInfo struct {
	Province string
	City     string
	District string
}

// LogisticsRepo defines the interface for logistics data access.
type LogisticsRepo interface {
	GetShippingRules(ctx context.Context, origin, destination string) ([]*ShippingRule, error)
}

// LogisticsUsecase is the business logic for logistics.
type LogisticsUsecase struct {
	repo LogisticsRepo
}

// NewLogisticsUsecase creates a new LogisticsUsecase.
func NewLogisticsUsecase(repo LogisticsRepo) *LogisticsUsecase {
	return &LogisticsUsecase{repo: repo}
}

// CalculateShippingCost calculates the shipping cost based on rules and item/address info.
func (uc *LogisticsUsecase) CalculateShippingCost(ctx context.Context, originAddress, destinationAddress *AddressInfo, items []*ItemInfo) (uint64, error) {
	// 1. Get applicable shipping rules
	// Simplified: using only city for origin/destination matching
	rules, err := uc.repo.GetShippingRules(ctx, originAddress.City, destinationAddress.City)
	if err != nil {
		return 0, err
	}
	if len(rules) == 0 {
		return 0, fmt.Errorf("no shipping rules found for %s to %s", originAddress.City, destinationAddress.City)
	}

	// 2. Calculate total weight
	totalWeightKg := 0.0
	for _, item := range items {
		totalWeightKg += item.WeightKg * float64(item.Quantity)
	}

	// 3. Apply rules (simplified: pick the first matching rule)
	var applicableRule *ShippingRule
	for _, rule := range rules {
		if totalWeightKg >= rule.MinWeight && (rule.MaxWeight == 0 || totalWeightKg <= rule.MaxWeight) {
			applicableRule = rule
			break
		}
	}

	if applicableRule == nil {
		return 0, fmt.Errorf("no applicable shipping rule found for total weight %.2fkg", totalWeightKg)
	}

	// 4. Calculate cost
	shippingCost := applicableRule.BaseCost
	if totalWeightKg > applicableRule.MinWeight {
		extraWeight := totalWeightKg - applicableRule.MinWeight
		shippingCost += uint64(extraWeight) * applicableRule.PerKgCost
	}

	return shippingCost, nil
}
