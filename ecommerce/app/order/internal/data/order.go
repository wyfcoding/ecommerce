package data

import (
	"context"
	"fmt"

	"ecommerce/ecommerce/app/order/internal/biz"

	// "github.com/smart-go/alipay-sdk"
	"gorm.io/gorm"
)

// GORM Models for order_basic and order_item would be defined here...

type orderRepo struct {
	db           *gorm.DB
	alipayClient *alipay.Client
}

func NewOrderRepo(db *gorm.DB) biz.OrderRepo {
	return &orderRepo{db: db}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.Order, items []*biz.OrderItem) (*biz.Order, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 创建订单主表记录 (biz.Order -> data.OrderBasic)
		// ... 转换逻辑 ...
		if err := tx.Create(&dbOrder).Error; err != nil {
			return err
		}

		// 2. 批量创建订单商品表记录 (biz.OrderItem -> data.OrderItem)
		// ... 转换和批量准备逻辑 ...
		if err := tx.Create(&dbOrderItems).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

func (r *orderRepo) GeneratePaymentURL(ctx context.Context, order *biz.Order) (string, error) {
	p := alipay.TradePagePay{}
	p.NotifyURL = "http://your-public-domain/v1/payment/alipay/notify" // 你的公网回调地址
	p.ReturnURL = "http://localhost:5173/payment/success"              // 前端支付成功跳转地址
	p.Subject = fmt.Sprintf("Genesis平台订单 - %d", order.OrderID)
	p.OutTradeNo = fmt.Sprintf("%d", order.OrderID)
	p.TotalAmount = fmt.Sprintf("%.2f", float64(order.PaymentAmount)/100.0)

	url, err := r.alipayClient.TradePagePay(p)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (r *orderRepo) VerifyPaymentNotification(ctx context.Context, data map[string]string) error {
	return r.alipayClient.VerifySign(data)
}

func (r *orderRepo) UpdateOrderStatus(ctx context.Context, orderID uint64, status int8) error {
	// 使用 GORM 更新订单状态，注意幂等性
	// tx.Model(&OrderBasic{}).Where("order_id = ? AND status = 10", orderID).Update("status", status)
	// ...
}
