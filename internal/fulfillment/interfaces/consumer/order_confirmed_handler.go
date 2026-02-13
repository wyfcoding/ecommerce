package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	"github.com/segmentio/kafka-go"
	orderv1 "github.com/wyfcoding/ecommerce/go-api/order/v1"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/application"
	"github.com/wyfcoding/ecommerce/internal/fulfillment/domain"
)

const orderConfirmedTopic = "order.confirmed"

// OrderConfirmedHandler 监听订单确认事件并自动创建履约单。
type OrderConfirmedHandler struct {
	commandService *application.CommandService
	queryService   *application.QueryService
	orderClient    orderv1.OrderServiceClient
	logger         *slog.Logger
}

func NewOrderConfirmedHandler(
	commandService *application.CommandService,
	queryService *application.QueryService,
	orderClient orderv1.OrderServiceClient,
	logger *slog.Logger,
) *OrderConfirmedHandler {
	return &OrderConfirmedHandler{
		commandService: commandService,
		queryService:   queryService,
		orderClient:    orderClient,
		logger:         logger,
	}
}

func (h *OrderConfirmedHandler) Handle(ctx context.Context, msg kafka.Message) error {
	if msg.Topic != orderConfirmedTopic {
		return nil
	}

	var payload struct {
		OrderID uint64 `json:"order_id"`
		OrderNo string `json:"order_no"`
		UserID  uint64 `json:"user_id"`
	}
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		h.logger.ErrorContext(ctx, "failed to unmarshal order confirmed event", "error", err)
		return err
	}
	if payload.OrderID == 0 {
		return nil
	}

	order, err := h.orderClient.GetOrderByID(ctx, &orderv1.GetOrderByIDRequest{Id: payload.OrderID})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to query order by id", "order_id", payload.OrderID, "error", err)
		return err
	}
	if order == nil || order.OrderNo == "" {
		return nil
	}

	existing, err := h.queryService.ListByOrderNo(ctx, order.OrderNo)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		// 幂等：该订单已创建履约单。
		return nil
	}

	items := make([]application.FulfillmentItemDTO, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, application.FulfillmentItemDTO{
			SKUID:       strconv.FormatUint(item.SkuId, 10),
			ProductName: item.ProductName,
			SKUName:     item.SkuName,
			ImageURL:    item.ProductImageUrl,
			Quantity:    item.Quantity,
			Location:    "",
			BatchNo:     "",
		})
	}

	_, err = h.commandService.CreateFulfillment(ctx, application.CreateFulfillmentCommand{
		OrderNo:     order.OrderNo,
		MerchantID:  0,
		StoreID:     0,
		WarehouseID: 0,
		Type:        domain.FulfillmentTypeNormal,
		Remark:      order.Remark,
		Address: application.ShippingAddress{
			ReceiverName:  order.ShippingAddress.GetRecipientName(),
			ReceiverPhone: order.ShippingAddress.GetPhoneNumber(),
			Province:      order.ShippingAddress.GetProvince(),
			City:          order.ShippingAddress.GetCity(),
			District:      order.ShippingAddress.GetDistrict(),
			Address:       order.ShippingAddress.GetDetailedAddress(),
			PostalCode:    order.ShippingAddress.GetPostalCode(),
		},
		Items: items,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to auto create fulfillment from order confirmed", "order_no", order.OrderNo, "error", err)
		return err
	}
	return nil
}
