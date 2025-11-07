package model

import "time"

// ReportType 报表类型
type ReportType string

const (
	ReportTypeSales      ReportType = "SALES"       // 销售报表
	ReportTypeOrder      ReportType = "ORDER"       // 订单报表
	ReportTypeUser       ReportType = "USER"        // 用户报表
	ReportTypeProduct    ReportType = "PRODUCT"     // 商品报表
	ReportTypeInventory  ReportType = "INVENTORY"   // 库存报表
	ReportTypeFinance    ReportType = "FINANCE"     // 财务报表
	ReportTypeMarketing  ReportType = "MARKETING"   // 营销报表
	ReportTypeCustom     ReportType = "CUSTOM"      // 自定义报表
)

// ReportPeriod 报表周期
type ReportPeriod string

const (
	ReportPeriodDaily   ReportPeriod = "DAILY"   // 日报
	ReportPeriodWeekly  ReportPeriod = "WEEKLY"  // 周报
	ReportPeriodMonthly ReportPeriod = "MONTHLY" // 月报
	ReportPeriodYearly  ReportPeriod = "YEARLY"  // 年报
	ReportPeriodCustom  ReportPeriod = "CUSTOM"  // 自定义周期
)

// ReportStatus 报表状态
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "PENDING"   // 待生成
	ReportStatusGenerating ReportStatus = "GENERATING" // 生成中
	ReportStatusCompleted ReportStatus = "COMPLETED" // 已完成
	ReportStatusFailed    ReportStatus = "FAILED"    // 失败
)

