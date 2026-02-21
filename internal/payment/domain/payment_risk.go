package domain

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

// RiskFactor 风险因子
type RiskFactor struct {
	FactorName  string  `json:"factor_name"`
	Weight      float64 `json:"weight"`
	Score       float64 `json:"score"`
	Description string  `json:"description"`
}

// RiskScore 风险评分
type RiskScore struct {
	ID         string                 `json:"id"`
	PaymentID  string                 `json:"payment_id"`
	UserID     string                 `json:"user_id"`
	DeviceID   string                 `json:"device_id"`
	SessionID  string                 `json:"session_id"`
	TotalScore float64                `json:"total_score"`
	RiskLevel  string                 `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	Factors    []*RiskFactor          `json:"factors"`
	Decision   string                 `json:"decision"` // APPROVE, REVIEW, REJECT
	Confidence float64                `json:"confidence"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// DeviceFingerprint 设备指纹
type DeviceFingerprint struct {
	ID               string       `json:"id"`
	UserID           string       `json:"user_id"`
	DeviceID         string       `json:"device_id"`
	Browser          string       `json:"browser"`
	OS               string       `json:"os"`
	ScreenResolution string       `json:"screen_resolution"`
	Timezone         string       `json:"timezone"`
	Language         string       `json:"language"`
	Plugins          []string     `json:"plugins"`
	Fonts            []string     `json:"fonts"`
	CanvasHash       string       `json:"canvas_hash"`
	WebGLHash        string       `json:"webgl_hash"`
	AudioHash        string       `json:"audio_hash"`
	IPAddress        string       `json:"ip_address"`
	Location         *GeoLocation `json:"location"`
	FirstSeen        time.Time    `json:"first_seen"`
	LastSeen         time.Time    `json:"last_seen"`
	TrustScore       float64      `json:"trust_score"`
	RiskIndicators   []string     `json:"risk_indicators"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// GeoLocation 地理位置
type GeoLocation struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Accuracy  float64 `json:"accuracy"`
}

// BehaviorAnalysis 行为分析
type BehaviorAnalysis struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	SessionID      string          `json:"session_id"`
	TypingPattern  *TypingPattern  `json:"typing_pattern"`
	MouseMovement  *MouseMovement  `json:"mouse_movement"`
	ScrollBehavior *ScrollBehavior `json:"scroll_behavior"`
	FormCompletion *FormCompletion `json:"form_completion"`
	TimePattern    *TimePattern    `json:"time_pattern"`
	RiskScore      float64         `json:"risk_score"`
	Anomalies      []string        `json:"anomalies"`
	Timestamp      time.Time       `json:"timestamp"`
}

// TypingPattern 打字模式
type TypingPattern struct {
	AvgKeystrokeInterval float64 `json:"avg_keystroke_interval"`
	KeystrokeVariance    float64 `json:"keystroke_variance"`
	BackspaceRate        float64 `json:"backspace_rate"`
	TypingSpeed          float64 `json:"typing_speed"`
}

// MouseMovement 鼠标移动
type MouseMovement struct {
	TotalDistance    float64 `json:"total_distance"`
	AvgSpeed         float64 `json:"avg_speed"`
	ClickFrequency   float64 `json:"click_frequency"`
	MovementVariance float64 `json:"movement_variance"`
}

// ScrollBehavior 滚动行为
type ScrollBehavior struct {
	ScrollDistance  float64 `json:"scroll_distance"`
	ScrollSpeed     float64 `json:"scroll_speed"`
	ScrollDirection string  `json:"scroll_direction"`
	ScrollPattern   string  `json:"scroll_pattern"`
}

// FormCompletion 表单完成
type FormCompletion struct {
	CompletionTime time.Duration `json:"completion_time"`
	FieldOrder     []string      `json:"field_order"`
	Corrections    int           `json:"corrections"`
	TabUsage       bool          `json:"tab_usage"`
}

// TimePattern 时间模式
type TimePattern struct {
	TimeOfDay          string        `json:"time_of_day"`
	DayOfWeek          string        `json:"day_of_week"`
	SessionDuration    time.Duration `json:"session_duration"`
	TimeBetweenActions time.Duration `json:"time_between_actions"`
}

// RiskEngine 风险引擎
type RiskEngine struct {
	deviceRepo   DeviceRepository
	behaviorRepo BehaviorRepository
	paymentRepo  PaymentRepository
	userRepo     UserRepository
	mu           sync.RWMutex
	config       *RiskConfig
	riskRules    []*RiskRule
}

// RiskConfig 风险配置
type RiskConfig struct {
	ThresholdLow    float64 `json:"threshold_low"`
	ThresholdMedium float64 `json:"threshold_medium"`
	ThresholdHigh   float64 `json:"threshold_high"`
	MaxScore        float64 `json:"max_score"`
	MinConfidence   float64 `json:"min_confidence"`
	EnableML        bool    `json:"enable_ml"`
	MLModelVersion  string  `json:"ml_model_version"`
}

// RiskRule 风险规则
type RiskRule struct {
	ID        string    `json:"id"`
	RuleName  string    `json:"rule_name"`
	Condition string    `json:"condition"`
	Action    string    `json:"action"` // SCORE, FLAG, BLOCK
	Weight    float64   `json:"weight"`
	Priority  int       `json:"priority"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRiskEngine 创建风险引擎
func NewRiskEngine(deviceRepo DeviceRepository, behaviorRepo BehaviorRepository,
	paymentRepo PaymentRepository, userRepo UserRepository) *RiskEngine {

	return &RiskEngine{
		deviceRepo:   deviceRepo,
		behaviorRepo: behaviorRepo,
		paymentRepo:  paymentRepo,
		userRepo:     userRepo,
		config: &RiskConfig{
			ThresholdLow:    20,
			ThresholdMedium: 40,
			ThresholdHigh:   60,
			MaxScore:        100,
			MinConfidence:   0.7,
			EnableML:        true,
			MLModelVersion:  "v1.0",
		},
		riskRules: make([]*RiskRule, 0),
	}
}

// Initialize 初始化风险引擎
func (re *RiskEngine) Initialize(ctx context.Context) error {
	// 加载风险规则
	rules, err := re.loadRiskRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to load risk rules: %w", err)
	}

	re.mu.Lock()
	re.riskRules = rules
	re.mu.Unlock()

	// 初始化机器学习模型
	if re.config.EnableML {
		err = re.initializeMLModel(ctx)
		if err != nil {
			return fmt.Errorf("failed to initialize ML model: %w", err)
		}
	}

	return nil
}

// AssessPaymentRisk 评估支付风险
func (re *RiskEngine) AssessPaymentRisk(ctx context.Context, payment *Payment) (*RiskScore, error) {
	// 收集风险数据
	riskData, err := re.collectRiskData(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to collect risk data: %w", err)
	}

	// 计算风险评分
	riskScore, err := re.calculateRiskScore(ctx, riskData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate risk score: %w", err)
	}

	// 确定风险等级
	riskScore.RiskLevel = re.determineRiskLevel(riskScore.TotalScore)

	// 做出决策
	riskScore.Decision = re.makeDecision(riskScore)

	// 计算置信度
	riskScore.Confidence = re.calculateConfidence(riskScore)

	// 保存风险评分
	err = re.saveRiskScore(ctx, riskScore)
	if err != nil {
		return nil, fmt.Errorf("failed to save risk score: %w", err)
	}

	return riskScore, nil
}

// collectRiskData 收集风险数据
func (re *RiskEngine) collectRiskData(ctx context.Context, payment *Payment) (*RiskData, error) {
	data := &RiskData{
		Payment:     payment,
		CollectedAt: time.Now(),
	}

	// 获取设备指纹
	deviceFingerprint, err := re.deviceRepo.GetDeviceFingerprint(ctx, payment.DeviceID)
	if err == nil {
		data.DeviceFingerprint = deviceFingerprint
	}

	// 获取行为分析
	behaviorAnalysis, err := re.behaviorRepo.GetBehaviorAnalysis(ctx, payment.SessionID)
	if err == nil {
		data.BehaviorAnalysis = behaviorAnalysis
	}

	// 获取用户历史
	userHistory, err := re.userRepo.GetUserPaymentHistory(ctx, fmt.Sprintf("%d", payment.UserID))
	if err == nil {
		data.UserHistory = userHistory
	}

	// 获取支付历史
	paymentHistory, err := re.paymentRepo.GetPaymentHistory(ctx, payment.UserID, 30) // 最近30天
	if err == nil {
		data.PaymentHistory = paymentHistory
	}

	return data, nil
}

// calculateRiskScore 计算风险评分
func (re *RiskEngine) calculateRiskScore(ctx context.Context, riskData *RiskData) (*RiskScore, error) {
	score := &RiskScore{
		ID:        generateRiskScoreID(),
		PaymentID: fmt.Sprintf("%d", riskData.Payment.ID),
		UserID:    fmt.Sprintf("%d", riskData.Payment.UserID),
		DeviceID:  riskData.Payment.DeviceID,
		SessionID: riskData.Payment.SessionID,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
		Factors:   make([]*RiskFactor, 0),
	}

	// 应用风险规则
	re.applyRiskRules(score, riskData)

	// 使用机器学习模型
	if re.config.EnableML {
		re.applyMLModel(score, riskData)
	}

	// 计算总分
	score.TotalScore = re.calculateTotalScore(score.Factors)

	// 确保分数在有效范围内
	if score.TotalScore > re.config.MaxScore {
		score.TotalScore = re.config.MaxScore
	}

	return score, nil
}

// applyRiskRules 应用风险规则
func (re *RiskEngine) applyRiskRules(score *RiskScore, riskData *RiskData) {
	re.mu.RLock()
	rules := re.riskRules
	re.mu.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// 检查条件
		if re.evaluateCondition(rule.Condition, riskData) {
			// 应用规则
			factor := &RiskFactor{
				FactorName:  rule.RuleName,
				Weight:      rule.Weight,
				Score:       re.calculateRuleScore(rule, riskData),
				Description: rule.Condition,
			}

			score.Factors = append(score.Factors, factor)
		}
	}
}

