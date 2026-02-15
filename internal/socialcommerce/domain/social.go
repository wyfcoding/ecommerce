// Package domain 社交电商领域模型
// 生成摘要：
// 1) 定义 GroupBuy（拼团）聚合根：管理拼团生命周期、成团条件
// 2) 定义 Bargain（砍价）聚合根：管理砍价进度、助力记录
// 3) 定义 Distribution（分销）关系：绑定上下级、计算佣金
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// --- 拼团 (Group Buying) ---

type GroupBuyStatus int8

const (
	GroupBuyStatusPending GroupBuyStatus = 1 // 待成团
	GroupBuyStatusSuccess GroupBuyStatus = 2 // 拼团成功
	GroupBuyStatusFailed  GroupBuyStatus = 3 // 拼团失败（超时/人数不足）
)

// GroupBuyActivity 拼团活动配置
type GroupBuyActivity struct {
	gorm.Model
	ActivityName   string          `gorm:"column:activity_name;type:varchar(128);not null"`
	SPUID          uint64          `gorm:"column:spu_id;index;not null"`
	SKUID          uint64          `gorm:"column:sku_id;index;not null"`
	GroupPrice     decimal.Decimal `gorm:"column:group_price;type:decimal(20,2);not null"` // 拼团价
	RequireNum     int             `gorm:"column:require_num;not null"`                    // 成团人数
	Duration       int             `gorm:"column:duration;not null"`                       // 成团有效时长(秒)
	StartTime      time.Time       `gorm:"column:start_time;index"`
	EndTime        time.Time       `gorm:"column:end_time;index"`
	Status         int8            `gorm:"column:status;default:1"` // 1上架 0下架
}

func (GroupBuyActivity) TableName() string { return "social_group_buy_activities" }

// GroupBuyInstance 拼团实例（开团记录）
type GroupBuyInstance struct {
	gorm.Model
	InstanceNo     string          `gorm:"column:instance_no;type:varchar(64);uniqueIndex;not null"`
	ActivityID     uint            `gorm:"column:activity_id;index;not null"`
	InitiatorID    uint64          `gorm:"column:initiator_id;index;not null"` // 团长
	CurrentNum     int             `gorm:"column:current_num;default:1"`
	RequireNum     int             `gorm:"column:require_num;not null"`
	Status         GroupBuyStatus  `gorm:"column:status;type:tinyint;default:1"`
	ExpireAt       time.Time       `gorm:"column:expire_at;index"`
	SuccessAt      *time.Time      `gorm:"column:success_at"`
	
	Members        []GroupBuyMember `gorm:"foreignKey:InstanceID"`
}

func (GroupBuyInstance) TableName() string { return "social_group_buy_instances" }

// GroupBuyMember 参团成员
type GroupBuyMember struct {
	gorm.Model
	InstanceID uint   `gorm:"column:instance_id;index;not null"`
	UserID     uint64 `gorm:"column:user_id;index;not null"`
	OrderID    string `gorm:"column:order_id;type:varchar(64);not null"` // 关联订单
	IsInitiator bool  `gorm:"column:is_initiator;default:false"`
	JoinTime   time.Time `gorm:"column:join_time"`
}

func (GroupBuyMember) TableName() string { return "social_group_buy_members" }

// --- 砍价 (Bargain) ---

type BargainStatus int8

const (
	BargainStatusProcessing BargainStatus = 1
	BargainStatusSuccess    BargainStatus = 2
	BargainStatusFailed     BargainStatus = 3
)

// BargainActivity 砍价活动
type BargainActivity struct {
	gorm.Model
	ActivityName string          `gorm:"column:activity_name;type:varchar(128)"`
	SPUID        uint64          `gorm:"column:spu_id"`
	OriginalPrice decimal.Decimal `gorm:"column:original_price;type:decimal(20,2)"`
	MinPrice      decimal.Decimal `gorm:"column:min_price;type:decimal(20,2)"` // 底价
	TotalCuts     int             `gorm:"column:total_cuts"`                    // 预计需砍次数
	StartTime     time.Time       `gorm:"column:start_time"`
	EndTime       time.Time       `gorm:"column:end_time"`
}

func (BargainActivity) TableName() string { return "social_bargain_activities" }

// BargainInstance 用户发起的砍价
type BargainInstance struct {
	gorm.Model
	ActivityID    uint            `gorm:"column:activity_id;index"`
	UserID        uint64          `gorm:"column:user_id;index"`
	CurrentPrice  decimal.Decimal `gorm:"column:current_price;type:decimal(20,2)"`
	CutAmount     decimal.Decimal `gorm:"column:cut_amount;type:decimal(20,2)"` // 已砍掉金额
	CutTimes      int             `gorm:"column:cut_times"`
	Status        BargainStatus   `gorm:"column:status;type:tinyint"`
	IsOrdered     bool            `gorm:"column:is_ordered;default:false"` // 是否已以此价格下单
}

func (BargainInstance) TableName() string { return "social_bargain_instances" }

// --- 分销 (Distribution) ---

// DistributionRelation 分销绑定关系
type DistributionRelation struct {
	gorm.Model
	UserID        uint64    `gorm:"column:user_id;uniqueIndex;not null"`
	ParentID      uint64    `gorm:"column:parent_id;index;not null"`
	RootID        uint64    `gorm:"column:root_id;index"` // 顶级分销商
	Level         int       `gorm:"column:level"` // 层级深度
	BindTime      time.Time `gorm:"column:bind_time"`
	ExpireAt      *time.Time `gorm:"column:expire_at"` // 绑定关系有效期
}

func (DistributionRelation) TableName() string { return "social_distribution_relations" }

// --- Logic ---

func (g *GroupBuyInstance) Join(userID uint64, orderID string) error {
	if g.Status != GroupBuyStatusPending {
		return errors.New("group buy is not pending")
	}
	if time.Now().After(g.ExpireAt) {
		g.Status = GroupBuyStatusFailed
		return errors.New("group buy expired")
	}
	if g.CurrentNum >= g.RequireNum {
		return errors.New("group buy full")
	}

	g.CurrentNum++
	g.Members = append(g.Members, GroupBuyMember{
		UserID:      userID,
		OrderID:     orderID,
		IsInitiator: false,
		JoinTime:    time.Now(),
	})

	if g.CurrentNum >= g.RequireNum {
		g.Status = GroupBuyStatusSuccess
		now := time.Now()
		g.SuccessAt = &now
	}
	return nil
}

// Repositories
type GroupBuyRepository interface {
	SaveActivity(ctx context.Context, act *GroupBuyActivity) error
	GetActivity(ctx context.Context, id uint) (*GroupBuyActivity, error)
	
	SaveInstance(ctx context.Context, ins *GroupBuyInstance) error
	GetInstance(ctx context.Context, instanceNo string) (*GroupBuyInstance, error)
}

type DistributionRepository interface {
	BindParent(ctx context.Context, userID, parentID uint64) error
	GetParent(ctx context.Context, userID uint64) (parentID uint64, err error)
}