// Report 报表
type Report struct {
	ID          uint64       `gorm:"primarykey" json:"id"`
	ReportNo    string       `gorm:"type:varchar(64);uniqueIndex;not null;comment:报表编号" json:"reportNo"`
	Type        ReportType   `gorm:"type:varchar(20);not null;index;comment:报表类型" json:"type"`
	Period      ReportPeriod `gorm:"type:varchar(20);not null;comment:报表周期" json:"period"`
	Title       string       `gorm:"type:varchar(255);not null;comment:报表标题" json:"title"`
	Description string       `gorm:"type:text;comment:报表描述" json:"description"`
	StartDate   time.Time    `gorm:"not null;index;comment:开始日期" json:"startDate"`
	EndDate     time.Time    `gorm:"not null;index;comment:结束日期" json:"endDate"`
	Status      ReportStatus `gorm:"type:varchar(20);not null;comment:报表状态" json:"status"`
	FileURL     string       `gorm:"type:varchar(500);comment:文件URL" json:"fileUrl"`
	FileFormat  string       `gorm:"type:varchar(20);comment:文件格式(PDF,EXCEL,CSV)" json:"fileFormat"`
	FileSize    int64        `gorm:"comment:文件大小(字节)" json:"fileSize"`
	Data        string       `gorm:"type:longtext;comment:报表数据JSON" json:"data"`
	GeneratedBy uint64       `gorm:"comment:生成人ID" json:"generatedBy"`
	GeneratedAt *time.Time   `gorm:"comment:生成时间" json:"generatedAt"`
	ErrorMsg    string       `gorm:"type:text;comment:错误信息" json:"errorMsg"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time   `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (Report) TableName() string {
	return "reports"
}

// SalesReport 销售报表数据
type SalesReport struct {
	// 基础指标
	TotalOrders      int64   `json:"totalOrders"`      // 总订单数
	TotalAmount      int64   `json:"totalAmount"`      // 总销售额(分)
	TotalRefund      int64   `json:"totalRefund"`      // 总退款额(分)
	NetAmount        int64   `json:"netAmount"`        // 净销售额(分)
	AvgOrderAmount   int64   `json:"avgOrderAmount"`   // 平均订单金额(分)
	
	// 用户指标
	TotalUsers       int64   `json:"totalUsers"`       // 总用户数
	NewUsers         int64   `json:"newUsers"`         // 新增用户数
	ActiveUsers      int64   `json:"activeUsers"`      // 活跃用户数
	PayingUsers      int64   `json:"payingUsers"`      // 付费用户数
	RepurchaseRate   float64 `json:"repurchaseRate"`   // 复购率
	
	// 商品指标
	TotalProducts    int64   `json:"totalProducts"`    // 总商品数
	SoldProducts     int64   `json:"soldProducts"`     // 已售商品数
	TotalQuantity    int64   `json:"totalQuantity"`    // 总销量
	
	// 转化指标
	ConversionRate   float64 `json:"conversionRate"`   // 转化率
	PaymentRate      float64 `json:"paymentRate"`      // 支付成功率
	RefundRate       float64 `json:"refundRate"`       // 退款率
	
	// 趋势数据
	DailyData        []DailySalesData `json:"dailyData"`        // 每日数据
	CategoryData     []CategorySalesData `json:"categoryData"`  // 分类数据
	ProductRanking   []ProductRankingData `json:"productRanking"` // 商品排行
	RegionData       []RegionSalesData `json:"regionData"`      // 地区数据
}

// DailySalesData 每日销售数据
type DailySalesData struct {
	Date        string `json:"date"`        // 日期
	OrderCount  int64  `json:"orderCount"`  // 订单数
	Amount      int64  `json:"amount"`      // 销售额(分)
	UserCount   int64  `json:"userCount"`   // 用户数
	ProductCount int64 `json:"productCount"` // 商品数
}

// CategorySalesData 分类销售数据
type CategorySalesData struct {
	CategoryID   uint64  `json:"categoryId"`   // 分类ID
	CategoryName string  `json:"categoryName"` // 分类名称
	OrderCount   int64   `json:"orderCount"`   // 订单数
	Amount       int64   `json:"amount"`       // 销售额(分)
	Percentage   float64 `json:"percentage"`   // 占比
}

// ProductRankingData 商品排行数据
type ProductRankingData struct {
	Rank        int    `json:"rank"`        // 排名
	ProductID   uint64 `json:"productId"`   // 商品ID
	ProductName string `json:"productName"` // 商品名称
	SalesCount  int64  `json:"salesCount"`  // 销量
	Amount      int64  `json:"amount"`      // 销售额(分)
}

// RegionSalesData 地区销售数据
type RegionSalesData struct {
	Province   string  `json:"province"`   // 省份
	City       string  `json:"city"`       // 城市
	OrderCount int64   `json:"orderCount"` // 订单数
	Amount     int64   `json:"amount"`     // 销售额(分)
	Percentage float64 `json:"percentage"` // 占比
}

// OrderReport 订单报表数据
type OrderReport struct {
	// 订单统计
	TotalOrders       int64   `json:"totalOrders"`       // 总订单数
	PendingOrders     int64   `json:"pendingOrders"`     // 待支付订单
	PaidOrders        int64   `json:"paidOrders"`        // 已支付订单
	ShippedOrders     int64   `json:"shippedOrders"`     // 已发货订单
	CompletedOrders   int64   `json:"completedOrders"`   // 已完成订单
	CancelledOrders   int64   `json:"cancelledOrders"`   // 已取消订单
	
	// 金额统计
	TotalAmount       int64   `json:"totalAmount"`       // 总金额(分)
	PaidAmount        int64   `json:"paidAmount"`        // 已支付金额(分)
	RefundAmount      int64   `json:"refundAmount"`      // 退款金额(分)
	
	// 时效统计
	AvgPaymentTime    float64 `json:"avgPaymentTime"`    // 平均支付时长(分钟)
	AvgShippingTime   float64 `json:"avgShippingTime"`   // 平均发货时长(小时)
	AvgDeliveryTime   float64 `json:"avgDeliveryTime"`   // 平均配送时长(天)
	
	// 趋势数据
	HourlyData        []HourlyOrderData `json:"hourlyData"`        // 每小时数据
	PaymentMethodData []PaymentMethodData `json:"paymentMethodData"` // 支付方式数据
}

// HourlyOrderData 每小时订单数据
type HourlyOrderData struct {
	Hour       int   `json:"hour"`       // 小时(0-23)
	OrderCount int64 `json:"orderCount"` // 订单数
	Amount     int64 `json:"amount"`     // 金额(分)
}

// PaymentMethodData 支付方式数据
type PaymentMethodData struct {
	Method     string  `json:"method"`     // 支付方式
	OrderCount int64   `json:"orderCount"` // 订单数
	Amount     int64   `json:"amount"`     // 金额(分)
	Percentage float64 `json:"percentage"` // 占比
}

// UserReport 用户报表数据
type UserReport struct {
	// 用户统计
	TotalUsers      int64   `json:"totalUsers"`      // 总用户数
	NewUsers        int64   `json:"newUsers"`        // 新增用户数
	ActiveUsers     int64   `json:"activeUsers"`     // 活跃用户数
	PayingUsers     int64   `json:"payingUsers"`     // 付费用户数
	
	// 用户行为
	AvgLoginFreq    float64 `json:"avgLoginFreq"`    // 平均登录频次
	AvgBrowseTime   float64 `json:"avgBrowseTime"`   // 平均浏览时长(分钟)
	AvgOrderCount   float64 `json:"avgOrderCount"`   // 平均订单数
	AvgOrderAmount  int64   `json:"avgOrderAmount"`  // 平均订单金额(分)
	
	// 用户价值
	TotalLTV        int64   `json:"totalLtv"`        // 总生命周期价值(分)
	AvgLTV          int64   `json:"avgLtv"`          // 平均生命周期价值(分)
	
	// 用户留存
	RetentionDay1   float64 `json:"retentionDay1"`   // 次日留存率
	RetentionDay7   float64 `json:"retentionDay7"`   // 7日留存率
	RetentionDay30  float64 `json:"retentionDay30"`  // 30日留存率
	
	// 趋势数据
	DailyUserData   []DailyUserData `json:"dailyUserData"`   // 每日用户数据
	UserLevelData   []UserLevelData `json:"userLevelData"`   // 用户等级数据
	UserRegionData  []UserRegionData `json:"userRegionData"` // 用户地区数据
}

// DailyUserData 每日用户数据
type DailyUserData struct {
	Date        string `json:"date"`        // 日期
	NewUsers    int64  `json:"newUsers"`    // 新增用户
	ActiveUsers int64  `json:"activeUsers"` // 活跃用户
	PayingUsers int64  `json:"payingUsers"` // 付费用户
}

// UserLevelData 用户等级数据
type UserLevelData struct {
	Level      string  `json:"level"`      // 等级
	UserCount  int64   `json:"userCount"`  // 用户数
	Percentage float64 `json:"percentage"` // 占比
}

// UserRegionData 用户地区数据
type UserRegionData struct {
	Province   string  `json:"province"`   // 省份
	UserCount  int64   `json:"userCount"`  // 用户数
	Percentage float64 `json:"percentage"` // 占比
}

// ProductReport 商品报表数据
type ProductReport struct {
	// 商品统计
	TotalProducts   int64   `json:"totalProducts"`   // 总商品数
	OnSaleProducts  int64   `json:"onSaleProducts"`  // 在售商品数
	SoldOutProducts int64   `json:"soldOutProducts"` // 售罄商品数
	
	// 销售统计
	TotalSales      int64   `json:"totalSales"`      // 总销量
	TotalAmount     int64   `json:"totalAmount"`     // 总销售额(分)
	AvgPrice        int64   `json:"avgPrice"`        // 平均价格(分)
	
	// 库存统计
	TotalStock      int64   `json:"totalStock"`      // 总库存
	LowStockCount   int64   `json:"lowStockCount"`   // 低库存商品数
	
	// 趋势数据
	CategoryData    []CategoryProductData `json:"categoryData"`    // 分类数据
	BrandData       []BrandProductData `json:"brandData"`       // 品牌数据
	TopProducts     []TopProductData `json:"topProducts"`     // 热销商品
	NewProducts     []NewProductData `json:"newProducts"`     // 新品数据
}

// CategoryProductData 分类商品数据
type CategoryProductData struct {
	CategoryID   uint64 `json:"categoryId"`   // 分类ID
	CategoryName string `json:"categoryName"` // 分类名称
	ProductCount int64  `json:"productCount"` // 商品数
	SalesCount   int64  `json:"salesCount"`   // 销量
	Amount       int64  `json:"amount"`       // 销售额(分)
}

// BrandProductData 品牌商品数据
type BrandProductData struct {
	BrandID      uint64 `json:"brandId"`      // 品牌ID
	BrandName    string `json:"brandName"`    // 品牌名称
	ProductCount int64  `json:"productCount"` // 商品数
	SalesCount   int64  `json:"salesCount"`   // 销量
	Amount       int64  `json:"amount"`       // 销售额(分)
}

// TopProductData 热销商品数据
type TopProductData struct {
	Rank        int    `json:"rank"`        // 排名
	ProductID   uint64 `json:"productId"`   // 商品ID
	ProductName string `json:"productName"` // 商品名称
	SalesCount  int64  `json:"salesCount"`  // 销量
	Amount      int64  `json:"amount"`      // 销售额(分)
	ViewCount   int64  `json:"viewCount"`   // 浏览量
}

// NewProductData 新品数据
type NewProductData struct {
	ProductID   uint64    `json:"productId"`   // 商品ID
	ProductName string    `json:"productName"` // 商品名称
	LaunchDate  time.Time `json:"launchDate"`  // 上架日期
	SalesCount  int64     `json:"salesCount"`  // 销量
	Amount      int64     `json:"amount"`      // 销售额(分)
}

// InventoryReport 库存报表数据
type InventoryReport struct {
	// 库存统计
	TotalStock      int64   `json:"totalStock"`      // 总库存
	TotalValue      int64   `json:"totalValue"`      // 总库存价值(分)
	LowStockCount   int64   `json:"lowStockCount"`   // 低库存商品数
	OutOfStockCount int64   `json:"outOfStockCount"` // 缺货商品数
	
	// 库存周转
	TurnoverRate    float64 `json:"turnoverRate"`    // 库存周转率
	AvgTurnoverDays float64 `json:"avgTurnoverDays"` // 平均周转天数
	
	// 趋势数据
	DailyStockData  []DailyStockData `json:"dailyStockData"`  // 每日库存数据
	CategoryStockData []CategoryStockData `json:"categoryStockData"` // 分类库存数据
	WarningProducts []WarningProductData `json:"warningProducts"` // 预警商品
}

// DailyStockData 每日库存数据
type DailyStockData struct {
	Date       string `json:"date"`       // 日期
	TotalStock int64  `json:"totalStock"` // 总库存
	InStock    int64  `json:"inStock"`    // 入库
	OutStock   int64  `json:"outStock"`   // 出库
}

// CategoryStockData 分类库存数据
type CategoryStockData struct {
	CategoryID   uint64 `json:"categoryId"`   // 分类ID
	CategoryName string `json:"categoryName"` // 分类名称
	Stock        int64  `json:"stock"`        // 库存
	Value        int64  `json:"value"`        // 价值(分)
}

// WarningProductData 预警商品数据
type WarningProductData struct {
	ProductID   uint64 `json:"productId"`   // 商品ID
	ProductName string `json:"productName"` // 商品名称
	Stock       int64  `json:"stock"`       // 当前库存
	SafeStock   int64  `json:"safeStock"`   // 安全库存
	WarningType string `json:"warningType"` // 预警类型(LOW,OUT)
}

// FinanceReport 财务报表数据
type FinanceReport struct {
	// 收入统计
	TotalRevenue    int64   `json:"totalRevenue"`    // 总收入(分)
	ProductRevenue  int64   `json:"productRevenue"`  // 商品收入(分)
	ServiceRevenue  int64   `json:"serviceRevenue"`  // 服务收入(分)
	OtherRevenue    int64   `json:"otherRevenue"`    // 其他收入(分)
	
	// 支出统计
	TotalExpense    int64   `json:"totalExpense"`    // 总支出(分)
	RefundExpense   int64   `json:"refundExpense"`   // 退款支出(分)
	CommissionExpense int64 `json:"commissionExpense"` // 佣金支出(分)
	OtherExpense    int64   `json:"otherExpense"`    // 其他支出(分)
	
	// 利润统计
	GrossProfit     int64   `json:"grossProfit"`     // 毛利润(分)
	NetProfit       int64   `json:"netProfit"`       // 净利润(分)
	ProfitMargin    float64 `json:"profitMargin"`    // 利润率
	
	// 趋势数据
	DailyFinanceData []DailyFinanceData `json:"dailyFinanceData"` // 每日财务数据
}

// DailyFinanceData 每日财务数据
type DailyFinanceData struct {
	Date     string `json:"date"`     // 日期
	Revenue  int64  `json:"revenue"`  // 收入(分)
	Expense  int64  `json:"expense"`  // 支出(分)
	Profit   int64  `json:"profit"`   // 利润(分)
}

// MarketingReport 营销报表数据
type MarketingReport struct {
	// 优惠券统计
	TotalCoupons    int64   `json:"totalCoupons"`    // 总优惠券数
	IssuedCoupons   int64   `json:"issuedCoupons"`   // 已发放优惠券
	UsedCoupons     int64   `json:"usedCoupons"`     // 已使用优惠券
	CouponUsageRate float64 `json:"couponUsageRate"` // 优惠券使用率
	CouponDiscount  int64   `json:"couponDiscount"`  // 优惠券优惠金额(分)
	
	// 促销统计
	TotalPromotions int64   `json:"totalPromotions"` // 总促销活动数
	ActivePromotions int64  `json:"activePromotions"` // 进行中活动数
	PromotionOrders int64   `json:"promotionOrders"` // 促销订单数
	PromotionAmount int64   `json:"promotionAmount"` // 促销金额(分)
	
	// 积分统计
	TotalPoints     int64   `json:"totalPoints"`     // 总积分发放
	UsedPoints      int64   `json:"usedPoints"`      // 已使用积分
	PointsValue     int64   `json:"pointsValue"`     // 积分价值(分)
	
	// ROI统计
	MarketingCost   int64   `json:"marketingCost"`   // 营销成本(分)
	MarketingRevenue int64  `json:"marketingRevenue"` // 营销收入(分)
	ROI             float64 `json:"roi"`             // 投资回报率
}
