package domain

import "time"

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
