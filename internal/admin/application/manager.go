package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
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

// RegisterAdmin 注册一个新的管理员。
func (m *AdminManager) RegisterAdmin(ctx context.Context, req *CreateUserRequest) (*domain.AdminUser, error) {
	// 检查是否存在
	if _, err := m.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("username exists")
	}

	// 密码哈希
	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	admin := &domain.AdminUser{
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: string(hashed),
		Status:       domain.UserStatusActive, // 默认启用
	}

	if err := m.userRepo.Create(ctx, admin); err != nil {
		return nil, err
	}

	// 分配角色
	if len(req.Roles) > 0 {
		if err := m.userRepo.AssignRole(ctx, admin.ID, req.Roles); err != nil {
			return nil, err
		}
		// 重新获取以包含角色信息
		return m.userRepo.GetByID(ctx, admin.ID)
	}

	return admin, nil
}

// Login 处理登录逻辑（验证 + 更新最后登录时间）。
func (m *AdminManager) Login(ctx context.Context, username, password string) (*domain.AdminUser, error) {
	user, err := m.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.Status != domain.UserStatusActive {
		return nil, errors.New("user is disabled")
	}

	if !security.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid password")
	}

	// 更新登录时间
	now := time.Now()
	user.LastLoginAt = &now
	if err := m.userRepo.Update(ctx, user); err != nil {
		m.logger.Warn("failed to update last login time", "err", err)
	}

	return user, nil
}

// UpdateAdmin 更新管理员信息。
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

	if roleIDs != nil { // empty slice means clear roles? assume yes or handle differently. Here assuming replace.
		if err := m.userRepo.AssignRole(ctx, id, roleIDs); err != nil {
			return nil, err
		}
		return m.userRepo.GetByID(ctx, id)
	}

	return user, nil
}

// DeleteAdmin 删除管理员。
func (m *AdminManager) DeleteAdmin(ctx context.Context, id uint) error {
	return m.userRepo.Delete(ctx, id)
}

// --- Role Management (Writes) ---

