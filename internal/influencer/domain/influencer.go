package domain

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type InfluencerStatus string

const (
	StatusPending  InfluencerStatus = "PENDING"
	StatusApproved InfluencerStatus = "APPROVED"
	StatusRejected InfluencerStatus = "REJECTED"
)

// Influencer 网红/KOL 实体
type Influencer struct {
	gorm.Model
	UserID        string           `gorm:"column:user_id;type:varchar(32);unique_index;not null"`
	InfluencerID  string           `gorm:"column:influencer_id;type:varchar(32);unique_index;not null"`
	Name          string           `gorm:"column:name;type:varchar(100);not null"`
	Platform      string           `gorm:"column:platform;type:varchar(50)"`
	Handle        string           `gorm:"column:handle;type:varchar(100)"`
	FollowerCount int32            `gorm:"column:follower_count"`
	Status        InfluencerStatus `gorm:"column:status;type:varchar(20);not null;default:'PENDING'"`
	TotalEarnings decimal.Decimal  `gorm:"column:total_earnings;type:decimal(32,16);default:0"`
}

// Campaign 推广活动实体
type Campaign struct {
	gorm.Model
	CampaignID     string          `gorm:"column:campaign_id;type:varchar(32);unique_index;not null"`
	InfluencerID   string          `gorm:"column:influencer_id;type:varchar(32);index;not null"`
	ProductID      string          `gorm:"column:product_id;type:varchar(32);index;not null"`
	CommissionRate decimal.Decimal `gorm:"column:commission_rate;type:decimal(10,4);not null"`
	StartAt        time.Time       `gorm:"column:start_at"`
	EndAt          time.Time       `gorm:"column:end_at"`
	Status         string          `gorm:"column:status;type:varchar(20);default:'ACTIVE'"`
}

func (Influencer) TableName() string { return "influencers" }
func (Campaign) TableName() string   { return "influencer_campaigns" }
