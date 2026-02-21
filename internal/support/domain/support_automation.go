package domain

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// AutomationType 自动化类型
type AutomationType string

const (
	AutoTicketRouting AutomationType = "TICKET_ROUTING" // 工单路由
	AutoResponse      AutomationType = "RESPONSE"       // 自动回复
	AutoEscalation    AutomationType = "ESCALATION"     // 自动升级
	AutoResolution    AutomationType = "RESOLUTION"     // 自动解决
	AutoFollowUp      AutomationType = "FOLLOW_UP"      // 自动跟进
)

// AutomationRule 自动化规则
type AutomationRule struct {
	ID          string         `json:"id"`
	RuleNo      string         `json:"rule_no"`
	Name        string         `json:"name"`
	Type        AutomationType `json:"type"`
	Description string         `json:"description"`

	// 触发条件
	Conditions     []*Condition `json:"conditions"`
	ConditionLogic string       `json:"condition_logic"` // AND, OR
	Priority       int          `json:"priority"`

	// 执行动作
	Actions     []*Action     `json:"actions"`
	ActionDelay time.Duration `json:"action_delay"`

	// 状态控制
	Status  string `json:"status"` // ACTIVE, INACTIVE, DRAFT
	Enabled bool   `json:"enabled"`
	RunOnce bool   `json:"run_once"` // 是否只运行一次

	// 执行统计
	ExecutionCount int        `json:"execution_count"`
	LastExecuted   *time.Time `json:"last_executed"`
	SuccessCount   int        `json:"success_count"`
	FailureCount   int        `json:"failure_count"`

	// 时间配置
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Condition 条件
type Condition struct {
	ID       string      `json:"id"`
	Field    string      `json:"field"`    // ticket.status, ticket.priority, customer.tier
	Operator string      `json:"operator"` // EQUALS, NOT_EQUALS, CONTAINS, GREATER_THAN, LESS_THAN
	Value    interface{} `json:"value"`
	Negate   bool        `json:"negate"`
}

// Action 动作
type Action struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // ASSIGN_TICKET, SEND_RESPONSE, CHANGE_STATUS, SEND_NOTIFICATION
	Parameters map[string]interface{} `json:"parameters"`
	Delay      time.Duration          `json:"delay"`
	RetryCount int                    `json:"retry_count"`
	RetryDelay time.Duration          `json:"retry_delay"`
}

// AutomationEngine 自动化引擎
type AutomationEngine struct {
	ruleRepo   RuleRepository
	ticketRepo TicketRepository
	mu         sync.RWMutex
	config     *AutomationConfig
	rules      map[string]*AutomationRule
	ruleQueue  *RuleQueue
}

// AutomationConfig 自动化配置
type AutomationConfig struct {
	CheckInterval    time.Duration `json:"check_interval"`
	MaxConcurrent    int           `json:"max_concurrent"`
	ExecutionTimeout time.Duration `json:"execution_timeout"`
	RetryCount       int           `json:"retry_count"`
	RetryDelay       time.Duration `json:"retry_delay"`
	EnableLogging    bool          `json:"enable_logging"`
}

// RuleQueue 规则队列
type RuleQueue struct {
	highPriority   []*AutomationRule
	normalPriority []*AutomationRule
	lowPriority    []*AutomationRule
	mu             sync.RWMutex
}

// NewAutomationEngine 创建自动化引擎
func NewAutomationEngine(ruleRepo RuleRepository, ticketRepo TicketRepository) *AutomationEngine {
	return &AutomationEngine{
		ruleRepo:   ruleRepo,
		ticketRepo: ticketRepo,
		config: &AutomationConfig{
			CheckInterval:    1 * time.Minute,
			MaxConcurrent:    10,
			ExecutionTimeout: 30 * time.Second,
			RetryCount:       3,
			RetryDelay:       5 * time.Second,
			EnableLogging:    true,
		},
		rules:     make(map[string]*AutomationRule),
		ruleQueue: &RuleQueue{},
	}
}

