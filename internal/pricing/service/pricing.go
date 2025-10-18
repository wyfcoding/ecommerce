package service

import (
	"context"
	"fmt"
	"time"

	"ecommerce/internal/pricing/client"
	"ecommerce/internal/pricing/model"
	"ecommerce/internal/pricing/repository"
)

// PricingService is the business logic for pricing.
type PricingService struct {
	repo          repository.PricingRepo
	aiModelClient client.AIModelClient // Added AI Model client
	// TODO: Add MarketingClient for coupon/promotion details
}

// NewPricingService creates a new PricingService.
func NewPricingService(repo repository.PricingRepo, aiModelClient client.AIModelClient) *PricingService {
	return &PricingService{repo: repo, aiModelClient: aiModelClient}
}

// CalculateFinalPrice calculates the final price for a list of SKUs, considering rules and discounts.
func (s *PricingService) CalculateFinalPrice(ctx context.Context, userID uint64, items []*model.SkuPriceInfo, couponCode string) (totalOriginalPrice, totalDiscountAmount, finalPrice uint64, err error) {
	totalOriginalPrice = 0
	totalDiscountAmount = 0

	// 1. Calculate total original price and apply base/special pricing rules
	skuIDs := make([]uint64, len(items))
	itemMap := make(map[uint64]*model.SkuPriceInfo)
	for i, item := range items {
		skuIDs[i] = item.SkuID
		itemMap[item.SkuID] = item
		totalOriginalPrice += item.OriginalPrice * uint64(item.Quantity)
	}

	// Get applicable price rules (e.g., special prices)
	priceRules, err := s.repo.GetPriceRulesForSKUs(ctx, skuIDs)
	if err != nil {
		return 0, 0, 0, err
	}

	// Apply price rules (simplified: assume highest priority rule applies)
	currentPrices := make(map[uint64]uint64) // skuID -> effective price
	for _, item := range items {
		currentPrices[item.SkuID] = item.OriginalPrice // Start with original price
	}

	for _, rule := range priceRules {
		if item, ok := itemMap[rule.SKUID]; ok {
			// For simplicity, just apply the rule price if it's a special price
			if rule.RuleType == "SPECIAL" {
				currentPrices[item.SKUID] = rule.Price
			}
		}
	}

	// Calculate price after base/special rules
	priceAfterRules := uint64(0)
	for skuID, price := range currentPrices {
		priceAfterRules += price * uint64(itemMap[skuID].Quantity)
	}

	// 2. Apply discounts (e.g., coupons, promotions)
	// This is a placeholder. In a real system, this would involve:
	// - Calling Marketing Service to validate and apply coupons/promotions.
	// - Handling discount stacking rules.
	if couponCode != "" {
		// Simulate a fixed discount for demonstration
		simulatedCouponDiscount := uint64(1000) // 10.00
		if priceAfterRules >= simulatedCouponDiscount {
			totalDiscountAmount += simulatedCouponDiscount
			// Save discount record
			_, err = s.repo.SaveDiscount(ctx, &model.Discount{
				DiscountType: "COUPON",
				DiscountID:   0, // Placeholder
				Amount:       simulatedCouponDiscount,
				AppliedTo:    "ORDER",
			})
			if err != nil {
				// Log error, but don't block price calculation
				fmt.Printf("failed to save coupon discount: %v\n", err)
			}
		}
	}

	finalPrice = priceAfterRules - totalDiscountAmount
	if finalPrice < 0 { // Ensure final price is not negative
		finalPrice = 0
	}

	return totalOriginalPrice, totalDiscountAmount, finalPrice, nil
}

// CalculateDynamicPrice calculates the dynamic price for a product using an AI model.
func (s *PricingService) CalculateDynamicPrice(ctx context.Context, productID, userID uint64, contextFeatures map[string]string) (uint64, string, error) {
	// Call AI Model Service to get dynamic price
	dynamicPrice, explanation, err := s.aiModelClient.CalculateDynamicPrice(ctx, productID, userID, contextFeatures)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get dynamic price from AI model: %w", err)
	}
	return dynamicPrice, explanation, nil
}
