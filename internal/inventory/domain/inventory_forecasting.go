package domain

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// ForecastMethod 预测方法
type ForecastMethod string

const (
	MethodMovingAverage        ForecastMethod = "MOVING_AVERAGE"        // 移动平均
	MethodExponentialSmoothing ForecastMethod = "EXPONENTIAL_SMOOTHING" // 指数平滑
	MethodARIMA                ForecastMethod = "ARIMA"                 // ARIMA
	MethodProphet              ForecastMethod = "PROPHET"               // Prophet
	MethodNeuralNetwork        ForecastMethod = "NEURAL_NETWORK"        // 神经网络
)

// ForecastPeriod 预测周期
type ForecastPeriod string

const (
	PeriodDaily     ForecastPeriod = "DAILY"     // 每日
	PeriodWeekly    ForecastPeriod = "WEEKLY"    // 每周
	PeriodMonthly   ForecastPeriod = "MONTHLY"   // 每月
	PeriodQuarterly ForecastPeriod = "QUARTERLY" // 每季度
	PeriodYearly    ForecastPeriod = "YEARLY"    // 每年
)

// ForecastResult 预测结果
type ForecastResult struct {
	ID             string         `json:"id"`
	ProductID      string         `json:"product_id"`
	ForecastMethod ForecastMethod `json:"forecast_method"`
	ForecastPeriod ForecastPeriod `json:"forecast_period"`

	// 预测数据
	ForecastData   []*ForecastPoint   `json:"forecast_data"`
	HistoricalData []*HistoricalPoint `json:"historical_data"`

	// 评估指标
	AccuracyMetrics *AccuracyMetrics `json:"accuracy_metrics"`
	ConfidenceLevel float64          `json:"confidence_level"`

	// 元数据
	GeneratedAt time.Time              `json:"generated_at"`
	ValidUntil  time.Time              `json:"valid_until"`
	Version     string                 `json:"version"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ForecastPoint 预测点
type ForecastPoint struct {
	Date          time.Time `json:"date"`
	ForecastValue float64   `json:"forecast_value"`
	LowerBound    float64   `json:"lower_bound"`
	UpperBound    float64   `json:"upper_bound"`
	Confidence    float64   `json:"confidence"`
}

// HistoricalPoint 历史点
type HistoricalPoint struct {
	Date            time.Time `json:"date"`
	ActualValue     float64   `json:"actual_value"`
	ForecastValue   float64   `json:"forecast_value"`
	Error           float64   `json:"error"`
	ErrorPercentage float64   `json:"error_percentage"`
}

// AccuracyMetrics 准确度指标
type AccuracyMetrics struct {
	MAE            float64 `json:"mae"`             // 平均绝对误差
	MSE            float64 `json:"mse"`             // 均方误差
	RMSE           float64 `json:"rmse"`            // 均方根误差
	MAPE           float64 `json:"mape"`            // 平均绝对百分比误差
	SMAPE          float64 `json:"smape"`           // 对称平均绝对百分比误差
	R2             float64 `json:"r2"`              // 决定系数
	MAD            float64 `json:"mad"`             // 平均绝对偏差
	TrackingSignal float64 `json:"tracking_signal"` // 跟踪信号
}

// ForecastEngine 预测引擎
type ForecastEngine struct {
	demandRepo   DemandRepository
	forecastRepo ForecastRepository
	mu           sync.RWMutex
	config       *ForecastConfig
	models       map[ForecastMethod]ForecastModel
}

// ForecastConfig 预测配置
type ForecastConfig struct {
	DefaultMethod    ForecastMethod `json:"default_method"`
	DefaultPeriod    ForecastPeriod `json:"default_period"`
	ForecastHorizon  int            `json:"forecast_horizon"`   // 预测步长
	MinHistoryLength int            `json:"min_history_length"` // 最小历史数据长度
	ConfidenceLevel  float64        `json:"confidence_level"`   // 置信水平
	SmoothingAlpha   float64        `json:"smoothing_alpha"`    // 平滑系数
	WindowSize       int            `json:"window_size"`        // 移动窗口大小
}

// ForecastModel 预测模型接口
type ForecastModel interface {
	Forecast(historicalData []float64, horizon int) ([]float64, error)
	Evaluate(actualData, forecastData []float64) *AccuracyMetrics
	GetConfidenceBounds(forecastData []float64, confidenceLevel float64) ([]float64, []float64)
}

// NewForecastEngine 创建预测引擎
func NewForecastEngine(demandRepo DemandRepository, forecastRepo ForecastRepository) *ForecastEngine {
	return &ForecastEngine{
		demandRepo:   demandRepo,
		forecastRepo: forecastRepo,
		config: &ForecastConfig{
			DefaultMethod:    MethodExponentialSmoothing,
			DefaultPeriod:    PeriodDaily,
			ForecastHorizon:  30,
			MinHistoryLength: 30,
			ConfidenceLevel:  0.95,
			SmoothingAlpha:   0.3,
			WindowSize:       7,
		},
		models: make(map[ForecastMethod]ForecastModel),
	}
}

// Initialize 初始化预测引擎
func (fe *ForecastEngine) Initialize(ctx context.Context) error {
	// 注册预测模型
	fe.registerModels()

	return nil
}

// registerModels 注册模型
func (fe *ForecastEngine) registerModels() {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	fe.models[MethodMovingAverage] = &MovingAverageModel{
		windowSize: fe.config.WindowSize,
	}

	fe.models[MethodExponentialSmoothing] = &ExponentialSmoothingModel{
		alpha: fe.config.SmoothingAlpha,
	}

	// 其他模型可以在这里注册
}

// GenerateForecast 生成预测
func (fe *ForecastEngine) GenerateForecast(ctx context.Context, productID string,
	method ForecastMethod, period ForecastPeriod, horizon int) (*ForecastResult, error) {

	// 获取历史数据
	historicalData, err := fe.getHistoricalData(ctx, productID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	// 检查数据量
	if len(historicalData) < fe.config.MinHistoryLength {
		return nil, fmt.Errorf("insufficient historical data: got %d, need at least %d",
			len(historicalData), fe.config.MinHistoryLength)
	}

	// 提取数值序列
	values := make([]float64, len(historicalData))
	dates := make([]time.Time, len(historicalData))

	for i, point := range historicalData {
		values[i] = point.Value
		dates[i] = point.Date
	}

	// 选择预测方法
	if method == "" {
		method = fe.config.DefaultMethod
	}

	// 选择预测模型
	model, exists := fe.models[method]
	if !exists {
		return nil, fmt.Errorf("forecast model not found: %s", method)
	}

	// 生成预测
	forecastValues, err := model.Forecast(values, horizon)
	if err != nil {
		return nil, fmt.Errorf("forecast failed: %w", err)
	}

	// 计算置信区间
	lowerBounds, upperBounds := model.GetConfidenceBounds(forecastValues, fe.config.ConfidenceLevel)

	// 生成预测日期
	forecastDates := fe.generateForecastDates(dates[len(dates)-1], period, horizon)

	// 创建预测结果
	result := &ForecastResult{
		ID:             generateForecastID(),
		ProductID:      productID,
		ForecastMethod: method,
		ForecastPeriod: period,
		GeneratedAt:    time.Now(),
		ValidUntil:     time.Now().Add(30 * 24 * time.Hour), // 30天有效期
		Version:        "1.0",
		Metadata:       make(map[string]interface{}),
	}

	// 创建预测点
	for i := 0; i < horizon; i++ {
		forecastPoint := &ForecastPoint{
			Date:          forecastDates[i],
			ForecastValue: forecastValues[i],
			LowerBound:    lowerBounds[i],
			UpperBound:    upperBounds[i],
			Confidence:    fe.config.ConfidenceLevel,
		}
		result.ForecastData = append(result.ForecastData, forecastPoint)
	}

	// 创建历史点（用于评估）
	for _, point := range historicalData {
		historicalPoint := &HistoricalPoint{
			Date:        point.Date,
			ActualValue: point.Value,
		}
		result.HistoricalData = append(result.HistoricalData, historicalPoint)
	}

	// 评估预测准确度
	fe.evaluateForecast(result, historicalData, forecastValues[:len(historicalData)])

	// 保存预测结果
	err = fe.forecastRepo.SaveForecast(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to save forecast: %w", err)
	}

	return result, nil
}

// getHistoricalData 获取历史数据
func (fe *ForecastEngine) getHistoricalData(ctx context.Context, productID string, period ForecastPeriod) ([]*DemandPoint, error) {
	// 获取需求数据
	demandData, err := fe.demandRepo.GetDemandData(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get demand data: %w", err)
	}

	// 根据周期聚合数据
	aggregatedData := fe.aggregateByPeriod(demandData.History, period)

	return aggregatedData, nil
}

// aggregateByPeriod 按周期聚合数据
func (fe *ForecastEngine) aggregateByPeriod(history []*DemandRecord, period ForecastPeriod) []*DemandPoint {
	if len(history) == 0 {
		return []*DemandPoint{}
	}

	// 按日期分组
	groupedData := make(map[string]*DemandPoint)

	for _, record := range history {
		// 根据周期确定分组键
		var groupKey string
		switch period {
		case PeriodDaily:
			groupKey = record.Date.Format("2006-01-02")
		case PeriodWeekly:
			year, week := record.Date.ISOWeek()
			groupKey = fmt.Sprintf("%d-W%02d", year, week)
		case PeriodMonthly:
			groupKey = record.Date.Format("2006-01")
		case PeriodQuarterly:
			quarter := (int(record.Date.Month())-1)/3 + 1
			groupKey = fmt.Sprintf("%d-Q%d", record.Date.Year(), quarter)
		case PeriodYearly:
			groupKey = fmt.Sprintf("%d", record.Date.Year())
		default:
			groupKey = record.Date.Format("2006-01-02")
		}

		if _, exists := groupedData[groupKey]; !exists {
			groupedData[groupKey] = &DemandPoint{
				Date:  record.Date,
				Value: 0,
			}
		}

		groupedData[groupKey].Value += record.Quantity
	}

	// 转换为切片并排序
	var result []*DemandPoint
	for _, point := range groupedData {
		result = append(result, point)
	}

	// 按日期排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Date.After(result[j].Date) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// generateForecastDates 生成预测日期
func (fe *ForecastEngine) generateForecastDates(lastDate time.Time, period ForecastPeriod, horizon int) []time.Time {
	var dates []time.Time

	currentDate := lastDate

	for i := 0; i < horizon; i++ {
		// 根据周期增加日期
		switch period {
		case PeriodDaily:
			currentDate = currentDate.AddDate(0, 0, 1)
		case PeriodWeekly:
			currentDate = currentDate.AddDate(0, 0, 7)
		case PeriodMonthly:
			currentDate = currentDate.AddDate(0, 1, 0)
		case PeriodQuarterly:
			currentDate = currentDate.AddDate(0, 3, 0)
		case PeriodYearly:
			currentDate = currentDate.AddDate(1, 0, 0)
		default:
			currentDate = currentDate.AddDate(0, 0, 1)
		}

		dates = append(dates, currentDate)
	}

	return dates
}

// evaluateForecast 评估预测
func (fe *ForecastEngine) evaluateForecast(result *ForecastResult, historicalData []*DemandPoint, forecastValues []float64) {
	if len(historicalData) == 0 || len(forecastValues) == 0 {
		return
	}

	// 提取实际值
	actualValues := make([]float64, len(historicalData))
	for i, point := range historicalData {
		actualValues[i] = point.Value
	}

	// 选择模型进行评估
	model, exists := fe.models[result.ForecastMethod]
	if !exists {
		return
	}

	// 计算评估指标
	metrics := model.Evaluate(actualValues, forecastValues)
	result.AccuracyMetrics = metrics
	result.ConfidenceLevel = fe.config.ConfidenceLevel

	// 更新历史点的预测值和误差
	for i := 0; i < len(historicalData) && i < len(forecastValues); i++ {
		result.HistoricalData[i].ForecastValue = forecastValues[i]
		result.HistoricalData[i].Error = actualValues[i] - forecastValues[i]

		if actualValues[i] != 0 {
			result.HistoricalData[i].ErrorPercentage = math.Abs(result.HistoricalData[i].Error / actualValues[i] * 100)
		}
	}
}

// UpdateForecast 更新预测
func (fe *ForecastEngine) UpdateForecast(ctx context.Context, forecastID string, newData []*DemandRecord) (*ForecastResult, error) {
	// 获取现有预测
	existingForecast, err := fe.forecastRepo.GetForecast(ctx, forecastID)
	if err != nil {
		return nil, fmt.Errorf("failed to get forecast: %w", err)
	}

	// 重新生成预测
	updatedForecast, err := fe.GenerateForecast(ctx, existingForecast.ProductID,
		existingForecast.ForecastMethod, existingForecast.ForecastPeriod,
		len(existingForecast.ForecastData))

	if err != nil {
		return nil, fmt.Errorf("failed to regenerate forecast: %w", err)
	}

	// 保留原始ID
	updatedForecast.ID = forecastID

	// 保存更新
	err = fe.forecastRepo.UpdateForecast(ctx, updatedForecast)
	if err != nil {
		return nil, fmt.Errorf("failed to update forecast: %w", err)
	}

	return updatedForecast, nil
}

// CompareMethods 比较预测方法
func (fe *ForecastEngine) CompareMethods(ctx context.Context, productID string,
	methods []ForecastMethod, period ForecastPeriod, horizon int) ([]*MethodComparison, error) {

	var comparisons []*MethodComparison

	for _, method := range methods {
		// 生成预测
		forecast, err := fe.GenerateForecast(ctx, productID, method, period, horizon)
		if err != nil {
			fmt.Printf("Failed to generate forecast with method %s: %v\n", method, err)
			continue
		}

		comparison := &MethodComparison{
			Method:          method,
			AccuracyMetrics: forecast.AccuracyMetrics,
			ConfidenceLevel: forecast.ConfidenceLevel,
			GeneratedAt:     forecast.GeneratedAt,
		}

		comparisons = append(comparisons, comparison)
	}

	return comparisons, nil
}

// GetBestMethod 获取最佳预测方法
func (fe *ForecastEngine) GetBestMethod(ctx context.Context, productID string,
	period ForecastPeriod) (ForecastMethod, *AccuracyMetrics, error) {

	// 所有可用方法
	methods := []ForecastMethod{
		MethodMovingAverage,
		MethodExponentialSmoothing,
		// 可以添加更多方法
	}

	// 比较方法
	comparisons, err := fe.CompareMethods(ctx, productID, methods, period, 30)
	if err != nil {
		return "", nil, fmt.Errorf("failed to compare methods: %w", err)
	}

	if len(comparisons) == 0 {
		return fe.config.DefaultMethod, nil, nil
	}

	// 选择MAPE最小的方法
	bestMethod := comparisons[0].Method
	bestMetrics := comparisons[0].AccuracyMetrics

	for _, comparison := range comparisons[1:] {
		if comparison.AccuracyMetrics != nil && bestMetrics != nil {
			if comparison.AccuracyMetrics.MAPE < bestMetrics.MAPE {
				bestMethod = comparison.Method
				bestMetrics = comparison.AccuracyMetrics
			}
		}
	}

	return bestMethod, bestMetrics, nil
}

// MovingAverageModel 移动平均模型
type MovingAverageModel struct {
	windowSize int
}

// Forecast 预测
func (mam *MovingAverageModel) Forecast(historicalData []float64, horizon int) ([]float64, error) {
	if len(historicalData) < mam.windowSize {
		return nil, fmt.Errorf("insufficient data for moving average")
	}

	var forecast []float64

	// 计算最后一个窗口的平均值
	var lastWindowSum float64
	for i := len(historicalData) - mam.windowSize; i < len(historicalData); i++ {
		lastWindowSum += historicalData[i]
	}
	lastWindowAvg := lastWindowSum / float64(mam.windowSize)

	// 使用最后一个窗口的平均值作为所有未来预测
	for i := 0; i < horizon; i++ {
		forecast = append(forecast, lastWindowAvg)
	}

	return forecast, nil
}

// Evaluate 评估
func (mam *MovingAverageModel) Evaluate(actualData, forecastData []float64) *AccuracyMetrics {
	if len(actualData) != len(forecastData) || len(actualData) == 0 {
		return nil
	}

	metrics := &AccuracyMetrics{}
	n := float64(len(actualData))

	var sumAE, sumSE, sumAPE float64
	var sumActual, sumForecast float64

	for i := 0; i < len(actualData); i++ {
		error := actualData[i] - forecastData[i]
		absError := math.Abs(error)
		percentageError := math.Abs(error/actualData[i]) * 100

		sumAE += absError
		sumSE += error * error
		sumAPE += percentageError

		sumActual += actualData[i]
		sumForecast += forecastData[i]
	}

	// 计算指标
	metrics.MAE = sumAE / n
	metrics.MSE = sumSE / n
	metrics.RMSE = math.Sqrt(metrics.MSE)
	metrics.MAPE = sumAPE / n

	// 计算SMAPE
	var sumSmape float64
	for i := 0; i < len(actualData); i++ {
		denominator := (math.Abs(actualData[i]) + math.Abs(forecastData[i])) / 2
		if denominator != 0 {
			sumSmape += math.Abs(actualData[i]-forecastData[i]) / denominator
		}
	}
	metrics.SMAPE = (sumSmape / n) * 100

	// 计算R2
	meanActual := sumActual / n
	var ssTotal, ssResidual float64
	for i := 0; i < len(actualData); i++ {
		ssTotal += (actualData[i] - meanActual) * (actualData[i] - meanActual)
		ssResidual += (actualData[i] - forecastData[i]) * (actualData[i] - forecastData[i])
	}

	if ssTotal != 0 {
		metrics.R2 = 1 - (ssResidual / ssTotal)
	}

	// 计算MAD
	metrics.MAD = metrics.MAE

	// 计算跟踪信号
	metrics.TrackingSignal = sumAE / metrics.MAD

	return metrics
}

// GetConfidenceBounds 获取置信区间
func (mam *MovingAverageModel) GetConfidenceBounds(forecastData []float64, confidenceLevel float64) ([]float64, []float64) {
	// 简化实现：使用固定百分比
	lowerBounds := make([]float64, len(forecastData))
	upperBounds := make([]float64, len(forecastData))

	for i, value := range forecastData {
		margin := value * 0.2 // 20%的边际
		lowerBounds[i] = math.Max(0, value-margin)
		upperBounds[i] = value + margin
	}

	return lowerBounds, upperBounds
}

// ExponentialSmoothingModel 指数平滑模型
type ExponentialSmoothingModel struct {
	alpha float64
}

// Forecast 预测
func (esm *ExponentialSmoothingModel) Forecast(historicalData []float64, horizon int) ([]float64, error) {
	if len(historicalData) == 0 {
		return nil, fmt.Errorf("insufficient data for exponential smoothing")
	}

	var forecast []float64

	// 计算初始预测
	lastValue := historicalData[len(historicalData)-1]

	// 简单指数平滑
	for i := 0; i < horizon; i++ {
		forecast = append(forecast, lastValue)
	}

	return forecast, nil
}

// Evaluate 评估
func (esm *ExponentialSmoothingModel) Evaluate(actualData, forecastData []float64) *AccuracyMetrics {
	// 重用移动平均模型的评估方法
	mam := &MovingAverageModel{windowSize: 1}
	return mam.Evaluate(actualData, forecastData)
}

// GetConfidenceBounds 获取置信区间
func (esm *ExponentialSmoothingModel) GetConfidenceBounds(forecastData []float64, confidenceLevel float64) ([]float64, []float64) {
	// 简化实现：使用固定百分比
	lowerBounds := make([]float64, len(forecastData))
	upperBounds := make([]float64, len(forecastData))

	for i, value := range forecastData {
		margin := value * 0.15 // 15%的边际
		lowerBounds[i] = math.Max(0, value-margin)
		upperBounds[i] = value + margin
	}

	return lowerBounds, upperBounds
}

// Data structures

type DemandPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type MethodComparison struct {
	Method          ForecastMethod   `json:"method"`
	AccuracyMetrics *AccuracyMetrics `json:"accuracy_metrics"`
	ConfidenceLevel float64          `json:"confidence_level"`
	GeneratedAt     time.Time        `json:"generated_at"`
}

// Repository interfaces

type ForecastRepository interface {
	SaveForecast(ctx context.Context, forecast *ForecastResult) error
	GetForecast(ctx context.Context, forecastID string) (*ForecastResult, error)
	GetForecastsByProduct(ctx context.Context, productID string, startDate, endDate time.Time) ([]*ForecastResult, error)
	GetLatestForecast(ctx context.Context, productID string) (*ForecastResult, error)
	UpdateForecast(ctx context.Context, forecast *ForecastResult) error
	DeleteForecast(ctx context.Context, forecastID string) error
}

// Helper functions

func generateForecastID() string {
	return fmt.Sprintf("FORECAST_%d", time.Now().UnixNano())
}
