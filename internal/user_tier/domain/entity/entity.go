package entity

import (
	"time"

	"gorm.io/gorm"
)

// TierLevel 等级等级
type TierLevel int32

const (
	TierLevelRegular TierLevel = 0
	TierLevelBronze  TierLevel = 1
	TierLevelSilver  TierLevel = 2
	TierLevelGold    TierLevel = 3
)

// UserTier 用户等级实体
type UserTier struct {
	gorm.Model
	UserID              uint64    `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`
	Level               TierLevel `gorm:"type:int;not null;default:0;comment:等级" json:"level"`
	LevelName           string    `gorm:"type:varchar(32);comment:等级名称" json:"level_name"`
	Score               int64     `gorm:"not null;default:0;comment:成长值" json:"score"`
	NextLevelScore      int64     `gorm:"not null;default:0;comment:下一级所需成长值" json:"next_level_score"`
	ProgressToNextLevel float64   `gorm:"type:decimal(5,2);default:0;comment:升级进度(%)" json:"progress_to_next_level"`
	DiscountRate        float64   `gorm:"type:decimal(5,2);default:100;comment:折扣率(%)" json:"discount_rate"`
	Points              float64   `gorm:"type:decimal(10,2);default:0;comment:当前积分" json:"points"`
}

// TierConfig 等级配置实体 (Renamed from TierBenefits to be more generic)
type TierConfig struct {
	gorm.Model
	Level            TierLevel `gorm:"uniqueIndex;not null;comment:等级" json:"level"`
	LevelName        string    `gorm:"type:varchar(32);not null;comment:等级名称" json:"level_name"`
	MinScore         int64     `gorm:"not null;default:0;comment:最低成长值" json:"min_score"`
	DiscountRate     float64   `gorm:"type:decimal(5,2);default:100;comment:折扣率(%)" json:"discount_rate"`
	ExtraPointsRate  float64   `gorm:"type:decimal(5,2);default:1.0;comment:积分倍率" json:"extra_points_rate"`
	FreeShipping     bool      `gorm:"not null;default:false;comment:包邮" json:"free_shipping"`
	PrioritySupport  bool      `gorm:"not null;default:false;comment:优先客服" json:"priority_support"`
	ExclusiveDeals   bool      `gorm:"not null;default:false;comment:专属优惠" json:"exclusive_deals"`
	BirthdayBonus    int64     `gorm:"not null;default:0;comment:生日奖励积分" json:"birthday_bonus"`
	AnniversaryBonus int64     `gorm:"not null;default:0;comment:周年奖励积分" json:"anniversary_bonus"`
}

// PointsAccount 积分账户实体
type PointsAccount struct {
	gorm.Model
	UserID  uint64 `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`
	Balance int64  `gorm:"not null;default:0;comment:积分余额" json:"balance"`
}

// PointsLog 积分日志实体
type PointsLog struct {
	gorm.Model
	UserID uint64 `gorm:"index;not null;comment:用户ID" json:"user_id"`
	Points int64  `gorm:"not null;comment:变动积分" json:"points"`
	Reason string `gorm:"type:varchar(255);comment:变动原因" json:"reason"`
	Type   string `gorm:"type:varchar(32);comment:类型(add/deduct)" json:"type"`
}

// Exchange 兑换商品实体
type Exchange struct {
	gorm.Model
	Name           string `gorm:"type:varchar(128);not null;comment:商品名称" json:"name"`
	Description    string `gorm:"type:varchar(255);comment:描述" json:"description"`
	RequiredPoints int64  `gorm:"not null;comment:所需积分" json:"required_points"`
	Stock          int32  `gorm:"not null;default:0;comment:库存" json:"stock"`
}

// ExchangeRecord 兑换记录实体
type ExchangeRecord struct {
	gorm.Model
	UserID     uint64 `gorm:"index;not null;comment:用户ID" json:"user_id"`
	ExchangeID uint64 `gorm:"index;not null;comment:兑换商品ID" json:"exchange_id"`
	Points     int64  `gorm:"not null;comment:消耗积分" json:"points"`
}

// UserStatistics 用户统计实体 (Optional, might be better in analytics or user service, but keeping here as per existing model)
type UserStatistics struct {
	gorm.Model
	UserID        uint64    `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`
	TotalSpent    int64     `gorm:"not null;default:0;comment:总消费(分)" json:"total_spent"`
	OrderCount    int64     `gorm:"not null;default:0;comment:订单数" json:"order_count"`
	ReviewCount   int64     `gorm:"not null;default:0;comment:评价数" json:"review_count"`
	ReferralCount int64     `gorm:"not null;default:0;comment:推荐数" json:"referral_count"`
	LastLoginTime time.Time `gorm:"comment:最后登录时间" json:"last_login_time"`
}
