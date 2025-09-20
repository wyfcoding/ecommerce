package biz

import (
	"context"

	orderV1 "ecommerce/ecommerce/api/order/v1"
)

// OrderGreeter 定义了与 order-service 通信的接口
type OrderGreeter interface {
	ListOrders(ctx context.Context, req *orderV1.ListOrdersRequest) (*orderV1.ListOrdersResponse, error)
	GetOrderDetail(ctx context.Context, req *orderV1.GetOrderDetailRequest) (*orderV1.GetOrderDetailResponse, error)
	ShipOrder(ctx context.Context, req *orderV1.ShipOrderRequest) (*orderV1.ShipOrderResponse, error)
}

// OrderUsecase 封装了订单管理的业务逻辑
type OrderUsecase struct {
	greeter OrderGreeter
}

// NewOrderUsecase 创建一个新的 OrderUsecase
func NewOrderUsecase(greeter OrderGreeter) *OrderUsecase {
	return &OrderUsecase{greeter: greeter}
}

func (uc *OrderUsecase) ListOrders(ctx context.Context, req *orderV1.ListOrdersRequest) (*orderV1.ListOrdersResponse, error) {
	return uc.greeter.ListOrders(ctx, req)
}

func (uc *OrderUsecase) GetOrderDetail(ctx context.Context, req *orderV1.GetOrderDetailRequest) (*orderV1.GetOrderDetailResponse, error) {
	// 在这里可以增加业务逻辑，比如从 context 获取 admin user id 用于操作日志记录
	return uc.greeter.GetOrderDetail(ctx, req)
}

func (uc *OrderUsecase) ShipOrder(ctx context.Context, req *orderV1.ShipOrderRequest) (*orderV1.ShipOrderResponse, error) {
	return uc.greeter.ShipOrder(ctx, req)
}
