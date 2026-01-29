package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// AdminQueryService 处理所有读操作（Query）
type AdminQueryService struct {
	userRepo     domain.AdminRepository
	roleRepo     domain.RoleRepository
	auditRepo    domain.AuditRepository
	settingRepo  domain.SettingRepository
	approvalRepo domain.ApprovalRepository
}

func NewAdminQueryService(
	userRepo domain.AdminRepository,
	roleRepo domain.RoleRepository,
	auditRepo domain.AuditRepository,
	settingRepo domain.SettingRepository,
	approvalRepo domain.ApprovalRepository,
) *AdminQueryService {
	return &AdminQueryService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		auditRepo:    auditRepo,
		settingRepo:  settingRepo,
		approvalRepo: approvalRepo,
	}
}

// --- Admin Queries ---

func (q *AdminQueryService) GetAdminProfile(ctx context.Context, id uint) (*domain.AdminUser, error) {
	return q.userRepo.GetByID(ctx, id)
}

func (q *AdminQueryService) ListAdmins(ctx context.Context, page, pageSize int) ([]*domain.AdminUser, int64, error) {
	return q.userRepo.List(ctx, page, pageSize)
}

func (q *AdminQueryService) CheckPermission(ctx context.Context, userID uint, requiredPerm string) (bool, error) {
	perms, err := q.userRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, p := range perms {
		if p == requiredPerm {
			return true, nil
		}
		if p == "*:*" {
			return true, nil
		}
	}
	return false, nil
}

// --- Role & Permission Queries ---

func (q *AdminQueryService) GetRole(ctx context.Context, id uint) (*domain.Role, error) {
	return q.roleRepo.GetRoleByID(ctx, id)
}

func (q *AdminQueryService) ListRoles(ctx context.Context) ([]*domain.Role, int64, error) {
	roles, err := q.roleRepo.ListRoles(ctx)
	if err != nil {
		return nil, 0, err
	}
	return roles, int64(len(roles)), nil
}

func (q *AdminQueryService) GetPermission(ctx context.Context, id uint) (*domain.Permission, error) {
	return q.roleRepo.GetPermissionByID(ctx, id)
}

func (q *AdminQueryService) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return q.roleRepo.ListPermissions(ctx)
}

// --- Setting Queries ---

func (q *AdminQueryService) GetSystemSetting(ctx context.Context, key string) (*domain.SystemSetting, error) {
	return q.settingRepo.GetByKey(ctx, key)
}

// --- Audit Queries ---

func (q *AdminQueryService) ListAuditLogs(ctx context.Context, adminID uint, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	filter := make(map[string]any)
	if adminID > 0 {
		filter["user_id"] = adminID
	}
	return q.auditRepo.Find(ctx, filter, page, pageSize)
}

// --- Approval Queries ---

func (q *AdminQueryService) GetApprovalRequest(ctx context.Context, id uint) (*domain.ApprovalRequest, error) {
	return q.approvalRepo.GetRequestByID(ctx, id)
}

func (q *AdminQueryService) ListPendingRequests(ctx context.Context, roleLimit string) ([]*domain.ApprovalRequest, error) {
	return q.approvalRepo.ListPendingRequests(ctx, roleLimit)
}
