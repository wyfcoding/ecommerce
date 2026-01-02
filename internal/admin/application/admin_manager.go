package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/jwt"
	"github.com/wyfcoding/pkg/security"
)

// AdminManager 处理所有写操作（Command）
type AdminManager struct {
	userRepo     domain.AdminRepository
	roleRepo     domain.RoleRepository
	auditRepo    domain.AuditRepository
	settingRepo  domain.SettingRepository
	approvalRepo domain.ApprovalRepository

	opsDeps SystemOpsDependencies
	logger  *slog.Logger
}

func NewAdminManager(
	userRepo domain.AdminRepository,
	roleRepo domain.RoleRepository,
	auditRepo domain.AuditRepository,
	settingRepo domain.SettingRepository,
	approvalRepo domain.ApprovalRepository,
	opsDeps SystemOpsDependencies,
	logger *slog.Logger,
) *AdminManager {
	return &AdminManager{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		auditRepo:    auditRepo,
		settingRepo:  settingRepo,
		approvalRepo: approvalRepo,
		opsDeps:      opsDeps,
		logger:       logger,
	}
}

// --- Auth & User Management (Writes) ---

func (m *AdminManager) RegisterAdmin(ctx context.Context, req *CreateUserRequest) (*domain.AdminUser, error) {
	if _, err := m.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("username exists")
	}

	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	admin := &domain.AdminUser{
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: string(hashed),
		Status:       domain.UserStatusActive,
	}
	admin.ID = uint(idgen.GenID())

	if err := m.userRepo.Create(ctx, admin); err != nil {
		return nil, err
	}

	if len(req.Roles) > 0 {
		if err := m.userRepo.AssignRole(ctx, admin.ID, req.Roles); err != nil {
			return nil, err
		}
		return m.userRepo.GetByID(ctx, admin.ID)
	}

	return admin, nil
}

// Login 处理登录并返回 JWT。
func (m *AdminManager) Login(ctx context.Context, username, password string, secret, issuer string, expiry time.Duration) (string, *domain.AdminUser, error) {
	user, err := m.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, errors.New("user not found")
	}

	if user.Status != domain.UserStatusActive {
		return "", nil, errors.New("user is disabled")
	}

	if !security.CheckPassword(password, user.PasswordHash) {
		return "", nil, errors.New("invalid password")
	}

	// 提取角色用于授权
	var roles []string
	for _, r := range user.Roles {
		roles = append(roles, r.Code)
	}

	// 【闭环】：调用统一的 Token 生成函数
	token, err := jwt.GenerateToken(
		uint64(user.ID),
		user.Username,
		roles,
		secret,
		issuer,
		expiry,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	go func() {
		now := time.Now()
		user.LastLoginAt = &now
		if err := m.userRepo.Update(context.Background(), user); err != nil {
			m.logger.Warn("failed to update last login time", "user_id", user.ID, "error", err)
		}
	}()

	return token, user, nil
}

// ... (UpdateAdmin and other methods remain unchanged)

func (m *AdminManager) UpdateAdmin(ctx context.Context, id uint, email, fullName string, roleIDs []uint) (*domain.AdminUser, error) {
	user, err := m.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if email != "" {
		user.Email = email
	}
	if fullName != "" {
		user.FullName = fullName
	}
	if err := m.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	if roleIDs != nil {
		if err := m.userRepo.AssignRole(ctx, id, roleIDs); err != nil {
			return nil, err
		}
		return m.userRepo.GetByID(ctx, id)
	}
	return user, nil
}

func (m *AdminManager) DeleteAdmin(ctx context.Context, id uint) error {
	return m.userRepo.Delete(ctx, id)
}

func (m *AdminManager) CreateRole(ctx context.Context, name, code, description string) (*domain.Role, error) {
	role := &domain.Role{Name: name, Code: code, Description: description}
	if err := m.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (m *AdminManager) UpdateRole(ctx context.Context, id uint, name, description string, permissionIDs []uint) (*domain.Role, error) {
	role, err := m.roleRepo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	if name != "" {
		role.Name = name
	}
	if description != "" {
		role.Description = description
	}
	if err := m.roleRepo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}
	if permissionIDs != nil {
		if err := m.roleRepo.AssignPermissions(ctx, id, permissionIDs); err != nil {
			return nil, err
		}
		return m.roleRepo.GetRoleByID(ctx, id)
	}
	return role, nil
}

func (m *AdminManager) DeleteRole(ctx context.Context, id uint) error {
	return m.roleRepo.DeleteRole(ctx, id)
}

