package service

import (
	"context"

	v1 "ecommerce/ecommerce/api/order/v1"
	"ecommerce/ecommerce/app/order/internal/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderService struct {
	v1.UnimplementedOrderServer
	uc *biz.OrderUsecase
}

func NewOrderService(uc *biz.OrderUsecase) *OrderService {
	return &OrderService{uc: uc}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *v1.CreateOrderRequest) (*v1.CreateOrderResponse, error) {
	// 1. 参数校验
	if req.UserId == 0 || len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and items are required")
	}

	// 2. 将 gRPC 请求模型转换为 biz 层的领域模型
	bizItems := make([]*biz.CreateOrderRequestItem, 0, len(req.Items))
	for _, item := range req.Items {
		bizItems = append(bizItems, &biz.CreateOrderRequestItem{
			SkuID:    item.SkuId,
			Quantity: item.Quantity,
		})
	}

	// 3. 调用业务逻辑层
	createdOrder, err := s.uc.CreateOrder(ctx, req.UserId, bizItems)
	if err != nil {
		// 这里可以根据 biz 层返回的错误类型，转换为不同的 gRPC 状态码
		// 例如，库存不足可以返回 codes.FailedPrecondition
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 4. 返回响应
	return &v1.CreateOrderResponse{
		OrderId: createdOrder.OrderID,
	}, nil
}

func (s *OrderService) GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrderDetail not implemented")
}
