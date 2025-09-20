package biz

import (
	"context"
	"ecommerce/ecommerce/pkg/snowflake"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"ecommerce/ecommerce/pkg/logging"
	"ecommerce/ecommerce/pkg/metrics"

	"go.uber.org/zap"
)

// --- Domain Models ---

type Order struct {
	ID              uint64
	OrderID         uint64
	UserID          uint64
	TotalAmount     uint64
	PaymentAmount   uint64
	ShippingFee     uint64
	Status          int8
	ShippingAddress json.RawMessage // JSON anp
}

type OrderItem struct {
	OrderID      uint64
	SkuID        uint64
	SpuID        uint64
	ProductTitle string
	ProductImage string
	Price        uint64
	Quantity     uint32
	SubTotal     uint64
}

// --- DTOs for Inter-Service Communication ---

type SkuInfo struct {
	SkuID uint64
	SpuID uint64
	Price uint64
	Stock uint32
	Title string
	Image string
}

type CreateOrderRequestItem struct {
	SkuID    uint64
	Quantity uint32
}

// --- Repo & Greeter Interfaces ---

type OrderRepo interface {
	CreateOrder(ctx context.Context, order *Order, items []*OrderItem) (*Order, error)
}

type ProductGreeter interface {
	GetSkuInfos(ctx context.Context, skuIDs []uint64) (map[uint64]*SkuInfo, error)
	LockStock(ctx context.Context, items []*OrderItem) error
	// UnlockStock(ctx context.Context, items []*OrderItem) error // 用于事务补偿
}

type CartGreeter interface {
	ClearCartItems(ctx context.Context, userID uint64, skuIDs []uint64) error
}

// --- Usecase ---

type OrderUsecase struct {
	repo    OrderRepo
	product ProductGreeter
	cart    CartGreeter
}

func NewOrderUsecase(repo OrderRepo, p Greeter, c Greeter) *OrderUsecase {
	return &OrderUsecase{repo: repo, product: p, cart: c}
}

// CreateOrder 是创建订单的核心业务逻辑
func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID uint64, reqItems []*CreateOrderRequestItem) (*Order, error) {
	// 使用带有 context 的 logger
	logger := logging.For(ctx)

	logger.Info("Starting to create order",
		zap.Uint64("user_id", userID),
		zap.Int("item_count", len(reqItems)),
	)
	// 1. 准备数据：提取所有 skuID
	skuIDs := make([]uint64, 0, len(reqItems))
	reqItemMap := make(map[uint64]*CreateOrderRequestItem, len(reqItems))
	for _, item := range reqItems {
		skuIDs = append(skuIDs, item.SkuID)
		reqItemMap[item.SkuID] = item
	}

	// 2. [跨服务调用] 调用商品服务，批量获取最新的商品信息
	skuInfos, err := uc.product.GetSkuInfos(ctx, skuIDs)
	if err != nil {
		logger.Error("Failed to get sku info from product-service",
			zap.Error(err),
			zap.Any("requested_sku_ids", skuIDs),
		)
		return nil, fmt.Errorf("failed to get sku info: %w", err)
	}
	if len(skuInfos) != len(reqItems) {
		return nil, errors.New("some products are not available")
	}

	// 3. 业务逻辑：校验库存，计算总价，生成订单商品项
	var totalAmount uint64 = 0
	orderItems := make([]*OrderItem, 0, len(reqItems))
	for _, skuInfo := range skuInfos {
		reqItem := reqItemMap[skuInfo.SkuID]
		// 校验库存
		if skuInfo.Stock < reqItem.Quantity {
			return nil, fmt.Errorf("sku %d has insufficient stock", skuInfo.SkuID)
		}

		subTotal := skuInfo.Price * uint64(reqItem.Quantity)
		totalAmount += subTotal

		orderItems = append(orderItems, &OrderItem{
			SkuID:        skuInfo.SkuID,
			SpuID:        skuInfo.SpuID,
			ProductTitle: skuInfo.Title,
			ProductImage: skuInfo.Image,
			Price:        skuInfo.Price, // 使用从后端获取的实时价格
			Quantity:     reqItem.Quantity,
			SubTotal:     subTotal,
		})
	}

	// 4. [跨服务调用] 锁定商品库存（需要实现分布式事务补偿逻辑）
	if err := uc.product.LockStock(ctx, orderItems); err != nil {
		return nil, fmt.Errorf("failed to lock stock: %w", err)
	}

	// 5. 生成订单
	order := &Order{
		OrderID:       snowflake.GenID(), // 使用我们之前定义的雪花算法ID生成器
		UserID:        userID,
		TotalAmount:   totalAmount,
		PaymentAmount: totalAmount, // 暂不考虑优惠
		ShippingFee:   0,           // 暂不考虑运费
		Status:        10,          // 10-待支付
		// ShippingAddress: ... // 地址信息需从 user-service 或请求中获取
	}

	// 6. [本地事务] 将订单信息和订单商品信息写入数据库
	createdOrder, err := uc.repo.CreateOrder(ctx, order, orderItems)
	if err != nil {
		// 如果数据库写入失败，需要调用 UnlockStock 进行库存补偿
		// uc.product.UnlockStock(ctx, orderItems)

		// 更新自定义指标：创建失败
		metrics.OrdersCreatedCounter.WithLabelValues("failure").Inc()

		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// 7. [跨服务调用] 清理购物车中已下单的商品
	if err := uc.cart.ClearCartItems(ctx, userID, skuIDs); err != nil {
		// 清理购物车失败不是关键错误，只需记录警告
		logger.Warn("Failed to clear cart items, requires manual intervention",
			zap.Uint64("user_id", userID),
			zap.Any("sku_ids_to_clear", skuIDs),
			zap.Error(err),
		)
	}

	// 更新自定义指标：创建成功
	metrics.OrdersCreatedCounter.WithLabelValues("success").Inc()

	logger.Info("Order created successfully",
		zap.Uint64("order_id", createdOrder.OrderID),
		zap.Uint64("user_id", userID),
	)

	return createdOrder, nil
}

func (uc *OrderUsecase) GetPaymentURL(ctx context.Context, userID, orderID uint64) (string, error) {
	// 1. 查找订单
	order, err := uc.repo.GetOrder(ctx, orderID)
	if err != nil {
		return "", err
	}
	// 2. 校验订单状态和归属
	if order.UserID != userID || order.Status != 10 { // 10-待支付
		return "", errors.New("invalid order status or permission denied")
	}

	// 3. 调用 repo 层生成URL
	return uc.repo.GeneratePaymentURL(ctx, order)
}

// ProcessPaymentNotification 处理支付成功的回调
func (uc *OrderUsecase) ProcessPaymentNotification(ctx context.Context, data map[string]string) error {
	// 1. 验签
	if err := uc.repo.VerifyPaymentNotification(ctx, data); err != nil {
		return err
	}

	// 2. 处理业务逻辑
	orderIDStr := data["out_trade_no"]
	tradeStatus := data["trade_status"]

	if tradeStatus == "TRADE_SUCCESS" {
		orderID, _ := strconv.ParseUint(orderIDStr, 10, 64)
		// 将订单状态更新为“待发货”(20)
		return uc.repo.UpdateOrderStatus(ctx, orderID, 20)
	}

	return nil // 其他状态不处理
}
