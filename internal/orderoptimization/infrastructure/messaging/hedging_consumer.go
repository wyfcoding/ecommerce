package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	inventorydomain "github.com/wyfcoding/ecommerce/internal/inventory/domain"
	executionv1 "github.com/wyfcoding/financialtrading/go-api/execution/v1"
)

// HedgingConsumer 消费库存对冲事件并触发交易
type HedgingConsumer struct {
	executionCli executionv1.ExecutionServiceClient
	logger       *slog.Logger
}

func NewHedgingConsumer(cli executionv1.ExecutionServiceClient, logger *slog.Logger) *HedgingConsumer {
	return &HedgingConsumer{
		executionCli: cli,
		logger:       logger,
	}
}

// HandleHedgeNeeded 处理对冲需求事件
func (c *HedgingConsumer) HandleHedgeNeeded(ctx context.Context, payload []byte) error {
	var event inventorydomain.HedgeNeededEvent
	// payload 应该是事件的 Data 部分或者整个结构，取决于 Publisher 实现。
	// 假设这里 payload 已经是 Event 结构体 JSON。
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal hedge event: %w", err)
	}

	config := event.HedgeConfig
	if config == nil || !config.AutoHedge {
		return nil
	}

	// 策略逻辑：
	// 库存增加 (ChangeQty > 0) -> 多头增加 -> 需要建立空头 (SELL)
	// 库存减少 (ChangeQty < 0) -> 多头减少 -> 需要平仓空头 (BUY)

	side := "BUY"
	if event.ChangeQty > 0 {
		side = "SELL"
	}

	// 计算对冲数量 = 变动量绝对值 * 对冲比例
	absChange := float64(event.ChangeQty)
	if absChange < 0 {
		absChange = -absChange
	}

	targetQty := absChange * config.HedgeRatio
	if targetQty < float64(config.MinHedgeQty) {
		c.logger.InfoContext(ctx, "change quantity below hedge threshold, skipping", "sku_id", event.SkuID, "change", event.ChangeQty, "target_hedge", targetQty)
		return nil
	}

	// 提交 TWAP 算法单，分散市场冲击
	// 默认 1 小时内完成交易
	duration := time.Hour
	req := &executionv1.SubmitAlgoOrderRequest{
		UserId:            "system-hedger", // 系统对冲账户
		Symbol:            config.InstrumentSymbol,
		Side:              side,
		TotalQuantity:     fmt.Sprintf("%.4f", targetQty),
		AlgoType:          "TWAP",
		StartTime:         time.Now().Unix(),
		EndTime:           time.Now().Add(duration).Unix(),
		ParticipationRate: "0.1", // 假设默认参与率 10%
	}

	c.logger.InfoContext(ctx, "submitting hedging algo order",
		"sku_id", event.SkuID,
		"symbol", config.InstrumentSymbol,
		"side", side,
		"qty", req.TotalQuantity,
	)

	resp, err := c.executionCli.SubmitAlgoOrder(ctx, req)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to submit hedging order", "error", err)
		return err
	}

	c.logger.InfoContext(ctx, "hedging order submitted successfully", "algo_id", resp.AlgoId)
	return nil
}
