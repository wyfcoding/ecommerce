package repository

import (
	"context"
	"ecommerce/internal/aftersales/domain/entity"
)

// AfterSalesRepository 售后仓储接口
type AfterSalesRepository interface {
	Create(ctx context.Context, afterSales *entity.AfterSales) error
	GetByID(ctx context.Context, id uint64) (*entity.AfterSales, error)
	GetByNo(ctx context.Context, no string) (*entity.AfterSales, error)
	Update(ctx context.Context, afterSales *entity.AfterSales) error
	List(ctx context.Context, query *AfterSalesQuery) ([]*entity.AfterSales, int64, error)

	// Log methods
	CreateLog(ctx context.Context, log *entity.AfterSalesLog) error
	ListLogs(ctx context.Context, afterSalesID uint64) ([]*entity.AfterSalesLog, error)
}

// AfterSalesQuery 查询条件
type AfterSalesQuery struct {
	UserID       uint64
	OrderID      uint64
	Type         entity.AfterSalesType
	Status       entity.AfterSalesStatus
	AfterSalesNo string
	Page         int
	PageSize     int
}
