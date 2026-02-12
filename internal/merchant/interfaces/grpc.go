package interfaces

import (
	"github.com/wyfcoding/ecommerce/go-api/merchant/v1"
	"github.com/wyfcoding/ecommerce/internal/merchant/application"
)

// GRPCHandler 提供商家服务的 gRPC 接线能力。
// 当前优先完成服务注册，具体方法后续按业务逐步实现。
type GRPCHandler struct {
	merchantv1.UnimplementedMerchantServiceServer
	commandService *application.CommandService
	queryService   *application.QueryService
}

// NewGRPCHandler 创建商家 gRPC 处理器。
func NewGRPCHandler(commandService *application.CommandService, queryService *application.QueryService) *GRPCHandler {
	return &GRPCHandler{
		commandService: commandService,
		queryService:   queryService,
	}
}
