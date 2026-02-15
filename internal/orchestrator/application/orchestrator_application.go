package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/orchestrator/domain"
	"github.com/wyfcoding/ecommerce/internal/orchestrator/infrastructure/grpcclient"
	"github.com/wyfcoding/pkg/saga"
)

type StartSagaCommand struct {
	SagaType    string
	BusinessKey string
	Payload     string
}

type OrchestratorApplicationService struct {
	repo    domain.OrchestratorRepository
	engine  *saga.Engine // 引用 pkg/saga 中的核心逻辑
	logger  *slog.Logger
	clients *grpcclient.ServiceClients
}

func NewOrchestratorApplicationService(
	repo domain.OrchestratorRepository,
	engine *saga.Engine,
	logger *slog.Logger,
	clients *grpcclient.ServiceClients,
) *OrchestratorApplicationService {
	return &OrchestratorApplicationService{
		repo:    repo,
		engine:  engine,
		logger:  logger,
		clients: clients,
	}
}

func (s *OrchestratorApplicationService) StartSaga(ctx context.Context, cmd StartSagaCommand) (string, error) {
	sagaID := fmt.Sprintf("SAGA-%d", time.Now().UnixNano())
	s.logger.Info("starting new saga transaction", "saga_id", sagaID, "type", cmd.SagaType)

	instance := &domain.SagaInstance{
		SagaID:        sagaID,
		SagaType:      cmd.SagaType,
		OriginalRefID: cmd.BusinessKey,
		ContextData:   cmd.Payload,
		Status:        domain.SagaStarted,
		StartTime:     time.Now(),
	}

	if err := s.repo.SaveInstance(ctx, instance); err != nil {
		return "", err
	}

	// 执行 Saga
	go s.executeOrderSaga(context.Background(), instance)

	return sagaID, nil
}

func (s *OrchestratorApplicationService) executeOrderSaga(ctx context.Context, instance *domain.SagaInstance) {
	s.logger.Info("executing order saga", "saga_id", instance.SagaID)

	// 此处仅作为示例演示核心逻辑，实际应从 instance.ContextData 解析业务参数
	// 假设 Payload 包含: OrderID, UserID, TotalAmount, Items
	engine := saga.NewEngine()

	// 步骤 1: 预占库存
	engine.AddStep("LockStock",
		func(ctx context.Context) error {
			s.logger.Info("step: LockStock", "saga_id", instance.SagaID)
			// TODO: 从 ContextData 解析 SKU 和数量
			// _, err := s.clients.InventoryClient.LockStock(ctx, &inventoryv1.LockStockRequest{
			// 	SkuId:    1001,
			// 	Quantity: 1,
			// 	Reason:   "Saga Order Lock: " + instance.OriginalRefID,
			// })
			return nil
		},
		func(ctx context.Context) error {
			s.logger.Warn("compensate: UnlockStock", "saga_id", instance.SagaID)
			// _, _ = s.clients.InventoryClient.UnlockStock(ctx, &inventoryv1.UnlockStockRequest{
			// 	SkuId:    1001,
			// 	Quantity: 1,
			// 	Reason:   "Saga Rollback: " + instance.OriginalRefID,
			// })
			return nil
		},
	)

	// 步骤 2: 冻结钱包余额
	engine.AddStep("FreezeBalance",
		func(ctx context.Context) error {
			s.logger.Info("step: FreezeBalance", "saga_id", instance.SagaID)
			// _, err := s.clients.WalletClient.FreezeBalance(ctx, &walletv1.FreezeBalanceRequest{
			// 	UserId:   instance.UserID,
			// 	Currency: "CNY",
			// 	Amount:   "100.00",
			// 	Reason:   "Order Freeze: " + instance.OriginalRefID,
			// })
			return nil
		},
		func(ctx context.Context) error {
			s.logger.Warn("compensate: UnfreezeBalance", "saga_id", instance.SagaID)
			// _, _ = s.clients.WalletClient.UnfreezeBalance(ctx, &walletv1.UnfreezeBalanceRequest{
			// 	UserId:   instance.UserID,
			// 	Currency: "CNY",
			// 	Amount:   "100.00",
			// 	Reason:   "Order Unfreeze: " + instance.OriginalRefID,
			// })
			return nil
		},
	)

	// 执行事务
	if err := engine.Execute(ctx); err != nil {
		s.logger.Error("saga execution failed", "saga_id", instance.SagaID, "error", err)
		instance.SetStatus(domain.SagaFailed)
		instance.LastError = err.Error()
	} else {
		s.logger.Info("saga execution success", "saga_id", instance.SagaID)
		instance.SetStatus(domain.SagaSucceeded)

		// 步骤 3: 最终确认订单 (单向同步或异步)
		// _, _ = s.clients.OrderClient.ConfirmOrder(ctx, &orderv1.ConfirmOrderRequest{
		// 	OrderId: instance.OriginalRefID,
		// })
	}

	_ = s.repo.SaveInstance(ctx, instance)
}

func (s *OrchestratorApplicationService) GetStatus(ctx context.Context, sagaID string) (*domain.SagaInstance, error) {
	return s.repo.FindInstanceByID(ctx, sagaID)
}
