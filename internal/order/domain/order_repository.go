package domain

import (
	"context"
	"fmt"
	"strconv"
	"time"
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
	FindByUserAndMerchant(ctx context.Context, userID uint64, merchantID uint64) ([]*Order, error)
	GetOrder(ctx context.Context, orderID string) (*Order, error) // 用于工作流
	Update(ctx context.Context, order *Order) error
	UpdateInTx(ctx context.Context, tx any, order *Order) error
	Delete(ctx context.Context, userID uint64, id uint64) error
	List(ctx context.Context, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Order, int64, error)
	ListByUserID(ctx context.Context, userID uint64, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Order, int64, error)

	// 搜索方法
	Search(ctx context.Context, params *ExportQueryParams) ([]*Order, error)
}

// OrderRepositoryAdapter 适配器，为工作流提供 GetOrder 方法
type OrderRepositoryAdapter struct {
	repo OrderRepository
}

// NewOrderRepositoryAdapter 创建适配器
func NewOrderRepositoryAdapter(repo OrderRepository) *OrderRepositoryAdapter {
	return &OrderRepositoryAdapter{repo: repo}
}

// GetOrder 获取订单（用于工作流）
func (a *OrderRepositoryAdapter) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order id is empty")
	}
	if numericID, err := strconv.ParseUint(orderID, 10, 64); err == nil {
		order, findErr := a.repo.FindByID(ctx, 0, numericID)
		if findErr != nil {
			return nil, findErr
		}
		if order != nil {
			return order, nil
		}
	}
	order, err := a.repo.FindByOrderNo(ctx, 0, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}
	return order, nil
}
