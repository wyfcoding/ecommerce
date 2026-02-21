package domain

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// OptimizationGoal 优化目标
type OptimizationGoal string

const (
	GoalMinimizeCost    OptimizationGoal = "MINIMIZE_COST"    // 最小化成本
	GoalMaximizeService OptimizationGoal = "MAXIMIZE_SERVICE" // 最大化服务水平
	GoalBalance         OptimizationGoal = "BALANCE"          // 平衡成本和服务
)

// OptimizationModel 优化模型
type OptimizationModel string

const (
	ModelEOQ          OptimizationModel = "EOQ"           // 经济订货批量
	ModelROP          OptimizationModel = "ROP"           // 再订货点
	ModelNewsboy      OptimizationModel = "NEWSBOY"       // 报童模型
	ModelMultiEchelon OptimizationModel = "MULTI_ECHELON" // 多级库存
)

// EOQParameters EOQ参数
type EOQParameters struct {
	AnnualDemand     float64 `json:"annual_demand"`     // 年需求量
	OrderingCost     float64 `json:"ordering_cost"`     // 订货成本
	HoldingCostRate  float64 `json:"holding_cost_rate"` // 持有成本率
	UnitCost         float64 `json:"unit_cost"`         // 单位成本
	LeadTime         int     `json:"lead_time"`         // 提前期（天）
}

// EOQResult EOQ结果
type EOQResult struct {
	OptimalOrderQuantity float64 `json:"optimal_order_quantity"` // 最优订货量
	TotalCost            float64 `json:"total_cost"`             // 总成本
	OrderingCost         float64 `json:"ordering_cost"`          // 订货成本
	HoldingCost          float64 `json:"holding_cost"`           // 持有成本
	ReorderPoint         float64 `json:"reorder_point"`          // 再订货点
	OrderFrequency       float64 `json:"order_frequency"`        // 订货频率
	CycleTime            float64 `json:"cycle_time"`             // 订货周期
}

// NewsboyParameters 报童模型参数
type NewsboyParameters struct {
	DemandMean        float64 `json:"demand_mean"`        // 需求均值
	DemandStdDev      float64 `json:"demand_std_dev"`     // 需求标准差
	UnitCost          float64 `json:"unit_cost"`          // 单位成本
	SellingPrice      float64 `json:"selling_price"`      // 销售价格
	SalvageValue      float64 `json:"salvage_value"`      // 残值
	ShortageCost      float64 `json:"shortage_cost"`      // 缺货成本
	ServiceLevel      float64 `json:"service_level"`      // 服务水平
}

// NewsboyResult 报童模型结果
type NewsboyResult struct {
	OptimalOrderQuantity float64 `json:"optimal_order_quantity"` // 最优订货量
	ExpectedProfit       float64 `json:"expected_profit"`        // 期望利润
	ExpectedSales        float64 `json:"expected_sales"`         // 期望销售量
	ExpectedLeftover     float64 `json:"expected_leftover"`     // 期望剩余量
	ExpectedShortage     float64 `json:"expected_shortage"`     // 期望缺货量
	ServiceLevel         float64 `json:"service_level"`          // 实际服务水平
}

// InventoryOptimizer 库存优化器
type InventoryOptimizer struct {
	demandRepo        DemandRepository
	costRepo          CostRepository
	mu                sync.RWMutex
	config            *OptimizationConfig
}

// OptimizationConfig 优化配置
type OptimizationConfig struct {
	DefaultServiceLevel float64 `json:"default_service_level"`
	MinServiceLevel     float64 `json:"min_service_level"`
	MaxServiceLevel     float64 `json:"max_service_level"`
	HoldingCostRate     float64 `json:"holding_cost_rate"`
	OrderingCost        float64 `json:"ordering_cost"`
	LeadTime            int     `json:"lead_time"`
}

