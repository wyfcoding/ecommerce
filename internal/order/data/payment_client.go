
package data

import (
	"context"

	paymentv1 "ecommerce/api/payment/v1"
	"ecommerce/internal/order/biz"

	"google.golang.org/grpc"
)

// paymentClient 实现了 biz.PaymentClient 接口
type paymentClient struct {
	client paymentv1.PaymentServiceClient
}

// NewPaymentClient 创建一个新的支付服务客户端
func NewPaymentClient(conn *grpc.ClientConn) biz.PaymentClient {
	return &paymentClient{client: paymentv1.NewPaymentServiceClient(conn)}
}

// CreatePayment 调用支付服务创建一个新的支付请求
func (pc *paymentClient) CreatePayment(ctx context.Context, userID uint64, orderID string, amount float64) (*biz.PaymentInfo, error) {
	req := &paymentv1.CreatePaymentRequest{
		OrderId: orderID,
		UserId:  userID,
		Amount:  amount,
		Method:  paymentv1.PaymentMethod_PAYMENT_METHOD_ALIPAY, // 示例：默认为支付宝
	}

	resp, err := pc.client.CreatePayment(ctx, req)
	if err != nil {
		return nil, err
	}

	return &biz.PaymentInfo{
		PaymentID:  resp.Payment.PaymentId,
		PaymentURL: resp.Payment.PaymentUrl,
	}, nil
}
