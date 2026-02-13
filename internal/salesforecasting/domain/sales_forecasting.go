package domain

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrModelNotFound      = errors.New("forecast model not found")
	ErrPredictionNotFound = errors.New("prediction not found")
	ErrModelNotReady      = errors.New("model is not ready for prediction")
	ErrInvalidDateRange   = errors.New("invalid date range")
	ErrInsufficientData   = errors.New("insufficient historical data")
)

type ModelType string

const (
	ModelTypeMovingAverage       ModelType = "MOVING_AVERAGE"
	ModelTypeExponentialSmoothing ModelType = "EXPONENTIAL_SMOOTHING"
	ModelTypeARIMA               ModelType = "ARIMA"
	ModelTypeProphet             ModelType = "PROPHET"
	ModelTypeLSTM                ModelType = "LSTM"
	ModelTypeEnsemble            ModelType = "ENSEMBLE"
)

type ModelStatus string

const (
	ModelStatusDraft    ModelStatus = "DRAFT"
	ModelStatusTraining ModelStatus = "TRAINING"
	ModelStatusReady    ModelStatus = "READY"
	ModelStatusFailed   ModelStatus = "FAILED"
)

type Granularity string

const (
	GranularityDaily    Granularity = "DAILY"
	GranularityWeekly   Granularity = "WEEKLY"
	GranularityMonthly  Granularity = "MONTHLY"
	GranularityQuarterly Granularity = "QUARTERLY"
)

