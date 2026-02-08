// Package domain 商家服务领域层
// 生成摘要：
// 1) 定义商家聚合根和实体
// 2) 定义商家状态机
// 3) 定义领域事件
// 假设：
// - 商家编号格式：M + 时间戳 + 随机数
// - 信用评分范围：0-100
package domain

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MerchantStatus 商家状态
type MerchantStatus int8

const (
	// MerchantStatusPending 待审核
	MerchantStatusPending MerchantStatus = 1
	// MerchantStatusApproved 已通过
	MerchantStatusApproved MerchantStatus = 2
	// MerchantStatusRejected 已拒绝
	MerchantStatusRejected MerchantStatus = 3
	// MerchantStatusDisabled 已禁用
	MerchantStatusDisabled MerchantStatus = 4
	// MerchantStatusSuspended 已暂停
	MerchantStatusSuspended MerchantStatus = 5
)

// String 返回状态的字符串表示
func (s MerchantStatus) String() string {
	switch s {
	case MerchantStatusPending:
		return "pending"
	case MerchantStatusApproved:
		return "approved"
	case MerchantStatusRejected:
		return "rejected"
	case MerchantStatusDisabled:
		return "disabled"
	case MerchantStatusSuspended:
		return "suspended"
	default:
		return "unknown"
	}
}

// MerchantType 商家类型
type MerchantType int8

const (
	// MerchantTypeIndividual 个人
	MerchantTypeIndividual MerchantType = 1
	// MerchantTypeEnterprise 企业
	MerchantTypeEnterprise MerchantType = 2
	// MerchantTypeBrand 品牌旗舰店
	MerchantTypeBrand MerchantType = 3
)

// MerchantLevel 商家等级
type MerchantLevel int8

const (
	// MerchantLevelBronze 青铜
	MerchantLevelBronze MerchantLevel = 1
	// MerchantLevelSilver 白银
	MerchantLevelSilver MerchantLevel = 2
	// MerchantLevelGold 黄金
	MerchantLevelGold MerchantLevel = 3
	// MerchantLevelPlatinum 铂金
	MerchantLevelPlatinum MerchantLevel = 4
	// MerchantLevelDiamond 钻石
	MerchantLevelDiamond MerchantLevel = 5
)

// Merchant 商家实体（聚合根）
type Merchant struct {
	gorm.Model
	// UserID 关联用户ID，商家必须先有用户账号
	UserID uint64 `gorm:"column:user_id;type:bigint unsigned;uniqueIndex;not null" json:"user_id"`
	// MerchantNo 商家编号，系统自动生成的唯一标识
	MerchantNo string `gorm:"column:merchant_no;type:varchar(32);uniqueIndex;not null" json:"merchant_no"`
	// Name 商家名称，对外展示的商家名
	Name string `gorm:"column:name;type:varchar(128);not null" json:"name"`
	// LegalName 法人姓名
	LegalName string `gorm:"column:legal_name;type:varchar(64);not null" json:"legal_name"`
	// LegalIDCard 法人身份证号（加密存储）
	LegalIDCard string `gorm:"column:legal_id_card;type:varchar(128);not null" json:"legal_id_card"`
	// ContactName 联系人姓名
	ContactName string `gorm:"column:contact_name;type:varchar(64);not null" json:"contact_name"`
	// ContactPhone 联系人电话
	ContactPhone string `gorm:"column:contact_phone;type:varchar(20);not null" json:"contact_phone"`
	// ContactEmail 联系人邮箱
	ContactEmail string `gorm:"column:contact_email;type:varchar(128)" json:"contact_email"`
	// Type 商家类型
	Type MerchantType `gorm:"column:type;type:tinyint;not null;default:1" json:"type"`
	// Status 商家状态
	Status MerchantStatus `gorm:"column:status;type:tinyint;not null;default:1;index" json:"status"`
	// Level 商家等级
	Level MerchantLevel `gorm:"column:level;type:tinyint;not null;default:1" json:"level"`
	// LogoURL 商家logo地址
	LogoURL string `gorm:"column:logo_url;type:varchar(512)" json:"logo_url"`
	// Description 商家简介
	Description string `gorm:"column:description;type:text" json:"description"`
	// CommissionRate 佣金费率（百分比，如 0.05 表示 5%）
	CommissionRate float64 `gorm:"column:commission_rate;type:decimal(5,4);not null;default:0.05" json:"commission_rate"`
	// CreditScore 信用评分（0-100）
	CreditScore float64 `gorm:"column:credit_score;type:decimal(5,2);not null;default:60" json:"credit_score"`
	// TotalSales 总销售额（分）
	TotalSales int64 `gorm:"column:total_sales;type:bigint;not null;default:0" json:"total_sales"`
	// TotalOrders 总订单数
	TotalOrders int32 `gorm:"column:total_orders;type:int;not null;default:0" json:"total_orders"`
	// Rating 综合评分（1-5）
	Rating float64 `gorm:"column:rating;type:decimal(2,1);not null;default:5.0" json:"rating"`
	// RejectReason 拒绝原因
	RejectReason string `gorm:"column:reject_reason;type:varchar(512)" json:"reject_reason"`
	// ApprovedAt 审核通过时间
	ApprovedAt *time.Time `gorm:"column:approved_at" json:"approved_at"`
	// 关联实体
	BusinessLicense *BusinessLicense `gorm:"foreignKey:MerchantID" json:"business_license,omitempty"`
	BankAccount     *BankAccount     `gorm:"foreignKey:MerchantID" json:"bank_account,omitempty"`
	Stores          []*Store         `gorm:"foreignKey:MerchantID" json:"stores,omitempty"`
	// 领域事件
	domainEvents []DomainEvent `gorm:"-" json:"-"`
}

