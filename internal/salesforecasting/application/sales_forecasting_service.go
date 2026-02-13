package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/salesforecasting/domain"
	"github.com/wyfcoding/pkg/idgen"
)

type SalesForecastingService struct {
	modelRepo      domain.ForecastModelRepository
	predictionRepo domain.SalesPredictionRepository
	accuracyRepo   domain.ForecastAccuracyRepository
	historicalRepo domain.HistoricalDataRepository
	idGenerator    idgen.Generator
	logger         *slog.Logger
	algorithms     map[domain.ModelType]domain.ForecastAlgorithm
}

func NewSalesForecastingService(
	modelRepo domain.ForecastModelRepository,
	predictionRepo domain.SalesPredictionRepository,
	accuracyRepo domain.ForecastAccuracyRepository,
	historicalRepo domain.HistoricalDataRepository,
	idGenerator idgen.Generator,
	logger *slog.Logger,
) *SalesForecastingService {
	svc := &SalesForecastingService{
		modelRepo:      modelRepo,
		predictionRepo: predictionRepo,
		accuracyRepo:   accuracyRepo,
		historicalRepo: historicalRepo,
		idGenerator:    idGenerator,
		logger:         logger,
		algorithms:     make(map[domain.ModelType]domain.ForecastAlgorithm),
	}

	svc.algorithms[domain.ModelTypeMovingAverage] = domain.NewMovingAverageAlgorithm(7)
	svc.algorithms[domain.ModelTypeExponentialSmoothing] = domain.NewExponentialSmoothingAlgorithm(0.3)

	return svc
}

type CreateModelCommand struct {
	Name         string
	Description  string
	ModelType    domain.ModelType
	Granularity  domain.Granularity
	ProductIDs   []string
	CategoryIDs  []string
	LookbackDays int
	ForecastDays int
	Parameters   map[string]string
}

func (s *SalesForecastingService) CreateModel(ctx context.Context, cmd *CreateModelCommand) (*domain.ForecastModel, error) {
	s.logger.InfoContext(ctx, "creating forecast model", "name", cmd.Name, "type", cmd.ModelType)

	model := domain.NewForecastModel(cmd.Name, cmd.ModelType, cmd.Granularity)
	model.ID = fmt.Sprintf("FM%d", s.idGenerator.Generate())
	model.Description = cmd.Description
	model.ProductIDs = cmd.ProductIDs
	model.CategoryIDs = cmd.CategoryIDs
	model.Parameters = cmd.Parameters

	if cmd.LookbackDays > 0 {
		model.LookbackDays = cmd.LookbackDays
	}
	if cmd.ForecastDays > 0 {
		model.ForecastDays = cmd.ForecastDays
	}

	if err := s.modelRepo.Save(ctx, model); err != nil {
		return nil, fmt.Errorf("failed to save model: %w", err)
	}

	s.logger.InfoContext(ctx, "forecast model created", "id", model.ID)
	return model, nil
}

func (s *SalesForecastingService) GetModel(ctx context.Context, modelID string) (*domain.ForecastModel, error) {
	return s.modelRepo.FindByID(ctx, modelID)
}

func (s *SalesForecastingService) ListModels(ctx context.Context, status domain.ModelStatus, page, pageSize int) ([]*domain.ForecastModel, int64, error) {
	offset := (page - 1) * pageSize
	if status != "" {
		models, err := s.modelRepo.FindByStatus(ctx, status)
		return models, int64(len(models)), err
	}
	return s.modelRepo.FindAll(ctx, pageSize, offset)
}

type TrainModelCommand struct {
	ModelID   string
	StartDate time.Time
	EndDate   time.Time
}

type TrainModelResult struct {
	ModelID      string
	Status       domain.ModelStatus
	TrainingRMSE float64
	TrainingMAE  float64
	Message      string
}

func (s *SalesForecastingService) TrainModel(ctx context.Context, cmd *TrainModelCommand) (*TrainModelResult, error) {
	s.logger.InfoContext(ctx, "training model", "model_id", cmd.ModelID)

	model, err := s.modelRepo.FindByID(ctx, cmd.ModelID)
	if err != nil {
		return nil, err
	}

	model.SetStatus(domain.ModelStatusTraining)
	if err := s.modelRepo.Update(ctx, model); err != nil {
		return nil, err
	}

	data, err := s.historicalRepo.GetSalesData(ctx, model.ProductIDs, cmd.StartDate, cmd.EndDate)
	if err != nil {
		model.SetStatus(domain.ModelStatusFailed)
		_ = s.modelRepo.Update(ctx, model)
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(data) < 10 {
		model.SetStatus(domain.ModelStatusFailed)
		_ = s.modelRepo.Update(ctx, model)
		return nil, domain.ErrInsufficientData
	}

	algorithm, exists := s.algorithms[model.ModelType]
	if !exists {
		model.SetStatus(domain.ModelStatusFailed)
		_ = s.modelRepo.Update(ctx, model)
		return nil, fmt.Errorf("unsupported model type: %s", model.ModelType)
	}

	if err := algorithm.Train(data, model.Parameters); err != nil {
		model.SetStatus(domain.ModelStatusFailed)
		_ = s.modelRepo.Update(ctx, model)
		return nil, fmt.Errorf("training failed: %w", err)
	}

	rmse, mae := s.calculateTrainingMetrics(data, algorithm)

	model.SetTrainingResult(rmse, mae)
	if err := s.modelRepo.Update(ctx, model); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "model training completed", "model_id", cmd.ModelID, "rmse", rmse, "mae", mae)

	return &TrainModelResult{
		ModelID:      model.ID,
		Status:       model.Status,
		TrainingRMSE: rmse,
		TrainingMAE:  mae,
		Message:      "Training completed successfully",
	}, nil
}