func (m *AdminManager) CreatePermission(ctx context.Context, name, code, permType, resource, action string, parentID uint) (*domain.Permission, error) {
	perm := &domain.Permission{Name: name, Code: code, Type: permType, Resource: resource, Action: action, ParentID: parentID}
	if err := m.roleRepo.CreatePermission(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (m *AdminManager) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
	return m.roleRepo.AssignPermissions(ctx, roleID, []uint{permissionID})
}

func (m *AdminManager) UpdateSystemSetting(ctx context.Context, key, value, description string) (*domain.SystemSetting, error) {
	setting := &domain.SystemSetting{Key: key, Value: value, Description: description}
	if err := m.settingRepo.Save(ctx, setting); err != nil {
		return nil, err
	}
	return setting, nil
}

func (m *AdminManager) LogAction(ctx context.Context, log *domain.AuditLog) {
	go func() {
		bgCtx := context.Background()
		if err := m.auditRepo.Save(bgCtx, log); err != nil {
			m.logger.Error("failed to save audit log", "error", err)
		}
	}()
}

func (m *AdminManager) CreateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	req.Status = domain.ApprovalStatusPending
	req.CurrentStep = 1

	// 根据业务类型和 Payload 内容决定审批流
	if err := m.determineApprovalFlow(req); err != nil {
		return err
	}

	if err := m.approvalRepo.CreateRequest(ctx, req); err != nil {
		return err
	}
	m.LogAction(ctx, &domain.AuditLog{
		UserID:   req.RequesterID,
		Action:   "workflow:create",
		Resource: "approval_request",
		TargetID: fmt.Sprintf("%d", req.ID),
		Status:   1,
		Payload:  req.Payload,
	})
	return nil
}

// determineApprovalFlow 决定审批所需的步骤数和初始/后续审批人角色
func (m *AdminManager) determineApprovalFlow(req *domain.ApprovalRequest) error {
	switch req.ActionType {
	case "ORDER_FORCE_REFUND":
		var payload struct {
			Amount float64 `json:"amount"`
		}
		if err := json.Unmarshal([]byte(req.Payload), &payload); err != nil {
			return fmt.Errorf("invalid payload for ORDER_FORCE_REFUND: %w", err)
		}
		// 金额 > 1000 需要两级审批：财务(FINANCE) -> 超管(SUPER_ADMIN)
		if payload.Amount > 1000 {
			req.TotalSteps = 2
			req.ApproverRole = "FINANCE" // 第一步
		} else {
			req.TotalSteps = 1
			req.ApproverRole = "FINANCE"
		}

	case "SYSTEM_CONFIG_UPDATE":
		req.TotalSteps = 1
		req.ApproverRole = "SUPER_ADMIN"

	default:
		// 默认通用流程
		req.TotalSteps = 1
		req.ApproverRole = "SUPER_ADMIN"
	}
	return nil
}

// calculateNextApprover 简单的流转逻辑，实际场景可能查库配置
func (m *AdminManager) calculateNextApprover(req *domain.ApprovalRequest) string {
	if req.ActionType == "ORDER_FORCE_REFUND" && req.CurrentStep == 2 {
		return "SUPER_ADMIN"
	}
	return "SUPER_ADMIN" // Default fallback
}

func (m *AdminManager) ApproveRequest(ctx context.Context, requestID, approverID uint, comment string) error {
	req, err := m.approvalRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusPending {
		return errors.New("request is not pending")
	}

	// 记录当前步骤的审批日志
	logEntry := &domain.ApprovalLog{
		RequestID:  req.ID,
		ApproverID: approverID,
		Action:     domain.ApprovalActionApprove,
		Comment:    comment,
	}
	if err := m.approvalRepo.AddLog(ctx, logEntry); err != nil {
		return err
	}

	// 判断是否还有后续步骤
	if req.CurrentStep < req.TotalSteps {
		req.CurrentStep++
		// 计算下一步的审批角色
		req.ApproverRole = m.calculateNextApprover(req)
		// 状态保持 Pending
		if err := m.approvalRepo.UpdateRequest(ctx, req); err != nil {
			return err
		}
		m.logger.InfoContext(ctx, "approval request moved to next step", "req_id", req.ID, "next_step", req.CurrentStep, "next_role", req.ApproverRole)
		return nil
	}

	// 最后一步完成，更新为已通过
	req.Status = domain.ApprovalStatusApproved
	now := time.Now()
	req.FinalizedAt = &now
	if err := m.approvalRepo.UpdateRequest(ctx, req); err != nil {
		return err
	}

	// 异步执行具体的业务操作
	go func() {
		bgCtx := context.Background()
		if err := m.executeOperation(bgCtx, req); err != nil {
			m.logger.Error("failed to execute operation", "reqID", req.ID, "error", err)
			// 记录失败状态
			req.Status = domain.ApprovalStatusFailed
			req.FailureReason = err.Error()
			_ = m.approvalRepo.UpdateRequest(bgCtx, req)
		}
	}()
	return nil
}

