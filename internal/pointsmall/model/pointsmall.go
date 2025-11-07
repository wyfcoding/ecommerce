package model

import "time"

// PointsProductStatus 积分商品状态
type PointsProductStatus string

const (
	PointsProductStatusOnSale   PointsProductStatus = "ON_SALE"   // 上架
	PointsProductStatusOffSale  PointsProductStatus = "OFF_SALE"  // 下架
	PointsProductStatusSoldOut  PointsProductStatus = "SOLD_OUT"  // 售罄
)

// ExchangeType 兑换类型
type ExchangeType string

const (
	ExchangeTypePoints      ExchangeType = "POINTS"       // 纯积分兑换
	ExchangeTypePointsCash  ExchangeType = "POINTS_CASH"  // 积分+现金
	ExchangeTypeLottery     ExchangeType = "LOTTERY"      // 积分抽奖
)

// PointsProduct 积分商品
type PointsProduct struct {
	ID              uint64              `gorm:"primarykey" json:"id"`
	ProductNo       string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:商品编号" json:"productNo"`
	Name            string              `gorm:"type:varchar(255);not null;comment:商品名称" json:"name"`
	Description     string              `gorm:"type:text;comment:商品描述" json:"description"`
	MainImageURL    string              `gorm:"type:varchar(255);comment:主图URL" json:"mainImageUrl"`
	GalleryImages   string              `gorm:"type:text;comment:画廊图片JSON" json:"galleryImages"`
	CategoryID      uint64              `gorm:"index;not null;comment:分类ID" json:"categoryId"`
	ExchangeType    ExchangeType        `gorm:"type:varchar(20);not null;comment:兑换类型" json:"exchangeType"`
	PointsPrice     int64               `gorm:"not null;comment:积分价格" json:"pointsPrice"`
	CashPrice       int64               `gorm:"comment:现金价格(分)" json:"cashPrice"`
	OriginalPrice   int64               `gorm:"comment:原价(分)" json:"originalPrice"`
	Stock           int32               `gorm:"not null;default:0;comment:库存" json:"stock"`
	SoldCount       int32               `gorm:"not null;default:0;comment:已兑换数量" json:"soldCount"`
	LimitPerUser    int32               `gorm:"not null;default:1;comment:每人限兑" json:"limitPerUser"`
	Status          PointsProductStatus `gorm:"type:varchar(20);not null;comment:商品状态" json:"status"`
	SortOrder       int32               `gorm:"not null;default:0;comment:排序" json:"sortOrder"`
	StartTime       *time.Time          `gorm:"comment:上架时间" json:"startTime"`
	EndTime         *time.Time          `gorm:"comment:下架时间" json:"endTime"`
	CreatedAt       time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt       *time.Time          `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (PointsProduct) TableName() string {
	return "points_products"
}

// PointsCategory 积分商品分类
type PointsCategory struct {
	ID        uint64     `gorm:"primarykey" json:"id"`
	Name      string     `gorm:"type:varchar(100);not null;comment:分类名称" json:"name"`
	Icon      string     `gorm:"type:varchar(255);comment:分类图标" json:"icon"`
	SortOrder int32      `gorm:"not null;default:0;comment:排序" json:"sortOrder"`
	IsVisible bool       `gorm:"not null;default:true;comment:是否可见" json:"isVisible"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (PointsCategory) TableName() string {
	return "points_categories"
}

// ExchangeOrderStatus 兑换订单状态
type ExchangeOrderStatus string

const (
	ExchangeOrderStatusPending    ExchangeOrderStatus = "PENDING"    // 待发货
	ExchangeOrderStatusShipped    ExchangeOrderStatus = "SHIPPED"    // 已发货
	ExchangeOrderStatusCompleted  ExchangeOrderStatus = "COMPLETED"  // 已完成
	ExchangeOrderStatusCancelled  ExchangeOrderStatus = "CANCELLED"  // 已取消
)

// ExchangeOrder 兑换订单
type ExchangeOrder struct {
	ID              uint64              `gorm:"primarykey" json:"id"`
	OrderNo         string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:订单编号" json:"orderNo"`
	UserID          uint64              `gorm:"index;not null;comment:用户ID" json:"userId"`
	ProductID       uint64              `gorm:"index;not null;comment:商品ID" json:"productId"`
	ProductName     string              `gorm:"type:varchar(255);not null;comment:商品名称" json:"productName"`
	ProductImage    string              `gorm:"type:varchar(255);comment:商品图片" json:"productImage"`
	Quantity        int32               `gorm:"not null;comment:兑换数量" json:"quantity"`
	PointsPrice     int64               `gorm:"not null;comment:积分价格" json:"pointsPrice"`
	CashPrice       int64               `gorm:"comment:现金价格(分)" json:"cashPrice"`
	TotalPoints     int64               `gorm:"not null;comment:总积分" json:"totalPoints"`
	TotalCash       int64               `gorm:"comment:总现金(分)" json:"totalCash"`
	Status          ExchangeOrderStatus `gorm:"type:varchar(20);not null;comment:订单状态" json:"status"`
	AddressID       uint64              `gorm:"comment:收货地址ID" json:"addressId"`
	RecipientName   string              `gorm:"type:varchar(100);comment:收货人" json:"recipientName"`
	RecipientPhone  string              `gorm:"type:varchar(20);comment:收货电话" json:"recipientPhone"`
	RecipientAddress string             `gorm:"type:varchar(500);comment:收货地址" json:"recipientAddress"`
	ExpressCompany  string              `gorm:"type:varchar(50);comment:快递公司" json:"expressCompany"`
	ExpressNo       string              `gorm:"type:varchar(100);comment:快递单号" json:"expressNo"`
	Remark          string              `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedAt       time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
	ShippedAt       *time.Time          `gorm:"comment:发货时间" json:"shippedAt"`
	CompletedAt     *time.Time          `gorm:"comment:完成时间" json:"completedAt"`
	DeletedAt       *time.Time          `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (ExchangeOrder) TableName() string {
	return "exchange_orders"
}

// LotteryActivity 抽奖活动
type LotteryActivity struct {
	ID              uint64     `gorm:"primarykey" json:"id"`
	Name            string     `gorm:"type:varchar(255);not null;comment:活动名称" json:"name"`
	Description     string     `gorm:"type:text;comment:活动描述" json:"description"`
	PointsPerDraw   int64      `gorm:"not null;comment:每次抽奖消耗积分" json:"pointsPerDraw"`
	MaxDrawsPerUser int32      `gorm:"not null;default:0;comment:每人最多抽奖次数,0为不限" json:"maxDrawsPerUser"`
	StartTime       time.Time  `gorm:"not null;comment:开始时间" json:"startTime"`
	EndTime         time.Time  `gorm:"not null;comment:结束时间" json:"endTime"`
	IsActive        bool       `gorm:"not null;default:true;comment:是否激活" json:"isActive"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt       *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (LotteryActivity) TableName() string {
	return "lottery_activities"
}

// LotteryPrize 抽奖奖品
type LotteryPrize struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	ActivityID  uint64     `gorm:"index;not null;comment:活动ID" json:"activityId"`
	Name        string     `gorm:"type:varchar(255);not null;comment:奖品名称" json:"name"`
	Type        string     `gorm:"type:varchar(20);not null;comment:奖品类型(POINTS,COUPON,PRODUCT)" json:"type"`
	Value       string     `gorm:"type:varchar(255);comment:奖品值" json:"value"`
	ImageURL    string     `gorm:"type:varchar(255);comment:奖品图片" json:"imageUrl"`
	Probability float64    `gorm:"not null;comment:中奖概率(0-1)" json:"probability"`
	TotalCount  int32      `gorm:"not null;comment:奖品总数" json:"totalCount"`
	RemainCount int32      `gorm:"not null;comment:剩余数量" json:"remainCount"`
	SortOrder   int32      `gorm:"not null;default:0;comment:排序" json:"sortOrder"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (LotteryPrize) TableName() string {
	return "lottery_prizes"
}

// LotteryRecord 抽奖记录
type LotteryRecord struct {
	ID         uint64     `gorm:"primarykey" json:"id"`
	ActivityID uint64     `gorm:"index;not null;comment:活动ID" json:"activityId"`
	UserID     uint64     `gorm:"index;not null;comment:用户ID" json:"userId"`
	PrizeID    uint64     `gorm:"comment:中奖奖品ID,0表示未中奖" json:"prizeId"`
	PrizeName  string     `gorm:"type:varchar(255);comment:奖品名称" json:"prizeName"`
	IsWinning  bool       `gorm:"not null;comment:是否中奖" json:"isWinning"`
	Points     int64      `gorm:"not null;comment:消耗积分" json:"points"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	DeletedAt  *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (LotteryRecord) TableName() string {
	return "lottery_records"
}

// PointsTask 积分任务
type PointsTask struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	Name        string     `gorm:"type:varchar(255);not null;comment:任务名称" json:"name"`
	Description string     `gorm:"type:text;comment:任务描述" json:"description"`
	Type        string     `gorm:"type:varchar(20);not null;comment:任务类型(DAILY,WEEKLY,ONCE)" json:"type"`
	Action      string     `gorm:"type:varchar(50);not null;comment:任务动作(LOGIN,SHARE,PURCHASE等)" json:"action"`
	Points      int64      `gorm:"not null;comment:奖励积分" json:"points"`
	Target      int32      `gorm:"not null;default:1;comment:目标次数" json:"target"`
	IsActive    bool       `gorm:"not null;default:true;comment:是否激活" json:"isActive"`
	SortOrder   int32      `gorm:"not null;default:0;comment:排序" json:"sortOrder"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (PointsTask) TableName() string {
	return "points_tasks"
}

// UserTaskProgress 用户任务进度
type UserTaskProgress struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	UserID      uint64     `gorm:"index:idx_user_task;not null;comment:用户ID" json:"userId"`
	TaskID      uint64     `gorm:"index:idx_user_task;not null;comment:任务ID" json:"taskId"`
	Progress    int32      `gorm:"not null;default:0;comment:当前进度" json:"progress"`
	IsCompleted bool       `gorm:"not null;default:false;comment:是否完成" json:"isCompleted"`
	CompletedAt *time.Time `gorm:"comment:完成时间" json:"completedAt"`
	ResetAt     time.Time  `gorm:"not null;comment:重置时间" json:"resetAt"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName 指定表名
func (UserTaskProgress) TableName() string {
	return "user_task_progress"
}
