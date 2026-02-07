package domain

import (
	"errors"
	"time"
)

// 定义Groupbuy模块的业务错误。
var (
	ErrGroupbuyNotFound   = errors.New("拼团活动不存在")
	ErrGroupbuyNotStarted = errors.New("拼团活动未开始")
	ErrGroupbuyEnded      = errors.New("拼团活动已结束")
	ErrGroupFull          = errors.New("拼团已满")
	ErrGroupNotFull       = errors.New("拼团人数未满")
)

// GroupbuyStatus 定义了拼团活动的生命周期状态。
type GroupbuyStatus int8

const (
	GroupbuyStatusPending  GroupbuyStatus = 0
	GroupbuyStatusOngoing  GroupbuyStatus = 1
	GroupbuyStatusEnded    GroupbuyStatus = 2
	GroupbuyStatusCanceled GroupbuyStatus = 3
)

// Groupbuy 实体代表一个拼团活动。
type Groupbuy struct {
	ID            uint           `json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Name          string         `json:"name"`
	ProductID     uint64         `json:"product_id"`
	SkuID         uint64         `json:"sku_id"`
	OriginalPrice uint64         `json:"original_price"`
	GroupPrice    uint64         `json:"group_price"`
	MinPeople     int32          `json:"min_people"`
	MaxPeople     int32          `json:"max_people"`
	TotalStock    int32          `json:"total_stock"`
	SoldCount     int32          `json:"sold_count"`
	StartTime     time.Time      `json:"start_time"`
	EndTime       time.Time      `json:"end_time"`
	Status        GroupbuyStatus `json:"status"`
	Description   string         `json:"description"`
}

// NewGroupbuy 创建并返回一个新的 Groupbuy 实体实例。
func NewGroupbuy(name string, productID, skuID, originalPrice, groupPrice uint64,
	minPeople, maxPeople, totalStock int32, startTime, endTime time.Time,
) *Groupbuy {
	return &Groupbuy{
		Name:          name,
		ProductID:     productID,
		SkuID:         skuID,
		OriginalPrice: originalPrice,
		GroupPrice:    groupPrice,
		MinPeople:     minPeople,
		MaxPeople:     maxPeople,
		TotalStock:    totalStock,
		SoldCount:     0,
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        GroupbuyStatusPending,
	}
}

// RemainingStock 计算拼团活动的剩余库存。
func (g *Groupbuy) RemainingStock() int32 {
	return g.TotalStock - g.SoldCount
}

// IsAvailable 检查拼团活动当前是否可用。
func (g *Groupbuy) IsAvailable() bool {
	now := time.Now()
	return g.Status == GroupbuyStatusOngoing &&
		now.After(g.StartTime) &&
		now.Before(g.EndTime) &&
		g.SoldCount < g.TotalStock
}

// Start 启动拼团活动。
func (g *Groupbuy) Start() {
	g.Status = GroupbuyStatusOngoing
}

// End 结束拼团活动。
func (g *Groupbuy) End() {
	g.Status = GroupbuyStatusEnded
}

// Cancel 取消拼团活动。
func (g *Groupbuy) Cancel() {
	g.Status = GroupbuyStatusCanceled
}

// GroupbuyTeam 实体代表一个拼团团队。
type GroupbuyTeam struct {
	ID            uint               `json:"id"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	GroupbuyID    uint64             `json:"groupbuy_id"`
	TeamNo        string             `json:"team_no"`
	LeaderID      uint64             `json:"leader_id"`
	CurrentPeople int32              `json:"current_people"`
	MaxPeople     int32              `json:"max_people"`
	Status        GroupbuyTeamStatus `json:"status"`
	ExpireAt      time.Time          `json:"expire_at"`
	SuccessAt     *time.Time         `json:"success_at"`
}

// GroupbuyTeamStatus 定义了拼团团队的生命周期状态。
type GroupbuyTeamStatus int8

const (
	GroupbuyTeamStatusOngoing   GroupbuyTeamStatus = 0
	GroupbuyTeamStatusSuccess   GroupbuyTeamStatus = 1
	GroupbuyTeamStatusFailed    GroupbuyTeamStatus = 2
	GroupbuyTeamStatusCancelled GroupbuyTeamStatus = 3
)

// NewGroupbuyTeam 创建并返回一个新的 GroupbuyTeam 实体实例。
func NewGroupbuyTeam(groupbuyID uint64, teamNo string, leaderID uint64, maxPeople int32, expireAt time.Time) *GroupbuyTeam {
	return &GroupbuyTeam{
		GroupbuyID:    groupbuyID,
		TeamNo:        teamNo,
		LeaderID:      leaderID,
		CurrentPeople: 1,
		MaxPeople:     maxPeople,
		Status:        GroupbuyTeamStatusOngoing,
		ExpireAt:      expireAt,
	}
}