type ForecastModel struct {
	ID             string            `json:"id"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	ModelType      ModelType         `json:"model_type"`
	Status         ModelStatus       `json:"status"`
	Granularity    Granularity       `json:"granularity"`
	ProductIDs     []string          `json:"product_ids"`
	CategoryIDs    []string          `json:"category_ids"`
	LookbackDays   int               `json:"lookback_days"`
	ForecastDays   int               `json:"forecast_days"`
	Parameters     map[string]string `json:"parameters"`
	TrainingRMSE   float64           `json:"training_rmse"`
	TrainingMAE    float64           `json:"training_mae"`
	LastTrainedAt  *time.Time        `json:"last_trained_at"`
}

type SalesPrediction struct {
	ID              string             `json:"id"`
	CreatedAt       time.Time          `json:"created_at"`
	ModelID         string             `json:"model_id"`
	ForecastStart   time.Time          `json:"forecast_start"`
	ForecastEnd     time.Time          `json:"forecast_end"`
	Items           []*PredictionItem  `json:"items"`
	Metadata        map[string]string  `json:"metadata"`
}

type PredictionItem struct {
	ProductID         string    `json:"product_id"`
	CategoryID        string    `json:"category_id"`
	Date              time.Time `json:"date"`
	PredictedQuantity float64   `json:"predicted_quantity"`
	PredictedRevenue  float64   `json:"predicted_revenue"`
	LowerBound        float64   `json:"lower_bound"`
	UpperBound        float64   `json:"upper_bound"`
	Confidence        float64   `json:"confidence"`
	ActualQuantity    *float64  `json:"actual_quantity"`
	ActualRevenue     *float64  `json:"actual_revenue"`
}

type ForecastAccuracy struct {
	ID               string              `json:"id"`
	ModelID          string              `json:"model_id"`
	CalculatedAt     time.Time           `json:"calculated_at"`
	MAPE             float64             `json:"mape"`
	RMSE             float64             `json:"rmse"`
	MAE              float64             `json:"mae"`
	MPE              float64             `json:"mpe"`
	TotalPredictions int                 `json:"total_predictions"`
	AccuracyByProduct []*AccuracyByProduct `json:"accuracy_by_product"`
}

type AccuracyByProduct struct {
	ProductID       string  `json:"product_id"`
	MAPE            float64 `json:"mape"`
	RMSE            float64 `json:"rmse"`
	PredictionCount int     `json:"prediction_count"`
}

type DemandForecast struct {
	ID                 string              `json:"id"`
	CreatedAt          time.Time           `json:"created_at"`
	ForecastStart      time.Time           `json:"forecast_start"`
	ForecastEnd        time.Time           `json:"forecast_end"`
	Items              []*DemandItem       `json:"items"`
	SeasonalityFactors []*SeasonalityFactor `json:"seasonality_factors"`
	PromotionEffects   []*PromotionEffect  `json:"promotion_effects"`
}

type DemandItem struct {
	ProductID           string    `json:"product_id"`
	Date                time.Time `json:"date"`
	BaseDemand          float64   `json:"base_demand"`
	SeasonalAdjustment  float64   `json:"seasonal_adjustment"`
	PromotionAdjustment float64   `json:"promotion_adjustment"`
	FinalDemand         float64   `json:"final_demand"`
	SafetyStock         float64   `json:"safety_stock"`
	ReorderPoint        float64   `json:"reorder_point"`
}

type SeasonalityFactor struct {
	Period      string  `json:"period"`
	Factor      float64 `json:"factor"`
	Description string  `json:"description"`
}

type PromotionEffect struct {
	PromotionType string  `json:"promotion_type"`
	LiftFactor    float64 `json:"lift_factor"`
	DurationDays  int     `json:"duration_days"`
}

type HistoricalDataPoint struct {
	ProductID string    `json:"product_id"`
	Date      time.Time `json:"date"`
	Quantity  float64   `json:"quantity"`
	Revenue   float64   `json:"revenue"`
}

func NewForecastModel(name string, modelType ModelType, granularity Granularity) *ForecastModel {
	return &ForecastModel{
		Name:        name,
		ModelType:   modelType,
		Status:      ModelStatusDraft,
		Granularity: granularity,
		ProductIDs:  []string{},
		CategoryIDs: []string{},
		Parameters:  make(map[string]string),
		LookbackDays: 90,
		ForecastDays: 30,
	}
}

func (m *ForecastModel) SetStatus(status ModelStatus) {
	m.Status = status
	m.UpdatedAt = time.Now()
}

func (m *ForecastModel) SetTrainingResult(rmse, mae float64) {
	m.TrainingRMSE = rmse
	m.TrainingMAE = mae
	now := time.Now()
	m.LastTrainedAt = &now
	m.Status = ModelStatusReady
	m.UpdatedAt = now
}

func (m *ForecastModel) IsReady() bool {
	return m.Status == ModelStatusReady
}

func NewSalesPrediction(modelID string, forecastStart, forecastEnd time.Time) *SalesPrediction {
	return &SalesPrediction{
		ModelID:       modelID,
		ForecastStart: forecastStart,
		ForecastEnd:   forecastEnd,
		Items:         []*PredictionItem{},
		Metadata:      make(map[string]string),
	}
}

func (p *SalesPrediction) AddItem(item *PredictionItem) {
	p.Items = append(p.Items, item)
}

func (p *SalesPrediction) CalculateAccuracy() *ForecastAccuracy {
	accuracy := &ForecastAccuracy{
		ModelID:          p.ModelID,
		CalculatedAt:     time.Now(),
		AccuracyByProduct: []*AccuracyByProduct{},
	}

	var sumAPE, sumSE, sumAE, sumPE float64
	var count int
	productStats := make(map[string]*AccuracyByProduct)

	for _, item := range p.Items {
		if item.ActualQuantity == nil || *item.ActualQuantity == 0 {
			continue
		}

		actual := *item.ActualQuantity
		predicted := item.PredictedQuantity

		ape := math.Abs((actual - predicted) / actual) * 100
		se := math.Pow(actual - predicted, 2)
		ae := math.Abs(actual - predicted)
		pe := ((actual - predicted) / actual) * 100

		sumAPE += ape
		sumSE += se
		sumAE += ae
		sumPE += pe
		count++

		if _, exists := productStats[item.ProductID]; !exists {
			productStats[item.ProductID] = &AccuracyByProduct{
				ProductID: item.ProductID,
			}
		}
		productStats[item.ProductID].MAPE += ape
		productStats[item.ProductID].RMSE += se
		productStats[item.ProductID].PredictionCount++
	}

	if count > 0 {
		accuracy.MAPE = sumAPE / float64(count)
		accuracy.RMSE = math.Sqrt(sumSE / float64(count))
		accuracy.MAE = sumAE / float64(count)
		accuracy.MPE = sumPE / float64(count)
		accuracy.TotalPredictions = count
	}

	for _, stats := range productStats {
		if stats.PredictionCount > 0 {
			stats.MAPE = stats.MAPE / float64(stats.PredictionCount)
			stats.RMSE = math.Sqrt(stats.RMSE / float64(stats.PredictionCount))
		}
		accuracy.AccuracyByProduct = append(accuracy.AccuracyByProduct, stats)
	}

	return accuracy
}

func NewDemandForecast(forecastStart, forecastEnd time.Time) *DemandForecast {
	return &DemandForecast{
		ForecastStart:      forecastStart,
		ForecastEnd:        forecastEnd,
		Items:              []*DemandItem{},
		SeasonalityFactors: []*SeasonalityFactor{},
		PromotionEffects:   []*PromotionEffect{},
	}
}

func (d *DemandItem) CalculateFinalDemand() {
	d.FinalDemand = d.BaseDemand * (1 + d.SeasonalAdjustment) * (1 + d.PromotionAdjustment)
}

func (d *DemandItem) CalculateSafetyStock(serviceLevel float64, leadTimeDays int, demandStdDev float64) {
	zScore := getZScore(serviceLevel)
	d.SafetyStock = zScore * demandStdDev * math.Sqrt(float64(leadTimeDays))
	d.ReorderPoint = d.FinalDemand*float64(leadTimeDays) + d.SafetyStock
}

func getZScore(serviceLevel float64) float64 {
	zScores := map[float64]float64{
		0.90: 1.28,
		0.95: 1.65,
		0.99: 2.33,
	}
	if z, ok := zScores[serviceLevel]; ok {
		return z
	}
	return 1.65
}

type ForecastModelRepository interface {
	Save(ctx context.Context, model *ForecastModel) error
	FindByID(ctx context.Context, id string) (*ForecastModel, error)
	FindByStatus(ctx context.Context, status ModelStatus) ([]*ForecastModel, error)
	FindAll(ctx context.Context, limit, offset int) ([]*ForecastModel, int64, error)
	Update(ctx context.Context, model *ForecastModel) error
}

type SalesPredictionRepository interface {
	Save(ctx context.Context, prediction *SalesPrediction) error
	FindByID(ctx context.Context, id string) (*SalesPrediction, error)
	FindByModelID(ctx context.Context, modelID string, limit, offset int) ([]*SalesPrediction, error)
	Update(ctx context.Context, prediction *SalesPrediction) error
}

type ForecastAccuracyRepository interface {
	Save(ctx context.Context, accuracy *ForecastAccuracy) error
	FindByModelID(ctx context.Context, modelID string) (*ForecastAccuracy, error)
}

type HistoricalDataRepository interface {
	GetSalesData(ctx context.Context, productIDs []string, start, end time.Time) ([]*HistoricalDataPoint, error)
	GetAggregatedData(ctx context.Context, productIDs []string, start, end time.Time, granularity Granularity) ([]*HistoricalDataPoint, error)
}

type ForecastAlgorithm interface {
	Train(data []*HistoricalDataPoint, params map[string]string) error
	Predict(startDate, endDate time.Time, productIDs []string) ([]*PredictionItem, error)
	GetName() ModelType
}

type ForecastService interface {
	CreateModel(ctx context.Context, model *ForecastModel) error
	TrainModel(ctx context.Context, modelID string, start, end time.Time) (*ForecastModel, error)
	Predict(ctx context.Context, modelID string, start, end time.Time, productIDs []string) (*SalesPrediction, error)
	CalculateAccuracy(ctx context.Context, predictionID string) (*ForecastAccuracy, error)
	GenerateDemandForecast(ctx context.Context, productIDs []string, days int) (*DemandForecast, error)
}

type MovingAverageAlgorithm struct {
	windowSize int
	data       []*HistoricalDataPoint
}

func NewMovingAverageAlgorithm(windowSize int) *MovingAverageAlgorithm {
	return &MovingAverageAlgorithm{windowSize: windowSize}
}

func (a *MovingAverageAlgorithm) Train(data []*HistoricalDataPoint, params map[string]string) error {
	if len(data) < a.windowSize {
		return fmt.Errorf("insufficient data: need at least %d points, got %d", a.windowSize, len(data))
	}
	a.data = data
	return nil
}

func (a *MovingAverageAlgorithm) Predict(startDate, endDate time.Time, productIDs []string) ([]*PredictionItem, error) {
	var items []*PredictionItem

	dataByProduct := make(map[string][]*HistoricalDataPoint)
	for _, d := range a.data {
		dataByProduct[d.ProductID] = append(dataByProduct[d.ProductID], d)
	}

	for productID, productData := range dataByProduct {
		if len(productIDs) > 0 {
			found := false
			for _, pid := range productIDs {
				if pid == productID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		var sum float64
		start := len(productData) - a.windowSize
		if start < 0 {
			start = 0
		}
		for i := start; i < len(productData); i++ {
			sum += productData[i].Quantity
		}
		avg := sum / float64(a.windowSize)

		stdDev := calculateStdDev(productData[start:], avg)

		current := startDate
		for !current.After(endDate) {
			items = append(items, &PredictionItem{
				ProductID:         productID,
				Date:              current,
				PredictedQuantity: avg,
				LowerBound:        avg - 1.96*stdDev,
				UpperBound:        avg + 1.96*stdDev,
				Confidence:        0.95,
			})
			current = current.AddDate(0, 0, 1)
		}
	}

	return items, nil
}

func (a *MovingAverageAlgorithm) GetName() ModelType {
	return ModelTypeMovingAverage
}

func calculateStdDev(data []*HistoricalDataPoint, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}
	var sumSquares float64
	for _, d := range data {
		diff := d.Quantity - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(data)))
}

type ExponentialSmoothingAlgorithm struct {
	alpha float64
	data  []*HistoricalDataPoint
}

func NewExponentialSmoothingAlgorithm(alpha float64) *ExponentialSmoothingAlgorithm {
	return &ExponentialSmoothingAlgorithm{alpha: alpha}
}

func (a *ExponentialSmoothingAlgorithm) Train(data []*HistoricalDataPoint, params map[string]string) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data: need at least 2 points")
	}
	a.data = data
	return nil
}

func (a *ExponentialSmoothingAlgorithm) Predict(startDate, endDate time.Time, productIDs []string) ([]*PredictionItem, error) {
	var items []*PredictionItem

	dataByProduct := make(map[string][]*HistoricalDataPoint)
	for _, d := range a.data {
		dataByProduct[d.ProductID] = append(dataByProduct[d.ProductID], d)
	}

	for productID, productData := range dataByProduct {
		if len(productIDs) > 0 {
			found := false
			for _, pid := range productIDs {
				if pid == productID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if len(productData) == 0 {
			continue
		}

		smoothed := productData[0].Quantity
		for i := 1; i < len(productData); i++ {
			smoothed = a.alpha*productData[i].Quantity + (1-a.alpha)*smoothed
		}

		variance := calculateVariance(productData, smoothed)

		current := startDate
		for !current.After(endDate) {
			items = append(items, &PredictionItem{
				ProductID:         productID,
				Date:              current,
				PredictedQuantity: smoothed,
				LowerBound:        smoothed - 1.96*math.Sqrt(variance),
				UpperBound:        smoothed + 1.96*math.Sqrt(variance),
				Confidence:        0.95,
			})
			current = current.AddDate(0, 0, 1)
		}
	}

	return items, nil
}

func (a *ExponentialSmoothingAlgorithm) GetName() ModelType {
	return ModelTypeExponentialSmoothing
}

func calculateVariance(data []*HistoricalDataPoint, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}
	var sumSquares float64
	for _, d := range data {
		diff := d.Quantity - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(data))
}
