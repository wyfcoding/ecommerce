package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ecommerce/pkg/snowflake"
	"go.uber.org/zap"
)

// OrderUsecase 封装了订单相关的业务逻辑。
type OrderUsecase struct {
	tx      Transaction
	repo    OrderRepo
	product ProductClient
	cart    CartClient
	log     *zap.SugaredLogger // 添加日志器
}

// NewOrderUsecase 是 OrderUsecase 的构造函数。
func NewOrderUsecase(tx Transaction, repo OrderRepo, product ProductClient, cart CartClient, logger *zap.SugaredLogger) *OrderUsecase {
	return &OrderUsecase{
		tx:      tx,
		repo:    repo,
		product: product,
		cart:    cart,
		log:     logger,
	}
}

// CreateOrderRequestItem 定义了创建订单时每个商品项的输入结构。
type CreateOrderRequestItem struct {
	SkuID    uint64
	Quantity uint32
}

// CreateOrder 是创建订单的核心业务流程。
// 它通过事务确保了订单创建、库存锁定等操作的原子性。
func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID uint64, reqItems []*CreateOrderRequestItem, shippingAddress json.RawMessage, paymentAmount uint64) (*Order, error) {
	var createdOrder *Order

	// 1. 准备数据：提取所有 skuID
	skuIDs := make([]uint64, 0, len(reqItems))
	reqItemMap := make(map[uint64]uint32, len(reqItems))
	for _, item := range reqItems {
		skuIDs = append(skuIDs, item.SkuID)
		reqItemMap[item.SkuID] = item.Quantity
	}

	// 2. 调用商品服务，获取最新的商品信息和价格
	skuInfos, err := uc.product.GetSkuInfos(ctx, skuIDs)
	if err != nil {
		uc.log.Errorf("CreateOrder: failed to get sku info from product-service: %v", err)
		return nil, fmt.Errorf("failed to get product information: %w", err)
	}
	if len(skuInfos) != len(reqItems) {
		return nil, fmt.Errorf("some products are invalid or out of stock")
	}

	// 3. 执行核心事务
	err = uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// 3a. 业务逻辑：校验库存，计算总价，生成订单商品项
		var totalAmount uint64 = 0
		orderItems := make([]*OrderItem, 0, len(reqItems))
		stockToLock := make(map[uint64]uint32)

		for _, skuInfo := range skuInfos {
			quantity := reqItemMap[skuInfo.SkuID]
			if skuInfo.Stock < quantity {
				return fmt.Errorf("product '%s' has insufficient stock", skuInfo.Title)
			}

			subTotal := skuInfo.Price * uint64(quantity)
			totalAmount += subTotal
			stockToLock[skuInfo.SkuID] = quantity

			orderItems = append(orderItems, &OrderItem{
				SkuID:        skuInfo.SkuID,
				SpuID:        skuInfo.SpuID,
				ProductTitle: skuInfo.Title,
				ProductImage: skuInfo.Image,
				Price:        skuInfo.Price,
				Quantity:     quantity,
				SubTotal:     subTotal,
			})
		}

		// 3b. 创建订单主记录
		order := &Order{
			ID:            snowflake.GenID(),
			UserID:        userID,
			TotalAmount:   totalAmount,
			PaymentAmount: paymentAmount, // 使用传入的 paymentAmount
			ShippingAddress: shippingAddress, // 传入的地址
			Status:        OrderStatusPendingPayment, // 使用常量
		}
		var err error
		createdOrder, err = uc.repo.CreateOrder(txCtx, order)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// 3c. 关联 OrderID 并批量创建订单商品记录
		for _, item := range orderItems {
			item.OrderID = createdOrder.ID
		}
		if err := uc.repo.CreateOrderItems(txCtx, orderItems); err != nil {
			return fmt.Errorf("failed to create order items: %w", err)
		}

		// 3d. 调用商品服务锁定库存
		if err := uc.product.LockStock(txCtx, stockToLock); err != nil {
			// 如果这里失败，整个事务会回滚，数据库中不会有订单记录
			return fmt.Errorf("failed to lock stock: %w", err)
		}

		return nil // 事务成功，提交
	})

	if err != nil {
		return nil, err // 事务执行失败
	}

	// 4. 事务成功后，执行非事务性操作：清空购物车
	// 这是一个最终一致性的操作，即使失败也不应影响订单的创建。
	if err := uc.cart.ClearCheckedItems(ctx, userID); err != nil {
		uc.log.Warnf("CreateOrder: failed to clear cart for user %d: %v", userID, err)
	}

	return createdOrder, nil
}

// GetOrder 获取订单详情。
func (uc *OrderUsecase) GetOrder(ctx context.Context, id uint64) (*Order, error) {
	return uc.repo.GetOrder(ctx, id)
}

// CreateOrderForFlashSale 为秒杀活动创建订单。
func (uc *OrderUsecase) CreateOrderForFlashSale(ctx context.Context, userID string, productID string, quantity int32, price float64) (string, error) {
	// 简化处理：直接创建订单，不涉及购物车和复杂的库存锁定
	// 库存已在秒杀服务中处理

	orderID := snowflake.GenID()
	order := &Order{
		ID:            orderID,
		UserID:        uint64(orderID), // Assuming userID can be converted to uint64 for now
		TotalAmount:   uint64(price * float64(quantity)),
		PaymentAmount: uint64(price * float64(quantity)),
		Status:        OrderStatusPendingPayment,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	orderItem := &OrderItem{
		OrderID:      orderID,
		ProductID:    productID,
		Quantity:     quantity,
		Price:        uint64(price),
		SubTotal:     uint64(price * float64(quantity)),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 使用事务确保订单和订单项的原子性
	err := uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		_, err := uc.repo.CreateOrder(txCtx, order)
		if err != nil {
			return fmt.Errorf("failed to create order for flash sale: %w", err)
		}
		if err := uc.repo.CreateOrderItems(txCtx, []*OrderItem{orderItem}); err != nil {
			return fmt.Errorf("failed to create order item for flash sale: %w", err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", orderID), nil
}

// CompensateCreateOrder 补偿创建订单操作，通常是取消订单。
func (uc *OrderUsecase) CompensateCreateOrder(ctx context.Context, orderID string) error {
	// 查找订单并将其状态设置为已取消
	parsedOrderID, err := strconv.ParseUint(orderID, 10, 64)
	if err != nil {
		return fmt.Errorf("无效的订单ID: %w", err)
	}

	order, err := uc.repo.GetOrder(ctx, parsedOrderID)
	if err != nil {
		// 如果订单不存在，可能已经被取消或从未成功创建，直接返回成功
		if errors.Is(err, ErrOrderNotFound) {
			return nil
		}
		return fmt.Errorf("获取订单失败: %w", err)
	}

	// 只有待支付或已创建的订单才能取消
	if order.Status == OrderStatusPendingPayment || order.Status == OrderStatusCreated {
		order.Status = OrderStatusCancelled
		if _, err := uc.repo.UpdateOrder(ctx, order); err != nil {
			return fmt.Errorf("更新订单状态为取消失败: %w", err)
		}
	}

	// TODO: 如果订单已经支付，可能需要触发退款流程

	return nil
}
