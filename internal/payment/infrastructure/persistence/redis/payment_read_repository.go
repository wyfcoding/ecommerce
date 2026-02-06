// 生成摘要：实现支付读模型 Redis 仓储，提供按支付ID/支付单号/订单ID的快速读取。
// 假设：支付单号与支付ID为全局唯一，缓存过期策略由调用方注入。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"

	"github.com/redis/go-redis/v9"
)

const (
	paymentDetailPrefix = "payment:detail:"
	paymentNoPrefix     = "payment:no:"
	paymentOrderPrefix  = "payment:order:"
)

// paymentReadRepository 基于 Redis 的支付读模型仓储。
type paymentReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewPaymentReadRepository 创建支付读模型仓储。
func NewPaymentReadRepository(client redis.UniversalClient, ttl time.Duration) domain.PaymentReadRepository {
	return &paymentReadRepository{
		client: client,
		ttl:    ttl,
	}
}

// Save 保存或更新支付读模型。
func (r *paymentReadRepository) Save(ctx context.Context, payment *domain.Payment) error {
	if payment == nil || payment.ID == 0 {
		return nil
	}

	data, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	paymentID := uint64(payment.ID)
	paymentIDKey := r.paymentIDKey(paymentID)
	paymentNoKey := r.paymentNoKey(payment.PaymentNo)
	orderIDKey := r.orderIDKey(payment.OrderID)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, paymentIDKey, data, r.ttl)
	if payment.PaymentNo != "" {
		pipe.Set(ctx, paymentNoKey, fmt.Sprintf("%d", paymentID), r.ttl)
	}
	if payment.OrderID > 0 {
		pipe.Set(ctx, orderIDKey, fmt.Sprintf("%d", paymentID), r.ttl)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// GetByID 根据支付ID获取读模型。
func (r *paymentReadRepository) GetByID(ctx context.Context, _ uint64, paymentID uint64) (*domain.Payment, error) {
	data, err := r.client.Get(ctx, r.paymentIDKey(paymentID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var payment domain.Payment
	if err := json.Unmarshal(data, &payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetByPaymentNo 根据支付单号获取读模型。
func (r *paymentReadRepository) GetByPaymentNo(ctx context.Context, userID uint64, paymentNo string) (*domain.Payment, error) {
	idStr, err := r.client.Get(ctx, r.paymentNoKey(paymentNo)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var paymentID uint64
	if _, err := fmt.Sscanf(idStr, "%d", &paymentID); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, userID, paymentID)
}

// GetByOrderID 根据订单ID获取读模型。
func (r *paymentReadRepository) GetByOrderID(ctx context.Context, userID uint64, orderID uint64) (*domain.Payment, error) {
	idStr, err := r.client.Get(ctx, r.orderIDKey(orderID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var paymentID uint64
	if _, err := fmt.Sscanf(idStr, "%d", &paymentID); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, userID, paymentID)
}

// Delete 删除读模型数据。
func (r *paymentReadRepository) Delete(ctx context.Context, _ uint64, paymentID uint64, paymentNo string, orderID uint64) error {
	keys := make([]string, 0, 3)
	if paymentID > 0 {
		keys = append(keys, r.paymentIDKey(paymentID))
	}
	if paymentNo != "" {
		keys = append(keys, r.paymentNoKey(paymentNo))
	}
	if orderID > 0 {
		keys = append(keys, r.orderIDKey(orderID))
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

// paymentIDKey 生成支付详情缓存键。
func (r *paymentReadRepository) paymentIDKey(paymentID uint64) string {
	return fmt.Sprintf("%s%d", paymentDetailPrefix, paymentID)
}

// paymentNoKey 生成支付单号映射缓存键。
func (r *paymentReadRepository) paymentNoKey(paymentNo string) string {
	return fmt.Sprintf("%s%s", paymentNoPrefix, paymentNo)
}

// orderIDKey 生成订单ID映射缓存键。
func (r *paymentReadRepository) orderIDKey(orderID uint64) string {
	return fmt.Sprintf("%s%d", paymentOrderPrefix, orderID)
}
