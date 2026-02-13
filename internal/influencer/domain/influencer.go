// 变更说明：完善KOL/网红营销领域模型，增加推广链接、佣金结算、效果追踪等完整功能
package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InfluencerStatus 网红状态
type InfluencerStatus string

const (
	StatusPending   InfluencerStatus = "PENDING"
	StatusApproved  InfluencerStatus = "APPROVED"
	StatusRejected  InfluencerStatus = "REJECTED"
	StatusSuspended InfluencerStatus = "SUSPENDED"
)

// Influencer 网红/KOL 聚合根
type Influencer struct {
	gorm.Model
	InfluencerID    string           `gorm:"column:influencer_id;type:varchar(32);uniqueIndex;not null" json:"influencer_id"`
	UserID          string           `gorm:"column:user_id;type:varchar(32);uniqueIndex;not null" json:"user_id"`
	Name            string           `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Avatar          string           `gorm:"column:avatar;type:varchar(512)" json:"avatar"`
	Bio             string           `gorm:"column:bio;type:text" json:"bio"`
	Platforms       []PlatformInfo   `gorm:"foreignKey:InfluencerID;references:InfluencerID" json:"platforms"`
	Status          InfluencerStatus `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	Level           string           `gorm:"column:level;type:varchar(20);not null;default:'BRONZE'" json:"level"`
	TotalFollowers  int64            `gorm:"column:total_followers;default:0" json:"total_followers"`
	TotalEarnings   decimal.Decimal  `gorm:"column:total_earnings;type:decimal(20,4);default:0" json:"total_earnings"`
	PendingEarnings decimal.Decimal  `gorm:"column:pending_earnings;type:decimal(20,4);default:0" json:"pending_earnings"`
	WithdrawnAmount decimal.Decimal  `gorm:"column:withdrawn_amount;type:decimal(20,4);default:0" json:"withdrawn_amount"`
	TotalOrders     int64            `gorm:"column:total_orders;default:0" json:"total_orders"`
	TotalSales      decimal.Decimal  `gorm:"column:total_sales;type:decimal(20,4);default:0" json:"total_sales"`
	ConversionRate  decimal.Decimal  `gorm:"column:conversion_rate;type:decimal(10,6);default:0" json:"conversion_rate"`
	BankAccount     string           `gorm:"column:bank_account;type:varchar(64)" json:"-"`
	BankName        string           `gorm:"column:bank_name;type:varchar(100)" json:"-"`
	ContactEmail    string           `gorm:"column:contact_email;type:varchar(128)" json:"contact_email"`
	ContactPhone    string           `gorm:"column:contact_phone;type:varchar(20)" json:"contact_phone"`
	ApprovedAt      *time.Time       `gorm:"column:approved_at" json:"approved_at"`
}

// PlatformInfo 平台信息
type PlatformInfo struct {
	gorm.Model
	InfluencerID  string          `gorm:"column:influencer_id;type:varchar(32);index;not null" json:"influencer_id"`
	Platform      string          `gorm:"column:platform;type:varchar(50);not null" json:"platform"`
	Handle        string          `gorm:"column:handle;type:varchar(100);not null" json:"handle"`
	ProfileURL    string          `gorm:"column:profile_url;type:varchar(512)" json:"profile_url"`
	FollowerCount int64           `gorm:"column:follower_count;default:0" json:"follower_count"`
	AvgViews      int64           `gorm:"column:avg_views;default:0" json:"avg_views"`
	AvgLikes      int64           `gorm:"column:avg_likes;default:0" json:"avg_likes"`
	AvgComments   int64           `gorm:"column:avg_comments;default:0" json:"avg_comments"`
	EngagementRate decimal.Decimal `gorm:"column:engagement_rate;type:decimal(10,6);default:0" json:"engagement_rate"`
	Verified      bool            `gorm:"column:verified;default:false" json:"verified"`
}