// evaluateCondition 评估条件
func (re *RiskEngine) evaluateCondition(condition string, riskData *RiskData) bool {
	// 简化实现
	// 实际应该使用规则引擎或表达式求值器

	switch condition {
	case "NEW_DEVICE":
		return riskData.DeviceFingerprint != nil &&
			riskData.DeviceFingerprint.TrustScore < 50
	case "HIGH_AMOUNT":
		return riskData.Payment.Amount > 10000
	case "UNUSUAL_LOCATION":
		return riskData.DeviceFingerprint != nil &&
			riskData.DeviceFingerprint.Location != nil &&
			riskData.DeviceFingerprint.Location.Country != "CN"
	case "RAPID_PAYMENTS":
		return len(riskData.PaymentHistory) > 10
	default:
		return false
	}
}

// calculateRuleScore 计算规则分数
func (re *RiskEngine) calculateRuleScore(rule *RiskRule, riskData *RiskData) float64 {
	// 根据规则类型计算分数
	switch rule.Action {
	case "SCORE":
		return rule.Weight * 10
	case "FLAG":
		return rule.Weight * 20
	case "BLOCK":
		return rule.Weight * 30
	default:
		return rule.Weight * 10
	}
}

// applyMLModel 应用机器学习模型
func (re *RiskEngine) applyMLModel(score *RiskScore, riskData *RiskData) {
	mlScore := re.calculateMLScore(riskData)
	if mlScore < 0 {
		mlScore = 0
	}
	if mlScore > re.config.MaxScore {
		mlScore = re.config.MaxScore
	}

	factor := &RiskFactor{
		FactorName:  "ML_MODEL",
		Weight:      0.3,
		Score:       mlScore,
		Description: fmt.Sprintf("Machine learning model prediction (%s)", re.config.MLModelVersion),
	}

	score.Factors = append(score.Factors, factor)
	if score.Metadata == nil {
		score.Metadata = make(map[string]interface{})
	}
	score.Metadata["ml_model_version"] = re.config.MLModelVersion
	score.Metadata["ml_score"] = mlScore
}

