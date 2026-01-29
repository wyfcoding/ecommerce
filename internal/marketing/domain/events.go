package domain

import (
	"time"
)

// CampaignCreatedEvent 营销活动创建事件
type CampaignCreatedEvent struct {
	CampaignID uint         `json:"campaign_id"`
	Name       string       `json:"name"`
	Type       CampaignType `json:"type"`
	StartTime  time.Time    `json:"start_time"`
	EndTime    time.Time    `json:"end_time"`
	Timestamp  time.Time    `json:"timestamp"`
}

// CampaignStatusUpdatedEvent 营销活动状态更新事件
type CampaignStatusUpdatedEvent struct {
	CampaignID uint           `json:"campaign_id"`
	OldStatus  CampaignStatus `json:"old_status"`
	NewStatus  CampaignStatus `json:"new_status"`
	Timestamp  time.Time      `json:"timestamp"`
}

// ParticipationRecordedEvent 用户参与记录事件
type ParticipationRecordedEvent struct {
	CampaignID uint      `json:"campaign_id"`
	UserID     uint64    `json:"user_id"`
	OrderID    uint64    `json:"order_id"`
	Discount   uint64    `json:"discount"`
	Timestamp  time.Time `json:"timestamp"`
}

// BannerCreatedEvent 广告位创建事件
type BannerCreatedEvent struct {
	BannerID  uint      `json:"banner_id"`
	Title     string    `json:"title"`
	Position  string    `json:"position"`
	Timestamp time.Time `json:"timestamp"`
}
