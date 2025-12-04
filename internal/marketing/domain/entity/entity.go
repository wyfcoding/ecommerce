package entity

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。

	"gorm.io/gorm" // 导入GORM库。
)

// 定义Marketing模块的业务错误。
var (
	ErrCampaignNotFound = errors.New("营销活动不存在") // 营销活动记录未找到。
	ErrCampaignEnded    = errors.New("营销活动已结束") // 营销活动已结束。
)

// CampaignType 定义了营销活动的类型。
type CampaignType string

const (
	CampaignTypeDiscount   CampaignType = "DISCOUNT"    // 折扣活动，例如满减、打折。
	CampaignTypeFullReduce CampaignType = "FULL_REDUCE" // 满减活动。
	CampaignTypeGift       CampaignType = "GIFT"        // 赠品活动。
	CampaignTypeBundling   CampaignType = "BUNDLING"    // 组合销售活动。
)

// CampaignStatus 定义了营销活动的生命周期状态。
type CampaignStatus int8

const (
	CampaignStatusPending  CampaignStatus = 0 // 未开始：活动已创建但尚未到开始时间。
	CampaignStatusOngoing  CampaignStatus = 1 // 进行中：活动正在进行。
	CampaignStatusEnded    CampaignStatus = 2 // 已结束：活动已过结束时间。
	CampaignStatusCanceled CampaignStatus = 3 // 已取消：活动被取消。
)

// JSONMap 定义了一个map类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的map[string]interface{}类型作为JSON字符串存储到数据库，并从数据库读取。
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口，将 JSONMap 转换为数据库可以存储的值（JSON字节数组）。
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m) // 将map编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 JSONMap。
func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m) // 将JSON字节数组解码为map。
}

// Campaign 实体是营销模块的聚合根。
// 它代表一个营销活动，包含了活动的名称、类型、时间、预算和规则等。
type Campaign struct {
	gorm.Model                  // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name         string         `gorm:"type:varchar(128);not null;comment:活动名称" json:"name"`         // 活动名称，不允许为空。
	CampaignType CampaignType   `gorm:"type:varchar(32);not null;comment:活动类型" json:"campaign_type"` // 活动类型。
	Description  string         `gorm:"type:text;comment:活动描述" json:"description"`                   // 活动描述。
	StartTime    time.Time      `gorm:"not null;comment:开始时间" json:"start_time"`                     // 活动开始时间。
	EndTime      time.Time      `gorm:"not null;comment:结束时间" json:"end_time"`                       // 活动结束时间。
	Budget       uint64         `gorm:"not null;default:0;comment:预算" json:"budget"`                 // 活动总预算。
	Spent        uint64         `gorm:"not null;default:0;comment:已花费" json:"spent"`                 // 活动已花费金额。
	TargetUsers  int64          `gorm:"not null;default:0;comment:目标用户数" json:"target_users"`        // 目标触达用户数。
	ReachedUsers int64          `gorm:"not null;default:0;comment:触达用户数" json:"reached_users"`       // 实际触达用户数。
	Status       CampaignStatus `gorm:"default:0;comment:状态" json:"status"`                          // 活动状态，默认为未开始。
	Rules        JSONMap        `gorm:"type:json;comment:规则配置" json:"rules"`                         // 活动规则配置，存储为JSON。
}

// NewCampaign 创建并返回一个新的 Campaign 实体实例。
// name: 活动名称。
// campaignType: 活动类型。
// description: 描述。
// startTime, endTime: 活动时间。
// budget: 预算。
// rules: 规则配置。
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
		Status:       CampaignStatusPending, // 默认状态为未开始。
		Rules:        JSONMap(rules),        // 初始化规则配置。
	}
}

// IsActive 检查营销活动当前是否处于活跃状态。
func (c *Campaign) IsActive() bool {
	now := time.Now()
	return c.Status == CampaignStatusOngoing && // 状态为进行中。
		now.After(c.StartTime) && // 当前时间在开始时间之后。
		now.Before(c.EndTime) // 当前时间在结束时间之前。
}

// Start 启动营销活动，将其状态设置为“进行中”。
func (c *Campaign) Start() {
	c.Status = CampaignStatusOngoing
}