// NewInventoryOptimizer 创建库存优化器
func NewInventoryOptimizer(demandRepo DemandRepository, costRepo CostRepository) *InventoryOptimizer {
	return &InventoryOptimizer{
		demandRepo: demandRepo,
		costRepo:   costRepo,
		config: &OptimizationConfig{
			DefaultServiceLevel: 0.95,
			MinServiceLevel:     0.90,
			MaxServiceLevel:     0.99,
			HoldingCostRate:     0.2,
			OrderingCost:       50,
			LeadTime:           7,
		},
	}
}

// OptimizeInventory 优化库存
func (io *InventoryOptimizer) OptimizeInventory(ctx context.Context, productID string, goal OptimizationGoal, model OptimizationModel) (*OptimizationResult, error) {
	// 获取产品数据
	product, err := io.getProductData(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product data: %w", err)
	}
	
	// 获取需求数据
	demandData, err := io.getDemandData(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get demand data: %w", err)
	}
	
	// 获取成本数据
	costData, err := io.getCostData(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost data: %w", err)
	}
	
	// 根据模型进行优化
	var result interface{}
	
	switch model {
	case ModelEOQ:
		result, err = io.optimizeEOQ(product, demandData, costData, goal)
	case ModelNewsboy:
		result, err = io.optimizeNewsboy(product, demandData, costData, goal)
	case ModelROP:
		result, err = io.optimizeROP(product, demandData, costData, goal)
	default:
		return nil, fmt.Errorf("unsupported optimization model: %s", model)
	}
	
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}
	
	// 创建优化结果
	optimizationResult := &OptimizationResult{
		ProductID:       productID,
		Goal:            goal,
		Model:           model,
		Result:          result,
		GeneratedAt:     time.Now(),
	}
	
	return optimizationResult, nil
}

// optimizeEOQ 优化EOQ
func (io *InventoryOptimizer) optimizeEOQ(product *Product, demandData *DemandData, costData *CostData, goal OptimizationGoal) (*EOQResult, error) {
	// 计算年需求量
	annualDemand := io.calculateAnnualDemand(demandData)
	
	// 获取成本参数
	orderingCost := costData.OrderingCost
	if orderingCost == 0 {
		orderingCost = io.config.OrderingCost
	}
	
	holdingCostRate := costData.HoldingCostRate
	if holdingCostRate == 0 {
		holdingCostRate = io.config.HoldingCostRate
	}
	
	unitCost := costData.UnitCost
	if unitCost == 0 {
		unitCost = product.Cost
	}
	
	// 计算EOQ
	eoq := math.Sqrt((2 * annualDemand * orderingCost) / (unitCost * holdingCostRate))
	
	// 计算总成本
	totalCost := (annualDemand / eoq) * orderingCost + (eoq / 2) * unitCost * holdingCostRate
	
	// 计算再订货点
	reorderPoint := io.calculateReorderPoint(demandData, costData.LeadTime)
	
	// 计算订货频率和周期
	orderFrequency := annualDemand / eoq
	cycleTime := 365 / orderFrequency
	
	result := &EOQResult{
		OptimalOrderQuantity: eoq,
		TotalCost:            totalCost,
		OrderingCost:         (annualDemand / eoq) * orderingCost,
		HoldingCost:          (eoq / 2) * unitCost * holdingCostRate,
		ReorderPoint:         reorderPoint,
		OrderFrequency:       orderFrequency,
		CycleTime:            cycleTime,
	}
	
	return result, nil
}

