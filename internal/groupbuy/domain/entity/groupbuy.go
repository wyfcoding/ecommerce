package entity

import (
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
)

// 定义Groupbuy模块的业务错误。
var (
	ErrGroupbuyNotFound   = errors.New("拼团活动不存在") // 拼团活动记录未找到。
	ErrGroupbuyNotStarted = errors.New("拼团活动未开始") // 拼团活动尚未开始。
	ErrGroupbuyEnded      = errors.New("拼团活动已结束") // 拼团活动已结束。
	ErrGroupFull          = errors.New("拼团已满")    // 拼团团队已满员。
	ErrGroupNotFull       = errors.New("拼团人数未满")  // 拼团团队人数未达到成团要求。
)

// GroupbuyStatus 定义了拼团活动的生命周期状态。
type GroupbuyStatus int8

const (
	GroupbuyStatusPending  GroupbuyStatus = 0 // 未开始：拼团活动已创建但尚未到开始时间。
	GroupbuyStatusOngoing  GroupbuyStatus = 1 // 进行中：拼团活动正在进行。
	GroupbuyStatusEnded    GroupbuyStatus = 2 // 已结束：拼团活动已过结束时间。
	GroupbuyStatusCanceled GroupbuyStatus = 3 // 已取消：拼团活动被取消。
)

// Groupbuy 实体代表一个拼团活动。
// 它包含了拼团商品的详细信息、价格、库存、时间范围和状态。
type Groupbuy struct {
	gorm.Model                   // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name          string         `gorm:"type:varchar(255);not null;comment:活动名称" json:"name"`      // 活动名称。
	ProductID     uint64         `gorm:"not null;comment:商品ID" json:"product_id"`                  // 关联的商品ID。
	SkuID         uint64         `gorm:"not null;comment:SKU ID" json:"sku_id"`                    // 关联的SKU ID。
	OriginalPrice uint64         `gorm:"not null;comment:原价(分)" json:"original_price"`             // 商品原价（单位：分）。
	GroupPrice    uint64         `gorm:"not null;comment:拼团价(分)" json:"group_price"`               // 拼团价格（单位：分）。
	MinPeople     int32          `gorm:"not null;comment:最小成团人数" json:"min_people"`                // 拼团成功所需的最小人数。
	MaxPeople     int32          `gorm:"not null;comment:最大成团人数" json:"max_people"`                // 拼团允许的最大人数。
	TotalStock    int32          `gorm:"not null;comment:总库存" json:"total_stock"`                  // 拼团活动的总库存。
	SoldCount     int32          `gorm:"not null;default:0;comment:已售数量" json:"sold_count"`        // 已售出的数量。
	StartTime     time.Time      `gorm:"not null;comment:开始时间" json:"start_time"`                  // 拼团活动开始时间。
	EndTime       time.Time      `gorm:"not null;comment:结束时间" json:"end_time"`                    // 拼团活动结束时间。
	Status        GroupbuyStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"` // 拼团活动状态，默认为未开始。
	Description   string         `gorm:"type:text;comment:活动描述" json:"description"`                // 活动描述。
}

// NewGroupbuy 创建并返回一个新的 Groupbuy 实体实例。
// name: 活动名称。
// productID, skuID: 商品ID和SKU ID。
// originalPrice, groupPrice: 原价和拼团价。
// minPeople, maxPeople: 最小/最大成团人数。
// totalStock: 总库存。
// startTime, endTime: 活动时间。
func NewGroupbuy(name string, productID, skuID, originalPrice, groupPrice uint64,
	minPeople, maxPeople, totalStock int32, startTime, endTime time.Time) *Groupbuy {
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
		Status:        GroupbuyStatusPending, // 默认状态为未开始。
	}
}

// RemainingStock 计算拼团活动的剩余库存。
func (g *Groupbuy) RemainingStock() int32 {
	return g.TotalStock - g.SoldCount
}

