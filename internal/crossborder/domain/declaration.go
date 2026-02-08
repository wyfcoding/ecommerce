package domain

import (
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DeclarationStatus int8

const (
	StatusPending   DeclarationStatus = 1
	StatusSubmitted DeclarationStatus = 2
	StatusCleared   DeclarationStatus = 3
	StatusRejected  DeclarationStatus = 4
)

func (s DeclarationStatus) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusSubmitted:
		return "SUBMITTED"
	case StatusCleared:
		return "CLEARED"
	case StatusRejected:
		return "REJECTED"
	}
	return "UNKNOWN"
}

// CustomsDeclaration 报关单
type CustomsDeclaration struct {
	gorm.Model
	DeclarationID string            `gorm:"column:declaration_id;type:varchar(32);unique_index;not null"`
	OrderID       string            `gorm:"column:order_id;type:varchar(32);index;not null"`
	UserID        string            `gorm:"column:user_id;type:varchar(32);index;not null"`
	LogisticsNo   string            `gorm:"column:logistics_no;type:varchar(64)"`
	DeclaredValue decimal.Decimal   `gorm:"column:declared_value;type:decimal(20,2);not null"`
	Currency      string            `gorm:"column:currency;type:varchar(3);not null"`
	DutyAmount    decimal.Decimal   `gorm:"column:duty_amount;type:decimal(20,2);not null"`
	TaxAmount     decimal.Decimal   `gorm:"column:tax_amount;type:decimal(20,2);not null"`
	Status        DeclarationStatus `gorm:"column:status;type:tinyint;not null;default:1"`
	RejectReason  string            `gorm:"column:reject_reason;type:varchar(255)"`

	Items []DeclarationItem `gorm:"foreignKey:DeclarationID;references:DeclarationID"`
}

type DeclarationItem struct {
	gorm.Model
	DeclarationID string          `gorm:"column:declaration_id;type:varchar(32);index;not null"`
	SKUID         string          `gorm:"column:sku_id;type:varchar(64);not null"`
	HSCode        string          `gorm:"column:hs_code;type:varchar(20)"`
	Price         decimal.Decimal `gorm:"column:price;type:decimal(20,2);not null"`
	Quantity      int32           `gorm:"column:quantity;not null"`
}

// HSCode 海关编码
type HSCode struct {
	gorm.Model
	Code        string          `gorm:"column:code;type:varchar(20);unique_index;not null"`
	Description string          `gorm:"column:description;type:varchar(255)"`
	DutyRate    decimal.Decimal `gorm:"column:duty_rate;type:decimal(5,4);not null"` // e.g. 0.1300 for 13%
	TaxRate     decimal.Decimal `gorm:"column:tax_rate;type:decimal(5,4);not null"`  // VAT rate
}

func (CustomsDeclaration) TableName() string { return "customs_declarations" }
func (DeclarationItem) TableName() string    { return "declaration_items" }
func (HSCode) TableName() string             { return "hs_codes" }

func NewDeclaration(id, orderID, userID, logisticsNo, currency string, declaredVal decimal.Decimal) *CustomsDeclaration {
	return &CustomsDeclaration{
		DeclarationID: id,
		OrderID:       orderID,
		UserID:        userID,
		LogisticsNo:   logisticsNo,
		Currency:      currency,
		DeclaredValue: declaredVal,
		Status:        StatusPending,
		DutyAmount:    decimal.Zero,
		TaxAmount:     decimal.Zero,
	}
}

func (d *CustomsDeclaration) AddItem(skuID, hsCode string, price decimal.Decimal, qty int32) {
	d.Items = append(d.Items, DeclarationItem{
		DeclarationID: d.DeclarationID,
		SKUID:         skuID,
		HSCode:        hsCode,
		Price:         price,
		Quantity:      qty,
	})
}

func (d *CustomsDeclaration) Submit() error {
	if d.Status != StatusPending {
		return errors.New("can only submit pending declaration")
	}
	d.Status = StatusSubmitted
	return nil
}

func (d *CustomsDeclaration) Clear() error {
	if d.Status != StatusSubmitted {
		return errors.New("can only clear submitted declaration")
	}
	d.Status = StatusCleared
	return nil
}

func (d *CustomsDeclaration) Reject(reason string) error {
	d.Status = StatusRejected
	d.RejectReason = reason
	return nil
}
