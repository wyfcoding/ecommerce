package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type PurchaseRequestCreatedEvent struct {
	RequestID   string `json:"request_id"`
	ApplicantID string `json:"applicant_id"`
	Reason      string `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *PurchaseRequestCreatedEvent) EventName() string     { return "procurement.pr.created" }
func (e *PurchaseRequestCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseRequestApprovedEvent struct {
	RequestID  string    `json:"request_id"`
	ApproverID string    `json:"approver_id"`
	Comment    string    `json:"comment"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *PurchaseRequestApprovedEvent) EventName() string     { return "procurement.pr.approved" }
func (e *PurchaseRequestApprovedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseRequestRejectedEvent struct {
	RequestID  string    `json:"request_id"`
	ApproverID string    `json:"approver_id"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *PurchaseRequestRejectedEvent) EventName() string     { return "procurement.pr.rejected" }
func (e *PurchaseRequestRejectedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseOrderCreatedEvent struct {
	OrderID     string    `json:"order_id"`
	SupplierID  string    `json:"supplier_id"`
	WarehouseID string    `json:"warehouse_id"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *PurchaseOrderCreatedEvent) EventName() string     { return "procurement.po.created" }
func (e *PurchaseOrderCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseOrderConfirmedEvent struct {
	OrderID   string    `json:"order_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *PurchaseOrderConfirmedEvent) EventName() string     { return "procurement.po.confirmed" }
func (e *PurchaseOrderConfirmedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseOrderShippedEvent struct {
	OrderID    string    `json:"order_id"`
	SupplierID string    `json:"supplier_id"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *PurchaseOrderShippedEvent) EventName() string     { return "procurement.po.shipped" }
func (e *PurchaseOrderShippedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseOrderReceivedEvent struct {
	OrderID     string    `json:"order_id"`
	WarehouseID string    `json:"warehouse_id"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *PurchaseOrderReceivedEvent) EventName() string     { return "procurement.po.received" }
func (e *PurchaseOrderReceivedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseOrderCompletedEvent struct {
	OrderID   string    `json:"order_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *PurchaseOrderCompletedEvent) EventName() string     { return "procurement.po.completed" }
func (e *PurchaseOrderCompletedEvent) OccurredAt() time.Time { return e.Timestamp }

type PurchaseOrderCancelledEvent struct {
	OrderID string    `json:"order_id"`
	Reason  string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *PurchaseOrderCancelledEvent) EventName() string     { return "procurement.po.cancelled" }
func (e *PurchaseOrderCancelledEvent) OccurredAt() time.Time { return e.Timestamp }
