package domain

import (
	"context"
	"fmt"
	"time"
)

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"      // 低风险
	RiskLevelMedium   RiskLevel = "MEDIUM"   // 中风险
	RiskLevelHigh     RiskLevel = "HIGH"     // 高风险
	RiskLevelCritical RiskLevel = "CRITICAL" // 严重风险
)

// RiskAction 风险处置动作
type RiskAction string

const (
	RiskActionPass      RiskAction = "PASS"      // 放行
	RiskActionReview    RiskAction = "REVIEW"    // 人工审核
	RiskActionBlock     RiskAction = "BLOCK"     // 拦截
	RiskActionChallenge RiskAction = "CHALLENGE" // 挑战（如验证码）
)

// RiskFactor 风险因子
type RiskFactor struct {
	FactorCode string  `json:"factor_code"` // 因子代码
	FactorName string  `json:"factor_name"` // 因子名称
	Weight     float64 `json:"weight"`      // 权重 (0-1)
	Score      float64 `json:"score"`       // 得分 (0-100)
	Reason     string  `json:"reason"`      // 原因说明
}

// RiskAssessment 风险评估结果
type RiskAssessment struct {
	OrderID        uint64       `json:"order_id"`
	OrderNo        string       `json:"order_no"`
	RiskScore      float64      `json:"risk_score"`      // 风险总分 (0-100)
	RiskLevel      RiskLevel    `json:"risk_level"`      // 风险等级
	RiskAction     RiskAction   `json:"risk_action"`     // 处置动作
	Factors        []RiskFactor `json:"factors"`         // 风险因子
	ReviewRequired bool         `json:"review_required"` // 是否需要人工审核
	ReviewReason   string       `json:"review_reason"`   // 审核原因
	BlockReason    string       `json:"block_reason"`    // 拦截原因
	Metadata       string       `json:"metadata"`        // 元数据 (JSON格式)
	AssessedAt     time.Time    `json:"assessed_at"`
	ExpiresAt      time.Time    `json:"expires_at"`
}

// RiskRule 风险规则
type RiskRule struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	RuleCode    string    `json:"rule_code"`   // 规则代码
	RuleName    string    `json:"rule_name"`   // 规则名称
	RuleType    string    `json:"rule_type"`   // 规则类型: SCORING, DECISION
	Enabled     bool      `json:"enabled"`     // 是否启用
	Priority    int       `json:"priority"`    // 优先级
	Condition   string    `json:"condition"`   // 条件表达式 (JSON)
	Action      string    `json:"action"`      // 执行动作 (JSON)
	Description string    `json:"description"` // 规则描述
	CreatorID   uint64    `json:"creator_id"`  // 创建者ID
}

// RiskEvent 风险事件
type RiskEvent struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	EventID     string    `json:"event_id"`    // 事件ID
	OrderID     uint64    `json:"order_id"`    // 订单ID
	EventType   string    `json:"event_type"`  // 事件类型
	RiskLevel   RiskLevel `json:"risk_level"`  // 风险等级
	Action      string    `json:"action"`      // 执行动作
	Description string    `json:"description"` // 事件描述
	Metadata    string    `json:"metadata"`    // 元数据 (JSON格式)
	UserID      uint64    `json:"user_id"`     // 操作用户ID
}

// RiskContext 风险评估上下文
type RiskContext struct {
	OrderID        uint64                 `json:"order_id"`
	OrderNo        string                 `json:"order_no"`
	UserID         uint64                 `json:"user_id"`
	IPAddress      string                 `json:"ip_address"`
	DeviceID       string                 `json:"device_id"`
	UserAgent      string                 `json:"user_agent"`
	TotalAmount    int64                  `json:"total_amount"`
	PaymentMethod  string                 `json:"payment_method"`
	ShippingMethod string                 `json:"shipping_method"`
	ShippingAddr   *ShippingAddress       `json:"shipping_address"`
	Items          []*OrderItem           `json:"items"`
	UserHistory    *UserRiskHistory       `json:"user_history"`
	SessionData    map[string]interface{} `json:"session_data"`
	Timestamp      time.Time              `json:"timestamp"`
}

