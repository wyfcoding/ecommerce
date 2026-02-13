package domain

import (
	"errors"
	"time"
)

var (
	ErrProductNotFound    = errors.New("积分商品不存在")
	ErrProductOffline     = errors.New("积分商品已下架")
	ErrInsufficientStock  = errors.New("库存不足")
	ErrInsufficientPoints = errors.New("积分不足")
	ErrExceedLimit        = errors.New("超过兑换限制")
	ErrOrderNotFound      = errors.New("积分订单不存在")
	ErrInvalidOrderStatus = errors.New("订单状态无效")
)

// PointsProductStatus 积分商品状态
type PointsProductStatus int8

const (
	PointsProductStatusOffline PointsProductStatus = 0 // 下架
	PointsProductStatusOnline  PointsProductStatus = 1 // 上架
)

// PointsProduct 积分商品实体
type PointsProduct struct {
	ID           uint               `json:"id"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	ImageURL     string             `json:"image_url"`
	Points       int64              `json:"points"`
	Stock        int32              `json:"stock"`
	SoldCount    int32              `json:"sold_count"`
	LimitPerUser int32              `json:"limit_per_user"`
	Status       PointsProductStatus `json:"status"`
}

// PointsOrderStatus 积分订单状态
type PointsOrderStatus int8

const (
	PointsOrderStatusPending   PointsOrderStatus = 0 // 待发货
	PointsOrderStatusShipped   PointsOrderStatus = 1 // 已发货
	PointsOrderStatusCompleted PointsOrderStatus = 2 // 已完成
	PointsOrderStatusCanceled  PointsOrderStatus = 3 // 已取消
)

// PointsOrder 积分订单实体
type PointsOrder struct {
	ID          uint              `json:"id"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	OrderNo     string            `json:"order_no"`
	UserID      uint64            `json:"user_id"`
	ProductID   uint64            `json:"product_id"`
	ProductName string            `json:"product_name"`
	Quantity    int32             `json:"quantity"`
	Points      int64             `json:"points"`
	TotalPoints int64             `json:"total_points"`
	Status      PointsOrderStatus `json:"status"`
	Address     string            `json:"address"`
	Phone       string            `json:"phone"`
	Receiver    string            `json:"receiver"`
	ShippedAt   *time.Time        `json:"shipped_at"`
	CompletedAt *time.Time        `json:"completed_at"`
}

// PointsAccount 积分账户实体
type PointsAccount struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      uint64    `json:"user_id"`
	TotalPoints int64     `json:"total_points"`
	UsedPoints  int64     `json:"used_points"`
}

// PointsTransaction 积分流水实体
type PointsTransaction struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      uint64    `json:"user_id"`
	Type        string    `json:"type"` // "earn" or "spend"
	Points      int64     `json:"points"`
	Description string    `json:"description"`
	RefID       string    `json:"ref_id"`
}

// NewPointsProduct 创建积分商品
func NewPointsProduct(name, description, imageURL string, points int64, stock, limitPerUser int32) *PointsProduct {
	return &PointsProduct{
		Name:         name,
		Description:  description,
		ImageURL:     imageURL,
		Points:       points,
		Stock:        stock,
		SoldCount:    0,
		LimitPerUser: limitPerUser,
		Status:       PointsProductStatusOffline,
	}
}

// IsAvailable 检查商品是否可兑换
func (p *PointsProduct) IsAvailable() bool {
	return p.Status == PointsProductStatusOnline && p.Stock > 0
}

// CanExchange 检查是否可以兑换指定数量
func (p *PointsProduct) CanExchange(quantity int32) error {
	if p.Status != PointsProductStatusOnline {
		return ErrProductOffline
	}
	if p.Stock < quantity {
		return ErrInsufficientStock
	}
	return nil
}

// DeductStock 扣减库存
func (p *PointsProduct) DeductStock(quantity int32) error {
	if err := p.CanExchange(quantity); err != nil {
		return err
	}
	p.Stock -= quantity
	p.SoldCount += quantity
	return nil
}

// RestoreStock 恢复库存
func (p *PointsProduct) RestoreStock(quantity int32) {
	p.Stock += quantity
	p.SoldCount -= quantity
}

// Online 上架
func (p *PointsProduct) Online() {
	p.Status = PointsProductStatusOnline
}

// Offline 下架
func (p *PointsProduct) Offline() {
	p.Status = PointsProductStatusOffline
}

// NewPointsAccount 创建积分账户
func NewPointsAccount(userID uint64) *PointsAccount {
	return &PointsAccount{
		UserID:      userID,
		TotalPoints: 0,
		UsedPoints:  0,
	}
}

// AvailablePoints 获取可用积分
func (a *PointsAccount) AvailablePoints() int64 {
	return a.TotalPoints - a.UsedPoints
}

// CanSpend 检查是否可以消费指定积分
func (a *PointsAccount) CanSpend(points int64) error {
	if a.AvailablePoints() < points {
		return ErrInsufficientPoints
	}
	return nil
}

// Spend 消费积分
func (a *PointsAccount) Spend(points int64) error {
	if err := a.CanSpend(points); err != nil {
		return err
	}
	a.UsedPoints += points
	return nil
}

// Earn 获得积分
func (a *PointsAccount) Earn(points int64) {
	a.TotalPoints += points
}

// Refund 退还积分
func (a *PointsAccount) Refund(points int64) {
	a.UsedPoints -= points
}

// NewPointsOrder 创建积分订单
func NewPointsOrder(orderNo string, userID, productID uint64, productName string, quantity int32, points int64, receiver, phone, address string) *PointsOrder {
	return &PointsOrder{
		OrderNo:     orderNo,
		UserID:      userID,
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		Points:      points,
		TotalPoints: points * int64(quantity),
		Status:      PointsOrderStatusPending,
		Receiver:    receiver,
		Phone:       phone,
		Address:     address,
	}
}

// Ship 发货
func (o *PointsOrder) Ship() error {
	if o.Status != PointsOrderStatusPending {
		return ErrInvalidOrderStatus
	}
	o.Status = PointsOrderStatusShipped
	now := time.Now()
	o.ShippedAt = &now
	return nil
}

// Complete 完成
func (o *PointsOrder) Complete() error {
	if o.Status != PointsOrderStatusShipped {
		return ErrInvalidOrderStatus
	}
	o.Status = PointsOrderStatusCompleted
	now := time.Now()
	o.CompletedAt = &now
	return nil
}

// Cancel 取消
func (o *PointsOrder) Cancel() error {
	if o.Status == PointsOrderStatusCompleted || o.Status == PointsOrderStatusCanceled {
		return ErrInvalidOrderStatus
	}
	o.Status = PointsOrderStatusCanceled
	return nil
}

// NewPointsTransaction 创建积分流水
func NewPointsTransaction(userID uint64, txType string, points int64, description, refID string) *PointsTransaction {
	return &PointsTransaction{
		UserID:      userID,
		Type:        txType,
		Points:      points,
		Description: description,
		RefID:       refID,
	}
}
