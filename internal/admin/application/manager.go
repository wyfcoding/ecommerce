package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

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

	// 【闭环】：调用增强版 Token 生成函数，保留 Roles 优化
	token, err := jwt.GenerateTokenWithRoles(
		uint64(user.ID),
		user.Username,
		roles,
		secret,
		issuer,
		expiry,
		nil,
	)

	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	go func() {
		now := time.Now()
		user.LastLoginAt = &now
		_ = m.userRepo.Update(context.Background(), user)
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
	req.TotalSteps = 1
	req.ApproverRole = "SUPER_ADMIN"
	if err := m.approvalRepo.CreateRequest(ctx, req); err != nil {
		return err
	}
	m.LogAction(ctx, &domain.AuditLog{UserID: req.RequesterID, Action: "workflow:create", Resource: "approval_request", TargetID: fmt.Sprintf("%d", req.ID), Status: 1})
	return nil
}

func (m *AdminManager) ApproveRequest(ctx context.Context, requestID, approverID uint, comment string) error {
	req, err := m.approvalRepo.GetRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.Status != domain.ApprovalStatusPending {
		return errors.New("request is not pending")
	}
	logEntry := &domain.ApprovalLog{RequestID: req.ID, ApproverID: approverID, Action: domain.ApprovalActionApprove, Comment: comment}
	if err := m.approvalRepo.AddLog(ctx, logEntry); err != nil {
		return err
	}
	req.Status = domain.ApprovalStatusApproved
	now := time.Now()
	req.FinalizedAt = &now
	if err := m.approvalRepo.UpdateRequest(ctx, req); err != nil {
		return err
	}
	go func() {
		bgCtx := context.Background()
		if err := m.executeOperation(bgCtx, req); err != nil {
			m.logger.Error("failed to execute operation", "reqID", req.ID, "error", err)
		}
	}()
	return nil
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
	m.logger.InfoContext(ctx, "Mocking Force Refund (Real gRPC call pending)", "payload", payload)
	return nil
}

func (m *AdminManager) handleConfigUpdate(ctx context.Context, payload string) error {
	m.logger.InfoContext(ctx, "Mocking Config Update", "payload", payload)
	return nil
}