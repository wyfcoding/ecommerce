package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrCampaignNotFound = errors.New("营销活动不存在")
	ErrCampaignEnded    = errors.New("营销活动已结束")
)

// CampaignType 活动类型
type CampaignType string

const (
	CampaignTypeDiscount   CampaignType = "DISCOUNT"    // 折扣活动
	CampaignTypeFullReduce CampaignType = "FULL_REDUCE" // 满减活动
	CampaignTypeGift       CampaignType = "GIFT"        // 赠品活动
	CampaignTypeBundling   CampaignType = "BUNDLING"    // 组合活动
)

// CampaignStatus 活动状态
type CampaignStatus int8

const (
	CampaignStatusPending  CampaignStatus = 0 // 未开始
	CampaignStatusOngoing  CampaignStatus = 1 // 进行中
	CampaignStatusEnded    CampaignStatus = 2 // 已结束
	CampaignStatusCanceled CampaignStatus = 3 // 已取消
)

// JSONMap defines a map that implements the sql.Scanner and driver.Valuer interfaces
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

// Campaign 营销活动实体
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

// NewCampaign 创建营销活动
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

// IsActive 是否激活
func (c *Campaign) IsActive() bool {
	now := time.Now()
	return c.Status == CampaignStatusOngoing && now.After(c.StartTime) && now.Before(c.EndTime)
}

// Start 开始活动
func (c *Campaign) Start() {
	c.Status = CampaignStatusOngoing
}

// End 结束活动
func (c *Campaign) End() {
	c.Status = CampaignStatusEnded
}

// Cancel 取消活动
func (c *Campaign) Cancel() {
	c.Status = CampaignStatusCanceled
}

// AddSpent 增加花费
func (c *Campaign) AddSpent(amount uint64) {
	c.Spent += amount
}

// IncrementReachedUsers 增加触达用户数
func (c *Campaign) IncrementReachedUsers() {
	c.ReachedUsers++
}

// RemainingBudget 剩余预算
func (c *Campaign) RemainingBudget() uint64 {
	if c.Spent >= c.Budget {
		return 0
	}
	return c.Budget - c.Spent
}

// CampaignParticipation 活动参与记录实体
type CampaignParticipation struct {
	gorm.Model
	CampaignID uint64 `gorm:"not null;index;comment:活动ID" json:"campaign_id"`
	UserID     uint64 `gorm:"not null;index;comment:用户ID" json:"user_id"`
	OrderID    uint64 `gorm:"index;comment:订单ID" json:"order_id"`
	Discount   uint64 `gorm:"not null;default:0;comment:优惠金额" json:"discount"`
}

// NewCampaignParticipation 创建活动参与记录
func NewCampaignParticipation(campaignID, userID, orderID, discount uint64) *CampaignParticipation {
	return &CampaignParticipation{
		CampaignID: campaignID,
		UserID:     userID,
		OrderID:    orderID,
		Discount:   discount,
	}
}

// Banner 广告横幅实体
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

// NewBanner 创建广告横幅
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

// IsActive 是否激活
func (b *Banner) IsActive() bool {
	now := time.Now()
	return b.Enabled && now.After(b.StartTime) && now.Before(b.EndTime)
}

// IncrementClick 增加点击数
func (b *Banner) IncrementClick() {
	b.ClickCount++
}

// Enable 启用
func (b *Banner) Enable() {
	b.Enabled = true
}

// Disable 禁用
func (b *Banner) Disable() {
	b.Enabled = false
}
