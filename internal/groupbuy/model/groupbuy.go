package model

import "time"

// GroupBuyActivityStatus 拼团活动状态
type GroupBuyActivityStatus string

const (
	GroupBuyActivityStatusPending   GroupBuyActivityStatus = "PENDING"   // 待开始
	GroupBuyActivityStatusOngoing   GroupBuyActivityStatus = "ONGOING"   // 进行中
	GroupBuyActivityStatusEnded     GroupBuyActivityStatus = "ENDED"     // 已结束
	GroupBuyActivityStatusCancelled GroupBuyActivityStatus = "CANCELLED" // 已取消
)

// GroupBuyActivity 拼团活动
type GroupBuyActivity struct {
	ID              uint64                 `gorm:"primarykey" json:"id"`
	Name            string                 `gorm:"type:varchar(255);not null;comment:活动名称" json:"name"`
	ProductID       uint64                 `gorm:"index;not null;comment:商品ID" json:"productId"`
	SKUID           uint64                 `gorm:"index;not null;comment:SKU ID" json:"skuId"`
	OriginalPrice   uint64                 `gorm:"not null;comment:原价(分)" json:"originalPrice"`
	GroupPrice      uint64                 `gorm:"not null;comment:拼团价(分)" json:"groupPrice"`
	RequiredMembers int32                  `gorm:"not null;comment:成团人数" json:"requiredMembers"`
	TimeLimit       int32                  `gorm:"not null;comment:拼团时限(小时)" json:"timeLimit"`
	StockQuantity   uint32                 `gorm:"not null;comment:活动库存" json:"stockQuantity"`
	SoldCount       uint32                 `gorm:"not null;default:0;comment:已售数量" json:"soldCount"`
	LimitPerUser    uint32                 `gorm:"not null;default:1;comment:每人限购" json:"limitPerUser"`
	Status          GroupBuyActivityStatus `gorm:"type:varchar(20);not null;comment:活动状态" json:"status"`
	StartTime       time.Time              `gorm:"not null;comment:开始时间" json:"startTime"`
	EndTime         time.Time              `gorm:"not null;comment:结束时间" json:"endTime"`
	CreatedAt       time.Time              `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time              `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (GroupBuyActivity) TableName() string {
	return "groupbuy_activities"
}

// GroupStatus 拼团状态
type GroupStatus string

const (
	GroupStatusRecruiting GroupStatus = "RECRUITING" // 招募中
	GroupStatusSuccess    GroupStatus = "SUCCESS"    // 拼团成功
	GroupStatusFailed     GroupStatus = "FAILED"     // 拼团失败
	GroupStatusCancelled  GroupStatus = "CANCELLED"  // 已取消
)

// Group 拼团
type Group struct {
	ID              uint64      `gorm:"primarykey" json:"id"`
	GroupNo         string      `gorm:"uniqueIndex;type:varchar(100);not null;comment:拼团编号" json:"groupNo"`
	ActivityID      uint64      `gorm:"index;not null;comment:活动ID" json:"activityId"`
	LeaderID        uint64      `gorm:"index;not null;comment:团长ID" json:"leaderId"`
	RequiredMembers int32       `gorm:"not null;comment:成团人数" json:"requiredMembers"`
	CurrentMembers  int32       `gorm:"not null;default:1;comment:当前人数" json:"currentMembers"`
	Status          GroupStatus `gorm:"type:varchar(20);not null;comment:拼团状态" json:"status"`
	ExpireTime      time.Time   `gorm:"not null;comment:过期时间" json:"expireTime"`
	SuccessTime     *time.Time  `gorm:"comment:成团时间" json:"successTime"`
	CreatedAt       time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (Group) TableName() string {
	return "groups"
}

// GroupMember 拼团成员
type GroupMember struct {
	ID         uint64    `gorm:"primarykey" json:"id"`
	GroupID    uint64    `gorm:"index;not null;comment:拼团ID" json:"groupId"`
	UserID     uint64    `gorm:"index;not null;comment:用户ID" json:"userId"`
	OrderID    uint64    `gorm:"index;not null;comment:订单ID" json:"orderId"`
	IsLeader   bool      `gorm:"not null;default:false;comment:是否团长" json:"isLeader"`
	JoinTime   time.Time `gorm:"not null;comment:加入时间" json:"joinTime"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName 指定表名
func (GroupMember) TableName() string {
	return "group_members"
}

// IsComplete 判断是否已成团
func (g *Group) IsComplete() bool {
	return g.CurrentMembers >= g.RequiredMembers
}

// IsExpired 判断是否已过期
func (g *Group) IsExpired() bool {
	return time.Now().After(g.ExpireTime)
}

// CanJoin 判断是否可以加入
func (g *Group) CanJoin() bool {
	return g.Status == GroupStatusRecruiting && 
		   !g.IsComplete() && 
		   !g.IsExpired()
}