// TableName 指定表名
func (Merchant) TableName() string {
	return "merchants"
}

// BusinessLicense 营业执照（值对象，独立存储）
type BusinessLicense struct {
	gorm.Model
	// MerchantID 商家ID
	MerchantID uint `gorm:"column:merchant_id;type:int unsigned;uniqueIndex;not null" json:"merchant_id"`
	// LicenseNo 执照编号
	LicenseNo string `gorm:"column:license_no;type:varchar(64);uniqueIndex;not null" json:"license_no"`
	// CompanyName 公司名称
	CompanyName string `gorm:"column:company_name;type:varchar(128);not null" json:"company_name"`
	// CompanyType 公司类型
	CompanyType string `gorm:"column:company_type;type:varchar(64)" json:"company_type"`
	// RegisteredCapital 注册资本
	RegisteredCapital string `gorm:"column:registered_capital;type:varchar(32)" json:"registered_capital"`
	// EstablishmentDate 成立日期
	EstablishmentDate string `gorm:"column:establishment_date;type:varchar(16)" json:"establishment_date"`
	// ValidUntil 有效期至
	ValidUntil string `gorm:"column:valid_until;type:varchar(16)" json:"valid_until"`
	// BusinessScope 经营范围
	BusinessScope string `gorm:"column:business_scope;type:text" json:"business_scope"`
	// RegisteredAddress 注册地址
	RegisteredAddress string `gorm:"column:registered_address;type:varchar(256)" json:"registered_address"`
	// LicenseImageURL 执照图片URL
	LicenseImageURL string `gorm:"column:license_image_url;type:varchar(512);not null" json:"license_image_url"`
}

// TableName 指定表名
func (BusinessLicense) TableName() string {
	return "merchant_business_licenses"
}

// BankAccount 银行账户（值对象，独立存储）
type BankAccount struct {
	gorm.Model
	// MerchantID 商家ID
	MerchantID uint `gorm:"column:merchant_id;type:int unsigned;uniqueIndex;not null" json:"merchant_id"`
	// AccountName 账户名称
	AccountName string `gorm:"column:account_name;type:varchar(128);not null" json:"account_name"`
	// AccountNo 账户号码（加密存储）
	AccountNo string `gorm:"column:account_no;type:varchar(128);not null" json:"account_no"`
	// BankName 银行名称
	BankName string `gorm:"column:bank_name;type:varchar(64);not null" json:"bank_name"`
	// BankBranch 开户支行
	BankBranch string `gorm:"column:bank_branch;type:varchar(128)" json:"bank_branch"`
	// BankCode 银行代码
	BankCode string `gorm:"column:bank_code;type:varchar(16)" json:"bank_code"`
}

// TableName 指定表名
func (BankAccount) TableName() string {
	return "merchant_bank_accounts"
}

// Store 店铺实体
type Store struct {
	gorm.Model
	// MerchantID 商家ID
	MerchantID uint `gorm:"column:merchant_id;type:int unsigned;index;not null" json:"merchant_id"`
	// StoreNo 店铺编号
	StoreNo string `gorm:"column:store_no;type:varchar(32);uniqueIndex;not null" json:"store_no"`
	// Name 店铺名称
	Name string `gorm:"column:name;type:varchar(128);not null" json:"name"`
	// LogoURL 店铺logo
	LogoURL string `gorm:"column:logo_url;type:varchar(512)" json:"logo_url"`
	// BannerURL 店铺banner
	BannerURL string `gorm:"column:banner_url;type:varchar(512)" json:"banner_url"`
	// Description 店铺简介
	Description string `gorm:"column:description;type:text" json:"description"`
	// Announcement 店铺公告
	Announcement string `gorm:"column:announcement;type:text" json:"announcement"`
	// Categories 经营类目（JSON数组）
	Categories string `gorm:"column:categories;type:text" json:"categories"`
	// Rating 店铺评分
	Rating float64 `gorm:"column:rating;type:decimal(2,1);not null;default:5.0" json:"rating"`
	// ProductCount 商品数量
	ProductCount int32 `gorm:"column:product_count;type:int;not null;default:0" json:"product_count"`
	// MonthlySales 月销售额（分）
	MonthlySales int64 `gorm:"column:monthly_sales;type:bigint;not null;default:0" json:"monthly_sales"`
	// IsOpen 是否营业
	IsOpen bool `gorm:"column:is_open;type:tinyint(1);not null;default:1" json:"is_open"`
	// BusinessHours 营业时间
	BusinessHours string `gorm:"column:business_hours;type:varchar(128)" json:"business_hours"`
	// Address 店铺地址
	Address string `gorm:"column:address;type:varchar(256)" json:"address"`
}

