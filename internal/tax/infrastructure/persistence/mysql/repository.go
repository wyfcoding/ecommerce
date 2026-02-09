package mysql

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/tax/domain"
	"gorm.io/gorm"
)

type TaxRepository struct {
	db *gorm.DB
}

func NewTaxRepository(db *gorm.DB) domain.TaxRepository {
	return &TaxRepository{db: db}
}

func (r *TaxRepository) FindActiveRules(ctx context.Context, country, region string, category string) ([]*domain.TaxRule, error) {
	var models []TaxRuleModel
	query := r.db.WithContext(ctx).Where("country_code = ? AND is_active = ?", country, true)

	if region != "" {
		query = query.Where("(region_code = ? OR region_code = '')", region)
	} else {
		query = query.Where("region_code = ''")
	}

	if category != "" {
		query = query.Where("(category = ? OR category = 'ALL')", category)
	}

	// 按优先级降序
	if err := query.Order("priority desc").Find(&models).Error; err != nil {
		return nil, err
	}

	rules := make([]*domain.TaxRule, len(models))
	for i, m := range models {
		rules[i] = toDomainTaxRule(&m)
	}
	return rules, nil
}

func (r *TaxRepository) SaveRule(ctx context.Context, rule *domain.TaxRule) error {
	model := TaxRuleModel{
		Name:        rule.Name,
		CountryCode: rule.CountryCode,
		RegionCode:  rule.RegionCode,
		TaxType:     int(rule.TaxType),
		Rate:        rule.Rate,
		FixedAmount: rule.FixedAmount,
		Category:    rule.Category,
		Priority:    rule.Priority,
		IsCompound:  rule.IsCompound,
		IsActive:    rule.IsActive,
		StartTime:   rule.StartTime,
		EndTime:     rule.EndTime,
	}
	// gorm.Model handling for ID
	if rule.ID > 0 {
		model.ID = uint(rule.ID)
	}

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return err
	}
	rule.ID = uint64(model.ID)
	return nil
}

func (r *TaxRepository) FindExemption(ctx context.Context, userID uint64) (*domain.TaxExemption, error) {
	var model TaxExemptionModel
	if err := r.db.WithContext(ctx).Where("user_id = ? AND status = 1", userID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomainTaxExemption(&model), nil
}

func (r *TaxRepository) SaveExemption(ctx context.Context, exemption *domain.TaxExemption) error {
	model := TaxExemptionModel{
		UserID:         exemption.UserID,
		Reason:         exemption.Reason,
		CertificateID:  exemption.CertificateID,
		CertificateImg: exemption.CertificateImg,
		ValidFrom:      exemption.ValidFrom,
		ValidTo:        exemption.ValidTo,
		Status:         exemption.Status,
	}
	if exemption.ID > 0 {
		model.ID = uint(exemption.ID)
	}
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return err
	}
	exemption.ID = uint64(model.ID)
	return nil
}

func (r *TaxRepository) SaveInvoice(ctx context.Context, invoice *domain.TaxInvoice) error {
	model := TaxInvoiceModel{
		OrderID:      invoice.OrderID,
		InvoiceNo:    invoice.InvoiceNo,
		TaxID:        invoice.TaxID,
		TotalNet:     invoice.TotalNet,
		TotalTax:     invoice.TotalTax,
		TotalGross:   invoice.TotalGross,
		CalculatedAt: invoice.CalculatedAt,
		TaxDetails:   invoice.TaxDetails,
	}
	if invoice.ID > 0 {
		model.ID = uint(invoice.ID)
	}
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return err
	}
	invoice.ID = uint64(model.ID)
	return nil
}

func (r *TaxRepository) FindByOrder(ctx context.Context, orderID uint64) (*domain.TaxInvoice, error) {
	var model TaxInvoiceModel
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&model).Error; err != nil {
		return nil, fmt.Errorf("invoice not found for order %d: %w", orderID, err)
	}
	return toDomainTaxInvoice(&model), nil
}