// UserRiskHistory 用户风险历史
type UserRiskHistory struct {
	UserID             uint64    `json:"user_id"`
	TotalOrders        int       `json:"total_orders"`
	TotalAmount        int64     `json:"total_amount"`
	AvgOrderAmount     float64   `json:"avg_order_amount"`
	RecentOrders       int       `json:"recent_orders"`        // 最近30天订单数
	RecentRefunds      int       `json:"recent_refunds"`       // 最近30天退款数
	RecentComplaints   int       `json:"recent_complaints"`    // 最近30天投诉数
	RiskOrderCount     int       `json:"risk_order_count"`     // 高风险订单数
	LastRiskOrderTime  time.Time `json:"last_risk_order_time"` // 最近高风险订单时间
	AccountAge         int       `json:"account_age"`          // 账号年龄（天）
	VerifiedStatus     string    `json:"verified_status"`      // 认证状态
	CreditScore        int       `json:"credit_score"`         // 信用分
	LastAssessmentTime time.Time `json:"last_assessment_time"` // 最近评估时间
}

// RiskEngine 风险引擎接口
type RiskEngine interface {
	// AssessOrderRisk 评估订单风险
	AssessOrderRisk(ctx context.Context, riskCtx *RiskContext) (*RiskAssessment, error)

	// ExecuteRiskRules 执行风险规则
	ExecuteRiskRules(ctx context.Context, riskCtx *RiskContext) ([]RiskFactor, error)

	// CalculateRiskScore 计算风险分数
	CalculateRiskScore(factors []RiskFactor) (float64, RiskLevel)

	// DetermineRiskAction 确定风险处置动作
	DetermineRiskAction(riskScore float64, riskLevel RiskLevel, factors []RiskFactor) (RiskAction, string)

	// RecordRiskEvent 记录风险事件
	RecordRiskEvent(ctx context.Context, event *RiskEvent) error
}

// OrderRiskManager 订单风险管理器
type OrderRiskManager struct {
	riskEngine   RiskEngine
	riskRepo     RiskRepository
	orderRepo    OrderRepository
	userRiskRepo UserRiskRepository
	thresholds   *RiskThresholds
}

// NewOrderRiskManager 创建订单风险管理器
func NewOrderRiskManager(riskEngine RiskEngine, riskRepo RiskRepository, orderRepo OrderRepository, userRiskRepo UserRiskRepository) *OrderRiskManager {
	return &OrderRiskManager{
		riskEngine:   riskEngine,
		riskRepo:     riskRepo,
		orderRepo:    orderRepo,
		userRiskRepo: userRiskRepo,
		thresholds: &RiskThresholds{
			LowThreshold:    30,
			MediumThreshold: 60,
			HighThreshold:   80,
			ReviewThreshold: 70,
			BlockThreshold:  90,
		},
	}
}

