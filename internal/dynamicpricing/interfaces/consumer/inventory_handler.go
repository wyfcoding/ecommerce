package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	kafkago "github.com/segmentio/kafka-go"
	inventoryv1 "github.com/wyfcoding/ecommerce/go-api/inventory/v1"
	productv1 "github.com/wyfcoding/ecommerce/go-api/product/v1"
	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/application"
	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

// InventoryHandler 负责处理来自库存服务的 Kafka 事件。
type InventoryHandler struct {
	cmd          *application.DynamicPricingCommandService
	inventoryCli inventoryv1.InventoryServiceClient
	productCli   productv1.ProductServiceClient
	logger       *slog.Logger
}

// NewInventoryHandler 创建并返回一个新的 InventoryHandler 实例。
func NewInventoryHandler(
	cmd *application.DynamicPricingCommandService,
	inventoryCli inventoryv1.InventoryServiceClient,
	productCli productv1.ProductServiceClient,
	logger *slog.Logger,
) *InventoryHandler {
	return &InventoryHandler{
		cmd:          cmd,
		inventoryCli: inventoryCli,
		productCli:   productCli,
		logger:       logger.With("module", "inventory_consumer"),
	}
}

// Handle 处理库存变动消息。
func (h *InventoryHandler) Handle(ctx context.Context, msg kafkago.Message) error {
	h.logger.Info("received inventory event", "topic", msg.Topic, "key", string(msg.Key))

	var skuID uint64

	// 解析事件 payload 以获取 SKU ID。
	// 这里适配 StockDeductedEvent 或 StockAddedEvent。
	var event struct {
		SkuID uint64 `json:"sku_id"`
	}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error("failed to unmarshal inventory event", "error", err)
		return nil // 不重试解析错误。
	}
	skuID = event.SkuID

	// 1. 获取商品基础价格。
	productResp, err := h.productCli.GetSKUByID(ctx, &productv1.GetSKUByIDRequest{Id: skuID})
	if err != nil {
		h.logger.Error("failed to get SKU info", "sku_id", skuID, "error", err)
		return err
	}

	// 2. 获取实时库存情况。
	inventoryResp, err := h.inventoryCli.GetInventory(ctx, &inventoryv1.GetInventoryRequest{SkuId: skuID})
	if err != nil {
		h.logger.Error("failed to get inventory", "sku_id", skuID, "error", err)
		return err
	}

	// 3. 构造调价请求。
	pricingReq := &domain.PricingRequest{
		SKUID:        skuID,
		BasePrice:    productResp.Price,
		CurrentStock: inventoryResp.Inventory.AvailableStock,
		TotalStock:   inventoryResp.Inventory.TotalStock,
		// Demand 等信息如果事件没给，暂取 0 或从 analytics 获取。
	}

	// 4. 执行动态调价逻辑。
	_, err = h.cmd.CalculatePrice(ctx, pricingReq)
	if err != nil {
		h.logger.Error("failed to calculate dynamic price on inventory change", "sku_id", skuID, "error", err)
		return err
	}

	return nil
}
