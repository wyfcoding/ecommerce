package biz

import (
	"context"
	"encoding/json"
	"fmt" // Added fmt import
	"strconv" // Added strconv import
	"time"

	"github.com/segmentio/kafka-go" // Added kafka import
)

// OrderStatus 定义订单状态常量。
const (
	OrderStatusPendingPayment = 1  // 待支付
	OrderStatusPaid           = 2  // 已支付
	OrderStatusShipped        = 3  // 已发货
	OrderStatusCompleted      = 4  // 已完成
	OrderStatusCancelled      = 5  // 已取消
	OrderStatusRefunded       = 6  // 已退款
)

// Order 是订单的业务领域模型。
type Order struct {
	ID              uint64
	UserID          uint64
	TotalAmount     uint64
	PaymentAmount   uint64
	ShippingFee     uint64
	Status          int8            // 订单状态，建议使用常量或枚举定义
	ShippingAddress json.RawMessage // 存储序列化后的地址信息
	CreatedAt       time.Time
}

// OrderItem 是订单商品的业务领域模型。
type OrderItem struct {
	ID           uint64
	OrderID      uint64
	SkuID        uint64
	SpuID        uint64
	ProductTitle string
	ProductImage string
	Price        uint64
	Quantity     uint32
	SubTotal     uint64
}

// OrderCreatedEvent 定义了订单创建事件的结构。
type OrderCreatedEvent struct {
	OrderID   uint64    `json:"order_id"`
	UserID    uint64    `json:"user_id"`
	TotalAmount uint64    `json:"total_amount"`
	CreatedAt time.Time `json:"created_at"`
	// 可以根据需要添加更多订单详情
}

// SkuInfo 是订单服务内部使用的商品SKU信息DTO。
type SkuInfo struct {
	SkuID uint64
	SpuID uint64
	Price uint64
	Stock uint32
	Title string
	Image string
}

// Transaction 定义了事务管理器的接口。
type Transaction interface {
	// ExecTx 在一个事务中执行传入的函数。
	ExecTx(context.Context, func(ctx context.Context) error) error
}

// OrderRepo 定义了订单数据仓库的接口。
type OrderRepo interface {
	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	CreateOrderItems(ctx context.Context, items []*OrderItem) error
	GetOrder(ctx context.Context, id uint64) (*Order, error)
	// ... 其他订单相关的数据库操作
}

// ProductClient 定义了订单服务依赖的商品服务客户端接口。
type ProductClient interface {
	GetSkuInfos(ctx context.Context, skuIDs []uint64) (map[uint64]*SkuInfo, error)
	LockStock(ctx context.Context, items map[uint64]uint32) error
	UnlockStock(ctx context.Context, items map[uint64]uint32) error
}

// CartClient 定义了订单服务依赖的购物车服务客户端接口。
	type CartClient interface {
	ClearCheckedItems(ctx context.Context, userID uint64) error
}

// CreateOrderRequestItem 是创建订单请求中的商品项。
type CreateOrderRequestItem struct {
	SkuID    uint64
	Quantity uint32
}

// OrderUsecase 是订单的业务用例。
type OrderUsecase struct {
	repo        OrderRepo
	productClient ProductClient
	cartClient    CartClient
	tx          Transaction
	kafkaProducer *kafka.Writer // Added Kafka producer
}

// NewOrderUsecase 创建一个新的 OrderUsecase。
func NewOrderUsecase(repo OrderRepo, productClient ProductClient, cartClient CartClient, tx Transaction, kafkaProducer *kafka.Writer) *OrderUsecase {
	return &OrderUsecase{
		repo:        repo,
		productClient: productClient,
		cartClient:    cartClient,
		tx:          tx,
		kafkaProducer: kafkaProducer,
	}
}

// CreateOrder 创建一个新订单。
func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID uint64, items []*CreateOrderRequestItem, addressID uint64, remark string, shippingAddress json.RawMessage, paymentAmount uint64) (*Order, error) {
	var createdOrder *Order
	err := uc.tx.ExecTx(ctx, func(txCtx context.Context) error {
		// 1. 获取 SKU 详情并验证库存
		skuIDs := make([]uint64, 0, len(items))
		itemMap := make(map[uint64]*CreateOrderRequestItem)
		for _, item := range items {
			skuIDs = append(skuIDs, item.SkuID)
			itemMap[item.SkuID] = item
		}

		skuInfos, err := uc.productClient.GetSkuInfos(txCtx, skuIDs)
		if err != nil {
			return err
		}

		var totalAmount uint64
		orderItems := make([]*OrderItem, 0, len(items))
		stockToLock := make(map[uint64]uint32)

		for _, skuID := range skuIDs {
			skuInfo, ok := skuInfos[skuID]
			if !ok {
				return fmt.Errorf("SKU %d not found", skuID)
			}
			reqItem := itemMap[skuID]

			if skuInfo.Stock < reqItem.Quantity {
				return fmt.Errorf("SKU %d stock not enough, available: %d, requested: %d", skuID, skuInfo.Stock, reqItem.Quantity)
			}

			subTotal := skuInfo.Price * uint64(reqItem.Quantity)
			totalAmount += subTotal

			orderItems = append(orderItems, &OrderItem{
				SkuID:        skuInfo.SkuID,
				SpuID:        skuInfo.SpuID,
				ProductTitle: skuInfo.Title,
				ProductImage: skuInfo.Image,
				Price:        skuInfo.Price,
				Quantity:     reqItem.Quantity,
				SubTotal:     subTotal,
			})
			stockToLock[skuID] = reqItem.Quantity
		}

		// TODO: 计算运费 shippingFee
		shippingFee := uint64(0) // 暂时设为0

		// 2. 创建订单主记录
		order := &Order{
			UserID:          userID,
			TotalAmount:     totalAmount,
			PaymentAmount:   paymentAmount, // 实际支付金额，可能包含优惠
			ShippingFee:     shippingFee,
			Status:          OrderStatusPendingPayment,
			ShippingAddress: shippingAddress,
		}
		createdOrder, err = uc.repo.CreateOrder(txCtx, order)
		if err != nil {
			return err
		}

		// 3. 创建订单商品记录
		for _, item := range orderItems {
			item.OrderID = createdOrder.ID
		}
		err = uc.repo.CreateOrderItems(txCtx, orderItems)
		if err != nil {
			return err
		}

		// 4. 锁定库存
		err = uc.productClient.LockStock(txCtx, stockToLock)
		if err != nil {
			return err
		}

		// 5. 清空购物车中已下单的商品
		err = uc.cartClient.ClearCheckedItems(txCtx, userID)
		if err != nil {
			// 购物车清理失败不影响订单创建，但需要记录日志或异步处理
			// 这里暂时返回错误，实际情况可能需要更复杂的处理
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

		// 6. 发布订单创建事件到 Kafka
		event := OrderCreatedEvent{
			OrderID:   createdOrder.ID,
			UserID:    createdOrder.UserID,
			TotalAmount: createdOrder.TotalAmount,
			CreatedAt: createdOrder.CreatedAt,
		}
		eventBytes, err := json.Marshal(event)
		if err != nil {
			// 记录错误，但不阻塞事务提交
			fmt.Printf("failed to marshal order created event: %v\n", err)
		}
	
		err = uc.kafkaProducer.WriteMessages(txCtx, kafka.Message{
			Key:   []byte(strconv.FormatUint(createdOrder.ID, 10)),
			Value: eventBytes,
		})
		if err != nil {
			// 记录错误，但不阻塞事务提交
			fmt.Printf("failed to publish order created event to Kafka: %v\n", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}