// End 结束营销活动，将其状态设置为“已结束”。
func (c *Campaign) End() {
	c.Status = CampaignStatusEnded
}

// Cancel 取消营销活动，将其状态设置为“已取消”。
func (c *Campaign) Cancel() {
	c.Status = CampaignStatusCanceled
}

// AddSpent 增加营销活动的已花费金额。
// amount: 增加的金额。
func (c *Campaign) AddSpent(amount uint64) {
	c.Spent += amount
}

// IncrementReachedUsers 增加营销活动的触达用户数。
func (c *Campaign) IncrementReachedUsers() {
	c.ReachedUsers++
}

// RemainingBudget 计算营销活动的剩余预算。
func (c *Campaign) RemainingBudget() uint64 {
	if c.Spent >= c.Budget {
		return 0
	}
	return c.Budget - c.Spent
}

// CampaignParticipation 实体代表用户参与营销活动的记录。
type CampaignParticipation struct {
	gorm.Model        // 嵌入gorm.Model。
	CampaignID uint64 `gorm:"not null;index;comment:活动ID" json:"campaign_id"`  // 关联的营销活动ID，索引字段。
	UserID     uint64 `gorm:"not null;index;comment:用户ID" json:"user_id"`      // 参与用户的ID，索引字段。
	OrderID    uint64 `gorm:"index;comment:订单ID" json:"order_id"`              // 关联的订单ID（如果参与与订单相关），索引字段。
	Discount   uint64 `gorm:"not null;default:0;comment:优惠金额" json:"discount"` // 用户因参与活动获得的优惠金额。
}

// NewCampaignParticipation 创建并返回一个新的 CampaignParticipation 实体实例。
// campaignID: 营销活动ID。
// userID: 用户ID。
// orderID: 订单ID。
// discount: 优惠金额。
func NewCampaignParticipation(campaignID, userID, orderID, discount uint64) *CampaignParticipation {
	return &CampaignParticipation{
		CampaignID: campaignID,
		UserID:     userID,
		OrderID:    orderID,
		Discount:   discount,
	}
}

// Banner 实体代表一个广告横幅。
// 它通常用于网站或App的显眼位置，吸引用户点击。
type Banner struct {
	gorm.Model           // 嵌入gorm.Model。
	Title      string    `gorm:"type:varchar(128);not null;comment:标题" json:"title"`        // Banner标题。
	ImageURL   string    `gorm:"type:varchar(255);not null;comment:图片URL" json:"image_url"` // Banner图片URL。
	LinkURL    string    `gorm:"type:varchar(255);comment:跳转URL" json:"link_url"`           // 点击Banner后跳转的链接。
	Position   string    `gorm:"type:varchar(32);not null;comment:位置" json:"position"`      // Banner展示的位置（例如，“首页顶部”，“商品详情页”）。
	Priority   int32     `gorm:"default:0;comment:优先级" json:"priority"`                     // Banner的展示优先级。
	StartTime  time.Time `gorm:"not null;comment:开始时间" json:"start_time"`                   // Banner的展示开始时间。
	EndTime    time.Time `gorm:"not null;comment:结束时间" json:"end_time"`                     // Banner的展示结束时间。
	ClickCount int64     `gorm:"default:0;comment:点击数" json:"click_count"`                  // Banner的点击次数。
	Enabled    bool      `gorm:"default:true;comment:是否启用" json:"enabled"`                  // Banner是否启用。
}

// NewBanner 创建并返回一个新的 Banner 实体实例。
// title: 标题。
// imageURL: 图片URL。
// linkURL: 跳转URL。
// position: 位置。
// priority: 优先级。
// startTime, endTime: 展示时间。
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
		Enabled:    true, // 默认启用。
	}
}

// IsActive 检查Banner当前是否处于活跃展示状态。
func (b *Banner) IsActive() bool {
	now := time.Now()
	return b.Enabled && // 确保Banner已启用。
		now.After(b.StartTime) && // 当前时间在开始时间之后。
		now.Before(b.EndTime) // 当前时间在结束时间之前。
}

// IncrementClick 增加Banner的点击次数。
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