// Campaign 推广活动
type Campaign struct {
	gorm.Model
	CampaignID       string          `gorm:"column:campaign_id;type:varchar(32);uniqueIndex;not null" json:"campaign_id"`
	InfluencerID     string          `gorm:"column:influencer_id;type:varchar(32);index;not null" json:"influencer_id"`
	MerchantID       string          `gorm:"column:merchant_id;type:varchar(32);index;not null" json:"merchant_id"`
	ProductID        string          `gorm:"column:product_id;type:varchar(32);index;not null" json:"product_id"`
	ProductName      string          `gorm:"column:product_name;type:varchar(255)" json:"product_name"`
	Title            string          `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description      string          `gorm:"column:description;type:text" json:"description"`
	CommissionType   string          `gorm:"column:commission_type;type:varchar(20);not null" json:"commission_type"`
	CommissionRate   decimal.Decimal `gorm:"column:commission_rate;type:decimal(10,4);not null" json:"commission_rate"`
	FixedCommission  decimal.Decimal `gorm:"column:fixed_commission;type:decimal(20,4)" json:"fixed_commission"`
	TargetSales      decimal.Decimal `gorm:"column:target_sales;type:decimal(20,4)" json:"target_sales"`
	ActualSales      decimal.Decimal `gorm:"column:actual_sales;type:decimal(20,4);default:0" json:"actual_sales"`
	TotalOrders      int64           `gorm:"column:total_orders;default:0" json:"total_orders"`
	TotalClicks      int64           `gorm:"column:total_clicks;default:0" json:"total_clicks"`
	TotalConversions int64           `gorm:"column:total_conversions;default:0" json:"total_conversions"`
	Status           string          `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	StartAt          time.Time       `gorm:"column:start_at;not null" json:"start_at"`
	EndAt            time.Time       `gorm:"column:end_at;not null" json:"end_at"`
	ApprovedAt       *time.Time      `gorm:"column:approved_at" json:"approved_at"`
}

// PromotionLink 推广链接
type PromotionLink struct {
	gorm.Model
	LinkID       string    `gorm:"column:link_id;type:varchar(32);uniqueIndex;not null" json:"link_id"`
	InfluencerID string    `gorm:"column:influencer_id;type:varchar(32);index;not null" json:"influencer_id"`
	CampaignID   string    `gorm:"column:campaign_id;type:varchar(32);index;not null" json:"campaign_id"`
	ProductID    string    `gorm:"column:product_id;type:varchar(32);index;not null" json:"product_id"`
	ShortCode    string    `gorm:"column:short_code;type:varchar(16);uniqueIndex;not null" json:"short_code"`
	OriginalURL  string    `gorm:"column:original_url;type:varchar(512);not null" json:"original_url"`
	ShortURL     string    `gorm:"column:short_url;type:varchar(128);not null" json:"short_url"`
	Platform     string    `gorm:"column:platform;type:varchar(50)" json:"platform"`
	TotalClicks  int64     `gorm:"column:total_clicks;default:0" json:"total_clicks"`
	UniqueClicks int64     `gorm:"column:unique_clicks;default:0" json:"unique_clicks"`
	ExpiresAt    *time.Time `gorm:"column:expires_at" json:"expires_at"`
}

