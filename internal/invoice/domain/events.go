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