// RetryFailedRequest 手动重试执行失败的审批请求
func (m *AdminManager) RetryFailedRequest(ctx context.Context, requestID uint) error {
	req, err := m.approvalRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusFailed {
		return errors.New("only failed requests can be retried")
	}

	req.Status = domain.ApprovalStatusApproved // 临时恢复为 Approved 状态进行重试
	req.RetryCount++

	if err := m.executeOperation(ctx, req); err != nil {
		req.Status = domain.ApprovalStatusFailed
		req.FailureReason = fmt.Sprintf("Retry %d failed: %s", req.RetryCount, err.Error())
		_ = m.approvalRepo.UpdateRequest(ctx, req)
		return err
	}

	req.Status = domain.ApprovalStatusApproved
	req.FailureReason = ""
	return m.approvalRepo.UpdateRequest(ctx, req)
}

func (m *AdminManager) RejectRequest(ctx context.Context, requestID, approverID uint, comment string) error {
	req, err := m.approvalRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusPending {
		return errors.New("request is not pending")
	}
	logEntry := &domain.ApprovalLog{RequestID: req.ID, ApproverID: approverID, Action: domain.ApprovalActionReject, Comment: comment}
	if err := m.approvalRepo.AddLog(ctx, logEntry); err != nil {
		return err
	}
	req.Status = domain.ApprovalStatusRejected
	now := time.Now()
	req.FinalizedAt = &now
	return m.approvalRepo.UpdateRequest(ctx, req)
}

func (m *AdminManager) executeOperation(ctx context.Context, req *domain.ApprovalRequest) error {
	m.logger.Info("executing approved operation", "type", req.ActionType, "req_id", req.ID)
	switch req.ActionType {
	case "ORDER_FORCE_REFUND":
		return m.handleForceRefund(ctx, req.Payload)
	case "SYSTEM_CONFIG_UPDATE":
		return m.handleConfigUpdate(ctx, req.Payload)
	default:
		return fmt.Errorf("unknown action type: %s", req.ActionType)
	}
}

func (m *AdminManager) handleForceRefund(ctx context.Context, payloadStr string) error {
	var payload struct {
		OrderID   string  `json:"orderId"`
		Amount    float64 `json:"amount"`
		Reason    string  `json:"reason"`
		UserID    uint64  `json:"userId"`
		PaymentID string  `json:"paymentId"`
	}
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return fmt.Errorf("unmarshal payload failed: %w", err)
	}

	orderID, _ := strconv.ParseUint(payload.OrderID, 10, 64)
	paymentTxID, _ := strconv.ParseUint(payload.PaymentID, 10, 64)
	refundAmountCents := int64(payload.Amount * 100)

	auditLog := &domain.AuditLog{
		Action:   "order:force_refund",
		Resource: "order",
		TargetID: payload.OrderID,
		Payload:  payloadStr,
		Status:   1,
	}

	// 1. 调用订单服务发起退款请求/取消订单
	orderClient := orderv1.NewOrderServiceClient(m.opsDeps.OrderClient)
	_, err := orderClient.RequestRefund(ctx, &orderv1.RequestRefundRequest{
		OrderId:      orderID,
		UserId:       payload.UserID,
		RefundAmount: refundAmountCents,
		Reason:       payload.Reason,
	})
	if err != nil {
		auditLog.Status = 0
		auditLog.Result = fmt.Sprintf("order service failed: %v", err)
		m.LogAction(ctx, auditLog)
		return fmt.Errorf("call order service failed: %w", err)
	}

	// 2. 如果需要直接操作支付网关退款
	if m.opsDeps.PaymentClient != nil {
		paymentClient := paymentv1.NewPaymentServiceClient(m.opsDeps.PaymentClient)
		_, err := paymentClient.RequestRefund(ctx, &paymentv1.RequestRefundRequest{
			PaymentTransactionId: paymentTxID,
			OrderId:              orderID,
			UserId:               payload.UserID,
			RefundAmount:         refundAmountCents,
			Reason:               payload.Reason,
		})
		if err != nil {
			auditLog.Status = 0
			auditLog.Result = fmt.Sprintf("payment service failed: %v", err)
			m.LogAction(ctx, auditLog)
			return fmt.Errorf("call payment service failed: %w", err)
		}
	}

	auditLog.Result = "Success"
	m.LogAction(ctx, auditLog)
	m.logger.Info("force refund executed successfully", "order_id", payload.OrderID)
	return nil
}

func (m *AdminManager) handleConfigUpdate(ctx context.Context, payloadStr string) error {
	var payload struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return fmt.Errorf("unmarshal payload failed: %w", err)
	}

	auditLog := &domain.AuditLog{
		Action:   "config:update",
		Resource: "system_setting",
		TargetID: payload.Key,
		Payload:  payloadStr,
		Status:   1,
	}

	// 实际更新数据库配置
	setting := &domain.SystemSetting{
		Key:         payload.Key,
		Value:       payload.Value,
		Description: payload.Description,
	}

	if err := m.settingRepo.Save(ctx, setting); err != nil {
		auditLog.Status = 0
		auditLog.Result = err.Error()
		m.LogAction(ctx, auditLog)
		return fmt.Errorf("save setting failed: %w", err)
	}

	auditLog.Result = "Success"
	m.LogAction(ctx, auditLog)
	m.logger.Info("system config updated", "key", payload.Key)
	return nil
}