// Initialize 初始化自动化引擎
func (ae *AutomationEngine) Initialize(ctx context.Context) error {
	// 加载规则
	rules, err := ae.ruleRepo.GetActiveRules(ctx)
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	ae.mu.Lock()
	for _, rule := range rules {
		ae.rules[rule.ID] = rule
	}
	ae.mu.Unlock()

	// 启动规则处理器
	go ae.startRuleProcessor(ctx)

	return nil
}

// startRuleProcessor 启动规则处理器
func (ae *AutomationEngine) startRuleProcessor(ctx context.Context) {
	fmt.Println("Automation rule processor started")

	ticker := time.NewTicker(ae.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Automation rule processor stopped")
			return

		case <-ticker.C:
			// 处理规则
			ae.processRules(ctx)
		}
	}
}

// processRules 处理规则
func (ae *AutomationEngine) processRules(ctx context.Context) {
	// 获取需要处理的工单
	tickets, err := ae.ticketRepo.GetTicketsForAutomation(ctx)
	if err != nil {
		fmt.Printf("Failed to get tickets for automation: %v\n", err)
		return
	}

	// 处理每个工单
	for _, ticket := range tickets {
		go ae.processTicket(ctx, ticket)
	}
}

// processTicket 处理工单
func (ae *AutomationEngine) processTicket(ctx context.Context, ticket *Ticket) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in ticket processing: %v\n", r)
		}
	}()

	// 获取适用的规则
	applicableRules := ae.getApplicableRules(ctx, ticket)

	// 按优先级排序
	sortedRules := ae.sortRulesByPriority(applicableRules)

	// 执行规则
	for _, rule := range sortedRules {
		err := ae.executeRule(ctx, rule, ticket)
		if err != nil {
			fmt.Printf("Failed to execute rule %s for ticket %s: %v\n",
				rule.ID, ticket.ID, err)
		}

		// 如果规则设置为只运行一次，则禁用
		if rule.RunOnce {
			rule.Enabled = false
			rule.UpdatedAt = time.Now()

			err = ae.ruleRepo.UpdateRule(ctx, rule)
			if err != nil {
				fmt.Printf("Failed to disable rule %s: %v\n", rule.ID, err)
			}
		}
	}
}

// getApplicableRules 获取适用的规则
func (ae *AutomationEngine) getApplicableRules(ctx context.Context, ticket *Ticket) []*AutomationRule {
	var applicableRules []*AutomationRule

	ae.mu.RLock()
	for _, rule := range ae.rules {
		if rule.Enabled && rule.Status == "ACTIVE" {
			// 检查有效期
			if !rule.ValidFrom.IsZero() && rule.ValidFrom.After(time.Now()) {
				continue
			}
			if !rule.ValidTo.IsZero() && rule.ValidTo.Before(time.Now()) {
				continue
			}

			// 检查条件
			if ae.checkConditions(ctx, rule, ticket) {
				applicableRules = append(applicableRules, rule)
			}
		}
	}
	ae.mu.RUnlock()

	return applicableRules
}

// checkConditions 检查条件
func (ae *AutomationEngine) checkConditions(ctx context.Context, rule *AutomationRule, ticket *Ticket) bool {
	if len(rule.Conditions) == 0 {
		return true
	}

	// 根据逻辑运算符检查条件
	switch rule.ConditionLogic {
	case "AND":
		for _, condition := range rule.Conditions {
			if !ae.checkCondition(ctx, condition, ticket) {
				return false
			}
		}
		return true

	case "OR":
		for _, condition := range rule.Conditions {
			if ae.checkCondition(ctx, condition, ticket) {
				return true
			}
		}
		return false

	default:
		// 默认为AND
		for _, condition := range rule.Conditions {
			if !ae.checkCondition(ctx, condition, ticket) {
				return false
			}
		}
		return true
	}
}

// checkCondition 检查单个条件
func (ae *AutomationEngine) checkCondition(ctx context.Context, condition *Condition, ticket *Ticket) bool {
	// 获取字段值
	value, err := ae.getFieldValue(ctx, condition.Field, ticket)
	if err != nil {
		return false
	}

	// 应用运算符
	result := ae.applyOperator(value, condition.Operator, condition.Value)

	// 应用否定
	if condition.Negate {
		return !result
	}

	return result
}