// calculateMLScore 计算机器学习分数
func (re *RiskEngine) calculateMLScore(riskData *RiskData) float64 {
	var score float64

	// 基于设备信任度
	if riskData.DeviceFingerprint != nil {
		score += (100 - riskData.DeviceFingerprint.TrustScore) * 0.3
	}

	// 基于支付金额
	amountScore := math.Min(float64(riskData.Payment.Amount)/10000, 1) * 30
	score += amountScore

	// 基于用户历史
	if riskData.UserHistory != nil {
		if riskData.UserHistory.FraudRate > 0.1 {
			score += 20
		}
	}

	return score
}

// calculateTotalScore 计算总分
func (re *RiskEngine) calculateTotalScore(factors []*RiskFactor) float64 {
	var total float64
	var totalWeight float64

	for _, factor := range factors {
		total += factor.Score * factor.Weight
		totalWeight += factor.Weight
	}

	if totalWeight > 0 {
		return total / totalWeight
	}

	return 0
}

// determineRiskLevel 确定风险等级
func (re *RiskEngine) determineRiskLevel(score float64) string {
	if score < re.config.ThresholdLow {
		return "LOW"
	} else if score < re.config.ThresholdMedium {
		return "MEDIUM"
	} else if score < re.config.ThresholdHigh {
		return "HIGH"
	} else {
		return "CRITICAL"
	}
}

