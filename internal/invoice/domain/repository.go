// Package domain 发票服务仓储接口
package domain

import "context"

// InvoiceRepository 发票仓储接口
type InvoiceRepository interface {
	Save(ctx context.Context, inv *Invoice) error
	Update(ctx context.Context, inv *Invoice) error
	FindByID(ctx context.Context, id uint) (*Invoice, error)
	FindByApplicationNo(ctx context.Context, no string) (*Invoice, error)
	FindByOrderNo(ctx context.Context, orderNo string) ([]*Invoice, error)
	List(ctx context.Context, filter *InvoiceFilter) ([]*Invoice, int64, error)
}

// InvoiceFilter 发票过滤条件
type InvoiceFilter struct {
	UserID     uint64
	MerchantID uint64
	OrderNo    string
	Status     *InvoiceStatus
	Page       int
	PageSize   int
}
