package domain

import (
	"errors"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type SupplierStatus int8

const (
	StatusActive    SupplierStatus = 1
	StatusInactive  SupplierStatus = 2
	StatusSuspended SupplierStatus = 3
)

func (s SupplierStatus) String() string {
	switch s {
	case StatusActive:
		return "ACTIVE"
	case StatusInactive:
		return "INACTIVE"
	case StatusSuspended:
		return "SUSPENDED"
	}
	return "UNKNOWN"
}

// Supplier 供应商聚合根
type Supplier struct {
	gorm.Model
	SupplierID   string          `gorm:"column:supplier_id;type:varchar(32);unique_index;not null"`
	Name         string          `gorm:"column:name;type:varchar(255);not null"`
	ContactName  string          `gorm:"column:contact_name;type:varchar(100)"`
	ContactPhone string          `gorm:"column:contact_phone;type:varchar(20)"`
	Email        string          `gorm:"column:email;type:varchar(255)"`
	Address      string          `gorm:"column:address;type:varchar(255)"`
	LicenseNo    string          `gorm:"column:license_no;type:varchar(50)"`
	Status       SupplierStatus  `gorm:"column:status;type:tinyint;not null;default:1"`
	Rating       decimal.Decimal `gorm:"column:rating;type:decimal(3,1);default:5.0"` // 0.0-5.0

	Supplies []ProductSupply `gorm:"foreignKey:SupplierID;references:SupplierID"`
}

// ProductSupply 供应商品
type ProductSupply struct {
	gorm.Model
	SupplierID   string          `gorm:"column:supplier_id;type:varchar(32);index;not null"`
	SKUID        string          `gorm:"column:sku_id;type:varchar(64);index;not null"`
	Price        decimal.Decimal `gorm:"column:price;type:decimal(20,2);not null"`
	LeadTimeDays int32           `gorm:"column:lead_time_days;not null"`
}

func (Supplier) TableName() string      { return "suppliers" }
func (ProductSupply) TableName() string { return "product_supplies" }

func NewSupplier(id, name, contact, phone, email, address, license string) *Supplier {
	return &Supplier{
		SupplierID:   id,
		Name:         name,
		ContactName:  contact,
		ContactPhone: phone,
		Email:        email,
		Address:      address,
		LicenseNo:    license,
		Status:       StatusActive,
		Rating:       decimal.NewFromFloat(5.0),
	}
}

func (s *Supplier) AddProductSupply(skuID string, price decimal.Decimal, leadTime int32) error {
	if s.Status != StatusActive {
		return errors.New("supplier is not active")
	}
	s.Supplies = append(s.Supplies, ProductSupply{
		SupplierID:   s.SupplierID,
		SKUID:        skuID,
		Price:        price,
		LeadTimeDays: leadTime,
	})
	return nil
}
