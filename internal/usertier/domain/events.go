package domain

import "time"

// UserTierUpgradedEvent 用户等级提升事件。
type UserTierUpgradedEvent struct {
	UserID    uint64    `json:"user_id"`
	OldTier   string    `json:"old_tier"`
	NewTier   string    `json:"new_tier"`
	Timestamp time.Time `json:"timestamp"`
}

// UserTierDowngradedEvent 用户等级下降事件。
type UserTierDowngradedEvent struct {
	UserID    uint64    `json:"user_id"`
	OldTier   string    `json:"old_tier"`
	NewTier   string    `json:"new_tier"`
	Timestamp time.Time `json:"timestamp"`
}

// UserPointsChangedEvent 用户积分变动事件。
type UserPointsChangedEvent struct {
	UserID    uint64    `json:"user_id"`
	Points    int32     `json:"points"`
	Change    int32     `json:"change"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}