// CommissionRecord 佣金记录
type CommissionRecord struct {
	gorm.Model
	RecordID       string          `gorm:"column:record_id;type:varchar(32);uniqueIndex;not null" json:"record_id"`
	InfluencerID   string          `gorm:"column:influencer_id;type:varchar(32);index;not null" json:"influencer_id"`
	CampaignID     string          `gorm:"column:campaign_id;type:varchar(32);index;not null" json:"campaign_id"`
	OrderID        string          `gorm:"column:order_id;type:varchar(32);index;not null" json:"order_id"`
	LinkID         string          `gorm:"column:link_id;type:varchar(32)" json:"link_id"`
	OrderAmount    decimal.Decimal `gorm:"column:order_amount;type:decimal(20,4);not null" json:"order_amount"`
	CommissionRate decimal.Decimal `gorm:"column:commission_rate;type:decimal(10,4);not null" json:"commission_rate"`
	Commission     decimal.Decimal `gorm:"column:commission;type:decimal(20,4);not null" json:"commission"`
	Status         string          `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	SettledAt      *time.Time      `gorm:"column:settled_at" json:"settled_at"`
}

// Withdrawal 提现记录
type Withdrawal struct {
	gorm.Model
	WithdrawalID  string          `gorm:"column:withdrawal_id;type:varchar(32);uniqueIndex;not null" json:"withdrawal_id"`
	InfluencerID  string          `gorm:"column:influencer_id;type:varchar(32);index;not null" json:"influencer_id"`
	Amount        decimal.Decimal `gorm:"column:amount;type:decimal(20,4);not null" json:"amount"`
	Status        string          `gorm:"column:status;type:varchar(20);not null;default:'PENDING'" json:"status"`
	BankAccount   string          `gorm:"column:bank_account;type:varchar(64)" json:"bank_account"`
	BankName      string          `gorm:"column:bank_name;type:varchar(100)" json:"bank_name"`
	TransactionID string          `gorm:"column:transaction_id;type:varchar(64)" json:"transaction_id"`
	ProcessedAt   *time.Time      `gorm:"column:processed_at" json:"processed_at"`
	RejectedAt    *time.Time      `gorm:"column:rejected_at" json:"rejected_at"`
	RejectReason  string          `gorm:"column:reject_reason;type:text" json:"reject_reason"`
}

// ClickRecord 点击记录
type ClickRecord struct {
	gorm.Model
	LinkID     string    `gorm:"column:link_id;type:varchar(32);index;not null" json:"link_id"`
	UserID     string    `gorm:"column:user_id;type:varchar(32);index" json:"user_id"`
	IPAddress  string    `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`
	UserAgent  string    `gorm:"column:user_agent;type:varchar(512)" json:"user_agent"`
	Platform   string    `gorm:"column:platform;type:varchar(50)" json:"platform"`
	Referrer   string    `gorm:"column:referrer;type:varchar(512)" json:"referrer"`
	Converted  bool      `gorm:"column:converted;default:false" json:"converted"`
	ClickTime  time.Time `gorm:"column:click_time;not null" json:"click_time"`
}

// InfluencerStats 网红统计
type InfluencerStats struct {
	InfluencerID       string          `json:"influencer_id"`
	TotalCampaigns     int64           `json:"total_campaigns"`
	ActiveCampaigns    int64           `json:"active_campaigns"`
	TotalClicks        int64           `json:"total_clicks"`
	TotalConversions   int64           `json:"total_conversions"`
	TotalSales         decimal.Decimal `json:"total_sales"`
	TotalCommission    decimal.Decimal `json:"total_commission"`
	AvgConversionRate  decimal.Decimal `json:"avg_conversion_rate"`
	AvgOrderValue      decimal.Decimal `json:"avg_order_value"`
	TopPerformingPlatform string        `json:"top_performing_platform"`
}

func (Influencer) TableName() string     { return "influencers" }
func (PlatformInfo) TableName() string   { return "influencer_platforms" }
func (Campaign) TableName() string       { return "influencer_campaigns" }
func (PromotionLink) TableName() string  { return "influencer_promotion_links" }
func (CommissionRecord) TableName() string { return "influencer_commission_records" }
func (Withdrawal) TableName() string     { return "influencer_withdrawals" }
func (ClickRecord) TableName() string    { return "influencer_click_records" }

// NewInfluencer 创建新网红
func NewInfluencer(userID, name, email string) *Influencer {
	return &Influencer{
		InfluencerID:    generateInfluencerID(),
		UserID:          userID,
		Name:            name,
		ContactEmail:    email,
		Status:          StatusPending,
		Level:           "BRONZE",
		TotalEarnings:   decimal.Zero,
		PendingEarnings: decimal.Zero,
		Platforms:       []PlatformInfo{},
	}
}

// AddPlatform 添加平台
func (i *Influencer) AddPlatform(platform, handle, profileURL string, followers int64) *PlatformInfo {
	info := &PlatformInfo{
		InfluencerID:  i.InfluencerID,
		Platform:      platform,
		Handle:        handle,
		ProfileURL:    profileURL,
		FollowerCount: followers,
	}
	i.Platforms = append(i.Platforms, *info)
	i.TotalFollowers += followers
	return info
}

// Approve 批准网红
func (i *Influencer) Approve() error {
	if i.Status != StatusPending {
		return errors.New("influencer is not pending")
	}
	i.Status = StatusApproved
	now := time.Now()
	i.ApprovedAt = &now
	return nil
}

// Reject 拒绝网红
func (i *Influencer) Reject() error {
	if i.Status != StatusPending {
		return errors.New("influencer is not pending")
	}
	i.Status = StatusRejected
	return nil
}

// Suspend 暂停网红
func (i *Influencer) Suspend() {
	i.Status = StatusSuspended
}

// AddEarnings 增加收益
func (i *Influencer) AddEarnings(amount decimal.Decimal) {
	i.PendingEarnings = i.PendingEarnings.Add(amount)
}

// SettleCommission 结算佣金
func (i *Influencer) SettleCommission(amount decimal.Decimal) {
	i.PendingEarnings = i.PendingEarnings.Sub(amount)
	i.TotalEarnings = i.TotalEarnings.Add(amount)
}

// UpdateLevel 更新等级
func (i *Influencer) UpdateLevel() {
	totalEarnings := i.TotalEarnings
	
	switch {
	case totalEarnings.GreaterThanOrEqual(decimal.NewFromInt(100000)):
		i.Level = "DIAMOND"
	case totalEarnings.GreaterThanOrEqual(decimal.NewFromInt(50000)):
		i.Level = "PLATINUM"
	case totalEarnings.GreaterThanOrEqual(decimal.NewFromInt(10000)):
		i.Level = "GOLD"
	case totalEarnings.GreaterThanOrEqual(decimal.NewFromInt(1000)):
		i.Level = "SILVER"
	default:
		i.Level = "BRONZE"
	}
}

// NewCampaign 创建推广活动
func NewCampaign(influencerID, merchantID, productID, title string, commissionType string, commissionRate decimal.Decimal, startAt, endAt time.Time) *Campaign {
	return &Campaign{
		CampaignID:     generateCampaignID(),
		InfluencerID:   influencerID,
		MerchantID:     merchantID,
		ProductID:      productID,
		Title:          title,
		CommissionType: commissionType,
		CommissionRate: commissionRate,
		StartAt:        startAt,
		EndAt:          endAt,
		Status:         "PENDING",
	}
}

// ApproveCampaign 批准活动
func (c *Campaign) ApproveCampaign() error {
	if c.Status != "PENDING" {
		return errors.New("campaign is not pending")
	}
	c.Status = "ACTIVE"
	now := time.Now()
	c.ApprovedAt = &now
	return nil
}

// RecordClick 记录点击
func (c *Campaign) RecordClick() {
	c.TotalClicks++
}

// RecordConversion 记录转化
func (c *Campaign) RecordConversion(orderAmount decimal.Decimal) {
	c.TotalConversions++
	c.TotalOrders++
	c.ActualSales = c.ActualSales.Add(orderAmount)
}

// CalculateCommission 计算佣金
func (c *Campaign) CalculateCommission(orderAmount decimal.Decimal) decimal.Decimal {
	if c.CommissionType == "PERCENTAGE" {
		return orderAmount.Mul(c.CommissionRate).Div(decimal.NewFromInt(100))
	}
	return c.FixedCommission
}

// NewPromotionLink 创建推广链接
func NewPromotionLink(influencerID, campaignID, productID, originalURL string) *PromotionLink {
	shortCode := generateShortCode()
	return &PromotionLink{
		LinkID:       generateLinkID(),
		InfluencerID: influencerID,
		CampaignID:   campaignID,
		ProductID:    productID,
		ShortCode:    shortCode,
		OriginalURL:  originalURL,
		ShortURL:     fmt.Sprintf("https://sp.shop/%s", shortCode),
	}
}

// RecordClick 记录链接点击
func (l *PromotionLink) RecordClick(unique bool) {
	l.TotalClicks++
	if unique {
		l.UniqueClicks++
	}
}

// NewCommissionRecord 创建佣金记录
func NewCommissionRecord(influencerID, campaignID, orderID, linkID string, orderAmount, commissionRate, commission decimal.Decimal) *CommissionRecord {
	return &CommissionRecord{
		RecordID:       generateRecordID(),
		InfluencerID:   influencerID,
		CampaignID:     campaignID,
		OrderID:        orderID,
		LinkID:         linkID,
		OrderAmount:    orderAmount,
		CommissionRate: commissionRate,
		Commission:     commission,
		Status:         "PENDING",
	}
}

// Settle 结算佣金
func (r *CommissionRecord) Settle() {
	r.Status = "SETTLED"
	now := time.Now()
	r.SettledAt = &now
}

// NewWithdrawal 创建提现
func NewWithdrawal(influencerID string, amount decimal.Decimal, bankAccount, bankName string) *Withdrawal {
	return &Withdrawal{
		WithdrawalID: generateWithdrawalID(),
		InfluencerID: influencerID,
		Amount:       amount,
		BankAccount:  bankAccount,
		BankName:     bankName,
		Status:       "PENDING",
	}
}

// Process 处理提现
func (w *Withdrawal) Process(transactionID string) {
	w.Status = "PROCESSED"
	w.TransactionID = transactionID
	now := time.Now()
	w.ProcessedAt = &now
}

// RejectWithdrawal 拒绝提现
func (w *Withdrawal) RejectWithdrawal(reason string) {
	w.Status = "REJECTED"
	w.RejectReason = reason
	now := time.Now()
	w.RejectedAt = &now
}

// 辅助函数
func generateInfluencerID() string {
	return fmt.Sprintf("INF%d", time.Now().UnixNano())
}

func generateCampaignID() string {
	return fmt.Sprintf("CP%d", time.Now().UnixNano())
}

func generateLinkID() string {
	return fmt.Sprintf("LK%d", time.Now().UnixNano())
}

func generateShortCode() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())[:8]
}

func generateRecordID() string {
	return fmt.Sprintf("CR%d", time.Now().UnixNano())
}

func generateWithdrawalID() string {
	return fmt.Sprintf("WD%d", time.Now().UnixNano())
}

// 错误定义
var (
	ErrInfluencerNotFound   = errors.New("influencer not found")
	ErrCampaignNotFound     = errors.New("campaign not found")
	ErrLinkNotFound         = errors.New("promotion link not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrInvalidStatus        = errors.New("invalid status")
)
