package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/tax/domain"
	"gorm.io/gorm"
)

// TaxRuleModel 税务规则 GORM 模型
type TaxRuleModel struct {
	gorm.Model
	Name        string    `gorm:"column:name;type:varchar(100);not null"`
	CountryCode string    `gorm:"column:country_code;type:varchar(2);index;not null"`
	RegionCode  string    `gorm:"column:region_code;type:varchar(10);index"`
	TaxType     int       `gorm:"column:tax_type;type:tinyint;not null"`
	Rate        float64   `gorm:"column:rate;type:decimal(5,4);not null"`
	FixedAmount int64     `gorm:"column:fixed_amount;type:bigint"`
	Category    string    `gorm:"column:category;type:varchar(50);index"`
	Priority    int       `gorm:"column:priority;type:int;default:0"`
	IsCompound  bool      `gorm:"column:is_compound;not null;default:false"`
	IsActive    bool      `gorm:"column:is_active;not null;default:true"`
	StartTime   time.Time `gorm:"column:start_time"`
	EndTime     time.Time `gorm:"column:end_time"`
}

func (TaxRuleModel) TableName() string {
	return "tax_rules"
}

// TaxInvoiceModel 税务发票 GORM 模型
type TaxInvoiceModel struct {
	gorm.Model
	OrderID      uint64    `gorm:"column:order_id;uniqueIndex;not null"`
	InvoiceNo    string    `gorm:"column:invoice_no;type:varchar(64);uniqueIndex;not null"`
	TaxID        string    `gorm:"column:tax_id;type:varchar(64)"`
	TotalNet     int64     `gorm:"column:total_net;not null"`
	TotalTax     int64     `gorm:"column:total_tax;not null"`
	TotalGross   int64     `gorm:"column:total_gross;not null"`
	CalculatedAt time.Time `gorm:"column:calculated_at;not null"`
	TaxDetails   string    `gorm:"column:tax_details;type:text"` // JSON string
}

func (TaxInvoiceModel) TableName() string {
	return "tax_invoices"
}

// TaxExemptionModel 税务减免 GORM 模型
type TaxExemptionModel struct {
	gorm.Model
	UserID         uint64    `gorm:"column:user_id;index;not null"`
	Reason         string    `gorm:"column:reason;type:varchar(255)"`
	CertificateID  string    `gorm:"column:certificate_id;type:varchar(100)"`
	CertificateImg string    `gorm:"column:certificate_img;type:varchar(512)"`
	ValidFrom      time.Time `gorm:"column:valid_from"`
	ValidTo        time.Time `gorm:"column:valid_to"`
	Status         int       `gorm:"column:status;type:tinyint;not null"`
}

func (TaxExemptionModel) TableName() string {
	return "tax_exemptions"
}

// Converters

func toDomainTaxRule(m *TaxRuleModel) *domain.TaxRule {
	return &domain.TaxRule{
		ID:          uint64(m.ID),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Name:        m.Name,
		CountryCode: m.CountryCode,
		RegionCode:  m.RegionCode,
		TaxType:     domain.TaxType(m.TaxType),
		Rate:        m.Rate,
		FixedAmount: m.FixedAmount,
		Category:    m.Category,
		Priority:    m.Priority,
		IsCompound:  m.IsCompound,
		IsActive:    m.IsActive,
		StartTime:   m.StartTime,
		EndTime:     m.EndTime,
	}
}

func toDomainTaxExemption(m *TaxExemptionModel) *domain.TaxExemption {
	return &domain.TaxExemption{
		ID:             uint64(m.ID),
		UserID:         m.UserID,
		Reason:         m.Reason,
		CertificateID:  m.CertificateID,
		CertificateImg: m.CertificateImg,
		ValidFrom:      m.ValidFrom,
		ValidTo:        m.ValidTo,
		Status:         m.Status,
	}
}

func toDomainTaxInvoice(m *TaxInvoiceModel) *domain.TaxInvoice {
	return &domain.TaxInvoice{
		ID:           uint64(m.ID),
		OrderID:      m.OrderID,
		InvoiceNo:    m.InvoiceNo,
		TaxID:        m.TaxID,
		TotalNet:     m.TotalNet,
		TotalTax:     m.TotalTax,
		TotalGross:   m.TotalGross,
		CalculatedAt: m.CalculatedAt,
		TaxDetails:   m.TaxDetails,
	}
}
