// Package domain 履约服务仓储接口
package domain

import "context"

// FulfillmentRepository 履约单仓储接口
type FulfillmentRepository interface {
	Save(ctx context.Context, f *Fulfillment) error
	Update(ctx context.Context, f *Fulfillment) error
	FindByID(ctx context.Context, id uint) (*Fulfillment, error)
	FindByFulfillmentNo(ctx context.Context, no string) (*Fulfillment, error)
	FindByOrderNo(ctx context.Context, orderNo string) ([]*Fulfillment, error)
	List(ctx context.Context, filter *FulfillmentFilter) ([]*Fulfillment, int64, error)
	Delete(ctx context.Context, id uint) error
}

// FulfillmentFilter 履约单过滤条件
type FulfillmentFilter struct {
	MerchantID  uint64
	WarehouseID uint64
	OrderNo     string
	Status      *FulfillmentStatus
	StartTime   *string
	EndTime     *string
	Page        int
	PageSize    int
}

// FulfillmentItemRepository 履约商品项仓储接口
type FulfillmentItemRepository interface {
	SaveBatch(ctx context.Context, items []FulfillmentItem) error
	UpdateBatch(ctx context.Context, items []FulfillmentItem) error
	FindByFulfillmentID(ctx context.Context, fulfillmentID uint) ([]FulfillmentItem, error)
	DeleteByFulfillmentID(ctx context.Context, fulfillmentID uint) error
}

// PackageRepository 包裹仓储接口
type PackageRepository interface {
	SaveBatch(ctx context.Context, packages []Package) error
	FindByFulfillmentID(ctx context.Context, fulfillmentID uint) ([]Package, error)
}

// PickingExceptionRepository 拣货异常仓储接口
type PickingExceptionRepository interface {
	Save(ctx context.Context, exception *PickingException) error
	FindByFulfillmentID(ctx context.Context, fulfillmentID uint) ([]PickingException, error)
}
