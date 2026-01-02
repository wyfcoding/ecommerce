package application

import (
	"context"
	"fmt"
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

	// 1. 获取本地成功订单
	localPayments, err := s.paymentRepo.FindSuccessPaymentsByDate(ctx, date)
	if err != nil {
		return err
	}

	localMap := make(map[string]*domain.Payment)
	for _, p := range localPayments {
		localMap[p.PaymentNo] = p
	}

	// 2. 遍历网关下载对账单并核对
	for gatewayType, gateway := range s.gateways {
		s.logger.InfoContext(ctx, "processing gateway bill", "gateway", gatewayType)

		billItems, err := gateway.DownloadBill(ctx, date)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to download bill", "gateway", gatewayType, "error", err)
			continue
		}

		for _, item := range billItems {
			record := &domain.ReconciliationRecord{
				OrderNo:       item.PaymentNo, // 账单中的单号
				GatewayAmount: item.Amount,
			}

			local, exists := localMap[item.PaymentNo]
			if !exists {
				// 外部有，本地无 -> 漏单 (MISSING_SYSTEM)
				record.Status = "MISSING_SYSTEM"
				record.Remark = "Transaction exists in gateway but not in local success list"
			} else {
				record.PaymentID = uint64(local.ID)
				record.SystemAmount = local.Amount
				record.DiffAmount = local.Amount - item.Amount

				if record.DiffAmount != 0 {
					record.Status = "MISMATCH"
					record.Remark = fmt.Sprintf("Amount mismatch: local %d, gateway %d", local.Amount, item.Amount)
				} else {
					record.Status = "MATCH"
				}
				// 标记已核对
				delete(localMap, item.PaymentNo)
			}

			_ = s.paymentRepo.SaveReconciliationRecord(ctx, record)
		}
	}

	// 3. 处理本地有，外部无的情况 -> 单边账 (MISSING_GATEWAY)
	for _, p := range localMap {
		record := &domain.ReconciliationRecord{
			PaymentID:    uint64(p.ID),
			OrderNo:      p.PaymentNo,
			SystemAmount: p.Amount,
			Status:       "MISSING_GATEWAY",
			Remark:       "Transaction exists in local but not found in gateway bill",
		}
		_ = s.paymentRepo.SaveReconciliationRecord(ctx, record)
	}

	s.logger.InfoContext(ctx, "reconciliation completed")
	return nil
}