// optimizeNewsboy 优化报童模型
func (io *InventoryOptimizer) optimizeNewsboy(product *Product, demandData *DemandData, costData *CostData, goal OptimizationGoal) (*NewsboyResult, error) {
	// 计算需求统计
	demandMean, demandStdDev := io.calculateDemandStatistics(demandData)
	
	// 获取成本参数
	unitCost := costData.UnitCost
	if unitCost == 0 {
		unitCost = product.Cost
	}
	
	sellingPrice := costData.SellingPrice
	if sellingPrice == 0 {
		sellingPrice = product.Price
	}
	
	salvageValue := costData.SalvageValue
	if salvageValue == 0 {
		salvageValue = unitCost * 0.5 // 默认残值为成本的一半
	}
	
	shortageCost := costData.ShortageCost
	if shortageCost == 0 {
		shortageCost = sellingPrice - unitCost // 缺货成本为利润损失
	}
	
	// 计算临界比
	criticalRatio := (sellingPrice - unitCost + shortageCost) / (sellingPrice - salvageValue + shortageCost)
	
	// 确保临界比在有效范围内
	if criticalRatio < 0 {
		criticalRatio = 0
	}
	if criticalRatio > 1 {
		criticalRatio = 1
	}
	
	// 计算最优订货量（假设正态分布）
	zScore := io.calculateZScore(criticalRatio)
	optimalOrderQuantity := demandMean + zScore*demandStdDev
	
	// 计算期望指标
	expectedSales, expectedLeftover, expectedShortage := io.calculateExpectedMetrics(demandMean, demandStdDev, optimalOrderQuantity)
	
	// 计算期望利润
	expectedProfit := expectedSales*sellingPrice + expectedLeftover*salvageValue - optimalOrderQuantity*unitCost - expectedShortage*shortageCost
	
	// 计算实际服务水平
	serviceLevel := 1 - (expectedShortage / demandMean)
	
	result := &NewsboyResult{
		OptimalOrderQuantity: optimalOrderQuantity,
		ExpectedProfit:       expectedProfit,
		ExpectedSales:        expectedSales,
		ExpectedLeftover:     expectedLeftover,
		ExpectedShortage:     expectedShortage,
		ServiceLevel:         serviceLevel,
	}
	
	return result, nil
}

// optimizeROP 优化再订货点
func (io *InventoryOptimizer) optimizeROP(product *Product, demandData *DemandData, costData *CostData, goal OptimizationGoal) (*ROPResult, error) {
	// 计算需求统计
	demandMean, demandStdDev := io.calculateDemandStatistics(demandData)
	
	// 获取提前期
	leadTime := costData.LeadTime
	if leadTime == 0 {
		leadTime = io.config.LeadTime
	}
	
	// 计算提前期需求
	leadTimeDemandMean := demandMean * float64(leadTime) / 365
	leadTimeDemandStdDev := demandStdDev * math.Sqrt(float64(leadTime)/365)
	
	// 计算服务水平
	serviceLevel := io.config.DefaultServiceLevel
	if goal == GoalMaximizeService {
		serviceLevel = io.config.MaxServiceLevel
	} else if goal == GoalMinimizeCost {
		serviceLevel = io.config.MinServiceLevel
	}
	
	// 计算Z分数
	zScore := io.calculateZScore(serviceLevel)
	
	// 计算再订货点
	reorderPoint := leadTimeDemandMean + zScore*leadTimeDemandStdDev
	
	// 计算安全库存
	safetyStock := zScore * leadTimeDemandStdDev
	
	result := &ROPResult{
		ReorderPoint:   reorderPoint,
		SafetyStock:    safetyStock,
		LeadTimeDemand: leadTimeDemandMean,
		ServiceLevel:   serviceLevel,
		ZScore:         zScore,
	}
	
	return result, nil
}

// calculateAnnualDemand 计算年需求量
func (io *InventoryOptimizer) calculateAnnualDemand(demandData *DemandData) float64 {
	// 基于历史数据计算年需求量
	if len(demandData.History) == 0 {
		return 0
	}
	
	var totalDemand float64
	for _, demand := range demandData.History {
		totalDemand += demand.Quantity
	}
	
	// 计算日均需求
	avgDailyDemand := totalDemand / float64(len(demandData.History))
	
	// 计算年需求量
	annualDemand := avgDailyDemand * 365
	
	return annualDemand
}

// calculateDemandStatistics 计算需求统计
func (io *InventoryOptimizer) calculateDemandStatistics(demandData *DemandData) (float64, float64) {
	if len(demandData.History) == 0 {
		return 0, 0
	}
	
	// 计算均值
	var sum float64
	for _, demand := range demandData.History {
		sum += demand.Quantity
	}
	mean := sum / float64(len(demandData.History))
	
	// 计算标准差
	var variance float64
	for _, demand := range demandData.History {
		diff := demand.Quantity - mean
		variance += diff * diff
	}
	variance /= float64(len(demandData.History))
	stdDev := math.Sqrt(variance)
	
	return mean, stdDev
}

