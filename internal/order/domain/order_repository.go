package domain

import (
	"context"
)

// OrderRepository 是订单模块的仓储接口。
type OrderRepository interface {
	// 事务支持
	BeginTx(ctx context.Context, userID uint64) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, userID uint64, fn func(tx any) error) error

	// --- 订单管理 (Order methods) ---

	Save(ctx context.Context, order *Order) error
	SaveInTx(ctx context.Context, tx any, order *Order) error
	FindByID(ctx context.Context, userID uint64, id uint64) (*Order, error)
	FindByOrderNo(ctx context.Context, userID uint64, orderNo string) (*Order, error)
	Update(ctx context.Context, order *Order) error
	UpdateInTx(ctx context.Context, tx any, order *Order) error
	Delete(ctx context.Context, userID uint64, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*Order, int64, error)
	ListByUserID(ctx context.Context, userID uint64, status *int, offset, limit int) ([]*Order, int64, error)
}
