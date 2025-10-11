package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "ecommerce/api/payment/v1"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidAmount = errors.New("invalid payment amount")
	ErrPaymentFailed = errors.New("payment failed")
)

// Payment 是支付的业务领域模型。
type Payment struct {
	ID            string
	OrderID       string
	UserID        uint64
	Amount        float64
	Currency      string
	Method        v1.PaymentMethod
	Status        v1.PaymentStatus
	TransactionID string // 来自支付网关的交易ID
	PaymentURL    string // 支付页面的URL
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PaymentRepo 定义了支付数据仓库需要实现的接口。
type PaymentRepo interface {
	CreatePayment(ctx context.Context, p *Payment) error
	GetPaymentByID(ctx context.Context, id string) (*Payment, error)
	UpdatePaymentStatus(ctx context.Context, id, transactionID string, status v1.PaymentStatus) error
}

// OrderClient 定义了与订单服务交互的接口。
type OrderClient interface {
	// 这里可以定义从订单服务获取信息的方法
	// GetOrderInfo(ctx context.Context, orderID string) (*Order, error)
}

// PaymentUsecase 是支付处理的业务逻辑。
type PaymentUsecase struct {
	repo        PaymentRepo
	orderClient OrderClient
	gateway     PaymentGatewayClient // 支付网关客户端
}

// NewPaymentUsecase 创建一个新的 PaymentUsecase。
func NewPaymentUsecase(repo PaymentRepo, orderClient OrderClient, gateway PaymentGatewayClient) *PaymentUsecase {
	return &PaymentUsecase{
		repo:        repo,
		orderClient: orderClient,
		gateway:     gateway,
	}
}

// CreatePayment 发起一个支付流程。
func (uc *PaymentUsecase) CreatePayment(ctx context.Context, orderID string, userID uint64, amount float64, currency string, method v1.PaymentMethod) (*Payment, error) {
	// 在真实业务中，这里可以先通过 orderClient 从订单服务获取订单信息，并进行校验
	// 比如检查订单状态是否为待支付，以及金额是否匹配等

	// 1. 调用外部支付网关获取支付URL和交易ID
	transactionID, paymentURL, err := uc.gateway.CreatePayment(ctx, amount, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment via gateway: %w", err)
	}

	// 2. 创建一个待支付的记录
	payment := &Payment{
		ID:            fmt.Sprintf("pay_%d", time.Now().UnixNano()), // 生成唯一的支付ID
		OrderID:       orderID,
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		Method:        method,
		Status:        v1.PaymentStatus_PAYMENT_STATUS_PENDING,
		TransactionID: transactionID,
		PaymentURL:    paymentURL,
	}

	// 3. 将支付记录保存到数据库
	if err := uc.repo.CreatePayment(ctx, payment); err != nil {
		// 如果数据库创建失败，理论上应该调用支付网关的接口取消这次支付，或者进行其他补偿操作
		return nil, fmt.Errorf("failed to save payment record: %w", err)
	}

	return payment, nil
}

// HandlePaymentCallback 处理来自支付网关的异步回调。
func (uc *PaymentUsecase) HandlePaymentCallback(ctx context.Context, paymentID string, gatewayData map[string]string) error {
	// 在真实业务中，这里需要：
	// 1. 验证回调的合法性（例如，检查签名）
	// 2. 解析 gatewayData，获取支付状态和第三方交易号

	// 模拟从回调数据中获取信息
	statusFromGateway := gatewayData["status"]
	transactionID := gatewayData["transaction_id"]

	// 3. 获取我们自己的支付记录
	payment, err := uc.repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment with id %s not found: %w", paymentID, err)
	}

	// 4. 如果状态已经是最终状态（成功），则直接返回，防止重复处理
	if payment.Status == v1.PaymentStatus_PAYMENT_STATUS_SUCCESS {
		return nil
	}

	// 5. 根据网关返回的状态，更新我们系统内的支付状态
	var newStatus v1.PaymentStatus
	if statusFromGateway == "SUCCESS" {
		newStatus = v1.PaymentStatus_PAYMENT_STATUS_SUCCESS
	} else {
		newStatus = v1.PaymentStatus_PAYMENT_STATUS_FAILED
	}

	if err := uc.repo.UpdatePaymentStatus(ctx, paymentID, transactionID, newStatus); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// 6. 如果支付成功，通过 orderClient 通知订单服务更新订单状态
	if newStatus == v1.PaymentStatus_PAYMENT_STATUS_SUCCESS {
		// err := uc.orderClient.UpdateOrderStatus(ctx, payment.OrderID, "PAID")
		// if err != nil { ... 处理通知失败的情况，例如加入重试队列 ... }
	}

	return nil
}