// IsFull 检查拼团团队是否已满员。
func (t *GroupbuyTeam) IsFull() bool {
	return t.CurrentPeople >= t.MaxPeople
}

// IsExpired 检查拼团团队是否已过期。
func (t *GroupbuyTeam) IsExpired() bool {
	return time.Now().After(t.ExpireAt)
}

// CanJoin 检查拼团团队是否可以加入。
func (t *GroupbuyTeam) CanJoin() bool {
	return t.Status == GroupbuyTeamStatusOngoing &&
		!t.IsFull() &&
		!t.IsExpired()
}

// Join 成员加入拼团团队。
func (t *GroupbuyTeam) Join() error {
	if !t.CanJoin() {
		return ErrGroupFull
	}

	t.CurrentPeople++

	if t.CurrentPeople >= t.MaxPeople {
		t.Success()
	}

	return nil
}

// Success 标记拼团团队为成功组建。
func (t *GroupbuyTeam) Success() {
	t.Status = GroupbuyTeamStatusSuccess
	now := time.Now()
	t.SuccessAt = &now
}

// Fail 标记拼团团队为失败。
func (t *GroupbuyTeam) Fail() {
	t.Status = GroupbuyTeamStatusFailed
}

// Cancel 取消拼团团队。
func (t *GroupbuyTeam) Cancel() {
	t.Status = GroupbuyTeamStatusCancelled
}

// GroupbuyOrder 实体代表一个拼团订单。
type GroupbuyOrder struct {
	ID          uint                `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	GroupbuyID  uint64              `json:"groupbuy_id"`
	TeamID      uint64              `json:"team_id"`
	TeamNo      string              `json:"team_no"`
	UserID      uint64              `json:"user_id"`
	ProductID   uint64              `json:"product_id"`
	SkuID       uint64              `json:"sku_id"`
	Price       uint64              `json:"price"`
	Quantity    int32               `json:"quantity"`
	TotalAmount uint64              `json:"total_amount"`
	IsLeader    bool                `json:"is_leader"`
	Status      GroupbuyOrderStatus `json:"status"`
	PaidAt      *time.Time          `json:"paid_at"`
	RefundedAt  *time.Time          `json:"refunded_at"`
}

// GroupbuyOrderStatus 定义了拼团订单的生命周期状态。
type GroupbuyOrderStatus int8

const (
	GroupbuyOrderStatusPending   GroupbuyOrderStatus = 0
	GroupbuyOrderStatusPaid      GroupbuyOrderStatus = 1
	GroupbuyOrderStatusSuccess   GroupbuyOrderStatus = 2
	GroupbuyOrderStatusFailed    GroupbuyOrderStatus = 3
	GroupbuyOrderStatusRefunded  GroupbuyOrderStatus = 4
	GroupbuyOrderStatusCancelled GroupbuyOrderStatus = 5
)

// NewGroupbuyOrder 创建并返回一个新的 GroupbuyOrder 实体实例。
func NewGroupbuyOrder(groupbuyID, teamID uint64, teamNo string, userID, productID, skuID, price uint64, quantity int32, isLeader bool) *GroupbuyOrder {
	return &GroupbuyOrder{
		GroupbuyID:  groupbuyID,
		TeamID:      teamID,
		TeamNo:      teamNo,
		UserID:      userID,
		ProductID:   productID,
		SkuID:       skuID,
		Price:       price,
		Quantity:    quantity,
		TotalAmount: price * uint64(quantity),
		IsLeader:    isLeader,
		Status:      GroupbuyOrderStatusPending,
	}
}

// Pay 支付订单。
func (o *GroupbuyOrder) Pay() {
	o.Status = GroupbuyOrderStatusPaid
	now := time.Now()
	o.PaidAt = &now
}

// Success 标记订单为拼团成功。
func (o *GroupbuyOrder) Success() {
	o.Status = GroupbuyOrderStatusSuccess
}

// Fail 标记订单为拼团失败。
func (o *GroupbuyOrder) Fail() {
	o.Status = GroupbuyOrderStatusFailed
}

// Refund 退款订单。
func (o *GroupbuyOrder) Refund() {
	o.Status = GroupbuyOrderStatusRefunded
	now := time.Now()
	o.RefundedAt = &now
}

// Cancel 取消订单。
func (o *GroupbuyOrder) Cancel() {
	o.Status = GroupbuyOrderStatusCancelled
}