// calculateReorderPoint 计算再订货点
func (io *InventoryOptimizer) calculateReorderPoint(demandData *DemandData, leadTime int) float64 {
	// 计算日均需求
	dailyDemand := io.calculateDailyDemand(demandData)
	
	// 计算提前期需求
	leadTimeDemand := dailyDemand * float64(leadTime)
	
	// 加上安全库存
	safetyStock := io.calculateSafetyStock(demandData, leadTime)
	
	return leadTimeDemand + safetyStock
}

// calculateDailyDemand 计算日均需求
func (io *InventoryOptimizer) calculateDailyDemand(demandData *DemandData) float64 {
	if len(demandData.History) == 0 {
		return 0
	}
	
	var totalDemand float64
	for _, demand := range demandData.History {
		totalDemand += demand.Quantity
	}
	
	return totalDemand / float64(len(demandData.History))
}

// calculateSafetyStock 计算安全库存
func (io *InventoryOptimizer) calculateSafetyStock(demandData *DemandData, leadTime int) float64 {
	// 计算需求标准差
	_, stdDev := io.calculateDemandStatistics(demandData)
	
	// 计算提前期需求标准差
	leadTimeStdDev := stdDev * math.Sqrt(float64(leadTime))
	
	// 计算Z分数（95%服务水平）
	zScore := 1.645
	
	// 计算安全库存
	safetyStock := zScore * leadTimeStdDev
	
	return safetyStock
}

// calculateZScore 计算Z分数
func (io *InventoryOptimizer) calculateZScore(probability float64) float64 {
	// 标准正态分布Z分数
	zScores := map[float64]float64{
		0.50: 0.000,
		0.60: 0.253,
		0.70: 0.524,
		0.75: 0.674,
		0.80: 0.842,
		0.85: 1.036,
		0.90: 1.282,
		0.95: 1.645,
		0.975: 1.960,
		0.99: 2.326,
		0.995: 2.576,
		0.999: 3.090,
	}
	
	if z, exists := zScores[probability]; exists {
		return z
	}
	
	// 线性插值
	var lowerProb, lowerZ, upperProb, upperZ float64
	
	for prob, z := range zScores {
		if prob <= probability && prob > lowerProb {
			lowerProb = prob
			lowerZ = z
		}
		if prob >= probability && (upperProb == 0 || prob < upperProb) {
			upperProb = prob
			upperZ = z
		}
	}
	
	if lowerProb == upperProb {
		return lowerZ
	}
	
	// 线性插值
	interpolatedZ := lowerZ + (probability-lowerProb)*(upperZ-lowerZ)/(upperProb-lowerProb)
	
	return interpolatedZ
}

// calculateExpectedMetrics 计算期望指标
func (io *InventoryOptimizer) calculateExpectedMetrics(mean, stdDev, orderQuantity float64) (float64, float64, float64) {
	// 计算标准化值
	z := (orderQuantity - mean) / stdDev
	
	// 计算标准正态分布的损失函数
	lossFunction := io.calculateLossFunction(z)
	
	// 计算期望销售量
	expectedSales := mean - stdDev*lossFunction
	
	// 计算期望剩余量
	expectedLeftover := orderQuantity - expectedSales
	
	// 计算期望缺货量
	expectedShortage := stdDev*lossFunction
	
	return expectedSales, expectedLeftover, expectedShortage
}

// calculateLossFunction 计算损失函数
func (io *InventoryOptimizer) calculateLossFunction(z float64) float64 {
	// 标准正态分布的损失函数
	// L(z) = φ(z) - z * (1 - Φ(z))
	// 其中φ(z)是标准正态分布的概率密度函数，Φ(z)是累积分布函数
	
	// 简化实现：使用近似公式
	if z < -8 {
		return -z
	}
	if z > 8 {
		return 0
	}
	
	// 使用近似公式
	phi := math.Exp(-z*z/2) / math.Sqrt(2*math.Pi)
	Phi := 0.5 * (1 + math.Erf(z/math.Sqrt(2)))
	
	loss := phi - z*(1-Phi)
	
	return loss
}