// CreateRole 创建一个新的角色。
func (m *AdminManager) CreateRole(ctx context.Context, name, code, description string) (*domain.Role, error) {
	role := &domain.Role{
		Name:        name,
		Code:        code,
		Description: description,
	}
	if err := m.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// UpdateRole 更新角色信息。
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

// DeleteRole 删除指定ID的角色。
func (m *AdminManager) DeleteRole(ctx context.Context, id uint) error {
	return m.roleRepo.DeleteRole(ctx, id)
}

// --- Permission Management (Writes) ---

// CreatePermission 创建一个新的权限项。
func (m *AdminManager) CreatePermission(ctx context.Context, name, code, permType, resource, action string, parentID uint) (*domain.Permission, error) {
	perm := &domain.Permission{
		Name:     name,
		Code:     code,
		Type:     permType,
		Resource: resource,
		Action:   action,
		ParentID: parentID,
	}
	if err := m.roleRepo.CreatePermission(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

// AssignPermissionToRole 为角色分配权限。
func (m *AdminManager) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint) error {
	// 这是一个追加还是替换？通常 AssignPermissions 是替换全部。这里简化为替换为单个（原逻辑如此），或者应该先获取再追加。
	// 为保持一致性，假设这只是演示，或者应该改为 AddPermissionToRole。
	// 按照原逻辑：`s.roleRepo.AssignPermissions(ctx, uint(roleID), []uint{uint(permissionID)})` 确实是替换。
	return m.roleRepo.AssignPermissions(ctx, roleID, []uint{permissionID})
}

// --- System Setting (Writes) ---

// UpdateSystemSetting 更新系统设置信息。
func (m *AdminManager) UpdateSystemSetting(ctx context.Context, key, value, description string) (*domain.SystemSetting, error) {
	setting := &domain.SystemSetting{
		Key:         key,
		Value:       value,
		Description: description,
	}
	if err := m.settingRepo.Save(ctx, setting); err != nil {
		return nil, err
	}
	return setting, nil
}

// --- Audit (Writes) ---

// LogAction 记录操作日志
func (m *AdminManager) LogAction(ctx context.Context, log *domain.AuditLog) {
	// 异步写入
	go func() {
		bgCtx := context.Background()
		if err := m.auditRepo.Save(bgCtx, log); err != nil {
			m.logger.Error("failed to save audit log", "error", err, "action", log.Action)
		}
	}()
}

// --- Workflow & System Ops (Writes/Execution) ---

// CreateRequest 提交新的审批申请
func (m *AdminManager) CreateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	req.Status = domain.ApprovalStatusPending
	req.CurrentStep = 1
	// 简单逻辑：假设所有流程都是1步审批，需要 "SUPER_ADMIN" 角色
	req.TotalSteps = 1
	req.ApproverRole = "SUPER_ADMIN"

	if err := m.approvalRepo.CreateRequest(ctx, req); err != nil {
		return err
	}

	m.LogAction(ctx, &domain.AuditLog{
		UserID:   req.RequesterID,
		Action:   "workflow:create",
		Resource: "approval_request",
		TargetID: fmt.Sprintf("%d", req.ID),
		Status:   1,
	})
	return nil
}

// ApproveRequest 审批通过
func (m *AdminManager) ApproveRequest(ctx context.Context, requestID, approverID uint, comment string) error {
	req, err := m.approvalRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusPending {
		return errors.New("request is not pending")
	}

	// 记录审批Log
	logEntry := &domain.ApprovalLog{
		RequestID:  req.ID,
		ApproverID: approverID,
		Action:     domain.ApprovalActionApprove,
		Comment:    comment,
	}
	if err := m.approvalRepo.AddLog(ctx, logEntry); err != nil {
		return err
	}

	// 更新状态
	req.Status = domain.ApprovalStatusApproved
	now := time.Now()
	req.FinalizedAt = &now
	if err := m.approvalRepo.UpdateRequest(ctx, req); err != nil {
		return err
	}

	// 自动执行操作
	go func() {
		bgCtx := context.Background()
		if err := m.executeOperation(bgCtx, req); err != nil {
			m.logger.Error("failed to execute operation after approval", "reqID", req.ID, "error", err)
		}
	}()

	return nil
}

// RejectRequest 审批拒绝
func (m *AdminManager) RejectRequest(ctx context.Context, requestID, approverID uint, comment string) error {
	req, err := m.approvalRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusPending {
		return errors.New("request is not pending")
	}

	logEntry := &domain.ApprovalLog{
		RequestID:  req.ID,
		ApproverID: approverID,
		Action:     domain.ApprovalActionReject,
		Comment:    comment,
	}
	if err := m.approvalRepo.AddLog(ctx, logEntry); err != nil {
		return err
	}

	req.Status = domain.ApprovalStatusRejected
	now := time.Now()
	req.FinalizedAt = &now
	return m.approvalRepo.UpdateRequest(ctx, req)
}

// executeOperation 执行系统操作 (原 SystemOpsService)
func (m *AdminManager) executeOperation(ctx context.Context, req *domain.ApprovalRequest) error {
	m.logger.Info("executing system operation", "type", req.ActionType, "payload", req.Payload)

	switch req.ActionType {
	case "ORDER_FORCE_REFUND":
		return m.handleForceRefund(ctx, req.Payload)
	case "SYSTEM_CONFIG_UPDATE":
		return m.handleConfigUpdate(ctx, req.Payload)
	default:
		return fmt.Errorf("unknown action type: %s", req.ActionType)
	}
}

func (m *AdminManager) handleForceRefund(ctx context.Context, payload string) error {
	// Mock implementation
	m.logger.InfoContext(ctx, "Mocking Force Refund (Real gRPC call pending)", "payload", payload)
	return nil
}

func (m *AdminManager) handleConfigUpdate(ctx context.Context, payload string) error {
	// Mock implementation
	m.logger.InfoContext(ctx, "Mocking Config Update", "payload", payload)
	return nil
}