// TableName 指定表名
func (Store) TableName() string {
	return "merchant_stores"
}

// MerchantSettings 商家设置
type MerchantSettings struct {
	gorm.Model
	// MerchantID 商家ID
	MerchantID uint `gorm:"column:merchant_id;type:int unsigned;uniqueIndex;not null" json:"merchant_id"`
	// AutoConfirmOrder 自动确认订单
	AutoConfirmOrder bool `gorm:"column:auto_confirm_order;type:tinyint(1);not null;default:0" json:"auto_confirm_order"`
	// AutoConfirmDays 自动确认天数
	AutoConfirmDays int32 `gorm:"column:auto_confirm_days;type:int;not null;default:7" json:"auto_confirm_days"`
	// EnableCOD 支持货到付款
	EnableCOD bool `gorm:"column:enable_cod;type:tinyint(1);not null;default:0" json:"enable_cod"`
	// EnableInvoice 支持开发票
	EnableInvoice bool `gorm:"column:enable_invoice;type:tinyint(1);not null;default:1" json:"enable_invoice"`
	// SupportedPayments 支持的支付方式（JSON数组）
	SupportedPayments string `gorm:"column:supported_payments;type:text" json:"supported_payments"`
	// FreeShippingThreshold 包邮门槛（分）
	FreeShippingThreshold int64 `gorm:"column:free_shipping_threshold;type:bigint;not null;default:0" json:"free_shipping_threshold"`
	// DefaultShippingFee 默认运费（分）
	DefaultShippingFee int64 `gorm:"column:default_shipping_fee;type:bigint;not null;default:0" json:"default_shipping_fee"`
	// EnableSameDayDelivery 支持当日达
	EnableSameDayDelivery bool `gorm:"column:enable_same_day_delivery;type:tinyint(1);not null;default:0" json:"enable_same_day_delivery"`
	// OrderNotification 订单通知
	OrderNotification bool `gorm:"column:order_notification;type:tinyint(1);not null;default:1" json:"order_notification"`
	// StockAlert 库存预警
	StockAlert bool `gorm:"column:stock_alert;type:tinyint(1);not null;default:1" json:"stock_alert"`
	// ReviewNotification 评价通知
	ReviewNotification bool `gorm:"column:review_notification;type:tinyint(1);not null;default:1" json:"review_notification"`
	// SettlementNotification 结算通知
	SettlementNotification bool `gorm:"column:settlement_notification;type:tinyint(1);not null;default:1" json:"settlement_notification"`
	// NotificationPhone 通知手机号
	NotificationPhone string `gorm:"column:notification_phone;type:varchar(20)" json:"notification_phone"`
	// NotificationEmail 通知邮箱
	NotificationEmail string `gorm:"column:notification_email;type:varchar(128)" json:"notification_email"`
}

// TableName 指定表名
func (MerchantSettings) TableName() string {
	return "merchant_settings"
}

// NewMerchant 创建商家实例
func NewMerchant(userID uint64, name, legalName, legalIDCard, contactName, contactPhone, contactEmail string, merchantType MerchantType) *Merchant {
	merchantNo := fmt.Sprintf("M%d%d", time.Now().UnixNano()/1000000, userID%1000)

	m := &Merchant{
		UserID:         userID,
		MerchantNo:     merchantNo,
		Name:           name,
		LegalName:      legalName,
		LegalIDCard:    legalIDCard,
		ContactName:    contactName,
		ContactPhone:   contactPhone,
		ContactEmail:   contactEmail,
		Type:           merchantType,
		Status:         MerchantStatusPending,
		Level:          MerchantLevelBronze,
		CommissionRate: 0.05, // 默认5%佣金
		CreditScore:    60,   // 默认信用分
		Rating:         5.0,
		domainEvents:   make([]DomainEvent, 0),
	}

	m.AddDomainEvent(&MerchantAppliedEvent{
		MerchantID:   0, // 待保存后更新
		MerchantNo:   merchantNo,
		UserID:       userID,
		Name:         name,
		MerchantType: merchantType,
		Timestamp:    time.Now(),
	})

	return m
}

