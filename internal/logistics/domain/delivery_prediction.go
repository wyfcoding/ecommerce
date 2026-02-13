package domain

import (
	"context"
	"errors"
	"math"
	"time"
)

var (
	ErrDeliveryPredictionNotFound = errors.New("delivery prediction not found")
	ErrInsufficientHistoryData    = errors.New("insufficient history data for prediction")
	ErrInvalidPredictionParams    = errors.New("invalid prediction parameters")
)

type PredictionType int8

const (
	PredictionTypeETA       PredictionType = 1
	PredictionTypeArrival   PredictionType = 2
	PredictionTypeDelayRisk PredictionType = 3
)

func (t PredictionType) String() string {
	switch t {
	case PredictionTypeETA:
		return "ETA"
	case PredictionTypeArrival:
		return "ARRIVAL"
	case PredictionTypeDelayRisk:
		return "DELAY_RISK"
	default:
		return "UNKNOWN"
	}
}

type DeliveryPrediction struct {
	ID               uint            `json:"id"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	PredictionNo     string          `json:"prediction_no"`
	LogisticsID      uint64          `json:"logistics_id"`
	TrackingNo       string          `json:"tracking_no"`
	PredictionType   PredictionType  `json:"prediction_type"`
	OriginLat        float64         `json:"origin_lat"`
	OriginLon        float64         `json:"origin_lon"`
	DestLat          float64         `json:"dest_lat"`
	DestLon          float64         `json:"dest_lon"`
	Distance         float64         `json:"distance"`
	CarrierCode      string          `json:"carrier_code"`
	ServiceType      string          `json:"service_type"`
	PredictedMinutes int             `json:"predicted_minutes"`
	PredictedTime    time.Time       `json:"predicted_time"`
	Confidence       float64         `json:"confidence"`
	MinMinutes       int             `json:"min_minutes"`
	MaxMinutes       int             `json:"max_minutes"`
	ActualMinutes    int             `json:"actual_minutes"`
	ActualTime       *time.Time      `json:"actual_time"`
	Accuracy         float64         `json:"accuracy"`
	ModelVersion     string          `json:"model_version"`
	Features         map[string]float64 `json:"features"`
	DelayRisk        float64         `json:"delay_risk"`
	DelayReasons     []string        `json:"delay_reasons"`
	WeatherFactor    float64         `json:"weather_factor"`
	TrafficFactor    float64         `json:"traffic_factor"`
	HolidayFactor    float64         `json:"holiday_factor"`
	Status           PredictionStatus `json:"status"`
}

type PredictionStatus int8

const (
	PredictionStatusPending   PredictionStatus = 1
	PredictionStatusActive    PredictionStatus = 2
	PredictionStatusCompleted PredictionStatus = 3
	PredictionStatusCancelled PredictionStatus = 4
)

func (s PredictionStatus) String() string {
	switch s {
	case PredictionStatusPending:
		return "PENDING"
	case PredictionStatusActive:
		return "ACTIVE"
	case PredictionStatusCompleted:
		return "COMPLETED"
	case PredictionStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type DeliveryHistory struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TrackingNo    string    `json:"tracking_no"`
	CarrierCode   string    `json:"carrier_code"`
	ServiceType   string    `json:"service_type"`
	OriginLat     float64   `json:"origin_lat"`
	OriginLon     float64   `json:"origin_lon"`
	DestLat       float64   `json:"dest_lat"`
	DestLon       float64   `json:"dest_lon"`
	Distance      float64   `json:"distance"`
	ShippedAt     time.Time `json:"shipped_at"`
	DeliveredAt   time.Time `json:"delivered_at"`
	ActualMinutes int       `json:"actual_minutes"`
	Weather       string    `json:"weather"`
	TrafficLevel  int       `json:"traffic_level"`
	IsHoliday     bool      `json:"is_holiday"`
	DayOfWeek     int       `json:"day_of_week"`
	HourOfDay     int       `json:"hour_of_day"`
}

type PredictionModel struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Description  string    `json:"description"`
	ModelType    string    `json:"model_type"`
	CarrierCode  string    `json:"carrier_code"`
	Features     []string  `json:"features"`
	Weights      map[string]float64 `json:"weights"`
	Bias         float64   `json:"bias"`
	MAE          float64   `json:"mae"`
	RMSE         float64   `json:"rmse"`
	MAPE         float64   `json:"mape"`
	SampleCount  int       `json:"sample_count"`
	LastTrained  *time.Time `json:"last_trained"`
	Enabled      bool      `json:"enabled"`
}

type PredictionConfig struct {
	DefaultModelType     string        `json:"default_model_type"`
	MinHistorySamples    int           `json:"min_history_samples"`
	ConfidenceThreshold  float64       `json:"confidence_threshold"`
	UpdateInterval       time.Duration `json:"update_interval"`
	EnableRealTimeUpdate bool          `json:"enable_real_time_update"`
	WeatherAPIEnabled    bool          `json:"weather_api_enabled"`
	TrafficAPIEnabled    bool          `json:"traffic_api_enabled"`
}

func DefaultPredictionConfig() *PredictionConfig {
	return &PredictionConfig{
		DefaultModelType:     "LINEAR_REGRESSION",
		MinHistorySamples:    100,
		ConfidenceThreshold:  0.7,
		UpdateInterval:       time.Hour,
		EnableRealTimeUpdate: true,
		WeatherAPIEnabled:    true,
		TrafficAPIEnabled:    true,
	}
}

func NewDeliveryPrediction(predictionNo string, logisticsID uint64, trackingNo string, originLat, originLon, destLat, destLon float64, carrierCode, serviceType string) *DeliveryPrediction {
	distance := calculateDistance(originLat, originLon, destLat, destLon)
	return &DeliveryPrediction{
		PredictionNo:   predictionNo,
		LogisticsID:    logisticsID,
		TrackingNo:     trackingNo,
		PredictionType: PredictionTypeETA,
		OriginLat:      originLat,
		OriginLon:      originLon,
		DestLat:        destLat,
		DestLon:        destLon,
		Distance:       distance,
		CarrierCode:    carrierCode,
		ServiceType:    serviceType,
		Confidence:     0.0,
		Features:       make(map[string]float64),
		DelayReasons:   make([]string, 0),
		WeatherFactor:  1.0,
		TrafficFactor:  1.0,
		HolidayFactor:  1.0,
		Status:         PredictionStatusPending,
	}
}

func (p *DeliveryPrediction) SetPrediction(minutes int, confidence float64) {
	now := time.Now()
	p.PredictedMinutes = minutes
	p.PredictedTime = now.Add(time.Duration(minutes) * time.Minute)
	p.Confidence = confidence
	p.MinMinutes = int(float64(minutes) * 0.9)
	p.MaxMinutes = int(float64(minutes) * 1.1)
	p.Status = PredictionStatusActive
}

func (p *DeliveryPrediction) SetActual(actualMinutes int, actualTime time.Time) {
	p.ActualMinutes = actualMinutes
	p.ActualTime = &actualTime
	p.Status = PredictionStatusCompleted
	p.calculateAccuracy()
}

func (p *DeliveryPrediction) calculateAccuracy() {
	if p.PredictedMinutes == 0 {
		return
	}
	diff := math.Abs(float64(p.ActualMinutes - p.PredictedMinutes))
	p.Accuracy = 1.0 - (diff / float64(p.PredictedMinutes))
	if p.Accuracy < 0 {
		p.Accuracy = 0
	}
}

func (p *DeliveryPrediction) SetDelayRisk(risk float64, reasons []string) {
	p.DelayRisk = risk
	p.DelayReasons = reasons
}

func (p *DeliveryPrediction) SetFactors(weather, traffic, holiday float64) {
	p.WeatherFactor = weather
	p.TrafficFactor = traffic
	p.HolidayFactor = holiday
}

func (p *DeliveryPrediction) IsDelayed() bool {
	return p.DelayRisk > 0.5
}

func (p *DeliveryPrediction) IsHighRisk() bool {
	return p.DelayRisk > 0.7
}

func (p *DeliveryPrediction) Cancel() {
	p.Status = PredictionStatusCancelled
}

type DeliveryPredictor struct {
	config    *PredictionConfig
	models    map[string]*PredictionModel
	repository PredictionRepository
}

func NewDeliveryPredictor(config *PredictionConfig, repo PredictionRepository) *DeliveryPredictor {
	return &DeliveryPredictor{
		config:    config,
		models:    make(map[string]*PredictionModel),
		repository: repo,
	}
}

func (d *DeliveryPredictor) LoadModels(ctx context.Context) error {
	if d.repository == nil {
		return nil
	}

	models, err := d.repository.FindEnabledModels(ctx)
	if err != nil {
		return err
	}

	for _, model := range models {
		d.models[model.CarrierCode] = model
	}

	return nil
}

func (d *DeliveryPredictor) Predict(ctx context.Context, prediction *DeliveryPrediction) error {
	model, exists := d.models[prediction.CarrierCode]
	if !exists {
		model, exists = d.models["default"]
		if !exists {
			return errors.New("no prediction model available")
		}
	}

	features := d.extractFeatures(prediction)
	prediction.Features = features

	minutes := d.calculateETA(model, features)
	confidence := d.calculateConfidence(model, features)

	prediction.SetPrediction(minutes, confidence)

	delayRisk, delayReasons := d.predictDelayRisk(model, features)
	prediction.SetDelayRisk(delayRisk, delayReasons)

	return nil
}

func (d *DeliveryPredictor) extractFeatures(prediction *DeliveryPrediction) map[string]float64 {
	features := make(map[string]float64)

	features["distance"] = prediction.Distance
	features["distance_log"] = math.Log1p(prediction.Distance)

	now := time.Now()
	features["day_of_week"] = float64(now.Weekday())
	features["hour_of_day"] = float64(now.Hour())
	features["is_weekend"] = 0
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		features["is_weekend"] = 1
	}

	features["weather_factor"] = prediction.WeatherFactor
	features["traffic_factor"] = prediction.TrafficFactor
	features["holiday_factor"] = prediction.HolidayFactor

	return features
}

func (d *DeliveryPredictor) calculateETA(model *PredictionModel, features map[string]float64) int {
	var result float64

	for featureName, value := range features {
		if weight, exists := model.Weights[featureName]; exists {
			result += weight * value
		}
	}

	result += model.Bias

	if result < 0 {
		result = 60
	}

	return int(result)
}

func (d *DeliveryPredictor) calculateConfidence(model *PredictionModel, features map[string]float64) float64 {
	confidence := 0.8

	if features["weather_factor"] > 1.2 {
		confidence -= 0.1
	}
	if features["traffic_factor"] > 1.3 {
		confidence -= 0.1
	}
	if features["holiday_factor"] > 1.0 {
		confidence -= 0.05
	}

	if confidence < 0.3 {
		confidence = 0.3
	}

	return confidence
}

func (d *DeliveryPredictor) predictDelayRisk(model *PredictionModel, features map[string]float64) (float64, []string) {
	var risk float64
	var reasons []string

	if features["weather_factor"] > 1.3 {
		risk += 0.3
		reasons = append(reasons, "恶劣天气")
	}

	if features["traffic_factor"] > 1.5 {
		risk += 0.25
		reasons = append(reasons, "交通拥堵")
	}

	if features["holiday_factor"] > 1.0 {
		risk += 0.15
		reasons = append(reasons, "节假日")
	}

	if features["is_weekend"] == 1 {
		risk += 0.1
	}

	if risk > 1.0 {
		risk = 1.0
	}

	return risk, reasons
}

func (d *DeliveryPredictor) UpdateWithActual(ctx context.Context, predictionID uint64, actualMinutes int, actualTime time.Time) error {
	if d.repository == nil {
		return nil
	}

	prediction, err := d.repository.FindPredictionByID(ctx, predictionID)
	if err != nil {
		return err
	}

	prediction.SetActual(actualMinutes, actualTime)

	return d.repository.UpdatePrediction(ctx, prediction)
}

func (d *DeliveryPredictor) RecordHistory(ctx context.Context, history *DeliveryHistory) error {
	if d.repository == nil {
		return nil
	}

	return d.repository.SaveHistory(ctx, history)
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

type PredictionRepository interface {
	SavePrediction(ctx context.Context, prediction *DeliveryPrediction) error
	UpdatePrediction(ctx context.Context, prediction *DeliveryPrediction) error
	FindPredictionByID(ctx context.Context, id uint64) (*DeliveryPrediction, error)
	FindPredictionByTrackingNo(ctx context.Context, trackingNo string) (*DeliveryPrediction, error)
	FindActivePredictions(ctx context.Context, limit int) ([]*DeliveryPrediction, error)

	SaveHistory(ctx context.Context, history *DeliveryHistory) error
	FindHistoryByCarrier(ctx context.Context, carrierCode string, limit int) ([]*DeliveryHistory, error)
	FindHistoryByRoute(ctx context.Context, originLat, originLon, destLat, destLon float64, limit int) ([]*DeliveryHistory, error)

	FindEnabledModels(ctx context.Context) ([]*PredictionModel, error)
	SaveModel(ctx context.Context, model *PredictionModel) error
}
