// Package domain 发票服务领域事件
package domain

import "time"

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// InvoiceAppliedEvent 发票申请事件
type InvoiceAppliedEvent struct {
	InvoiceID     uint64    `json:"invoice_id"`
	ApplicationNo string    `json:"application_no"`
	OrderNo       string    `json:"order_no"`
	UserID        uint64    `json:"user_id"`
	Amount        int64     `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *InvoiceAppliedEvent) EventName() string     { return "invoice.applied" }
func (e *InvoiceAppliedEvent) OccurredAt() time.Time { return e.Timestamp }

// InvoiceIssuedEvent 发票开具事件
type InvoiceIssuedEvent struct {
	InvoiceID   uint64    `json:"invoice_id"`
	InvoiceCode string    `json:"invoice_code"`
	InvoiceNo   string    `json:"invoice_no"`
	PDFUrl      string    `json:"pdf_url"`
	IsRed       bool      `json:"is_red"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *InvoiceIssuedEvent) EventName() string     { return "invoice.issued" }
func (e *InvoiceIssuedEvent) OccurredAt() time.Time { return e.Timestamp }

// InvoiceFailedEvent 开票失败事件
type InvoiceFailedEvent struct {
	InvoiceID uint64    `json:"invoice_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *InvoiceFailedEvent) EventName() string     { return "invoice.failed" }
func (e *InvoiceFailedEvent) OccurredAt() time.Time { return e.Timestamp }

// InvoiceBlueAppliedEvent 蓝冲申请事件
type InvoiceBlueAppliedEvent struct {
	InvoiceID     uint64    `json:"invoice_id"`
	ApplicationNo string    `json:"application_no"`
	Reason        string    `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *InvoiceBlueAppliedEvent) EventName() string     { return "invoice.blue_applied" }
func (e *InvoiceBlueAppliedEvent) OccurredAt() time.Time { return e.Timestamp }

// InvoiceCancelledEvent 发票取消事件
type InvoiceCancelledEvent struct {
	InvoiceID     uint64    `json:"invoice_id"`
	ApplicationNo string    `json:"application_no"`
	Reason        string    `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *InvoiceCancelledEvent) EventName() string     { return "invoice.cancelled" }
func (e *InvoiceCancelledEvent) OccurredAt() time.Time { return e.Timestamp }

// InvoiceRedAppliedEvent 红冲申请事件
type InvoiceRedAppliedEvent struct {
	InvoiceID       uint64    `json:"invoice_id"`
	RedInvoiceID    uint64    `json:"red_invoice_id"`
	Reason          string    `json:"reason"`
	Timestamp       time.Time `json:"timestamp"`
}

func (e *InvoiceRedAppliedEvent) EventName() string     { return "invoice.red_applied" }
func (e *InvoiceRedAppliedEvent) OccurredAt() time.Time { return e.Timestamp }

// InvoiceVerifiedEvent 发票验真事件
type InvoiceVerifiedEvent struct {
	InvoiceID   uint64    `json:"invoice_id"`
	InvoiceCode string    `json:"invoice_code"`
	InvoiceNo   string    `json:"invoice_no"`
	Valid       bool      `json:"valid"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *InvoiceVerifiedEvent) EventName() string     { return "invoice.verified" }
func (e *InvoiceVerifiedEvent) OccurredAt() time.Time { return e.Timestamp }
