package data

import (
	"context"
	"ecommerce/internal/order/biz"
	"encoding/json"
)

type orderRepo struct {
	*Data
}

// NewOrderRepo 是 orderRepo 的构造函数。
func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{Data: data}
}

// toBizOrder 将数据库模型 data.Order 转换为业务领域模型 biz.Order。
func (r *orderRepo) toBizOrder(o *Order) *biz.Order {
	if o == nil {
		return nil
	}
	return &biz.Order{
		ID:              o.ID,
		UserID:          o.UserID,
		TotalAmount:     o.TotalAmount,
		PaymentAmount:   o.PaymentAmount,
		ShippingFee:     o.ShippingFee,
		Status:          o.Status,
		ShippingAddress: json.RawMessage(o.ShippingAddress),
		CreatedAt:       o.CreatedAt,
	}
}

// CreateOrder 在数据库中创建一条订单主记录。
func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.Order) (*biz.Order, error) {
	// 从 context 中获取事务
	tx := GetDBFromContext(ctx) // 使用 GetDBFromContext
	if tx == nil {
		tx = r.db // 如果没有事务，则使用常规的 db 连接
	}

	po := &Order{
		UserID:          order.UserID,
		TotalAmount:     order.TotalAmount,
		PaymentAmount:   order.PaymentAmount,
		ShippingFee:     order.ShippingFee,
		Status:          order.Status,
		ShippingAddress: order.ShippingAddress, // 确保这里是 []byte
	}
	if err := tx.Create(po).Error; err != nil {
		return nil, err
	}
	order.ID = po.ID // Assign the generated ID back to biz.Order
	return r.toBizOrder(po), nil
}

// CreateOrderItems 批量在数据库中创建订单商品记录。
func (r *orderRepo) CreateOrderItems(ctx context.Context, items []*biz.OrderItem) error {
	tx := GetDBFromContext(ctx) // 使用 GetDBFromContext
	if tx == nil {
		tx = r.db
	}

	pos := make([]*OrderItem, 0, len(items))
	for _, item := range items {
		pos = append(pos, &OrderItem{
			OrderID:      item.OrderID,
			SkuID:        item.SkuID,
			SpuID:        item.SpuID,
			ProductTitle: item.ProductTitle,
			ProductImage: item.ProductImage,
			Price:        item.Price,
			Quantity:     item.Quantity,
			SubTotal:     item.SubTotal,
		})
	}
	return tx.Create(&pos).Error
}

// GetOrder 从数据库中获取订单信息。
func (r *orderRepo) GetOrder(ctx context.Context, id uint64) (*biz.Order, error) {
	var o Order // 将 model.Order 改为 Order
	if err := r.db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return r.toBizOrder(&o), nil
}

// CreateOrderForFlashSale creates a new order record in the database for flash sale scenarios.
func (r *orderRepo) CreateOrderForFlashSale(ctx context.Context, order *biz.Order) (*biz.Order, error) {
	tx := GetDBFromContext(ctx)
	if tx == nil {
		tx = r.db
	}

	po := &Order{
		ID:            order.ID,
		UserID:        order.UserID,
		TotalAmount:   order.TotalAmount,
		PaymentAmount: order.PaymentAmount,
		Status:        order.Status,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}
	if err := tx.Create(po).Error; err != nil {
		return nil, err
	}
	return r.toBizOrder(po), nil
}

// CompensateCreateOrder updates the status of a given order to 'Cancelled' for Saga compensation.
func (r *orderRepo) CompensateCreateOrder(ctx context.Context, orderID uint64) error {
	tx := GetDBFromContext(ctx)
	if tx == nil {
		tx = r.db
	}

	result := tx.WithContext(ctx).Model(&Order{}).Where("id = ?", orderID).Update("status", biz.OrderStatusCancelled)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
