package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// AdminQueryService 处理所有读操作（Query）
type AdminQueryService struct {
	userRepo        domain.AdminRepository
	roleRepo        domain.RoleRepository
	auditRepo       domain.AuditRepository
	settingRepo     domain.SettingRepository
	approvalRepo    domain.ApprovalRepository
	userReadRepo    domain.AdminUserReadRepository
	settingReadRepo domain.SettingReadRepository
	auditSearchRepo domain.AuditLogSearchRepository
}

func NewAdminQueryService(
	userRepo domain.AdminRepository,
	roleRepo domain.RoleRepository,
	auditRepo domain.AuditRepository,
	settingRepo domain.SettingRepository,
	approvalRepo domain.ApprovalRepository,
	userReadRepo domain.AdminUserReadRepository,
	settingReadRepo domain.SettingReadRepository,
	auditSearchRepo domain.AuditLogSearchRepository,
) *AdminQueryService {
	return &AdminQueryService{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		auditRepo:       auditRepo,
		settingRepo:     settingRepo,
		approvalRepo:    approvalRepo,
		userReadRepo:    userReadRepo,
		settingReadRepo: settingReadRepo,
		auditSearchRepo: auditSearchRepo,
	}
}

// --- Admin Queries ---

func (q *AdminQueryService) GetAdminProfile(ctx context.Context, id uint) (*domain.AdminUser, error) {
	if q.userReadRepo != nil {
		if cached, err := q.userReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	user, err := q.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user != nil && q.userReadRepo != nil {
		_ = q.userReadRepo.Save(ctx, user)
	}
	return user, nil
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
	if q.settingReadRepo != nil {
		if cached, err := q.settingReadRepo.GetByKey(ctx, key); err == nil && cached != nil {
			return cached, nil
		}
	}
	setting, err := q.settingRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if setting != nil && q.settingReadRepo != nil {
		_ = q.settingReadRepo.Save(ctx, setting)
	}
	return setting, nil
}

// --- Audit Queries ---

func (q *AdminQueryService) ListAuditLogs(ctx context.Context, adminID uint, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	filter := make(map[string]any)
	if adminID > 0 {
		filter["user_id"] = adminID
	}
	if q.auditSearchRepo != nil {
		var userID *uint
		if adminID > 0 {
			userID = &adminID
		}
		list, total, err := q.auditSearchRepo.Search(ctx, userID, nil, nil, (page-1)*pageSize, pageSize)
		if err == nil {
			return list, total, nil
		}
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