// IsAvailable 检查拼团活动当前是否可用（正在进行中且有库存）。
func (g *Groupbuy) IsAvailable() bool {
	now := time.Now()
	return g.Status == GroupbuyStatusOngoing && // 状态为进行中。
		now.After(g.StartTime) && // 当前时间在开始时间之后。
		now.Before(g.EndTime) && // 当前时间在结束时间之前。
		g.SoldCount < g.TotalStock // 仍有剩余库存。
}

// Start 启动拼团活动，将其状态设置为“进行中”。
func (g *Groupbuy) Start() {
	g.Status = GroupbuyStatusOngoing
}

// End 结束拼团活动，将其状态设置为“已结束”。
func (g *Groupbuy) End() {
	g.Status = GroupbuyStatusEnded
}

// Cancel 取消拼团活动，将其状态设置为“已取消”。
func (g *Groupbuy) Cancel() {
	g.Status = GroupbuyStatusCanceled
}

// GroupbuyTeam 实体代表一个拼团团队。
// 它是用户发起或加入拼团活动后形成的具体团队。
type GroupbuyTeam struct {
	gorm.Model                       // 嵌入gorm.Model。
	GroupbuyID    uint64             `gorm:"not null;index;comment:拼团活动ID" json:"groupbuy_id"`                  // 关联的拼团活动ID，索引字段。
	TeamNo        string             `gorm:"type:varchar(64);uniqueIndex;not null;comment:拼团编号" json:"team_no"` // 拼团团队的唯一编号，唯一索引。
	LeaderID      uint64             `gorm:"not null;comment:团长用户ID" json:"leader_id"`                          // 团长的用户ID。
	CurrentPeople int32              `gorm:"not null;default:1;comment:当前人数" json:"current_people"`             // 当前团队中的人数。
	MaxPeople     int32              `gorm:"not null;comment:最大人数" json:"max_people"`                           // 拼团团队的最大人数。
	Status        GroupbuyTeamStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`          // 团队状态，默认为拼团中。
	ExpireAt      time.Time          `gorm:"not null;comment:过期时间" json:"expire_at"`                            // 团队的过期时间。
	SuccessAt     *time.Time         `gorm:"comment:成团时间" json:"success_at"`                                    // 团队成功组建的时间。
}

// GroupbuyTeamStatus 定义了拼团团队的生命周期状态。
type GroupbuyTeamStatus int8

const (
	GroupbuyTeamStatusOngoing   GroupbuyTeamStatus = 0 // 拼团中：团队正在组建，等待更多成员加入。
	GroupbuyTeamStatusSuccess   GroupbuyTeamStatus = 1 // 拼团成功：团队人数已达到成团要求。
	GroupbuyTeamStatusFailed    GroupbuyTeamStatus = 2 // 拼团失败：团队未在规定时间内达到成团人数。
	GroupbuyTeamStatusCancelled GroupbuyTeamStatus = 3 // 已取消：团队被取消。
)

// NewGroupbuyTeam 创建并返回一个新的 GroupbuyTeam 实体实例。
// groupbuyID: 关联的拼团活动ID。
// teamNo: 团队编号。
// leaderID: 团长用户ID。
// maxPeople: 最大人数。
// expireAt: 过期时间。
func NewGroupbuyTeam(groupbuyID uint64, teamNo string, leaderID uint64, maxPeople int32, expireAt time.Time) *GroupbuyTeam {
	return &GroupbuyTeam{
		GroupbuyID:    groupbuyID,
		TeamNo:        teamNo,
		LeaderID:      leaderID,
		CurrentPeople: 1, // 团长发起时，人数为1。
		MaxPeople:     maxPeople,
		Status:        GroupbuyTeamStatusOngoing, // 默认状态为拼团中。
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
	return t.Status == GroupbuyTeamStatusOngoing && // 团队状态为进行中。
		!t.IsFull() && // 团队未满员。
		!t.IsExpired() // 团队未过期。
}

// Join 成员加入拼团团队。
// 如果团队已满或已过期，则返回错误。
func (t *GroupbuyTeam) Join() error {
	if !t.CanJoin() {
		return ErrGroupFull // 团队已满或不能加入。
	}

	t.CurrentPeople++ // 增加当前人数。

	// 检查是否达到成团人数，如果达到则更新状态为成功。
	if t.CurrentPeople >= t.MaxPeople {
		t.Success()
	}

	return nil
}

// Success 标记拼团团队为成功组建。
func (t *GroupbuyTeam) Success() {
	t.Status = GroupbuyTeamStatusSuccess // 状态更新为拼团成功。
	now := time.Now()
	t.SuccessAt = &now // 记录成团时间。
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
// 它记录了用户参与拼团的订单信息。
type GroupbuyOrder struct {
	gorm.Model                      // 嵌入gorm.Model。
	GroupbuyID  uint64              `gorm:"not null;index;comment:拼团活动ID" json:"groupbuy_id"`         // 关联的拼团活动ID，索引字段。
	TeamID      uint64              `gorm:"not null;index;comment:拼团团队ID" json:"team_id"`             // 关联的拼团团队ID，索引字段。
	TeamNo      string              `gorm:"type:varchar(64);not null;comment:拼团编号" json:"team_no"`    // 关联的拼团编号。
	UserID      uint64              `gorm:"not null;index;comment:用户ID" json:"user_id"`               // 下单用户ID，索引字段。
	ProductID   uint64              `gorm:"not null;comment:商品ID" json:"product_id"`                  // 商品ID。
	SkuID       uint64              `gorm:"not null;comment:SKU ID" json:"sku_id"`                    // SKU ID。
	Price       uint64              `gorm:"not null;comment:单价(分)" json:"price"`                      // 商品单价（拼团价）。
	Quantity    int32               `gorm:"not null;comment:数量" json:"quantity"`                      // 购买数量。
	TotalAmount uint64              `gorm:"not null;comment:总金额(分)" json:"total_amount"`              // 订单总金额。
	IsLeader    bool                `gorm:"not null;default:false;comment:是否团长" json:"is_leader"`     // 是否为团长订单。
	Status      GroupbuyOrderStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"` // 订单状态，默认为待支付。
	PaidAt      *time.Time          `gorm:"comment:支付时间" json:"paid_at"`                              // 支付时间。
	RefundedAt  *time.Time          `gorm:"comment:退款时间" json:"refunded_at"`                          // 退款时间。
}