// Approve 审核通过
func (m *Merchant) Approve(commissionRate float64, operator, remark string) error {
	if m.Status != MerchantStatusPending {
		return fmt.Errorf("merchant status is not pending, current: %s", m.Status.String())
	}

	m.Status = MerchantStatusApproved
	m.CommissionRate = commissionRate
	now := time.Now()
	m.ApprovedAt = &now

	m.AddDomainEvent(&MerchantApprovedEvent{
		MerchantID:     uint64(m.ID),
		MerchantNo:     m.MerchantNo,
		CommissionRate: commissionRate,
		Operator:       operator,
		Remark:         remark,
		Timestamp:      now,
	})

	return nil
}

// Reject 审核拒绝
func (m *Merchant) Reject(reason, operator string) error {
	if m.Status != MerchantStatusPending {
		return fmt.Errorf("merchant status is not pending, current: %s", m.Status.String())
	}

	m.Status = MerchantStatusRejected
	m.RejectReason = reason

	m.AddDomainEvent(&MerchantRejectedEvent{
		MerchantID: uint64(m.ID),
		MerchantNo: m.MerchantNo,
		Reason:     reason,
		Operator:   operator,
		Timestamp:  time.Now(),
	})

	return nil
}

// Disable 禁用商家
func (m *Merchant) Disable(reason, operator string) error {
	if m.Status == MerchantStatusDisabled {
		return fmt.Errorf("merchant is already disabled")
	}

	m.Status = MerchantStatusDisabled

	m.AddDomainEvent(&MerchantDisabledEvent{
		MerchantID: uint64(m.ID),
		MerchantNo: m.MerchantNo,
		Reason:     reason,
		Operator:   operator,
		Timestamp:  time.Now(),
	})

	return nil
}

// Enable 启用商家
func (m *Merchant) Enable(operator string) error {
	if m.Status != MerchantStatusDisabled {
		return fmt.Errorf("merchant is not disabled, current: %s", m.Status.String())
	}

	m.Status = MerchantStatusApproved

	m.AddDomainEvent(&MerchantEnabledEvent{
		MerchantID: uint64(m.ID),
		MerchantNo: m.MerchantNo,
		Operator:   operator,
		Timestamp:  time.Now(),
	})

	return nil
}

// UpdateLevel 更新商家等级
func (m *Merchant) UpdateLevel(level MerchantLevel) {
	oldLevel := m.Level
	m.Level = level

	m.AddDomainEvent(&MerchantLevelChangedEvent{
		MerchantID: uint64(m.ID),
		MerchantNo: m.MerchantNo,
		OldLevel:   oldLevel,
		NewLevel:   level,
		Timestamp:  time.Now(),
	})
}

// UpdateCreditScore 更新信用评分
func (m *Merchant) UpdateCreditScore(score float64) {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	m.CreditScore = score
}

// AddSales 增加销售额和订单数
func (m *Merchant) AddSales(amount int64, orderCount int32) {
	m.TotalSales += amount
	m.TotalOrders += orderCount
}

// IsActive 判断商家是否可正常经营
func (m *Merchant) IsActive() bool {
	return m.Status == MerchantStatusApproved
}

// CanCreateStore 判断是否可以创建店铺
func (m *Merchant) CanCreateStore() bool {
	return m.IsActive()
}

// AddDomainEvent 添加领域事件
func (m *Merchant) AddDomainEvent(event DomainEvent) {
	if m.domainEvents == nil {
		m.domainEvents = make([]DomainEvent, 0)
	}
	m.domainEvents = append(m.domainEvents, event)
}

// GetDomainEvents 获取领域事件
func (m *Merchant) GetDomainEvents() []DomainEvent {
	return m.domainEvents
}

// ClearDomainEvents 清除领域事件
func (m *Merchant) ClearDomainEvents() {
	m.domainEvents = make([]DomainEvent, 0)
}

// NewStore 创建店铺
func NewStore(merchantID uint, name, logoURL, bannerURL, description, businessHours, address string, categories []string) *Store {
	storeNo := fmt.Sprintf("S%d%d", time.Now().UnixNano()/1000000, merchantID%1000)

	return &Store{
		MerchantID:    merchantID,
		StoreNo:       storeNo,
		Name:          name,
		LogoURL:       logoURL,
		BannerURL:     bannerURL,
		Description:   description,
		BusinessHours: businessHours,
		Address:       address,
		Rating:        5.0,
		IsOpen:        true,
	}
}

// Open 开店
func (s *Store) Open() {
	s.IsOpen = true
}

// Close 关店
func (s *Store) Close() {
	s.IsOpen = false
}
