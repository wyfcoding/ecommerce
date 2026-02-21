package application

import (
	"context"
	"google.golang.org/protobuf/proto"
)

type OrderServiceClient interface {
	CreateOrder(ctx context.Context, req proto.Message) (proto.Message, error)
	CompensateCreateOrder(ctx context.Context, req proto.Message) error
}

type InventoryServiceClient interface {
	DeductStock(ctx context.Context, req proto.Message) (proto.Message, error)
	ReleaseStock(ctx context.Context, req proto.Message) error
}
