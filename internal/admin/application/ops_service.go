package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"google.golang.org/grpc"
)

// SystemOpsService 负责执行经过审批的系统操作
// 这里会调用其他微服务的 GRPC 接口
type SystemOpsService struct {
	logger *slog.Logger
	deps   SystemOpsDependencies
}

// OrderClient 结构体定义。
type OrderClient interface {
	CancelOrder(ctx context.Context, orderID uint64, reason string) error
	// ... 其他需要的接口
}

// SystemOpsDependencies 结构体定义。
type SystemOpsDependencies struct {
	OrderClient   *grpc.ClientConn
	PaymentClient *grpc.ClientConn
	UserClient    *grpc.ClientConn
}

// NewSystemOpsService 定义了 NewSystemOps 相关的服务逻辑。
func NewSystemOpsService(deps SystemOpsDependencies, logger *slog.Logger) *SystemOpsService {
	return &SystemOpsService{
		logger: logger,
		deps:   deps,
	}
}

func (s *SystemOpsService) ExecuteOperation(ctx context.Context, req *domain.ApprovalRequest) error {
	s.logger.Info("executing system operation", "type", req.ActionType, "payload", req.Payload)

	switch req.ActionType {
	case "ORDER_FORCE_REFUND":
		return s.handleForceRefund(ctx, req.Payload)
	case "SYSTEM_CONFIG_UPDATE":
		return s.handleConfigUpdate(ctx, req.Payload)
	default:
		return fmt.Errorf("unknown action type: %s", req.ActionType)
	}
}

func (s *SystemOpsService) handleForceRefund(ctx context.Context, payload string) error {
	// 实际逻辑：解析 payload -> 调用 Order/Payment Service gRPC
	// payload 应该是 JSON: {"order_id": 123, "reason": "fraud"}
	s.logger.InfoContext(ctx, "Mocking Force Refund (Real gRPC call pending)", "payload", payload)

	// Example:
	// client := orderpb.NewOrderServiceClient(s.deps.OrderClient)
	// _, err := client.CancelOrder(ctx, &orderpb.CancelOrderRequest{Id: ...})
	// return err

	return nil
}

func (s *SystemOpsService) handleConfigUpdate(ctx context.Context, payload string) error {
	// 实际逻辑：更新 Etcd / ConfigMap / DB
	s.logger.InfoContext(ctx, "Mocking Config Update", "payload", payload)
	return nil
}
