package application

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/crossborder/domain"
)

type CrossBorderService struct {
	repo domain.CrossBorderRepository
}

func NewCrossBorderService(repo domain.CrossBorderRepository) *CrossBorderService {
	return &CrossBorderService{repo: repo}
}

// CalculateDuty 计算预计税费
func (s *CrossBorderService) CalculateDuty(ctx context.Context, items []struct {
	HSCode string
	Price  float64
	Qty    int32
}, destCountry string) (float64, float64, error) {
	var totalDuty, totalTax decimal.Decimal

	for _, item := range items {
		hs, err := s.repo.GetHSCode(ctx, item.HSCode)
		if err != nil {
			// fallback default rate if not found, or error
			// for simplicity: continue with 10% duty, 10% tax
			hs = &domain.HSCode{
				DutyRate: decimal.NewFromFloat(0.1),
				TaxRate:  decimal.NewFromFloat(0.1),
			}
		}

		amount := decimal.NewFromFloat(item.Price).Mul(decimal.NewFromInt32(item.Qty))

		duty := amount.Mul(hs.DutyRate)
		tax := amount.Add(duty).Mul(hs.TaxRate) // Tax often calculated on Value + Duty

		totalDuty = totalDuty.Add(duty)
		totalTax = totalTax.Add(tax)
	}

	return totalDuty.InexactFloat64(), totalTax.InexactFloat64(), nil
}

func (s *CrossBorderService) CreateDeclaration(ctx context.Context, orderID, userID, logisticsNo, currency string, declVal float64, items []struct {
	SKUID  string
	HSCode string
	Price  float64
	Qty    int32
}) (string, error) {
	id := fmt.Sprintf("DEC%d", time.Now().UnixNano())
	decl := domain.NewDeclaration(id, orderID, userID, logisticsNo, currency, decimal.NewFromFloat(declVal))

	// Calculate estimated duty
	var calcItems []struct {
		HSCode string
		Price  float64
		Qty    int32
	}
	for _, it := range items {
		calcItems = append(calcItems, struct {
			HSCode string
			Price  float64
			Qty    int32
		}{it.HSCode, it.Price, it.Qty})

		decl.AddItem(it.SKUID, it.HSCode, decimal.NewFromFloat(it.Price), it.Qty)
	}

	duty, tax, _ := s.CalculateDuty(ctx, calcItems, "DEFAULT")
	decl.DutyAmount = decimal.NewFromFloat(duty)
	decl.TaxAmount = decimal.NewFromFloat(tax)

	if err := s.repo.SaveDeclaration(ctx, decl); err != nil {
		return "", err
	}
	return id, nil
}

func (s *CrossBorderService) UpdateStatus(ctx context.Context, declIDStr, statusStr, reason string) error {
	decl, err := s.repo.GetDeclaration(ctx, declIDStr)
	if err != nil {
		return err
	}

	switch statusStr {
	case "SUBMITTED":
		_ = decl.Submit()
	case "CLEARED":
		_ = decl.Clear()
	case "REJECTED":
		_ = decl.Reject(reason)
	}

	return s.repo.SaveDeclaration(ctx, decl)
}
