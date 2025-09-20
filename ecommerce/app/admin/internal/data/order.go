package data

import (
	"context"

	orderV1 "ecommerce/ecommerce/api/order/v1"
	"ecommerce/ecommerce/app/admin/internal/biz"

	"google.golang.org/grpc"
)

type orderGreeter struct {
	client orderV1.OrderClient
}

// NewOrderGreeter 创建一个 order-service 的客户端
func NewOrderGreeter(conn *grpc.ClientConn) biz.OrderGreeter {
	return &orderGreeter{client: orderV1.NewOrderClient(conn)}
}

func (g *orderGreeter) ListOrders(ctx context.Context, req *orderV1.ListOrdersRequest) (*orderV1.ListOrdersResponse, error) {
	// 假设 order-service 已有此接口
	return g.client.ListOrders(ctx, req)
}

func (g *orderGreeter) GetOrderDetail(ctx context.Context, req *orderV1.GetOrderDetailRequest) (*orderV1.GetOrderDetailResponse, error) {
	return g.client.GetOrderDetail(ctx, req)
}

func (g *orderGreeter) ShipOrder(ctx context.Context, req *orderV1.ShipOrderRequest) (*orderV1.ShipOrderResponse, error) {
	// 假设 order-service 已有此接口
	return g.client.ShipOrder(ctx, req)
}
