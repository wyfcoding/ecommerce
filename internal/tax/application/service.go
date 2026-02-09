package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/tax/domain"
)

type TaxService struct {
	repo   domain.TaxRepository
	logger *slog.Logger
}

func NewTaxService(repo domain.TaxRepository, logger *slog.Logger) *TaxService {
	return &TaxService{
		repo:   repo,
		logger: logger,
	}
}

func (s *TaxService) CalculateOrderTax(ctx context.Context, userID uint64, country, region, category string, amount int64) (*domain.TaxCalculationResult, error) {
	// 1. Check exemption
	exemption, err := s.repo.FindExemption(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check exemption", "error", err)
	}
	if exemption != nil {
		s.logger.InfoContext(ctx, "tax exemption applied", "user_id", userID, "reason", exemption.Reason)
		return &domain.TaxCalculationResult{
			TotalTaxAmount: 0,
			Currency:       "USD", // Default or passed in
			Items:          []*domain.TaxDetailItem{},
		}, nil
	}

	// 2. Find rules
	rules, err := s.repo.FindActiveRules(ctx, country, region, category)
	if err != nil {
		return nil, err
	}

	// 3. Calculate
	result := &domain.TaxCalculationResult{
		Currency: "USD", // Should arguably come from request
	}

	baseAmount := amount
	// Simple calculation: sum up applicable taxes
	// Complex logic (compound tax) can be added here
	for _, rule := range rules {
		taxAmount := rule.CalculateTax(baseAmount)
		if taxAmount > 0 {
			item := &domain.TaxDetailItem{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				TaxType:    rule.TaxType,
				BaseAmount: baseAmount,
				Rate:       rule.Rate,
				Amount:     taxAmount,
			}
			result.Items = append(result.Items, item)
			result.TotalTaxAmount += taxAmount

			if rule.IsCompound {
				baseAmount += taxAmount
			}
		}
	}

	return result, nil
}

func (s *TaxService) RecordInvoice(ctx context.Context, orderID uint64, result *domain.TaxCalculationResult) error {
	detailsJSON, _ := json.Marshal(result.Items)

	invoice := &domain.TaxInvoice{
		OrderID:      orderID,
		InvoiceNo:    fmt.Sprintf("INV-%d-%d", orderID, time.Now().Unix()), // Simple generation
		TotalNet:     0,                                                    // Needs base amount from somewhere, maybe result should include it
		TotalTax:     result.TotalTaxAmount,
		TotalGross:   0, // Net + Tax
		CalculatedAt: time.Now(),
		TaxDetails:   string(detailsJSON),
	}

	return s.repo.SaveInvoice(ctx, invoice)
}