// getProductData 获取产品数据
func (io *InventoryOptimizer) getProductData(ctx context.Context, productID string) (*Product, error) {
	// TODO: 实现产品数据获取
	// 实际应该从仓储中获取
	
	return &Product{
		ID:    productID,
		Name:  "Sample Product",
		Cost:  100,
		Price: 150,
	}, nil
}

// getDemandData 获取需求数据
func (io *InventoryOptimizer) getDemandData(ctx context.Context, productID string) (*DemandData, error) {
	// 从需求仓储获取数据
	demandData, err := io.demandRepo.GetDemandData(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get demand data: %w", err)
	}
	
	return demandData, nil
}

// getCostData 获取成本数据
func (io *InventoryOptimizer) getCostData(ctx context.Context, productID string) (*CostData, error) {
	// 从成本仓储获取数据
	costData, err := io.costRepo.GetCostData(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost data: %w", err)
	}
	
	return costData, nil
}

// Data structures

type OptimizationResult struct {
	ProductID       string            `json:"product_id"`
	Goal            OptimizationGoal  `json:"goal"`
	Model           OptimizationModel `json:"model"`
	Result          interface{}       `json:"result"`
	GeneratedAt     time.Time         `json:"generated_at"`
}

type ROPResult struct {
	ReorderPoint    float64 `json:"reorder_point"`
	SafetyStock     float64 `json:"safety_stock"`
	LeadTimeDemand  float64 `json:"lead_time_demand"`
	ServiceLevel    float64 `json:"service_level"`
	ZScore          float64 `json:"z_score"`
}

type DemandData struct {
	ProductID       string           `json:"product_id"`
	History         []*DemandRecord  `json:"history"`
	Forecast        []*DemandForecast `json:"forecast"`
	Seasonality     *Seasonality     `json:"seasonality"`
	Trend           *Trend           `json:"trend"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type DemandRecord struct {
	Date            time.Time `json:"date"`
	Quantity        float64   `json:"quantity"`
	Timestamp       time.Time `json:"timestamp"`
}

type DemandForecast struct {
	Date            time.Time `json:"date"`
	Quantity        float64   `json:"quantity"`
	Confidence      float64   `json:"confidence"`
}

type Seasonality struct {
	Type            string    `json:"type"` // DAILY, WEEKLY, MONTHLY, YEARLY
	Pattern         []float64 `json:"pattern"`
	Strength        float64   `json:"strength"`
}

type Trend struct {
	Type            string    `json:"type"` // LINEAR, EXPONENTIAL, LOGARITHMIC
	Slope           float64   `json:"slope"`
	Intercept       float64   `json:"intercept"`
	R2              float64   `json:"r2"`
}

type CostData struct {
	ProductID       string    `json:"product_id"`
	UnitCost        float64   `json:"unit_cost"`
	OrderingCost    float64   `json:"ordering_cost"`
	HoldingCostRate float64   `json:"holding_cost_rate"`
	SellingPrice    float64   `json:"selling_price"`
	SalvageValue    float64   `json:"salvage_value"`
	ShortageCost    float64   `json:"shortage_cost"`
	LeadTime        int       `json:"lead_time"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Repository interfaces

type DemandRepository interface {
	GetDemandData(ctx context.Context, productID string) (*DemandData, error)
	SaveDemandData(ctx context.Context, data *DemandData) error
	UpdateDemandData(ctx context.Context, data *DemandData) error
	DeleteDemandData(ctx context.Context, productID string) error
}

type CostRepository interface {
	GetCostData(ctx context.Context, productID string) (*CostData, error)
	SaveCostData(ctx context.Context, data *CostData) error
	UpdateCostData(ctx context.Context, data *CostData) error
	DeleteCostData(ctx context.Context, productID string) error
}