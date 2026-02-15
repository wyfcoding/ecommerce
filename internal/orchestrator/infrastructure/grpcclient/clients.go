package grpcclient

import (
	"google.golang.org/grpc"

	inventoryv1 "github.com/wyfcoding/ecommerce/go-api/inventory/v1"
	logisticsv1 "github.com/wyfcoding/ecommerce/go-api/logistics/v1"
	orderv1 "github.com/wyfcoding/ecommerce/go-api/order/v1"
	walletv1 "github.com/wyfcoding/ecommerce/go-api/wallet/v1"
)

// ServiceClients 用于统一管理编排器依赖的外部服务 gRPC 客户端。
type ServiceClients struct {
	OrderConn     *grpc.ClientConn `service:"order"`
	InventoryConn *grpc.ClientConn `service:"inventory"`
	WalletConn    *grpc.ClientConn `service:"wallet"`
	LogisticsConn *grpc.ClientConn `service:"logistics"`

	OrderClient     orderv1.OrderServiceClient
	InventoryClient inventoryv1.InventoryServiceClient
	WalletClient    walletv1.WalletServiceClient
	LogisticsClient logisticsv1.LogisticsServiceClient
}

// Init 完成具体的 Client 初始化（基于已连接的 Conn）。
func (c *ServiceClients) Init() {
	if c.OrderConn != nil {
		c.OrderClient = orderv1.NewOrderServiceClient(c.OrderConn)
	}
	if c.InventoryConn != nil {
		c.InventoryClient = inventoryv1.NewInventoryServiceClient(c.InventoryConn)
	}
	if c.WalletConn != nil {
		c.WalletClient = walletv1.NewWalletServiceClient(c.WalletConn)
	}
	if c.LogisticsConn != nil {
		c.LogisticsClient = logisticsv1.NewLogisticsServiceClient(c.LogisticsConn)
	}
}
