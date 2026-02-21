package domain

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

// WorkflowState 工作流状态
type WorkflowState string

const (
	StateCreated    WorkflowState = "CREATED"    // 已创建
	StateValidated  WorkflowState = "VALIDATED"  // 已验证
	StatePaid       WorkflowState = "PAID"       // 已支付
	StateConfirmed  WorkflowState = "CONFIRMED"  // 已确认
	StateProcessing WorkflowState = "PROCESSING" // 处理中
	StateShipped    WorkflowState = "SHIPPED"    // 已发货
	StateDelivered  WorkflowState = "DELIVERED"  // 已送达
	StateCompleted  WorkflowState = "COMPLETED"  // 已完成
	StateCancelled  WorkflowState = "CANCELLED"  // 已取消
	StateRefunded   WorkflowState = "REFUNDED"   // 已退款
)

// WorkflowAction 工作流动作
type WorkflowAction string

const (
	ActionValidate       WorkflowAction = "VALIDATE"        // 验证订单
	ActionProcessPayment WorkflowAction = "PROCESS_PAYMENT" // 处理支付
	ActionConfirm        WorkflowAction = "CONFIRM"         // 确认订单
	ActionProcess        WorkflowAction = "PROCESS"         // 处理订单
	ActionShip           WorkflowAction = "SHIP"            // 发货
	ActionDeliver        WorkflowAction = "DELIVER"         // 送达
	ActionComplete       WorkflowAction = "COMPLETE"        // 完成订单
	ActionCancel         WorkflowAction = "CANCEL"          // 取消订单
	ActionRefund         WorkflowAction = "REFUND"          // 退款
)

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID         string         `json:"id"`
	StepName   string         `json:"step_name"`
	Action     WorkflowAction `json:"action"`
	FromState  WorkflowState  `json:"from_state"`
	ToState    WorkflowState  `json:"to_state"`
	Handler    string         `json:"handler"`     // 处理器名称
	Timeout    time.Duration  `json:"timeout"`     // 超时时间
	RetryCount int            `json:"retry_count"` // 重试次数
	IsAsync    bool           `json:"is_async"`    // 是否异步
	Conditions []string       `json:"conditions"`  // 执行条件
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// WorkflowInstance 工作流实例
type WorkflowInstance struct {
	ID            string          `json:"id"`
	OrderID       string          `json:"order_id"`
	CurrentState  WorkflowState   `json:"current_state"`
	PreviousState WorkflowState   `json:"previous_state"`
	Steps         []*StepInstance `json:"steps"`
	Status        string          `json:"status"` // ACTIVE, COMPLETED, FAILED
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	CompletedAt   *time.Time      `json:"completed_at"`
}

// StepInstance 步骤实例
type StepInstance struct {
	ID          string                 `json:"id"`
	StepID      string                 `json:"step_id"`
	StepName    string                 `json:"step_name"`
	Action      WorkflowAction         `json:"action"`
	Status      string                 `json:"status"` // PENDING, IN_PROGRESS, COMPLETED, FAILED
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Error       string                 `json:"error"`
	RetryCount  int                    `json:"retry_count"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
	workflowRepo WorkflowRepository
	orderRepo    OrderRepository
	stepHandlers map[WorkflowAction]StepHandler
	mu           sync.RWMutex
	config       *WorkflowConfig
	workflows    map[string]*WorkflowDefinition
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	DefaultTimeout   time.Duration `json:"default_timeout"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	AsyncWorkers     int           `json:"async_workers"`
	EnableMonitoring bool          `json:"enable_monitoring"`
}

// StepHandler 步骤处理器接口
type StepHandler interface {
	Execute(ctx context.Context, order *Order, step *WorkflowStep) error
	Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error)
	Rollback(ctx context.Context, order *Order, step *WorkflowStep) error
}

