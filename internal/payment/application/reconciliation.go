package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// ReconciliationService 提供支付对账核心功能，确保系统账务与网关账单的一致性。
type ReconciliationService struct {
	paymentRepo domain.PaymentRepository                     // 支付仓储
	gateways    map[domain.GatewayType]domain.PaymentGateway // 各支付渠道驱动映射
	logger      *slog.Logger                                 // 日志记录器
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

// RunDailyReconciliation 执行指定日期的每日自动对账流程。
// 核心逻辑：获取本地成功单 -> 下载网关对账单 -> 双向核对（平账/差错/掉单）-> 持久化对账记录。
func (s *ReconciliationService) RunDailyReconciliation(ctx context.Context, date time.Time) error {
	s.logger.InfoContext(ctx, "starting daily reconciliation", "date", date.Format("2006-01-02"))

	// 1. 获取本地数据库中该日期的所有“已成功”支付记录
	localPayments, err := s.paymentRepo.FindSuccessPaymentsByDate(ctx, date)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to fetch local success payments", "error", err)
		return err
	}

	localMap := make(map[string]*domain.Payment)
	for _, p := range localPayments {
		localMap[p.PaymentNo] = p
	}

	// 2. 遍历所有已启用的支付网关，执行勾对
	for gatewayType, gateway := range s.gateways {
		s.logger.InfoContext(ctx, "processing gateway bill", "gateway", gatewayType)

		billItems, err := gateway.DownloadBill(ctx, date)
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to download bill", "gateway", gatewayType, "error", err)
			continue
		}

		for _, item := range billItems {
			record := &domain.ReconciliationRecord{
				OrderNo:       item.PaymentNo,
				GatewayAmount: item.Amount,
			}

			local, exists := localMap[item.PaymentNo]
			if !exists {
				// 外部有记录，本地系统无记录：判定为“掉单”或“存疑单”
				record.Status = "MISSING_SYSTEM"
				record.Remark = "transaction exists in gateway but not in local success list"
			} else {
				record.PaymentID = uint64(local.Model.ID)
				record.SystemAmount = local.Amount
				record.DiffAmount = local.Amount - item.Amount

				if record.DiffAmount != 0 {
					// 金额不匹配：判定为“差错单”
					record.Status = "MISMATCH"
					record.Remark = fmt.Sprintf("amount mismatch: local %d, gateway %d", local.Amount, item.Amount)
				} else {
					// 金额完全一致
					record.Status = "MATCH"
				}
				// 从待核对 Map 中移除，剩余的即为“单边账”
				delete(localMap, item.PaymentNo)
			}

			if err := s.paymentRepo.SaveReconciliationRecord(ctx, record); err != nil {
				s.logger.ErrorContext(ctx, "failed to save reconciliation record", "order_no", record.OrderNo, "error", err)
			}
		}
	}

	// 3. 处理“单边账”：本地显示成功，但网关账单中未发现
	for _, p := range localMap {
		record := &domain.ReconciliationRecord{
			PaymentID:    uint64(p.Model.ID),
			OrderNo:      p.PaymentNo,
			SystemAmount: p.Amount,
			Status:       "MISSING_GATEWAY",
			Remark:       "transaction exists in local but not found in gateway bill",
		}
		if err := s.paymentRepo.SaveReconciliationRecord(ctx, record); err != nil {
			s.logger.ErrorContext(ctx, "failed to save missing_gateway record", "order_no", record.OrderNo, "error", err)
		}
	}

	s.logger.InfoContext(ctx, "reconciliation completed successfully", "date", date.Format("2006-01-02"))
	return nil
}
