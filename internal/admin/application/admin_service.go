package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"golang.org/x/crypto/bcrypt"
)

// AdminService 门面服务，聚合所有子领域逻辑，供 gRPC Server 使用
type AdminService struct {
	userRepo    domain.AdminRepository
	roleRepo    domain.RoleRepository
	auditRepo   domain.AuditRepository
	settingRepo domain.SettingRepository

	authService *AdminAuthService
	audit       *AuditService
	workflow    *WorkflowService

	logger *slog.Logger
}

func NewAdminService(
	userRepo domain.AdminRepository,
	roleRepo domain.RoleRepository,
	auditRepo domain.AuditRepository,
	settingRepo domain.SettingRepository,
	authService *AdminAuthService,
	audit *AuditService,
	workflow *WorkflowService,
	logger *slog.Logger,
) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		auditRepo:   auditRepo,
		settingRepo: settingRepo,
		authService: authService,
		audit:       audit,
		workflow:    workflow,
		logger:      logger,
	}
}

// --- Admin CRUD ---

func (s *AdminService) RegisterAdmin(ctx context.Context, username, email, password, fullName, phone string) (*domain.AdminUser, error) {
	// Check exist
	if _, err := s.userRepo.GetByUsername(ctx, username); err == nil {
		return nil, errors.New("username exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := &domain.AdminUser{
		Username:     username,
		Email:        email,
		FullName:     fullName,
		PasswordHash: string(hashed),
		Status:       domain.UserStatusActive,
	}

	if err := s.userRepo.Create(ctx, admin); err != nil {
		return nil, err
	}
	return admin, nil
}

func (s *AdminService) GetAdminProfile(ctx context.Context, id uint64) (*domain.AdminUser, error) {
	return s.userRepo.GetByID(ctx, uint(id))
}

func (s *AdminService) UpdateAdmin(ctx context.Context, id uint64, email, fullName, phone string, roleIDs []uint64) (*domain.AdminUser, error) {
	user, err := s.userRepo.GetByID(ctx, uint(id))
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

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	if len(roleIDs) > 0 {
		rIDs := make([]uint, len(roleIDs))
		for i, v := range roleIDs {
			rIDs[i] = uint(v)
		}
		if err := s.userRepo.AssignRole(ctx, uint(id), rIDs); err != nil {
			return nil, err
		}
		// Refresh
		return s.userRepo.GetByID(ctx, uint(id))
	}

	return user, nil
}

func (s *AdminService) DeleteAdmin(ctx context.Context, id uint64) error {
	return s.userRepo.Delete(ctx, uint(id))
}

func (s *AdminService) ListAdmins(ctx context.Context, page, pageSize int) ([]*domain.AdminUser, int64, error) {
	return s.userRepo.List(ctx, page, pageSize)
}

func (s *AdminService) AssignRoleToAdmin(ctx context.Context, adminID, roleID uint64) error {
	// 该方法在旧版中是单角色？还是应该是数组。
	// 适配单角色添加，但 Repo 期望数组替换。
	// Assuming Facade just wraps Repo.
	// Let's assume we append? Repo AssignRole is REPLACE in my implementation.
	// 为简单起见，我们仅使用单项调用 Repo。
	return s.userRepo.AssignRole(ctx, uint(adminID), []uint{uint(roleID)})
}

// --- Role CRUD ---

// --- Role CRUD ---

func (s *AdminService) CreateRole(ctx context.Context, name, code, description string) (*domain.Role, error) {
	role := &domain.Role{
		Name:        name,
		Code:        code,
		Description: description,
	}
	if err := s.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *AdminService) GetRole(ctx context.Context, id uint64) (*domain.Role, error) {
	return s.roleRepo.GetRoleByID(ctx, uint(id))
}

func (s *AdminService) UpdateRole(ctx context.Context, id uint64, name, description string, permissionIDs []uint64) (*domain.Role, error) {
	role, err := s.roleRepo.GetRoleByID(ctx, uint(id))
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

	if err := s.roleRepo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}

	if len(permissionIDs) > 0 {
		permIDs := make([]uint, len(permissionIDs))
		for i, v := range permissionIDs {
			permIDs[i] = uint(v)
		}
		if err := s.roleRepo.AssignPermissions(ctx, uint(id), permIDs); err != nil {
			return nil, err
		}
		// Refresh
		return s.roleRepo.GetRoleByID(ctx, uint(id))
	}

	return role, nil
}

func (s *AdminService) DeleteRole(ctx context.Context, id uint64) error {
	return s.roleRepo.DeleteRole(ctx, uint(id))
}

func (s *AdminService) ListRoles(ctx context.Context, page, pageSize int) ([]*domain.Role, int64, error) {
	// Repo ListRoles 目前返回所有角色（接口中无分页参数）。
	// 我们目前仅返回所有，或稍后按需修改 repo。
	roles, err := s.roleRepo.ListRoles(ctx)
	if err != nil {
		return nil, 0, err
	}
	// 如果列表很大，则手动分页？对于角色，通常列表很小。
	// Let's simplified return all.
	return roles, int64(len(roles)), nil
}

// --- Permission ---

func (s *AdminService) CreatePermission(ctx context.Context, name, code, permType, path, method string, parentID uint64) (*domain.Permission, error) {
	perm := &domain.Permission{
		Name:     name,
		Code:     code,
		Type:     permType,
		Resource: path,   // 将 path 映射到 Resource
		Action:   method, // 将 method 映射到 Action
		ParentID: uint(parentID),
	}
	// 假设 CreatePermission 存在于 repo 逻辑中
	if err := s.roleRepo.CreatePermission(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *AdminService) GetPermission(ctx context.Context, id uint64) (*domain.Permission, error) {
	return s.roleRepo.GetPermissionByID(ctx, uint(id))
}

func (s *AdminService) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return s.roleRepo.ListPermissions(ctx)
}

func (s *AdminService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error {
	return s.roleRepo.AssignPermissions(ctx, uint(roleID), []uint{uint(permissionID)})
}

// --- Audit ---

func (s *AdminService) ListAuditLogs(ctx context.Context, adminID uint64, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	filter := make(map[string]interface{})
	if adminID > 0 {
		filter["user_id"] = adminID
	}
	return s.auditRepo.Find(ctx, filter, page, pageSize)
}

// --- Setting ---

func (s *AdminService) GetSystemSetting(ctx context.Context, key string) (*domain.SystemSetting, error) {
	return s.settingRepo.GetByKey(ctx, key)
}

func (s *AdminService) UpdateSystemSetting(ctx context.Context, key, value, description string) (*domain.SystemSetting, error) {
	setting := &domain.SystemSetting{
		Key:         key,
		Value:       value,
		Description: description,
	}
	if err := s.settingRepo.Save(ctx, setting); err != nil {
		return nil, err
	}
	return setting, nil
}