// WorkflowDefinition 工作流定义
type WorkflowDefinition struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Steps       []*WorkflowStep `json:"steps"`
	StartState  WorkflowState   `json:"start_state"`
	EndState    WorkflowState   `json:"end_state"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// NewWorkflowEngine 创建工作流引擎
func NewWorkflowEngine(workflowRepo WorkflowRepository, orderRepo OrderRepository) *WorkflowEngine {
	// 创建适配器
	adapter := NewWorkflowOrderRepositoryAdapter(orderRepo)

	return &WorkflowEngine{
		workflowRepo: workflowRepo,
		orderRepo:    adapter,
		stepHandlers: make(map[WorkflowAction]StepHandler),
		config: &WorkflowConfig{
			DefaultTimeout:   30 * time.Minute,
			MaxRetries:       3,
			RetryDelay:       5 * time.Minute,
			AsyncWorkers:     10,
			EnableMonitoring: true,
		},
		workflows: make(map[string]*WorkflowDefinition),
	}
}

// Initialize 初始化工作流引擎
func (we *WorkflowEngine) Initialize(ctx context.Context) error {
	// 加载工作流定义
	definitions, err := we.workflowRepo.GetWorkflowDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("failed to load workflow definitions: %w", err)
	}

	we.mu.Lock()
	for _, definition := range definitions {
		we.workflows[definition.ID] = definition
	}
	we.mu.Unlock()

	// 注册步骤处理器
	we.registerStepHandlers()

	// 启动异步工作器
	go we.startAsyncWorkers(ctx)

	return nil
}

// registerStepHandlers 注册步骤处理器
func (we *WorkflowEngine) registerStepHandlers() {
	we.mu.Lock()
	defer we.mu.Unlock()

	we.stepHandlers[ActionValidate] = &ValidationHandler{}
	we.stepHandlers[ActionProcessPayment] = &PaymentHandler{}
	we.stepHandlers[ActionConfirm] = &ConfirmationHandler{}
	we.stepHandlers[ActionProcess] = &ProcessingHandler{}
	we.stepHandlers[ActionShip] = &ShippingHandler{}
	we.stepHandlers[ActionDeliver] = &DeliveryHandler{}
	we.stepHandlers[ActionComplete] = &CompletionHandler{}
	we.stepHandlers[ActionCancel] = &CancellationHandler{}
	we.stepHandlers[ActionRefund] = &RefundHandler{}
}

// startAsyncWorkers 启动异步工作器
func (we *WorkflowEngine) startAsyncWorkers(ctx context.Context) {
	for i := 0; i < we.config.AsyncWorkers; i++ {
		go we.asyncWorker(ctx, i)
	}
}

// asyncWorker 异步工作器
func (we *WorkflowEngine) asyncWorker(ctx context.Context, workerID int) {
	fmt.Printf("Workflow async worker %d started\n", workerID)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Workflow async worker %d stopped\n", workerID)
			return

		case <-ticker.C:
			// 处理待处理的异步步骤
			we.processPendingSteps(ctx)
		}
	}
}

// processPendingSteps 处理待处理步骤
func (we *WorkflowEngine) processPendingSteps(ctx context.Context) {
	// 获取待处理的步骤实例
	pendingSteps, err := we.workflowRepo.GetPendingSteps(ctx)
	if err != nil {
		fmt.Printf("Failed to get pending steps: %v\n", err)
		return
	}

	for _, stepInstance := range pendingSteps {
		// 获取工作流实例
		workflowInstance, err := we.workflowRepo.GetWorkflowInstance(ctx, stepInstance.ID[:len(stepInstance.ID)-len(stepInstance.StepID)-1])
		if err != nil {
			fmt.Printf("Failed to get workflow instance: %v\n", err)
			continue
		}

		// 获取订单
		order, err := we.orderRepo.GetOrder(ctx, workflowInstance.OrderID)
		if err != nil {
			fmt.Printf("Failed to get order: %v\n", err)
			continue
		}

		// 获取步骤定义
		step, err := we.getStepDefinition(stepInstance.StepID)
		if err != nil {
			fmt.Printf("Failed to get step definition: %v\n", err)
			continue
		}

		// 执行步骤
		go we.executeStep(ctx, order, workflowInstance, stepInstance, step)
	}
}

// StartWorkflow 启动工作流
func (we *WorkflowEngine) StartWorkflow(ctx context.Context, order *Order, workflowID string) (*WorkflowInstance, error) {
	// 获取工作流定义
	workflowDef, err := we.getWorkflowDefinition(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow definition: %w", err)
	}

	// 创建工作流实例
	workflowInstance := &WorkflowInstance{
		ID:            generateWorkflowInstanceID(),
		OrderID:       fmt.Sprintf("%d", order.ID),
		CurrentState:  workflowDef.StartState,
		PreviousState: "",
		Status:        "ACTIVE",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 创建步骤实例
	for _, step := range workflowDef.Steps {
		stepInstance := &StepInstance{
			ID:         fmt.Sprintf("%s_%s", workflowInstance.ID, step.ID),
			StepID:     step.ID,
			StepName:   step.StepName,
			Action:     step.Action,
			Status:     "PENDING",
			RetryCount: 0,
			UpdatedAt:  time.Now(),
			Metadata:   make(map[string]interface{}),
		}
		workflowInstance.Steps = append(workflowInstance.Steps, stepInstance)
	}

	// 保存工作流实例
	err = we.workflowRepo.SaveWorkflowInstance(ctx, workflowInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to save workflow instance: %w", err)
	}

	// 执行第一个步骤
	err = we.executeNextStep(ctx, order, workflowInstance)
	if err != nil {
		return nil, fmt.Errorf("failed to execute first step: %w", err)
	}

	return workflowInstance, nil
}

// executeNextStep 执行下一步
func (we *WorkflowEngine) executeNextStep(ctx context.Context, order *Order, workflowInstance *WorkflowInstance) error {
	// 查找下一个待执行的步骤
	var nextStepInstance *StepInstance
	var nextStepDef *WorkflowStep

	for _, stepInstance := range workflowInstance.Steps {
		if stepInstance.Status == "PENDING" {
			// 获取步骤定义
			stepDef, err := we.getStepDefinition(stepInstance.StepID)
			if err != nil {
				return fmt.Errorf("failed to get step definition: %w", err)
			}

			// 检查状态转换
			if stepDef.FromState == workflowInstance.CurrentState {
				nextStepInstance = stepInstance
				nextStepDef = stepDef
				break
			}
		}
	}

	if nextStepInstance == nil {
		// 没有更多步骤，工作流完成
		workflowInstance.Status = "COMPLETED"
		completedAt := time.Now()
		workflowInstance.CompletedAt = &completedAt
		workflowInstance.UpdatedAt = time.Now()

		err := we.workflowRepo.UpdateWorkflowInstance(ctx, workflowInstance)
		if err != nil {
			return fmt.Errorf("failed to complete workflow: %w", err)
		}

		return nil
	}

	// 执行步骤
	if nextStepDef.IsAsync {
		// 异步执行
		nextStepInstance.Status = "IN_PROGRESS"
		startedAt := time.Now()
		nextStepInstance.StartedAt = &startedAt
		nextStepInstance.UpdatedAt = time.Now()

		err := we.workflowRepo.UpdateStepInstance(ctx, nextStepInstance)
		if err != nil {
			return fmt.Errorf("failed to update step instance: %w", err)
		}

		// 异步执行会在工作器中处理
	} else {
		// 同步执行
		err := we.executeStep(ctx, order, workflowInstance, nextStepInstance, nextStepDef)
		if err != nil {
			return fmt.Errorf("failed to execute step: %w", err)
		}
	}

	return nil
}

// executeStep 执行步骤
func (we *WorkflowEngine) executeStep(ctx context.Context, order *Order, workflowInstance *WorkflowInstance,
	stepInstance *StepInstance, stepDef *WorkflowStep) error {

	// 更新步骤状态
	stepInstance.Status = "IN_PROGRESS"
	startedAt := time.Now()
	stepInstance.StartedAt = &startedAt
	stepInstance.UpdatedAt = time.Now()

	err := we.workflowRepo.UpdateStepInstance(ctx, stepInstance)
	if err != nil {
		return fmt.Errorf("failed to update step instance: %w", err)
	}

	// 获取处理器
	handler, exists := we.stepHandlers[stepDef.Action]
	if !exists {
		return fmt.Errorf("handler not found for action: %s", stepDef.Action)
	}

	// 验证条件
	canExecute, err := handler.Validate(ctx, order, stepDef)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if !canExecute {
		stepInstance.Status = "FAILED"
		stepInstance.Error = "validation failed"
		stepInstance.UpdatedAt = time.Now()

		err = we.workflowRepo.UpdateStepInstance(ctx, stepInstance)
		if err != nil {
			return fmt.Errorf("failed to update step instance: %w", err)
		}

		return fmt.Errorf("step validation failed")
	}

	// 执行步骤
	var executionErr error
	for retry := 0; retry <= stepDef.RetryCount; retry++ {
		executionErr = handler.Execute(ctx, order, stepDef)
		if executionErr == nil {
			break
		}

		if retry < stepDef.RetryCount {
			time.Sleep(we.config.RetryDelay)
			stepInstance.RetryCount++
		}
	}

	// 步骤执行成功后先落库，避免仅内存状态推进导致后续步骤读取旧状态。
	if executionErr == nil {
		if err = we.orderRepo.Update(ctx, order); err != nil {
			executionErr = fmt.Errorf("failed to persist order state: %w", err)
		}
	}

	// 更新步骤状态
	if executionErr != nil {
		stepInstance.Status = "FAILED"
		stepInstance.Error = executionErr.Error()
		stepInstance.UpdatedAt = time.Now()

		// 回滚
		rollbackErr := handler.Rollback(ctx, order, stepDef)
		if rollbackErr != nil {
			fmt.Printf("Failed to rollback step: %v\n", rollbackErr)
		}
	} else {
		stepInstance.Status = "COMPLETED"
		completedAt := time.Now()
		stepInstance.CompletedAt = &completedAt
		stepInstance.UpdatedAt = time.Now()

		// 更新工作流状态
		workflowInstance.PreviousState = workflowInstance.CurrentState
		workflowInstance.CurrentState = stepDef.ToState
		workflowInstance.UpdatedAt = time.Now()
	}

	// 保存更新
	err = we.workflowRepo.UpdateStepInstance(ctx, stepInstance)
	if err != nil {
		return fmt.Errorf("failed to update step instance: %w", err)
	}

	err = we.workflowRepo.UpdateWorkflowInstance(ctx, workflowInstance)
	if err != nil {
		return fmt.Errorf("failed to update workflow instance: %w", err)
	}

	if executionErr != nil {
		return fmt.Errorf("step execution failed: %w", executionErr)
	}

	// 执行下一步
	err = we.executeNextStep(ctx, order, workflowInstance)
	if err != nil {
		return fmt.Errorf("failed to execute next step: %w", err)
	}

	return nil
}

// getWorkflowDefinition 获取工作流定义
func (we *WorkflowEngine) getWorkflowDefinition(workflowID string) (*WorkflowDefinition, error) {
	we.mu.RLock()
	workflowDef, exists := we.workflows[workflowID]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow definition not found: %s", workflowID)
	}

	return workflowDef, nil
}

// getStepDefinition 获取步骤定义
func (we *WorkflowEngine) getStepDefinition(stepID string) (*WorkflowStep, error) {
	we.mu.RLock()
	defer we.mu.RUnlock()

	for _, workflowDef := range we.workflows {
		for _, step := range workflowDef.Steps {
			if step.ID == stepID {
				return step, nil
			}
		}
	}

	return nil, fmt.Errorf("step definition not found: %s", stepID)
}

// GetWorkflowStatus 获取工作流状态
func (we *WorkflowEngine) GetWorkflowStatus(ctx context.Context, orderID string) (*WorkflowInstance, error) {
	return we.workflowRepo.GetWorkflowInstanceByOrder(ctx, orderID)
}

// CancelWorkflow 取消工作流
func (we *WorkflowEngine) CancelWorkflow(ctx context.Context, orderID string, reason string) error {
	// 获取工作流实例
	workflowInstance, err := we.workflowRepo.GetWorkflowInstanceByOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get workflow instance: %w", err)
	}

	// 获取订单
	order, err := we.orderRepo.GetOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	// 执行取消步骤
	cancelStep := &WorkflowStep{
		ID:         "CANCEL_STEP",
		StepName:   "Cancel Order",
		Action:     ActionCancel,
		FromState:  workflowInstance.CurrentState,
		ToState:    StateCancelled,
		Handler:    "CancellationHandler",
		Timeout:    we.config.DefaultTimeout,
		RetryCount: we.config.MaxRetries,
	}

	// 执行取消
	handler, exists := we.stepHandlers[ActionCancel]
	if !exists {
		return fmt.Errorf("cancellation handler not found")
	}

	err = handler.Execute(ctx, order, cancelStep)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// 更新工作流状态
	workflowInstance.CurrentState = StateCancelled
	workflowInstance.Status = "COMPLETED"
	completedAt := time.Now()
	workflowInstance.CompletedAt = &completedAt
	workflowInstance.UpdatedAt = time.Now()

	// 保存更新
	err = we.workflowRepo.UpdateWorkflowInstance(ctx, workflowInstance)
	if err != nil {
		return fmt.Errorf("failed to update workflow instance: %w", err)
	}

	return nil
}

// Step Handlers

type ValidationHandler struct{}

func (vh *ValidationHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order has no items")
	}
	if order.ActualAmount <= 0 {
		return fmt.Errorf("order amount must be positive")
	}

	needsShipping := false
	for _, item := range order.Items {
		if item.ProductType != pb.ProductType_VIRTUAL {
			needsShipping = true
			break
		}
	}
	if needsShipping && order.ShippingAddress == nil {
		return fmt.Errorf("shipping address is required for physical goods")
	}

	order.AddLog("Workflow", "VALIDATE", order.Status.String(), order.Status.String(), "Order validation passed")
	order.UpdatedAt = time.Now()
	return nil
}

func (vh *ValidationHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可验证
	return order.Status == pb.OrderStatus_PENDING_PAYMENT, nil
}

func (vh *ValidationHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚验证
	return nil
}

type PaymentHandler struct{}

func (ph *PaymentHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	paymentMethod := order.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "WORKFLOW_AUTO"
	}

	if err := order.Pay(ctx, paymentMethod, "Workflow"); err != nil {
		return err
	}
	// 预售单在工作流支付环节直接补齐尾款，确保主链路可继续推进。
	if order.Status == pb.OrderStatus_PENDING_BALANCE {
		if err := order.PayBalance(ctx, paymentMethod, "Workflow"); err != nil {
			return err
		}
	}
	order.UpdatedAt = time.Now()
	return nil
}

func (ph *PaymentHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可支付
	return order.Status == pb.OrderStatus_PENDING_PAYMENT, nil
}

func (ph *PaymentHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚支付
	return nil
}

type ConfirmationHandler struct{}

func (ch *ConfirmationHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	order.AddLog("Workflow", "CONFIRM", order.Status.String(), order.Status.String(), "Order confirmed by workflow")
	order.UpdatedAt = time.Now()
	return nil
}

func (ch *ConfirmationHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可确认
	return order.Status == pb.OrderStatus_PAID, nil
}

func (ch *ConfirmationHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚确认
	return nil
}

type ProcessingHandler struct{}

func (ph *ProcessingHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	order.AddLog("Workflow", "PROCESS", order.Status.String(), order.Status.String(), "Order processing started")
	order.UpdatedAt = time.Now()
	return nil
}

func (ph *ProcessingHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可处理
	return order.Status == pb.OrderStatus_PAID, nil
}

func (ph *ProcessingHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚处理
	return nil
}

type ShippingHandler struct{}

func (sh *ShippingHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	if order.TrackingNumber == "" {
		order.TrackingNumber = fmt.Sprintf("WF-%d", time.Now().UnixNano())
	}
	if order.LogisticsCompany == "" {
		order.LogisticsCompany = "WORKFLOW"
	}
	if err := order.Ship(ctx, "Workflow"); err != nil {
		return err
	}
	order.UpdatedAt = time.Now()
	return nil
}

func (sh *ShippingHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可发货
	// 使用 PAID 状态，因为 PROCESSING 可能不存在
	return order.Status == pb.OrderStatus_PAID, nil
}

func (sh *ShippingHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚发货
	return nil
}

type DeliveryHandler struct{}

func (dh *DeliveryHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	if err := order.Deliver(ctx, "Workflow"); err != nil {
		return err
	}
	order.UpdatedAt = time.Now()
	return nil
}

func (dh *DeliveryHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可送达
	return order.Status == pb.OrderStatus_SHIPPED, nil
}

func (dh *DeliveryHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚送达
	return nil
}

type CompletionHandler struct{}

func (ch *CompletionHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	if err := order.Complete(ctx, "Workflow"); err != nil {
		return err
	}
	order.UpdatedAt = time.Now()
	return nil
}

func (ch *CompletionHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可完成
	return order.Status == pb.OrderStatus_DELIVERED, nil
}

func (ch *CompletionHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚完成
	return nil
}

type CancellationHandler struct{}

func (ch *CancellationHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	if order.Status == pb.OrderStatus_CANCELLED {
		return nil
	}
	reason := "Workflow cancellation"
	if step != nil && len(step.Conditions) > 0 && step.Conditions[0] != "" {
		reason = step.Conditions[0]
	}
	if err := order.Cancel(ctx, "Workflow", reason); err != nil {
		return err
	}
	order.UpdatedAt = time.Now()
	return nil
}

func (ch *CancellationHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可取消
	// 通常只有特定状态下的订单可以取消
	cancellableStates := []pb.OrderStatus{
		pb.OrderStatus_PENDING_PAYMENT,
		pb.OrderStatus_PAID,
		pb.OrderStatus_ALLOCATING,
		pb.OrderStatus_PENDING_BALANCE,
	}

	for _, state := range cancellableStates {
		if order.Status == state {
			return true, nil
		}
	}

	return false, nil
}

func (ch *CancellationHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚取消
	return nil
}

type RefundHandler struct{}

func (rh *RefundHandler) Execute(ctx context.Context, order *Order, step *WorkflowStep) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	reason := "Workflow auto refund"
	if step != nil && len(step.Conditions) > 0 && step.Conditions[0] != "" {
		reason = step.Conditions[0]
	}

	switch order.Status {
	case pb.OrderStatus_REFUND_REQUESTED:
		if err := order.ApproveRefund(ctx, "Workflow"); err != nil {
			return err
		}
	case pb.OrderStatus_PAID, pb.OrderStatus_SHIPPED, pb.OrderStatus_DELIVERED:
		if err := order.RequestRefund(ctx, "Workflow", reason); err != nil {
			return err
		}
		if err := order.ApproveRefund(ctx, "Workflow"); err != nil {
			return err
		}
	case pb.OrderStatus_CANCELLED:
		if order.PaymentStatus == pb.PaymentStatus_UNPAID || order.PaidAt == nil {
			order.AddLog("Workflow", "REFUND_SKIPPED", order.Status.String(), order.Status.String(), "Order unpaid, no refund required")
		} else {
			order.PaymentStatus = pb.PaymentStatus_REFUND_SUCCESS
			order.RefundAmount = order.ActualAmount
			order.RefundReason = reason
			order.AddLog("Workflow", "REFUND", order.Status.String(), order.Status.String(), "Refund marked successful")
		}
	default:
		return fmt.Errorf("order status %s cannot be refunded", order.Status.String())
	}

	order.UpdatedAt = time.Now()
	return nil
}

func (rh *RefundHandler) Validate(ctx context.Context, order *Order, step *WorkflowStep) (bool, error) {
	// 检查订单是否可退款
	return order.Status == pb.OrderStatus_CANCELLED, nil
}

func (rh *RefundHandler) Rollback(ctx context.Context, order *Order, step *WorkflowStep) error {
	// 回滚退款
	return nil
}

// Helper functions

func generateWorkflowInstanceID() string {
	return fmt.Sprintf("WORKFLOW_%d", time.Now().UnixNano())
}

// Repository interfaces

// WorkflowOrderRepositoryAdapter 适配器，为工作流提供 OrderRepository 接口
type WorkflowOrderRepositoryAdapter struct {
	repo OrderRepository
}

// NewWorkflowOrderRepositoryAdapter 创建适配器
func NewWorkflowOrderRepositoryAdapter(repo OrderRepository) *WorkflowOrderRepositoryAdapter {
	return &WorkflowOrderRepositoryAdapter{repo: repo}
}

// 实现 OrderRepository 接口的所有方法
func (a *WorkflowOrderRepositoryAdapter) BeginTx(ctx context.Context, userID uint64) any {
	return a.repo.BeginTx(ctx, userID)
}

func (a *WorkflowOrderRepositoryAdapter) CommitTx(tx any) error {
	return a.repo.CommitTx(tx)
}

func (a *WorkflowOrderRepositoryAdapter) RollbackTx(tx any) error {
	return a.repo.RollbackTx(tx)
}

func (a *WorkflowOrderRepositoryAdapter) WithTx(ctx context.Context, userID uint64, fn func(tx any) error) error {
	return a.repo.WithTx(ctx, userID, fn)
}

func (a *WorkflowOrderRepositoryAdapter) Save(ctx context.Context, order *Order) error {
	return a.repo.Save(ctx, order)
}

func (a *WorkflowOrderRepositoryAdapter) SaveInTx(ctx context.Context, tx any, order *Order) error {
	return a.repo.SaveInTx(ctx, tx, order)
}

func (a *WorkflowOrderRepositoryAdapter) FindByID(ctx context.Context, userID uint64, id uint64) (*Order, error) {
	return a.repo.FindByID(ctx, userID, id)
}

func (a *WorkflowOrderRepositoryAdapter) FindByOrderNo(ctx context.Context, userID uint64, orderNo string) (*Order, error) {
	return a.repo.FindByOrderNo(ctx, userID, orderNo)
}

func (a *WorkflowOrderRepositoryAdapter) FindByUserAndMerchant(ctx context.Context, userID uint64, merchantID uint64) ([]*Order, error) {
	return a.repo.FindByUserAndMerchant(ctx, userID, merchantID)
}

func (a *WorkflowOrderRepositoryAdapter) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order id is empty")
	}
	// 优先按主键 ID 查询，兼容 workflow 存储的数字 order_id。
	if numericID, err := strconv.ParseUint(orderID, 10, 64); err == nil {
		order, findErr := a.repo.FindByID(ctx, 0, numericID)
		if findErr != nil {
			return nil, findErr
		}
		if order != nil {
			return order, nil
		}
	}

	// 回退按业务单号查询，兼容历史 workflow 使用 order_no 的场景。
	order, err := a.repo.FindByOrderNo(ctx, 0, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}
	return order, nil
}

func (a *WorkflowOrderRepositoryAdapter) Update(ctx context.Context, order *Order) error {
	return a.repo.Update(ctx, order)
}

func (a *WorkflowOrderRepositoryAdapter) UpdateInTx(ctx context.Context, tx any, order *Order) error {
	return a.repo.UpdateInTx(ctx, tx, order)
}

func (a *WorkflowOrderRepositoryAdapter) Delete(ctx context.Context, userID uint64, id uint64) error {
	return a.repo.Delete(ctx, userID, id)
}

func (a *WorkflowOrderRepositoryAdapter) List(ctx context.Context, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Order, int64, error) {
	return a.repo.List(ctx, status, offset, limit, startTime, endTime, sortBy)
}

func (a *WorkflowOrderRepositoryAdapter) ListByUserID(ctx context.Context, userID uint64, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Order, int64, error) {
	return a.repo.ListByUserID(ctx, userID, status, offset, limit, startTime, endTime, sortBy)
}

func (a *WorkflowOrderRepositoryAdapter) Search(ctx context.Context, params *ExportQueryParams) ([]*Order, error) {
	return a.repo.Search(ctx, params)
}

type WorkflowRepository interface {
	SaveWorkflowInstance(ctx context.Context, instance *WorkflowInstance) error
	GetWorkflowInstance(ctx context.Context, instanceID string) (*WorkflowInstance, error)
	GetWorkflowInstanceByOrder(ctx context.Context, orderID string) (*WorkflowInstance, error)
	UpdateWorkflowInstance(ctx context.Context, instance *WorkflowInstance) error
	DeleteWorkflowInstance(ctx context.Context, instanceID string) error

	SaveStepInstance(ctx context.Context, step *StepInstance) error
	GetStepInstance(ctx context.Context, stepID string) (*StepInstance, error)
	UpdateStepInstance(ctx context.Context, step *StepInstance) error
	DeleteStepInstance(ctx context.Context, stepID string) error
	GetPendingSteps(ctx context.Context) ([]*StepInstance, error)

	GetWorkflowDefinitions(ctx context.Context) ([]*WorkflowDefinition, error)
	SaveWorkflowDefinition(ctx context.Context, definition *WorkflowDefinition) error
	UpdateWorkflowDefinition(ctx context.Context, definition *WorkflowDefinition) error
	DeleteWorkflowDefinition(ctx context.Context, definitionID string) error
}
