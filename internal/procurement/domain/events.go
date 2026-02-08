package domain

import "time"

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type PurchaseOrderConfirmedEvent struct {
	OrderID   string
	Timestamp time.Time
}

func (e *PurchaseOrderConfirmedEvent) EventName() string     { return "procurement.po.confirmed" }
func (e *PurchaseOrderConfirmedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseOrderReceivedEvent struct {
	OrderID   string
	Timestamp time.Time
}

func (e *PurchaseOrderReceivedEvent) EventName() string     { return "procurement.po.received" }
func (e *PurchaseOrderReceivedEvent) OccurredAt() time.Time { return e.Timestamp }