// makeDecision 做出决策
func (re *RiskEngine) makeDecision(riskScore *RiskScore) string {
	switch riskScore.RiskLevel {
	case "LOW":
		return "APPROVE"
	case "MEDIUM":
		return "REVIEW"
	case "HIGH":
		return "REVIEW"
	case "CRITICAL":
		return "REJECT"
	default:
		return "REVIEW"
	}
}

// calculateConfidence 计算置信度
func (re *RiskEngine) calculateConfidence(riskScore *RiskScore) float64 {
	// 基于因素数量和一致性计算置信度
	factorCount := len(riskScore.Factors)
	if factorCount == 0 {
		return 0
	}

	// 计算因素一致性
	var consistency float64
	for _, factor := range riskScore.Factors {
		consistency += factor.Weight
	}

	// 置信度 = 一致性 * 因素数量因子
	confidence := consistency * math.Min(float64(factorCount)/10, 1)

	// 确保在有效范围内
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// saveRiskScore 保存风险评分
func (re *RiskEngine) saveRiskScore(ctx context.Context, riskScore *RiskScore) error {
	if riskScore == nil {
		return fmt.Errorf("risk score is nil")
	}
	if riskScore.Metadata == nil {
		riskScore.Metadata = make(map[string]interface{})
	}
	riskScore.Metadata["saved_at"] = time.Now().UTC().Format(time.RFC3339Nano)

	paymentID, err := strconv.ParseUint(riskScore.PaymentID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid payment id %s: %w", riskScore.PaymentID, err)
	}
	userID, err := strconv.ParseUint(riskScore.UserID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user id %s: %w", riskScore.UserID, err)
	}

	remark := fmt.Sprintf(
		"risk_score=%.2f level=%s decision=%s confidence=%.2f factors=%d",
		riskScore.TotalScore, riskScore.RiskLevel, riskScore.Decision, riskScore.Confidence, len(riskScore.Factors),
	)
	log := &PaymentLog{
		PaymentID: paymentID,
		UserID:    userID,
		Action:    "RISK_ASSESS",
		OldStatus: "",
		NewStatus: "",
		Remark:    remark,
	}
	if err := re.paymentRepo.SaveLog(ctx, log); err != nil {
		return fmt.Errorf("failed to save risk assessment log: %w", err)
	}
	return nil
}

// loadRiskRules 加载风险规则
func (re *RiskEngine) loadRiskRules(ctx context.Context) ([]*RiskRule, error) {
	// 简化实现：创建默认规则
	rules := []*RiskRule{
		{
			ID:        "RULE_001",
			RuleName:  "New Device Detection",
			Condition: "NEW_DEVICE",
			Action:    "SCORE",
			Weight:    0.2,
			Priority:  1,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "RULE_002",
			RuleName:  "High Amount Detection",
			Condition: "HIGH_AMOUNT",
			Action:    "SCORE",
			Weight:    0.3,
			Priority:  2,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "RULE_003",
			RuleName:  "Unusual Location Detection",
			Condition: "UNUSUAL_LOCATION",
			Action:    "SCORE",
			Weight:    0.25,
			Priority:  3,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "RULE_004",
			RuleName:  "Rapid Payments Detection",
			Condition: "RAPID_PAYMENTS",
			Action:    "SCORE",
			Weight:    0.25,
			Priority:  4,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return rules, nil
}

// initializeMLModel 初始化机器学习模型
func (re *RiskEngine) initializeMLModel(ctx context.Context) error {
	// 规范模型配置并执行一次轻量预热，确保运行期参数可用。
	if re.config.MLModelVersion == "" {
		re.config.MLModelVersion = "v1.0"
	}
	if re.config.MinConfidence <= 0 || re.config.MinConfidence > 1 {
		re.config.MinConfidence = 0.7
	}
	_ = re.calculateMLScore(&RiskData{
		Payment: &Payment{
			Amount: 0,
		},
		DeviceFingerprint: &DeviceFingerprint{
			TrustScore: 100,
		},
		UserHistory: &UserPaymentHistory{
			FraudRate: 0,
		},
	})

	return nil
}

// Data structures

type RiskData struct {
	Payment           *Payment            `json:"payment"`
	DeviceFingerprint *DeviceFingerprint  `json:"device_fingerprint"`
	BehaviorAnalysis  *BehaviorAnalysis   `json:"behavior_analysis"`
	UserHistory       *UserPaymentHistory `json:"user_history"`
	PaymentHistory    []*Payment          `json:"payment_history"`
	CollectedAt       time.Time           `json:"collected_at"`
}

type UserPaymentHistory struct {
	UserID          string    `json:"user_id"`
	TotalPayments   int       `json:"total_payments"`
	TotalAmount     float64   `json:"total_amount"`
	AvgAmount       float64   `json:"avg_amount"`
	MaxAmount       float64   `json:"max_amount"`
	FraudRate       float64   `json:"fraud_rate"`
	ChargebackRate  float64   `json:"chargeback_rate"`
	LastPaymentDate time.Time `json:"last_payment_date"`
}

// Repository interfaces

type DeviceRepository interface {
	GetDeviceFingerprint(ctx context.Context, deviceID string) (*DeviceFingerprint, error)
	SaveDeviceFingerprint(ctx context.Context, fingerprint *DeviceFingerprint) error
	UpdateDeviceFingerprint(ctx context.Context, fingerprint *DeviceFingerprint) error
}

type BehaviorRepository interface {
	GetBehaviorAnalysis(ctx context.Context, sessionID string) (*BehaviorAnalysis, error)
	SaveBehaviorAnalysis(ctx context.Context, analysis *BehaviorAnalysis) error
	UpdateBehaviorAnalysis(ctx context.Context, analysis *BehaviorAnalysis) error
}

// Repositories are already defined in payment.go or other domain files

type UserRepository interface {
	GetUserPaymentHistory(ctx context.Context, userID string) (*UserPaymentHistory, error)
	SaveUserPaymentHistory(ctx context.Context, history *UserPaymentHistory) error
	UpdateUserPaymentHistory(ctx context.Context, history *UserPaymentHistory) error
}

// Helper functions

func generateRiskScoreID() string {
	return fmt.Sprintf("RISK_%d", time.Now().UnixNano())
}