func (s *SalesForecastingService) calculateTrainingMetrics(data []*domain.HistoricalDataPoint, algorithm domain.ForecastAlgorithm) (float64, float64) {
	if len(data) < 2 {
		return 0, 0
	}

	trainSize := int(float64(len(data)) * 0.8)
	trainData := data[:trainSize]
	testData := data[trainSize:]

	_ = algorithm.Train(trainData, nil)

	predictions, err := algorithm.Predict(testData[0].Date, testData[len(testData)-1].Date, nil)
	if err != nil {
		return 0, 0
	}

	predMap := make(map[string]map[time.Time]float64)
	for _, p := range predictions {
		if predMap[p.ProductID] == nil {
			predMap[p.ProductID] = make(map[time.Time]float64)
		}
		predMap[p.ProductID][p.Date] = p.PredictedQuantity
	}

	var sumSE, sumAE float64
	var count int

	for _, d := range testData {
		if predByDate, ok := predMap[d.ProductID]; ok {
			if pred, ok := predByDate[d.Date]; ok {
				diff := d.Quantity - pred
				sumSE += diff * diff
				sumAE += abs(diff)
				count++
			}
		}
	}

	if count == 0 {
		return 0, 0
	}

	rmse := sqrt(sumSE / float64(count))
	mae := sumAE / float64(count)

	return rmse, mae
}

type PredictSalesCommand struct {
	ModelID                  string
	StartDate                time.Time
	EndDate                  time.Time
	ProductIDs               []string
	IncludeConfidenceInterval bool
}

