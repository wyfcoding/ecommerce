package domain

import "time"

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
	ID                  uint      `json:"id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	UserID              uint64    `json:"user_id"`                // 用户ID，唯一索引，不允许为空。
	Level               TierLevel `json:"level"`                  // 用户当前等级，默认为普通会员。
	LevelName           string    `json:"level_name"`             // 等级名称，例如“普通会员”，“黄金会员”。
	Score               int64     `json:"score"`                  // 用户当前的成长值。
	NextLevelScore      int64     `json:"next_level_score"`       // 升级到下一级所需的成长值。
	ProgressToNextLevel float64   `json:"progress_to_next_level"` // 升级进度百分比。
	DiscountRate        float64   `json:"discount_rate"`          // 折扣率。
	Points              float64   `json:"points"`                 // 当前积分。
}

// TierConfig 实体定义了不同等级的配置和权益。
// 它是系统级别的配置，而非用户个人数据。
type TierConfig struct {
	ID               uint      `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Level            TierLevel `json:"level"`
	LevelName        string    `json:"level_name"`
	MinScore         int64     `json:"min_score"`
	DiscountRate     float64   `json:"discount_rate"`
	ExtraPointsRate  float64   `json:"extra_points_rate"`
	FreeShipping     bool      `json:"free_shipping"`
	PrioritySupport  bool      `json:"priority_support"`
	ExclusiveDeals   bool      `json:"exclusive_deals"`
	BirthdayBonus    int64     `json:"birthday_bonus"`
	AnniversaryBonus int64     `json:"anniversary_bonus"`
}

// PointsAccount 积分账户实体。
// 记录用户的积分余额。
type PointsAccount struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uint64    `json:"user_id"`
	Balance   int64     `json:"balance"`
}

// PointsLog 积分日志实体。
// 记录用户积分的每一次变动。
type PointsLog struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uint64    `json:"user_id"`
	Points    int64     `json:"points"`
	Reason    string    `json:"reason"`
	Type      string    `json:"type"`
}
