package domain

import (
	"time" // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
)

// TierLevel 定义了用户等级的级别。
type TierLevel int32

const (
	TierLevelRegular TierLevel = 0 // 普通会员。
	TierLevelBronze  TierLevel = 1 // 青铜会员。
	TierLevelSilver  TierLevel = 2 // 白银会员。
	TierLevelGold    TierLevel = 3 // 黄金会员。
)

// UserTier 实体是用户等级模块的聚合根。
// 它记录了用户的当前等级、成长值、升级进度和所享受的折扣等信息。
type UserTier struct {
	gorm.Model                    // 嵌入gorm.Model。
	UserID              uint64    `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`                          // 用户ID，唯一索引，不允许为空。
	Level               TierLevel `gorm:"type:int;not null;default:0;comment:等级" json:"level"`                       // 用户当前等级，默认为普通会员。
	LevelName           string    `gorm:"type:varchar(32);comment:等级名称" json:"level_name"`                           // 等级名称，例如“普通会员”，“黄金会员”。
	Score               int64     `gorm:"not null;default:0;comment:成长值" json:"score"`                               // 用户当前的成长值。
	NextLevelScore      int64     `gorm:"not null;default:0;comment:下一级所需成长值" json:"next_level_score"`               // 升级到下一级所需的成长值。
	ProgressToNextLevel float64   `gorm:"type:decimal(5,2);default:0;comment:升级进度(%)" json:"progress_to_next_level"` // 升级进度百分比。
	DiscountRate        float64   `gorm:"type:decimal(5,2);default:100;comment:折扣率(%)" json:"discount_rate"`         // 该等级享受的折扣率（例如90表示9折）。
	Points              float64   `gorm:"type:decimal(10,2);default:0;comment:当前积分" json:"points"`                   // 用户当前的积分。
}

// TierConfig 实体定义了不同等级的配置和权益。
// 它是系统级别的配置，而非用户个人数据。
type TierConfig struct {
	gorm.Model                 // 嵌入gorm.Model。
	Level            TierLevel `gorm:"uniqueIndex;not null;comment:等级" json:"level"`                        // 等级，唯一索引，不允许为空。
	LevelName        string    `gorm:"type:varchar(32);not null;comment:等级名称" json:"level_name"`            // 等级名称。
	MinScore         int64     `gorm:"not null;default:0;comment:最低成长值" json:"min_score"`                   // 达到此等级所需的最低成长值。
	DiscountRate     float64   `gorm:"type:decimal(5,2);default:100;comment:折扣率(%)" json:"discount_rate"`   // 该等级享受的折扣率。
	ExtraPointsRate  float64   `gorm:"type:decimal(5,2);default:1.0;comment:积分倍率" json:"extra_points_rate"` // 该等级额外积分倍率。
	FreeShipping     bool      `gorm:"not null;default:false;comment:包邮" json:"free_shipping"`              // 是否享受包邮。
	PrioritySupport  bool      `gorm:"not null;default:false;comment:优先客服" json:"priority_support"`         // 是否享受优先客服。
	ExclusiveDeals   bool      `gorm:"not null;default:false;comment:专属优惠" json:"exclusive_deals"`          // 是否享受专属优惠。
	BirthdayBonus    int64     `gorm:"not null;default:0;comment:生日奖励积分" json:"birthday_bonus"`             // 生日奖励积分。
	AnniversaryBonus int64     `gorm:"not null;default:0;comment:周年奖励积分" json:"anniversary_bonus"`          // 周年奖励积分。
}

// PointsAccount 积分账户实体。
// 记录用户的积分余额。
type PointsAccount struct {
	gorm.Model        // 嵌入gorm.Model。
	UserID     uint64 `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"` // 用户ID，唯一索引，不允许为空。
	Balance    int64  `gorm:"not null;default:0;comment:积分余额" json:"balance"`   // 积分余额。
}

// PointsLog 积分日志实体。
// 记录用户积分的每一次变动。
type PointsLog struct {
	gorm.Model        // 嵌入gorm.Model。
	UserID     uint64 `gorm:"index;not null;comment:用户ID" json:"user_id"`          // 关联的用户ID，索引字段。
	Points     int64  `gorm:"not null;comment:变动积分" json:"points"`                 // 积分变动数量（正数增加，负数减少）。
	Reason     string `gorm:"type:varchar(255);comment:变动原因" json:"reason"`        // 积分变动的原因。
	Type       string `gorm:"type:varchar(32);comment:类型(add/deduct)" json:"type"` // 变动类型，例如“add”（增加）或“deduct”（扣除）。
}

// Exchange 兑换商品实体。
// 定义了积分商城中可供兑换的商品。
type Exchange struct {
	gorm.Model            // 嵌入gorm.Model。
	Name           string `gorm:"type:varchar(128);not null;comment:商品名称" json:"name"` // 兑换商品名称。
	Description    string `gorm:"type:varchar(255);comment:描述" json:"description"`     // 兑换商品描述。
	RequiredPoints int64  `gorm:"not null;comment:所需积分" json:"required_points"`        // 兑换此商品所需的积分数量。
	Stock          int32  `gorm:"not null;default:0;comment:库存" json:"stock"`          // 兑换商品的库存。
}

// ExchangeRecord 兑换记录实体。
// 记录了用户的兑换历史。
type ExchangeRecord struct {
	gorm.Model        // 嵌入gorm.Model。
	UserID     uint64 `gorm:"index;not null;comment:用户ID" json:"user_id"`       // 兑换用户ID，索引字段。
	ExchangeID uint64 `gorm:"index;not null;comment:兑换商品ID" json:"exchange_id"` // 兑换的商品ID，索引字段。
	Points     int64  `gorm:"not null;comment:消耗积分" json:"points"`              // 兑换消耗的积分。
}

// UserStatistics 用户统计实体。
// 记录用户的各项统计数据，例如总消费、订单数等。
// 备注：此实体可能是可选的，其数据也可能更适合由分析服务或用户服务维护。
// 目前根据现有模型保留在此。
type UserStatistics struct {
	gorm.Model              // 嵌入gorm.Model。
	UserID        uint64    `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`     // 用户ID，唯一索引，不允许为空。
	TotalSpent    int64     `gorm:"not null;default:0;comment:总消费(分)" json:"total_spent"` // 累计总消费金额（单位：分）。
	OrderCount    int64     `gorm:"not null;default:0;comment:订单数" json:"order_count"`    // 累计订单数量。
	ReviewCount   int64     `gorm:"not null;default:0;comment:评价数" json:"review_count"`   // 累计评论数量。
	ReferralCount int64     `gorm:"not null;default:0;comment:推荐数" json:"referral_count"` // 累计推荐用户数量。
	LastLoginTime time.Time `gorm:"comment:最后登录时间" json:"last_login_time"`                // 最后一次登录时间。
}
