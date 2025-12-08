package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// ReconciliationService 对账服务
type ReconciliationService struct {
	paymentRepo domain.PaymentRepository
	gateways    map[domain.GatewayType]domain.PaymentGateway
	logger      *slog.Logger
}

// NewReconciliationService 构造函数
func NewReconciliationService(
	paymentRepo domain.PaymentRepository,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	logger *slog.Logger,
) *ReconciliationService {
	return &ReconciliationService{
		paymentRepo: paymentRepo,
		gateways:    gateways,
		logger:      logger,
	}
}

// RunDailyReconciliation 执行每日对账任务
func (s *ReconciliationService) RunDailyReconciliation(ctx context.Context, date time.Time) error {
	s.logger.InfoContext(ctx, "starting daily reconciliation", "date", date.Format("2006-01-02"))

	// 1. 下载各渠道对账单
	for gatewayType := range s.gateways {
		s.logger.InfoContext(ctx, "downloading bill", "gateway", gatewayType)
		// 模拟下载...
	}

	// 2. 获取本地成功订单
	// payments, _ := s.paymentRepo.FindSuccessPaymentsByDate(date)

	// 3. 逐笔核对
	// for _, p := range payments { ... }

	s.logger.InfoContext(ctx, "reconciliation completed")
	return nil
}