// getFieldValue 获取字段值
func (ae *AutomationEngine) getFieldValue(ctx context.Context, field string, ticket *Ticket) (interface{}, error) {
	// 解析字段路径
	// 格式: object.field 或 object.subobject.field

	switch field {
	case "ticket.status":
		return ticket.Status, nil
	case "ticket.priority":
		return ticket.Priority, nil
	case "ticket.category":
		return ticket.Category, nil
	case "ticket.created_at":
		return ticket.CreatedAt, nil
	case "ticket.updated_at":
		return ticket.UpdatedAt, nil
	case "ticket.response_time":
		return ae.calculateResponseTime(ticket), nil
	case "ticket.resolution_time":
		return ae.calculateResolutionTime(ticket), nil
	case "customer.tier":
		// TODO: 获取客户信息 (暂未注入 UserRepository)
		return "BRONZE", nil
	case "customer.lifetime_value":
		return 0.0, nil
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}
}

// calculateResponseTime 计算响应时间
func (ae *AutomationEngine) calculateResponseTime(ticket *Ticket) time.Duration {
	// 目前 Ticket 实体不直接持有 Responses 列表，需从 repo 获取或简化逻辑
	// 暂返回创建至今的时间
	return time.Since(ticket.CreatedAt)
}

// calculateResolutionTime 计算解决时间
func (ae *AutomationEngine) calculateResolutionTime(ticket *Ticket) time.Duration {
	if ticket.Status != TicketStatusResolved && ticket.Status != TicketStatusClosed {
		return time.Since(ticket.CreatedAt)
	}

	return ticket.UpdatedAt.Sub(ticket.CreatedAt)
}

// applyOperator 应用运算符
func (ae *AutomationEngine) applyOperator(value interface{}, operator string, expected interface{}) bool {
	switch operator {
	case "EQUALS":
		return ae.equals(value, expected)
	case "NOT_EQUALS":
		return !ae.equals(value, expected)
	case "CONTAINS":
		return ae.contains(value, expected)
	case "GREATER_THAN":
		return ae.greaterThan(value, expected)
	case "LESS_THAN":
		return ae.lessThan(value, expected)
	case "IN":
		return ae.in(value, expected)
	case "NOT_IN":
		return !ae.in(value, expected)
	default:
		return false
	}
}

// equals 等于
func (ae *AutomationEngine) equals(value, expected interface{}) bool {
	return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", expected)
}

// contains 包含
func (ae *AutomationEngine) contains(value, expected interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	expectedStr := fmt.Sprintf("%v", expected)

	return len(valueStr) > 0 && len(expectedStr) > 0 &&
		ae.containsString(valueStr, expectedStr)
}

// containsString 字符串包含
func (ae *AutomationEngine) containsString(str, substr string) bool {
	return len(str) >= len(substr) &&
		ae.indexOf(str, substr) >= 0
}

// indexOf 查找子字符串
func (ae *AutomationEngine) indexOf(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// greaterThan 大于
func (ae *AutomationEngine) greaterThan(value, expected interface{}) bool {
	valueNum, ok1 := ae.toFloat(value)
	expectedNum, ok2 := ae.toFloat(expected)

	if !ok1 || !ok2 {
		return false
	}

	return valueNum > expectedNum
}

// lessThan 小于
func (ae *AutomationEngine) lessThan(value, expected interface{}) bool {
	valueNum, ok1 := ae.toFloat(value)
	expectedNum, ok2 := ae.toFloat(expected)

	if !ok1 || !ok2 {
		return false
	}

	return valueNum < expectedNum
}

// in 在集合中
func (ae *AutomationEngine) in(value, expected interface{}) bool {
	// 期望值应该是切片
	expectedSlice, ok := expected.([]interface{})
	if !ok {
		return false
	}

	valueStr := fmt.Sprintf("%v", value)

	for _, item := range expectedSlice {
		if fmt.Sprintf("%v", item) == valueStr {
			return true
		}
	}

	return false
}

// toFloat 转换为浮点数
func (ae *AutomationEngine) toFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		// 尝试解析字符串
		var f float64
		_, err := fmt.Sscanf(v, "%f", &f)
		return f, err == nil
	default:
		return 0, false
	}
}