// GroupbuyOrderStatus 定义了拼团订单的生命周期状态。
type GroupbuyOrderStatus int8

const (
	GroupbuyOrderStatusPending   GroupbuyOrderStatus = 0 // 待支付：订单已创建，等待用户支付。
	GroupbuyOrderStatusPaid      GroupbuyOrderStatus = 1 // 已支付：订单已支付成功。
	GroupbuyOrderStatusSuccess   GroupbuyOrderStatus = 2 // 拼团成功：拼团成功后的订单状态。
	GroupbuyOrderStatusFailed    GroupbuyOrderStatus = 3 // 拼团失败：拼团失败后的订单状态。
	GroupbuyOrderStatusRefunded  GroupbuyOrderStatus = 4 // 已退款：订单已退款。
	GroupbuyOrderStatusCancelled GroupbuyOrderStatus = 5 // 已取消：订单已取消。
)

// NewGroupbuyOrder 创建并返回一个新的 GroupbuyOrder 实体实例。
// groupbuyID, teamID: 关联的拼团活动和团队ID。
// teamNo: 团队编号。
// userID: 用户ID。
// productID, skuID: 商品ID和SKU ID。
// price: 单价。
// quantity: 数量。
// isLeader: 是否为团长订单。
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
		TotalAmount: price * uint64(quantity), // 计算总金额。
		IsLeader:    isLeader,
		Status:      GroupbuyOrderStatusPending, // 默认状态为待支付。
	}
}

// Pay 支付订单，将其状态设置为“已支付”，并记录支付时间。
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

// Refund 退款订单，将其状态设置为“已退款”，并记录退款时间。
func (o *GroupbuyOrder) Refund() {
	o.Status = GroupbuyOrderStatusRefunded
	now := time.Now()
	o.RefundedAt = &now
}

// Cancel 取消订单，将其状态设置为“已取消”。
func (o *GroupbuyOrder) Cancel() {
	o.Status = GroupbuyOrderStatusCancelled
}