// AssessOrder 评估订单风险
func (m *OrderRiskManager) AssessOrder(ctx context.Context, order *Order, riskCtx *RiskContext) (*RiskAssessment, error) {
	// 获取用户风险历史
	userHistory, err := m.userRiskRepo.GetUserRiskHistory(ctx, order.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user risk history: %w", err)
	}

	// 构建风险评估上下文
	if riskCtx == nil {
		riskCtx = &RiskContext{
			OrderID:       uint64(order.ID),
			OrderNo:       order.OrderNo,
			UserID:        order.UserID,
			TotalAmount:   order.ActualAmount,
			PaymentMethod: order.PaymentMethod,
			ShippingAddr:  order.ShippingAddress,
			Items:         order.Items,
			UserHistory:   userHistory,
			Timestamp:     time.Now(),
		}
	} else {
		riskCtx.UserHistory = userHistory
	}

	// 执行风险评估
	assessment, err := m.riskEngine.AssessOrderRisk(ctx, riskCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to assess order risk: %w", err)
	}

	// 保存风险评估结果
	err = m.riskRepo.SaveAssessment(ctx, assessment)
	if err != nil {
		return nil, fmt.Errorf("failed to save risk assessment: %w", err)
	}

	// 更新订单风险评分
	order.RiskScore = assessment.RiskScore

	// 根据风险等级设置订单标签
	m.applyRiskTags(ctx, order, assessment)

	// 记录风险事件
	event := &RiskEvent{
		EventID:   generateEventID(),
		OrderID:   uint64(order.ID),
		EventType: "ORDER_RISK_ASSESSMENT",
		RiskLevel: assessment.RiskLevel,
		Action:    string(assessment.RiskAction),
		Description: fmt.Sprintf("订单风险评估: 分数=%.2f, 等级=%s, 动作=%s",
			assessment.RiskScore, assessment.RiskLevel, assessment.RiskAction),
		Metadata:  assessment.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = m.riskEngine.RecordRiskEvent(ctx, event)
	if err != nil {
		// 记录错误但不中断流程
		fmt.Printf("Failed to record risk event: %v\n", err)
	}

	return assessment, nil
}

// applyRiskTags 应用风险标签
func (m *OrderRiskManager) applyRiskTags(ctx context.Context, order *Order, assessment *RiskAssessment) {
	// TODO: 根据风险评估结果为订单添加相应的标签
	// 例如：高风险订单添加"高风险"标签，需要审核的订单添加"待审核"标签
}

// HandleRiskAction 处理风险处置动作
func (m *OrderRiskManager) HandleRiskAction(ctx context.Context, order *Order, assessment *RiskAssessment) error {
	switch assessment.RiskAction {
	case RiskActionPass:
		// 放行订单，继续正常流程
		return m.handlePassAction(ctx, order, assessment)

	case RiskActionReview:
		// 需要人工审核
		return m.handleReviewAction(ctx, order, assessment)

	case RiskActionBlock:
		// 拦截订单
		return m.handleBlockAction(ctx, order, assessment)

	case RiskActionChallenge:
		// 需要额外验证
		return m.handleChallengeAction(ctx, order, assessment)

	default:
		return fmt.Errorf("unknown risk action: %s", assessment.RiskAction)
	}
}

// handlePassAction 处理放行动作
func (m *OrderRiskManager) handlePassAction(ctx context.Context, order *Order, assessment *RiskAssessment) error {
	// 记录放行日志
	order.AddLog("RiskSystem", "RISK_PASS", order.Status.String(), order.Status.String(),
		fmt.Sprintf("订单风险评估通过，分数: %.2f", assessment.RiskScore))

	// 更新订单状态
	return nil
}

// handleReviewAction 处理审核动作
func (m *OrderRiskManager) handleReviewAction(ctx context.Context, order *Order, assessment *RiskAssessment) error {
	// 标记订单为待审核状态
	order.AddLog("RiskSystem", "RISK_REVIEW", order.Status.String(), order.Status.String(),
		fmt.Sprintf("订单需要人工审核，原因: %s", assessment.ReviewReason))

	// 创建审核任务
	reviewTask := &RiskReviewTask{
		OrderID:      uint64(order.ID),
		OrderNo:      order.OrderNo,
		RiskScore:    assessment.RiskScore,
		RiskLevel:    assessment.RiskLevel,
		ReviewReason: assessment.ReviewReason,
		Status:       "PENDING",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := m.riskRepo.SaveReviewTask(ctx, reviewTask)
	if err != nil {
		return fmt.Errorf("failed to save review task: %w", err)
	}

	// TODO: 发送审核通知

	return nil
}

// handleBlockAction 处理拦截动作
func (m *OrderRiskManager) handleBlockAction(ctx context.Context, order *Order, assessment *RiskAssessment) error {
	// 记录拦截日志
	order.AddLog("RiskSystem", "RISK_BLOCK", order.Status.String(), order.Status.String(),
		fmt.Sprintf("订单被风险系统拦截，原因: %s", assessment.BlockReason))

	// 取消订单
	err := order.Cancel(ctx, "RiskSystem", assessment.BlockReason)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// 更新订单状态
	err = m.orderRepo.Update(ctx, order)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// handleChallengeAction 处理挑战动作
func (m *OrderRiskManager) handleChallengeAction(ctx context.Context, order *Order, assessment *RiskAssessment) error {
	// 创建挑战任务（如发送验证码）
	challengeTask := &RiskChallengeTask{
		OrderID:       uint64(order.ID),
		OrderNo:       order.OrderNo,
		ChallengeType: "CAPTCHA", // 或其他挑战类型
		Status:        "PENDING",
		ExpiresAt:     time.Now().Add(5 * time.Minute),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := m.riskRepo.SaveChallengeTask(ctx, challengeTask)
	if err != nil {
		return fmt.Errorf("failed to save challenge task: %w", err)
	}

	// TODO: 发送挑战请求（如发送验证码）

	return nil
}

// GetRiskAssessment 获取风险评估结果
func (m *OrderRiskManager) GetRiskAssessment(ctx context.Context, orderID uint64) (*RiskAssessment, error) {
	assessment, err := m.riskRepo.GetAssessmentByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk assessment: %w", err)
	}

	return assessment, nil
}

// UpdateRiskRules 更新风险规则
func (m *OrderRiskManager) UpdateRiskRules(ctx context.Context, rules []*RiskRule) error {
	for _, rule := range rules {
		err := m.riskRepo.SaveRule(ctx, rule)
		if err != nil {
			return fmt.Errorf("failed to save rule %s: %w", rule.RuleCode, err)
		}
	}

	return nil
}

// GetRiskStatistics 获取风险统计
func (m *OrderRiskManager) GetRiskStatistics(ctx context.Context, startTime, endTime time.Time) (*RiskStatistics, error) {
	stats, err := m.riskRepo.GetRiskStatistics(ctx, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk statistics: %w", err)
	}

	return stats, nil
}

// RiskThresholds 风险阈值配置
type RiskThresholds struct {
	LowThreshold    float64 `json:"low_threshold"`    // 低风险阈值
	MediumThreshold float64 `json:"medium_threshold"` // 中风险阈值
	HighThreshold   float64 `json:"high_threshold"`   // 高风险阈值
	ReviewThreshold float64 `json:"review_threshold"` // 需要审核的阈值
	BlockThreshold  float64 `json:"block_threshold"`  // 需要拦截的阈值
}

// RiskReviewTask 风险审核任务
type RiskReviewTask struct {
	ID           uint       `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	OrderID      uint64     `json:"order_id"`
	OrderNo      string     `json:"order_no"`
	RiskScore    float64    `json:"risk_score"`
	RiskLevel    RiskLevel  `json:"risk_level"`
	ReviewReason string     `json:"review_reason"`
	ReviewerID   uint64     `json:"reviewer_id"`
	ReviewResult string     `json:"review_result"` // PASS, BLOCK
	ReviewNote   string     `json:"review_note"`
	Status       string     `json:"status"` // PENDING, REVIEWING, COMPLETED
	CompletedAt  *time.Time `json:"completed_at"`
}

// RiskChallengeTask 风险挑战任务
type RiskChallengeTask struct {
	ID            uint       `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	OrderID       uint64     `json:"order_id"`
	OrderNo       string     `json:"order_no"`
	ChallengeType string     `json:"challenge_type"`
	ChallengeData string     `json:"challenge_data"` // JSON格式的挑战数据
	ResponseData  string     `json:"response_data"`  // JSON格式的响应数据
	Status        string     `json:"status"`         // PENDING, VERIFIED, FAILED, EXPIRED
	ExpiresAt     time.Time  `json:"expires_at"`
	VerifiedAt    *time.Time `json:"verified_at"`
}

// RiskStatistics 风险统计
type RiskStatistics struct {
	TotalOrders       int64            `json:"total_orders"`
	RiskOrders        int64            `json:"risk_orders"`
	BlockedOrders     int64            `json:"blocked_orders"`
	ReviewedOrders    int64            `json:"reviewed_orders"`
	AvgRiskScore      float64          `json:"avg_risk_score"`
	HighRiskRate      float64          `json:"high_risk_rate"`
	BlockRate         float64          `json:"block_rate"`
	FalsePositiveRate float64          `json:"false_positive_rate"`
	TopRiskFactors    []RiskFactorStat `json:"top_risk_factors"`
}

// RiskFactorStat 风险因子统计
type RiskFactorStat struct {
	FactorCode string  `json:"factor_code"`
	FactorName string  `json:"factor_name"`
	Count      int64   `json:"count"`
	AvgScore   float64 `json:"avg_score"`
	ImpactRate float64 `json:"impact_rate"`
}

// Helper functions

func generateEventID() string {
	return fmt.Sprintf("RISK_EVENT_%d", time.Now().UnixNano())
}

// RiskRepository 风险仓储接口
type RiskRepository interface {
	// 风险评估管理
	SaveAssessment(ctx context.Context, assessment *RiskAssessment) error
	GetAssessmentByOrderID(ctx context.Context, orderID uint64) (*RiskAssessment, error)
	GetAssessmentsByTimeRange(ctx context.Context, startTime, endTime time.Time, page, pageSize int) ([]*RiskAssessment, int64, error)

	// 风险规则管理
	SaveRule(ctx context.Context, rule *RiskRule) error
	GetRuleByCode(ctx context.Context, ruleCode string) (*RiskRule, error)
	GetRulesByType(ctx context.Context, ruleType string, enabledOnly bool) ([]*RiskRule, error)
	UpdateRule(ctx context.Context, rule *RiskRule) error
	DeleteRule(ctx context.Context, ruleID uint64) error

	// 风险事件管理
	SaveEvent(ctx context.Context, event *RiskEvent) error
	GetEventsByOrderID(ctx context.Context, orderID uint64, page, pageSize int) ([]*RiskEvent, int64, error)
	GetEventsByTimeRange(ctx context.Context, startTime, endTime time.Time, eventType string, page, pageSize int) ([]*RiskEvent, int64, error)

	// 审核任务管理
	SaveReviewTask(ctx context.Context, task *RiskReviewTask) error
	GetReviewTaskByOrderID(ctx context.Context, orderID uint64) (*RiskReviewTask, error)
	GetPendingReviewTasks(ctx context.Context, page, pageSize int) ([]*RiskReviewTask, int64, error)
	UpdateReviewTask(ctx context.Context, task *RiskReviewTask) error

	// 挑战任务管理
	SaveChallengeTask(ctx context.Context, task *RiskChallengeTask) error
	GetChallengeTaskByOrderID(ctx context.Context, orderID uint64) (*RiskChallengeTask, error)
	UpdateChallengeTask(ctx context.Context, task *RiskChallengeTask) error

	// 统计分析
	GetRiskStatistics(ctx context.Context, startTime, endTime time.Time) (*RiskStatistics, error)
	GetTopRiskFactors(ctx context.Context, startTime, endTime time.Time, limit int) ([]RiskFactorStat, error)
}

// UserRiskRepository 用户风险仓储接口
type UserRiskRepository interface {
	GetUserRiskHistory(ctx context.Context, userID uint64) (*UserRiskHistory, error)
	UpdateUserRiskHistory(ctx context.Context, history *UserRiskHistory) error
	GetHighRiskUsers(ctx context.Context, threshold float64, limit int) ([]uint64, error)
	GetUserRiskTrend(ctx context.Context, userID uint64, days int) ([]*RiskTrendPoint, error)
}

// RiskTrendPoint 风险趋势点
type RiskTrendPoint struct {
	Date       time.Time `json:"date"`
	RiskScore  float64   `json:"risk_score"`
	OrderCount int       `json:"order_count"`
	RiskCount  int       `json:"risk_count"`
}