// sortRulesByPriority 按优先级排序规则
func (ae *AutomationEngine) sortRulesByPriority(rules []*AutomationRule) []*AutomationRule {
	// 按优先级降序排序
	sorted := make([]*AutomationRule, len(rules))
	copy(sorted, rules)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority < sorted[j].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// executeRule 执行规则
func (ae *AutomationEngine) executeRule(ctx context.Context, rule *AutomationRule, ticket *Ticket) error {
	// 记录开始时间
	startTime := time.Now()

	// 执行所有动作
	for _, action := range rule.Actions {
		err := ae.executeAction(ctx, action, ticket)
		if err != nil {
			// 更新规则统计
			rule.FailureCount++
			rule.UpdatedAt = time.Now()

			// 保存更新
			ae.ruleRepo.UpdateRule(ctx, rule)

			return fmt.Errorf("failed to execute action %s: %w", action.ID, err)
		}

		// 如果有延迟，等待
		if action.Delay > 0 {
			time.Sleep(action.Delay)
		}
	}

	// 更新规则统计
	rule.ExecutionCount++
	rule.SuccessCount++
	lastExecuted := time.Now()
	rule.LastExecuted = &lastExecuted
	rule.UpdatedAt = time.Now()

	// 保存更新
	err := ae.ruleRepo.UpdateRule(ctx, rule)
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	// 记录执行日志
	if ae.config.EnableLogging {
		executionTime := time.Since(startTime)
		fmt.Printf("Rule %s executed successfully for ticket %s in %v\n",
			rule.ID, ticket.ID, executionTime)
	}

	return nil
}

// executeAction 执行动作
func (ae *AutomationEngine) executeAction(ctx context.Context, action *Action, ticket *Ticket) error {
	switch action.Type {
	case "ASSIGN_TICKET":
		return ae.assignTicket(ctx, action.Parameters, ticket)
	case "SEND_RESPONSE":
		return ae.sendResponse(ctx, action.Parameters, ticket)
	case "CHANGE_STATUS":
		return ae.changeStatus(ctx, action.Parameters, ticket)
	case "SEND_NOTIFICATION":
		return ae.sendNotification(ctx, action.Parameters, ticket)
	case "ADD_TAG":
		return ae.addTag(ctx, action.Parameters, ticket)
	case "SET_PRIORITY":
		return ae.setPriority(ctx, action.Parameters, ticket)
	case "ESCALATE_TICKET":
		return ae.escalateTicket(ctx, action.Parameters, ticket)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// assignTicket 分配工单
func (ae *AutomationEngine) assignTicket(ctx context.Context, params map[string]interface{}, ticket *Ticket) error {
	// 获取分配目标
	assigneeID, ok := params["assignee_id"].(uint64)
	if !ok {
		// 尝试从字符串解析
		fmt.Sscanf(fmt.Sprintf("%v", params["assignee_id"]), "%d", &assigneeID)
	}

	// 更新工单
	ticket.Assign(assigneeID)

	// 保存更新
	err := ae.ticketRepo.UpdateTicket(ctx, ticket)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// 记录活动 (暂未使用，跳过)
	return nil
}

// sendResponse 发送回复
func (ae *AutomationEngine) sendResponse(ctx context.Context, params map[string]interface{}, ticket *Ticket) error {
	// 获取回复内容
	content, ok := params["content"].(string)
	if !ok {
		return fmt.Errorf("content not found in parameters")
	}

	// 创建回复
	response := NewMessage(ticket.ID, 0, "SYSTEM", content, MessageTypeText, false)

	// 保存回复
	err := ae.ticketRepo.SaveMessage(ctx, response)
	if err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	// 更新工单
	ticket.UpdatedAt = time.Now()

	err = ae.ticketRepo.UpdateTicket(ctx, ticket)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	return nil
}

// changeStatus 更改状态
func (ae *AutomationEngine) changeStatus(ctx context.Context, params map[string]interface{}, ticket *Ticket) error {
	// 获取新状态
	newStatus, ok := params["status"].(string)
	if !ok {
		return fmt.Errorf("status not found in parameters")
	}

	// 更新工单
	switch strings.ToUpper(newStatus) {
	case "RESOLVED":
		ticket.Resolve()
	case "CLOSED":
		ticket.Close()
	default:
		// 其他状态映射暂略
	}

	// 保存更新
	err := ae.ticketRepo.UpdateTicket(ctx, ticket)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// 记录活动 (暂未使用)
	return nil
}

// sendNotification 发送通知
func (ae *AutomationEngine) sendNotification(ctx context.Context, params map[string]interface{}, ticket *Ticket) error {
	// 目前暂无通过数据库表持久化通知的需求，跳过
	return nil
}

// addTag 添加标签
func (ae *AutomationEngine) addTag(ctx context.Context, params map[string]interface{}, ticket *Ticket) error {
	// 目前暂不支持标签，跳过
	return nil
}

// setPriority 设置优先级
func (ae *AutomationEngine) setPriority(ctx context.Context, params map[string]interface{}, ticket *Ticket) error {
	// 获取新优先级
	newPriority, ok := params["priority"].(string)
	if !ok {
		return fmt.Errorf("priority not found in parameters")
	}

	// 更新工单
	// 转换优先级
	var priority TicketPriority
	switch strings.ToUpper(newPriority) {
	case "LOW":
		priority = TicketPriorityLow
	case "MEDIUM":
		priority = TicketPriorityMedium
	case "HIGH":
		priority = TicketPriorityHigh
	case "URGENT":
		priority = TicketPriorityUrgent
	}
	ticket.Priority = priority

	// 保存更新
	err := ae.ticketRepo.UpdateTicket(ctx, ticket)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	// 记录活动 (暂未使用)
	return nil
}

// escalateTicket 升级工单
func (ae *AutomationEngine) escalateTicket(ctx context.Context, params map[string]interface{}, ticket *Ticket) error {
	// 升级逻辑目前简化为提高优先级
	ticket.Priority = TicketPriorityUrgent

	// 保存更新
	err := ae.ticketRepo.UpdateTicket(ctx, ticket)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	return nil
}

// CreateRule 创建规则
func (ae *AutomationEngine) CreateRule(ctx context.Context, rule *AutomationRule) error {
	// 验证规则
	if err := ae.validateRule(rule); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	// 设置默认值
	if rule.RuleNo == "" {
		rule.RuleNo = generateRuleNo()
	}

	if rule.Status == "" {
		rule.Status = "DRAFT"
	}

	if rule.ConditionLogic == "" {
		rule.ConditionLogic = "AND"
	}

	if rule.Priority == 0 {
		rule.Priority = 5
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// 保存规则
	err := ae.ruleRepo.SaveRule(ctx, rule)
	if err != nil {
		return fmt.Errorf("failed to save rule: %w", err)
	}

	// 添加到规则列表
	ae.mu.Lock()
	ae.rules[rule.ID] = rule
	ae.mu.Unlock()

	return nil
}

// validateRule 验证规则
func (ae *AutomationEngine) validateRule(rule *AutomationRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	if rule.Type == "" {
		return fmt.Errorf("rule type is required")
	}

	if len(rule.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	return nil
}

// UpdateRule 更新规则
func (ae *AutomationEngine) UpdateRule(ctx context.Context, rule *AutomationRule) error {
	// 验证规则
	if err := ae.validateRule(rule); err != nil {
		return fmt.Errorf("rule validation failed: %w", err)
	}

	rule.UpdatedAt = time.Now()

	// 保存更新
	err := ae.ruleRepo.UpdateRule(ctx, rule)
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	// 更新规则列表
	ae.mu.Lock()
	ae.rules[rule.ID] = rule
	ae.mu.Unlock()

	return nil
}

// EnableRule 启用规则
func (ae *AutomationEngine) EnableRule(ctx context.Context, ruleID string) error {
	// 获取规则
	rule, err := ae.ruleRepo.GetRule(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("failed to get rule: %w", err)
	}

	// 启用规则
	rule.Enabled = true
	rule.Status = "ACTIVE"
	rule.UpdatedAt = time.Now()

	// 保存更新
	err = ae.ruleRepo.UpdateRule(ctx, rule)
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	// 更新规则列表
	ae.mu.Lock()
	ae.rules[rule.ID] = rule
	ae.mu.Unlock()

	return nil
}

// DisableRule 禁用规则
func (ae *AutomationEngine) DisableRule(ctx context.Context, ruleID string) error {
	// 获取规则
	rule, err := ae.ruleRepo.GetRule(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("failed to get rule: %w", err)
	}

	// 禁用规则
	rule.Enabled = false
	rule.Status = "INACTIVE"
	rule.UpdatedAt = time.Now()

	// 保存更新
	err = ae.ruleRepo.UpdateRule(ctx, rule)
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	// 更新规则列表
	ae.mu.Lock()
	ae.rules[rule.ID] = rule
	ae.mu.Unlock()

	return nil
}

// GetRule 获取规则
func (ae *AutomationEngine) GetRule(ctx context.Context, ruleID string) (*AutomationRule, error) {
	return ae.ruleRepo.GetRule(ctx, ruleID)
}

// GetRulesByType 按类型获取规则
func (ae *AutomationEngine) GetRulesByType(ctx context.Context, ruleType AutomationType) ([]*AutomationRule, error) {
	return ae.ruleRepo.GetRulesByType(ctx, ruleType)
}

// GetRuleStatistics 获取规则统计
func (ae *AutomationEngine) GetRuleStatistics(ctx context.Context) (*RuleStatistics, error) {
	// 获取所有规则
	rules, err := ae.ruleRepo.GetAllRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	// 计算统计
	stats := &RuleStatistics{
		GeneratedAt:     time.Now(),
		TotalRules:      len(rules),
		ActiveRules:     0,
		InactiveRules:   0,
		DraftRules:      0,
		RulesByType:     make(map[string]int),
		TotalExecutions: 0,
		SuccessRate:     0,
	}

	for _, rule := range rules {
		// 按状态统计
		switch rule.Status {
		case "ACTIVE":
			stats.ActiveRules++
		case "INACTIVE":
			stats.InactiveRules++
		case "DRAFT":
			stats.DraftRules++
		}

		// 按类型统计
		stats.RulesByType[string(rule.Type)]++

		// 执行统计
		stats.TotalExecutions += rule.ExecutionCount

		// 成功率
		if rule.ExecutionCount > 0 {
			successRate := float64(rule.SuccessCount) / float64(rule.ExecutionCount) * 100
			stats.SuccessRate += successRate
		}
	}

	// 计算平均成功率
	if stats.TotalRules > 0 {
		stats.SuccessRate /= float64(stats.TotalRules)
	}

	return stats, nil
}

// Data structures

type RuleStatistics struct {
	GeneratedAt     time.Time      `json:"generated_at"`
	TotalRules      int            `json:"total_rules"`
	ActiveRules     int            `json:"active_rules"`
	InactiveRules   int            `json:"inactive_rules"`
	DraftRules      int            `json:"draft_rules"`
	RulesByType     map[string]int `json:"rules_by_type"`
	TotalExecutions int            `json:"total_executions"`
	SuccessRate     float64        `json:"success_rate"`
}

// Helper functions

func generateRuleID() string {
	return fmt.Sprintf("RULE_%d", time.Now().UnixNano())
}

func generateRuleNo() string {
	return fmt.Sprintf("R%d", time.Now().UnixNano())
}

func generateActivityID() string {
	return fmt.Sprintf("ACTIVITY_%d", time.Now().UnixNano())
}

func generateResponseID() string {
	return fmt.Sprintf("RESPONSE_%d", time.Now().UnixNano())
}

func generateNotificationID() string {
	return fmt.Sprintf("NOTIFICATION_%d", time.Now().UnixNano())
}

// Repository interfaces

type RuleRepository interface {
	SaveRule(ctx context.Context, rule *AutomationRule) error
	GetRule(ctx context.Context, ruleID string) (*AutomationRule, error)
	GetRulesByType(ctx context.Context, ruleType AutomationType) ([]*AutomationRule, error)
	GetActiveRules(ctx context.Context) ([]*AutomationRule, error)
	GetAllRules(ctx context.Context) ([]*AutomationRule, error)
	UpdateRule(ctx context.Context, rule *AutomationRule) error
	DeleteRule(ctx context.Context, ruleID string) error
}
