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
	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/security"
)

// AdminCommandService 处理所有写操作（Command）
type AdminCommandService struct {
	userRepo     domain.AdminRepository
	roleRepo     domain.RoleRepository
	auditRepo    domain.AuditRepository
	settingRepo  domain.SettingRepository
	approvalRepo domain.ApprovalRepository

	opsDeps   SystemOpsDependencies
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
}

func NewAdminCommandService(
	userRepo domain.AdminRepository,
	roleRepo domain.RoleRepository,
	auditRepo domain.AuditRepository,
	settingRepo domain.SettingRepository,
	approvalRepo domain.ApprovalRepository,
	opsDeps SystemOpsDependencies,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *AdminCommandService {
	return &AdminCommandService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		auditRepo:    auditRepo,
		settingRepo:  settingRepo,
		approvalRepo: approvalRepo,
		opsDeps:      opsDeps,
		publisher:    publisher,
		logger:       logger,
	}
}

// --- Auth & User Management (Writes) ---

func (m *AdminCommandService) RegisterAdmin(ctx context.Context, req *CreateUserRequest) (*domain.AdminUser, error) {
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

	if err := m.userRepo.WithTx(ctx, func(tx any) error {
		if err := m.userRepo.CreateInTx(ctx, tx, admin); err != nil {
			return err
		}
		// 发布领域事件
		if err := m.publisher.PublishInTx(ctx, tx, domain.AdminUserCreatedEventType, fmt.Sprintf("%d", admin.ID), &domain.AdminUserCreatedEvent{
			UserID:    admin.ID,
			Username:  admin.Username,
			Email:     admin.Email,
			Timestamp: time.Now(),
		}); err != nil {
			return err
		}

		if len(req.Roles) > 0 {
			if err := m.userRepo.AssignRoleInTx(ctx, tx, admin.ID, req.Roles); err != nil {
				return err
			}
			_ = m.publisher.PublishInTx(ctx, tx, domain.RoleAssignedEventType, fmt.Sprintf("%d", admin.ID), &domain.RoleAssignedEvent{
				UserID:    admin.ID,
				RoleIDs:   req.Roles,
				Operator:  admin.Username,
				Timestamp: time.Now(),
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(req.Roles) > 0 {
		return m.userRepo.GetByID(ctx, admin.ID)
	}
	return admin, nil
}

// Login 处理登录并返回 JWT。
func (m *AdminCommandService) Login(ctx context.Context, username, password string, secret, issuer string, expiry time.Duration) (string, *domain.AdminUser, error) {
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

func (m *AdminCommandService) UpdateAdmin(ctx context.Context, id uint, email, fullName string, roleIDs []uint) (*domain.AdminUser, error) {
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
	if err := m.userRepo.WithTx(ctx, func(tx any) error {
		if err := m.userRepo.UpdateInTx(ctx, tx, user); err != nil {
			return err
		}
		if err := m.publisher.PublishInTx(ctx, tx, domain.AdminUserUpdatedEventType, fmt.Sprintf("%d", user.ID), &domain.AdminUserUpdatedEvent{
			UserID:    user.ID,
			Username:  user.Username,
			Timestamp: time.Now(),
		}); err != nil {
			return err
		}
		if roleIDs != nil {
			if err := m.userRepo.AssignRoleInTx(ctx, tx, id, roleIDs); err != nil {
				return err
			}
			_ = m.publisher.PublishInTx(ctx, tx, domain.RoleAssignedEventType, fmt.Sprintf("%d", user.ID), &domain.RoleAssignedEvent{
				UserID:    user.ID,
				RoleIDs:   roleIDs,
				Operator:  user.Username,
				Timestamp: time.Now(),
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if roleIDs != nil {
		return m.userRepo.GetByID(ctx, id)
	}
	return user, nil
}

func (m *AdminCommandService) DeleteAdmin(ctx context.Context, id uint) error {
	return m.userRepo.WithTx(ctx, func(tx any) error {
		if err := m.userRepo.DeleteInTx(ctx, tx, id); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.AdminUserDisabledEventType, fmt.Sprintf("%d", id), &domain.AdminUserDisabledEvent{
			UserID:    id,
			Username:  "",
			Reason:    "deleted",
			Timestamp: time.Now(),
		})
	})
}

func (m *AdminCommandService) CreateRole(ctx context.Context, name, code, description string) (*domain.Role, error) {
	role := &domain.Role{Name: name, Code: code, Description: description}
	if err := m.roleRepo.WithTx(ctx, func(tx any) error {
		if err := m.roleRepo.CreateRoleInTx(ctx, tx, role); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.RoleCreatedEventType, fmt.Sprintf("%d", role.ID), &domain.RoleCreatedEvent{
			RoleID:    role.ID,
			Name:      role.Name,
			Code:      role.Code,
			Timestamp: time.Now(),
		})
	}); err != nil {
		return nil, err
	}
	return role, nil
}

func (m *AdminCommandService) UpdateRole(ctx context.Context, id uint, name, description string, permissionIDs []uint) (*domain.Role, error) {
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
	if err := m.roleRepo.WithTx(ctx, func(tx any) error {
		if err := m.roleRepo.UpdateRoleInTx(ctx, tx, role); err != nil {
			return err
		}
		if permissionIDs != nil {
			if err := m.roleRepo.AssignPermissionsInTx(ctx, tx, id, permissionIDs); err != nil {
				return err
			}
		}
		return m.publisher.PublishInTx(ctx, tx, domain.RoleUpdatedEventType, fmt.Sprintf("%d", role.ID), &domain.RoleUpdatedEvent{
			RoleID:    role.ID,
			Name:      role.Name,
			Code:      role.Code,
			Timestamp: time.Now(),
		})
	}); err != nil {
		return nil, err
	}
	if permissionIDs != nil {
		return m.roleRepo.GetRoleByID(ctx, id)
	}
	return role, nil
}

func (m *AdminCommandService) DeleteRole(ctx context.Context, id uint) error {
	return m.roleRepo.WithTx(ctx, func(tx any) error {
		if err := m.roleRepo.DeleteRoleInTx(ctx, tx, id); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.RoleDeletedEventType, fmt.Sprintf("%d", id), &domain.RoleDeletedEvent{
			RoleID:    id,
			Timestamp: time.Now(),
		})
	})
}

func (m *AdminCommandService) CreatePermission(ctx context.Context, name, code, permType, resource, action string, parentID uint) (*domain.Permission, error) {
	perm := &domain.Permission{Name: name, Code: code, Type: permType, Resource: resource, Action: action, ParentID: parentID}
	if err := m.roleRepo.WithTx(ctx, func(tx any) error {
		if err := m.roleRepo.CreatePermissionInTx(ctx, tx, perm); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.PermissionCreatedEventType, fmt.Sprintf("%d", perm.ID), &domain.PermissionCreatedEvent{
			PermissionID: perm.ID,
			Name:         perm.Name,
			Code:         perm.Code,
			Timestamp:    time.Now(),
		})
	}); err != nil {
		return nil, err
	}
	return perm, nil
}

func (m *AdminCommandService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
	return m.roleRepo.WithTx(ctx, func(tx any) error {
		if err := m.roleRepo.AssignPermissionsInTx(ctx, tx, roleID, []uint{permissionID}); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.RoleUpdatedEventType, fmt.Sprintf("%d", roleID), &domain.RoleUpdatedEvent{
			RoleID:    roleID,
			Name:      "",
			Code:      "",
			Timestamp: time.Now(),
		})
	})
}

func (m *AdminCommandService) UpdateSystemSetting(ctx context.Context, key, value, description string) (*domain.SystemSetting, error) {
	var oldValue string
	if current, err := m.settingRepo.GetByKey(ctx, key); err == nil && current != nil {
		oldValue = current.Value
	}
	setting := &domain.SystemSetting{Key: key, Value: value, Description: description}
	if err := m.settingRepo.WithTx(ctx, func(tx any) error {
		if err := m.settingRepo.SaveInTx(ctx, tx, setting); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.SystemSettingUpdatedEventType, key, &domain.SystemSettingUpdatedEvent{
			Key:       key,
			OldValue:  oldValue,
			NewValue:  value,
			Operator:  "",
			Timestamp: time.Now(),
		})
	}); err != nil {
		return nil, err
	}
	return setting, nil
}

func (m *AdminCommandService) LogAction(ctx context.Context, log *domain.AuditLog) {
	go func() {
		bgCtx := context.Background()
		if err := m.auditRepo.WithTx(bgCtx, func(tx any) error {
			if err := m.auditRepo.SaveInTx(bgCtx, tx, log); err != nil {
				return err
			}
			return m.publisher.PublishInTx(bgCtx, tx, domain.AuditLogCreatedEventType, fmt.Sprintf("%d", log.ID), &domain.AuditLogCreatedEvent{
				LogID:     log.ID,
				UserID:    log.UserID,
				Action:    log.Action,
				Resource:  log.Resource,
				Timestamp: time.Now(),
			})
		}); err != nil {
			m.logger.Error("failed to save audit log", "error", err)
		}
	}()
}

func (m *AdminCommandService) CreateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	req.Status = domain.ApprovalStatusPending
	req.CurrentStep = 1

	// 根据业务类型和 Payload 内容决定审批流
	if err := m.determineApprovalFlow(req); err != nil {
		return err
	}

	if err := m.approvalRepo.WithTx(ctx, func(tx any) error {
		if err := m.approvalRepo.CreateRequestInTx(ctx, tx, req); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ApprovalRequestCreatedEventType, fmt.Sprintf("%d", req.ID), &domain.ApprovalRequestCreatedEvent{
			RequestID:   req.ID,
			RequesterID: req.RequesterID,
			ActionType:  req.ActionType,
			Timestamp:   time.Now(),
		})
	}); err != nil {
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
func (m *AdminCommandService) determineApprovalFlow(req *domain.ApprovalRequest) error {
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
func (m *AdminCommandService) calculateNextApprover(req *domain.ApprovalRequest) string {
	if req.ActionType == "ORDER_FORCE_REFUND" && req.CurrentStep == 2 {
		return "SUPER_ADMIN"
	}
	return "SUPER_ADMIN" // Default fallback
}

func (m *AdminCommandService) ApproveRequest(ctx context.Context, requestID, approverID uint, comment string) error {
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

	var approvedFinal bool
	if err := m.approvalRepo.WithTx(ctx, func(tx any) error {
		if err := m.approvalRepo.AddLogInTx(ctx, tx, logEntry); err != nil {
			return err
		}

		// 判断是否还有后续步骤
		if req.CurrentStep < req.TotalSteps {
			req.CurrentStep++
			// 计算下一步的审批角色
			req.ApproverRole = m.calculateNextApprover(req)
			// 状态保持 Pending
			if err := m.approvalRepo.UpdateRequestInTx(ctx, tx, req); err != nil {
				return err
			}
			return nil
		}

		// 最后一步完成，更新为已通过
		req.Status = domain.ApprovalStatusApproved
		now := time.Now()
		req.FinalizedAt = &now
		if err := m.approvalRepo.UpdateRequestInTx(ctx, tx, req); err != nil {
			return err
		}
		approvedFinal = true
		return m.publisher.PublishInTx(ctx, tx, domain.ApprovalRequestApprovedEventType, fmt.Sprintf("%d", req.ID), &domain.ApprovalRequestApprovedEvent{
			RequestID:  req.ID,
			ApproverID: approverID,
			Comment:    comment,
			Timestamp:  time.Now(),
		})
	}); err != nil {
		return err
	}

	if !approvedFinal {
		m.logger.InfoContext(ctx, "approval request moved to next step", "req_id", req.ID, "next_step", req.CurrentStep, "next_role", req.ApproverRole)
		return nil
	}

	// 异步执行具体的业务操作
	go func() {
		bgCtx := context.Background()
		if err := m.executeOperation(bgCtx, req); err != nil {
			m.logger.Error("failed to execute operation", "reqID", req.ID, "error", err)
			// 记录失败状态
			req.Status = domain.ApprovalStatusFailed
			req.FailureReason = err.Error()
			if err := m.approvalRepo.UpdateRequest(bgCtx, req); err != nil {
				m.logger.Error("failed to update request status to failed", "reqID", req.ID, "error", err)
			}
		}
	}()
	return nil
}

// RetryFailedRequest 手动重试执行失败的审批请求
func (m *AdminCommandService) RetryFailedRequest(ctx context.Context, requestID uint) error {
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
		if err := m.approvalRepo.UpdateRequest(ctx, req); err != nil {
			m.logger.ErrorContext(ctx, "failed to update request status during retry", "reqID", req.ID, "error", err)
		}
		return err
	}

	req.Status = domain.ApprovalStatusApproved
	req.FailureReason = ""
	return m.approvalRepo.UpdateRequest(ctx, req)
}

func (m *AdminCommandService) RejectRequest(ctx context.Context, requestID, approverID uint, comment string) error {
	req, err := m.approvalRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusPending {
		return errors.New("request is not pending")
	}
	logEntry := &domain.ApprovalLog{RequestID: req.ID, ApproverID: approverID, Action: domain.ApprovalActionReject, Comment: comment}
	return m.approvalRepo.WithTx(ctx, func(tx any) error {
		if err := m.approvalRepo.AddLogInTx(ctx, tx, logEntry); err != nil {
			return err
		}
		req.Status = domain.ApprovalStatusRejected
		now := time.Now()
		req.FinalizedAt = &now
		if err := m.approvalRepo.UpdateRequestInTx(ctx, tx, req); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ApprovalRequestRejectedEventType, fmt.Sprintf("%d", req.ID), &domain.ApprovalRequestRejectedEvent{
			RequestID:  req.ID,
			ApproverID: approverID,
			Reason:     comment,
			Timestamp:  time.Now(),
		})
	})
}

func (m *AdminCommandService) executeOperation(ctx context.Context, req *domain.ApprovalRequest) error {
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

func (m *AdminCommandService) handleForceRefund(ctx context.Context, payloadStr string) error {
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

func (m *AdminCommandService) handleConfigUpdate(ctx context.Context, payloadStr string) error {
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
