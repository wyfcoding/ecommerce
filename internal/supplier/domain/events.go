package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type SupplierCreatedEvent struct {
	SupplierID   string    `json:"supplier_id"`
	Name         string    `json:"name"`
	ContactName  string    `json:"contact_name"`
	ContactPhone string    `json:"contact_phone"`
	Email        string    `json:"email"`
	Timestamp    time.Time `json:"timestamp"`
}

func (e *SupplierCreatedEvent) EventName() string     { return "supplier.created" }
func (e *SupplierCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

type SupplierStatusChangedEvent struct {
	SupplierID string         `json:"supplier_id"`
	OldStatus  SupplierStatus `json:"old_status"`
	NewStatus  SupplierStatus `json:"new_status"`
	Reason     string         `json:"reason"`
	Timestamp  time.Time      `json:"timestamp"`
}

func (e *SupplierStatusChangedEvent) EventName() string     { return "supplier.status_changed" }
func (e *SupplierStatusChangedEvent) OccurredAt() time.Time { return e.Timestamp }

type SupplierRatingUpdatedEvent struct {
	SupplierID string          `json:"supplier_id"`
	OldRating  decimal.Decimal `json:"old_rating"`
	NewRating  decimal.Decimal `json:"new_rating"`
	Timestamp  time.Time       `json:"timestamp"`
}

func (e *SupplierRatingUpdatedEvent) EventName() string     { return "supplier.rating_updated" }
func (e *SupplierRatingUpdatedEvent) OccurredAt() time.Time { return e.Timestamp }

type ProductSupplyAddedEvent struct {
	SupplierID   string          `json:"supplier_id"`
	SKUID        string          `json:"sku_id"`
	Price        decimal.Decimal `json:"price"`
	LeadTimeDays int32           `json:"lead_time_days"`
	Timestamp    time.Time       `json:"timestamp"`
}

func (e *ProductSupplyAddedEvent) EventName() string     { return "supplier.product_supply_added" }
func (e *ProductSupplyAddedEvent) OccurredAt() time.Time { return e.Timestamp }

type ProductSupplyUpdatedEvent struct {
	SupplierID string          `json:"supplier_id"`
	SKUID      string          `json:"sku_id"`
	OldPrice   decimal.Decimal `json:"old_price"`
	NewPrice   decimal.Decimal `json:"new_price"`
	Timestamp  time.Time       `json:"timestamp"`
}

func (e *ProductSupplyUpdatedEvent) EventName() string     { return "supplier.product_supply_updated" }
func (e *ProductSupplyUpdatedEvent) OccurredAt() time.Time { return e.Timestamp }