func (s *SalesForecastingService) PredictSales(ctx context.Context, cmd *PredictSalesCommand) (*domain.SalesPrediction, error) {
	s.logger.InfoContext(ctx, "predicting sales", "model_id", cmd.ModelID)

	model, err := s.modelRepo.FindByID(ctx, cmd.ModelID)
	if err != nil {
		return nil, err
	}

	if !model.IsReady() {
		return nil, domain.ErrModelNotReady
	}

	algorithm, exists := s.algorithms[model.ModelType]
	if !exists {
		return nil, fmt.Errorf("unsupported model type: %s", model.ModelType)
	}

	lookbackStart := cmd.StartDate.AddDate(0, 0, -model.LookbackDays)
	data, err := s.historicalRepo.GetSalesData(ctx, model.ProductIDs, lookbackStart, cmd.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	if err := algorithm.Train(data, model.Parameters); err != nil {
		return nil, fmt.Errorf("failed to train algorithm: %w", err)
	}

	productIDs := cmd.ProductIDs
	if len(productIDs) == 0 {
		productIDs = model.ProductIDs
	}

	items, err := algorithm.Predict(cmd.StartDate, cmd.EndDate, productIDs)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	prediction := domain.NewSalesPrediction(model.ID, cmd.StartDate, cmd.EndDate)
	prediction.ID = fmt.Sprintf("SP%d", s.idGenerator.Generate())
	for _, item := range items {
		prediction.AddItem(item)
	}

	if err := s.predictionRepo.Save(ctx, prediction); err != nil {
		return nil, fmt.Errorf("failed to save prediction: %w", err)
	}

	s.logger.InfoContext(ctx, "sales prediction completed", "prediction_id", prediction.ID, "items", len(items))
	return prediction, nil
}

func (s *SalesForecastingService) GetPrediction(ctx context.Context, predictionID string) (*domain.SalesPrediction, error) {
	return s.predictionRepo.FindByID(ctx, predictionID)
}

type UpdateActualSalesCommand struct {
	PredictionID string
	Items        []*ActualSalesItem
}

type ActualSalesItem struct {
	ProductID      string
	Date           time.Time
	ActualQuantity float64
	ActualRevenue  float64
}

func (s *SalesForecastingService) UpdateActualSales(ctx context.Context, cmd *UpdateActualSalesCommand) error {
	s.logger.InfoContext(ctx, "updating actual sales", "prediction_id", cmd.PredictionID)

	prediction, err := s.predictionRepo.FindByID(ctx, cmd.PredictionID)
	if err != nil {
		return err
	}

	for _, item := range cmd.Items {
		for _, predItem := range prediction.Items {
			if predItem.ProductID == item.ProductID && predItem.Date.Equal(item.Date) {
				predItem.ActualQuantity = &item.ActualQuantity
				predItem.ActualRevenue = &item.ActualRevenue
				break
			}
		}
	}

	return s.predictionRepo.Update(ctx, prediction)
}

func (s *SalesForecastingService) CalculateAccuracy(ctx context.Context, modelID string, start, end time.Time) (*domain.ForecastAccuracy, error) {
	s.logger.InfoContext(ctx, "calculating forecast accuracy", "model_id", modelID)

	predictions, err := s.predictionRepo.FindByModelID(ctx, modelID, 100, 0)
	if err != nil {
		return nil, err
	}

	var allItems []*domain.PredictionItem
	for _, pred := range predictions {
		if (pred.ForecastStart.After(start) || pred.ForecastStart.Equal(start)) &&
			(pred.ForecastEnd.Before(end) || pred.ForecastEnd.Equal(end)) {
			allItems = append(allItems, pred.Items...)
		}
	}

	tempPrediction := &domain.SalesPrediction{Items: allItems}
	accuracy := tempPrediction.CalculateAccuracy()
	accuracy.ID = fmt.Sprintf("FA%d", s.idGenerator.Generate())
	accuracy.ModelID = modelID

	if err := s.accuracyRepo.Save(ctx, accuracy); err != nil {
		return nil, fmt.Errorf("failed to save accuracy: %w", err)
	}

	return accuracy, nil
}

type GetDemandForecastCommand struct {
	ProductIDs            []string
	CategoryIDs           []string
	ForecastDays          int
	IncludeSeasonality    bool
	IncludePromotionEffect bool
}

func (s *SalesForecastingService) GetDemandForecast(ctx context.Context, cmd *GetDemandForecastCommand) (*domain.DemandForecast, error) {
	s.logger.InfoContext(ctx, "generating demand forecast", "products", len(cmd.ProductIDs), "days", cmd.ForecastDays)

	now := time.Now()
	forecastEnd := now.AddDate(0, 0, cmd.ForecastDays)

	forecast := domain.NewDemandForecast(now, forecastEnd)
	forecast.ID = fmt.Sprintf("DF%d", s.idGenerator.Generate())

	lookbackStart := now.AddDate(0, 0, -90)
	data, err := s.historicalRepo.GetSalesData(ctx, cmd.ProductIDs, lookbackStart, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	algorithm := domain.NewMovingAverageAlgorithm(7)
	if err := algorithm.Train(data, nil); err != nil {
		return nil, fmt.Errorf("failed to train algorithm: %w", err)
	}

	items, err := algorithm.Predict(now, forecastEnd, cmd.ProductIDs)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	for _, item := range items {
		demandItem := &domain.DemandItem{
			ProductID:  item.ProductID,
			Date:       item.Date,
			BaseDemand: item.PredictedQuantity,
		}

		if cmd.IncludeSeasonality {
			demandItem.SeasonalAdjustment = s.calculateSeasonalityFactor(item.Date)
		}

		if cmd.IncludePromotionEffect {
			demandItem.PromotionAdjustment = s.estimatePromotionEffect(item.Date)
		}

		demandItem.CalculateFinalDemand()
		demandItem.CalculateSafetyStock(0.95, 7, item.PredictedQuantity*0.2)

		forecast.Items = append(forecast.Items, demandItem)
	}

	if cmd.IncludeSeasonality {
		forecast.SeasonalityFactors = s.getSeasonalityFactors()
	}

	s.logger.InfoContext(ctx, "demand forecast generated", "id", forecast.ID, "items", len(forecast.Items))
	return forecast, nil
}

func (s *SalesForecastingService) calculateSeasonalityFactor(date time.Time) float64 {
	weekday := date.Weekday()
	switch weekday {
	case time.Saturday, time.Sunday:
		return 0.2
	case time.Friday:
		return 0.1
	default:
		return 0
	}
}

func (s *SalesForecastingService) estimatePromotionEffect(date time.Time) float64 {
	month := date.Month()
	if month == time.November || month == time.December {
		return 0.3
	}
	if month == time.June || month == time.July {
		return 0.15
	}
	return 0
}

func (s *SalesForecastingService) getSeasonalityFactors() []*domain.SeasonalityFactor {
	return []*domain.SeasonalityFactor{
		{Period: "WEEKEND", Factor: 1.2, Description: "Weekend sales boost"},
		{Period: "HOLIDAY", Factor: 1.5, Description: "Holiday season boost"},
		{Period: "SUMMER", Factor: 1.15, Description: "Summer promotion period"},
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	var lo, hi float64 = 0, x
	for i := 0; i < 100; i++ {
		mid := (lo + hi) / 2
		if mid*mid > x {
			hi = mid
		} else {
			lo = mid
		}
	}
	return (lo + hi) / 2
}
