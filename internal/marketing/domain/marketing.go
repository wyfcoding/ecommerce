package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrCampaignNotFound = errors.New("campaign not found")
	ErrCampaignEnded    = errors.New("campaign ended")
)

// CampaignType 定义了营销活动的类型。
type CampaignType string

const (
	CampaignTypeDiscount   CampaignType = "DISCOUNT"
	CampaignTypeFullReduce CampaignType = "FULL_REDUCE"
	CampaignTypeGift       CampaignType = "GIFT"
	CampaignTypeBundling   CampaignType = "BUNDLING"
)

// CampaignStatus 定义了营销活动的生命周期状态。
type CampaignStatus int8

const (
	CampaignStatusPending  CampaignStatus = 0
	CampaignStatusOngoing  CampaignStatus = 1
	CampaignStatusEnded    CampaignStatus = 2
	CampaignStatusCanceled CampaignStatus = 3
)

// JSONMap 定义了一个map类型，用于JSON存储。
type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

// Campaign 实体是营销模块的聚合根。
type Campaign struct {
	gorm.Model
	Name         string         `gorm:"type:varchar(128);not null;comment:活动名称" json:"name"`
	CampaignType CampaignType   `gorm:"type:varchar(32);not null;comment:活动类型" json:"campaign_type"`
	Description  string         `gorm:"type:text;comment:活动描述" json:"description"`
	StartTime    time.Time      `gorm:"not null;comment:开始时间" json:"start_time"`
	EndTime      time.Time      `gorm:"not null;comment:结束时间" json:"end_time"`
	Budget       uint64         `gorm:"not null;default:0;comment:预算" json:"budget"`
	Spent        uint64         `gorm:"not null;default:0;comment:已花费" json:"spent"`
	TargetUsers  int64          `gorm:"not null;default:0;comment:目标用户数" json:"target_users"`
	ReachedUsers int64          `gorm:"not null;default:0;comment:触达用户数" json:"reached_users"`
	Status       CampaignStatus `gorm:"default:0;comment:状态" json:"status"`
	Rules        JSONMap        `gorm:"type:json;comment:规则配置" json:"rules"`
}

// NewCampaign 创建并返回一个新的 Campaign 实体实例。
func NewCampaign(name string, campaignType CampaignType, description string, startTime, endTime time.Time, budget uint64, rules map[string]interface{}) *Campaign {
	return &Campaign{
		Name:         name,
		CampaignType: campaignType,
		Description:  description,
		StartTime:    startTime,
		EndTime:      endTime,
		Budget:       budget,
		Spent:        0,
		TargetUsers:  0,
		ReachedUsers: 0,
		Status:       CampaignStatusPending,
		Rules:        JSONMap(rules),
	}
}

// IsActive 检查营销活动当前是否处于活跃状态。
func (c *Campaign) IsActive() bool {
	now := time.Now()
	// 注意：只要是 ongoing 且在时间范围内
	return c.Status == CampaignStatusOngoing &&
		now.After(c.StartTime) &&
		now.Before(c.EndTime)
}

// Start 启动营销活动。
func (c *Campaign) Start() {
	c.Status = CampaignStatusOngoing
}

// End 结束营销活动。
func (c *Campaign) End() {
	c.Status = CampaignStatusEnded
}

// Cancel 取消营销活动。
func (c *Campaign) Cancel() {
	c.Status = CampaignStatusCanceled
}

// AddSpent 增加已花费金额。
func (c *Campaign) AddSpent(amount uint64) {
	c.Spent += amount
}

// IncrementReachedUsers 增加触达用户数。
func (c *Campaign) IncrementReachedUsers() {
	c.ReachedUsers++
}

// RemainingBudget 计算剩余预算。
func (c *Campaign) RemainingBudget() uint64 {
	if c.Spent >= c.Budget {
		return 0
	}
	return c.Budget - c.Spent
}

// CampaignParticipation 实体代表用户参与营销活动的记录。
type CampaignParticipation struct {
	gorm.Model
	CampaignID uint64 `gorm:"not null;index;comment:活动ID" json:"campaign_id"`
	UserID     uint64 `gorm:"not null;index;comment:用户ID" json:"user_id"`
	OrderID    uint64 `gorm:"index;comment:订单ID" json:"order_id"`
	Discount   uint64 `gorm:"not null;default:0;comment:优惠金额" json:"discount"`
}

// NewCampaignParticipation 创建并返回一个新的 CampaignParticipation 实体实例。
func NewCampaignParticipation(campaignID, userID, orderID, discount uint64) *CampaignParticipation {
	return &CampaignParticipation{
		CampaignID: campaignID,
		UserID:     userID,
		OrderID:    orderID,
		Discount:   discount,
	}
}

// Banner 实体代表一个广告横幅。
type Banner struct {
	gorm.Model
	Title      string    `gorm:"type:varchar(128);not null;comment:标题" json:"title"`
	ImageURL   string    `gorm:"type:varchar(255);not null;comment:图片URL" json:"image_url"`
	LinkURL    string    `gorm:"type:varchar(255);comment:跳转URL" json:"link_url"`
	Position   string    `gorm:"type:varchar(32);not null;comment:位置" json:"position"`
	Priority   int32     `gorm:"default:0;comment:优先级" json:"priority"`
	StartTime  time.Time `gorm:"not null;comment:开始时间" json:"start_time"`
	EndTime    time.Time `gorm:"not null;comment:结束时间" json:"end_time"`
	ClickCount int64     `gorm:"default:0;comment:点击数" json:"click_count"`
	Enabled    bool      `gorm:"default:true;comment:是否启用" json:"enabled"`
}

// NewBanner 创建并返回一个新的 Banner 实体实例。
func NewBanner(title, imageURL, linkURL, position string, priority int32, startTime, endTime time.Time) *Banner {
	return &Banner{
		Title:      title,
		ImageURL:   imageURL,
		LinkURL:    linkURL,
		Position:   position,
		Priority:   priority,
		StartTime:  startTime,
		EndTime:    endTime,
		ClickCount: 0,
		Enabled:    true,
	}
}

// IsActive 检查Banner是否活跃。
func (b *Banner) IsActive() bool {
	now := time.Now()
	return b.Enabled &&
		now.After(b.StartTime) &&
		now.Before(b.EndTime)
}

// IncrementClick 增加点击次数。
func (b *Banner) IncrementClick() {
	b.ClickCount++
}

// Enable 启用Banner。
func (b *Banner) Enable() {
	b.Enabled = true
}

// Disable 禁用Banner。
func (b *Banner) Disable() {
	b.Enabled = false
}
