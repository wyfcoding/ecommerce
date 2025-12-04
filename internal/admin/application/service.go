package application

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/admin/domain/entity"     // 导入领域实体。
	"github.com/wyfcoding/ecommerce/internal/admin/domain/repository" // 导入领域仓储接口。
	"github.com/wyfcoding/ecommerce/pkg/jwt"                          // 导入JWT工具包。

	"golang.org/x/crypto/bcrypt" // 导入密码哈希库。
	"log/slog"                   // 导入结构化日志库。
)

// AdminService 结构体定义了后台管理相关的应用服务。
// 它协调领域层和基础设施层，实现后台用户的注册、登录、权限管理等业务逻辑。
type AdminService struct {
	repo   repository.AdminRepository // 依赖AdminRepository接口，用于数据持久化操作。
	logger *slog.Logger               // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewAdminService 创建并返回一个新的 AdminService 实例。
func NewAdminService(repo repository.AdminRepository, logger *slog.Logger) *AdminService {
	return &AdminService{
		repo:   repo,
		logger: logger,
	}
}

// --- Admin methods ---

// RegisterAdmin 注册一个新的后台管理员用户。
// ctx: 上下文，用于控制操作的生命周期。
// username, email, password, realName, phone: 管理员的注册信息。
func (s *AdminService) RegisterAdmin(ctx context.Context, username, email, password, realName, phone string) (*entity.Admin, error) {
	// 检查用户名或邮箱是否已存在，避免重复注册。
	if _, err := s.repo.GetAdminByUsername(ctx, username); err == nil {
		return nil, entity.ErrUsernameExists
	}
	if _, err := s.repo.GetAdminByEmail(ctx, email); err == nil {
		return nil, entity.ErrEmailExists
	}

	// 对管理员密码进行哈希处理，存储哈希值而不是明文密码，提高安全性。
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建管理员实体。
	admin := entity.NewAdmin(username, email, string(hashedPassword), realName, phone)
	// 通过仓储接口将管理员实体保存到数据库。
	if err := s.repo.CreateAdmin(ctx, admin); err != nil {
		s.logger.ErrorContext(ctx, "failed to create admin", "username", username, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "admin created successfully", "admin_id", admin.ID, "username", username)

	return admin, nil
}

// Login 管理员登录。
// ctx: 上下文。
// username, password: 管理员输入的用户名和密码。
// ip: 登录请求的IP地址，用于记录登录日志。
// 返回生成的JWT令牌和可能发生的错误。
func (s *AdminService) Login(ctx context.Context, username, password, ip string) (string, error) {
	// 根据用户名从数据库获取管理员信息。
	admin, err := s.repo.GetAdminByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid credentials") // 凭证无效。
	}

	// 检查管理员账户状态。
	if !admin.IsActive() {
		return "", errors.New("account is inactive") // 账户未激活。
	}
	if admin.IsLocked() {
		return "", errors.New("account is locked") // 账户已锁定。
	}

	// 比较输入密码与存储的哈希密码是否匹配。
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		// 密码不匹配，记录登录失败。
		admin.RecordLoginFailure()
		s.repo.UpdateAdmin(ctx, admin) // 更新管理员信息，例如失败次数。
		s.repo.CreateLoginLog(ctx, &entity.LoginLog{
			AdminID: uint64(admin.ID),
			IP:      ip,
			Success: false,
			Reason:  "invalid password",
		})
		return "", errors.New("invalid credentials") // 凭证无效。
	}

	// 登录成功，记录登录成功状态。
	admin.RecordLoginSuccess(ip)
	s.repo.UpdateAdmin(ctx, admin) // 更新管理员信息。
	s.repo.CreateLoginLog(ctx, &entity.LoginLog{
		AdminID: uint64(admin.ID),
		IP:      ip,
		Success: true,
	})

	// 生成 JWT 令牌。
	// TODO: 生产环境应从配置中加载密钥和签发者信息。
	token, err := jwt.GenerateToken(uint64(admin.ID), admin.Username, "your-secret-key", "admin-service", 24*time.Hour, nil)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetAdminProfile 获取指定ID的管理员用户详情。
// ctx: 上下文。
// id: 管理员用户ID。
func (s *AdminService) GetAdminProfile(ctx context.Context, id uint64) (*entity.Admin, error) {
	return s.repo.GetAdminByID(ctx, id)
}

// ListAdmins 列出所有管理员用户，支持分页。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回管理员用户列表、总数和可能发生的错误。
func (s *AdminService) ListAdmins(ctx context.Context, page, pageSize int) ([]*entity.Admin, int64, error) {
	return s.repo.ListAdmins(ctx, page, pageSize)
}

// --- Role methods ---

// CreateRole 创建一个新的角色。
// ctx: 上下文。
// name, code, description: 角色的名称、编码和描述。
func (s *AdminService) CreateRole(ctx context.Context, name, code, description string) (*entity.Role, error) {
	// 检查角色编码是否已存在。
	if _, err := s.repo.GetRoleByCode(ctx, code); err == nil {
		return nil, entity.ErrRoleCodeExists
	}

	// 创建角色实体。
	role := &entity.Role{
		Name:        name,
		Code:        code,
		Description: description,
		Status:      1, // 默认状态为激活。
	}

	// 通过仓储接口保存角色。
	if err := s.repo.CreateRole(ctx, role); err != nil {
		s.logger.ErrorContext(ctx, "failed to create role", "role_code", code, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "role created successfully", "role_id", role.ID, "role_code", code)

	return role, nil
}

// ListRoles 列出所有角色，支持分页。
// ctx: 上下文。
// page, pageSize: 分页参数。
func (s *AdminService) ListRoles(ctx context.Context, page, pageSize int) ([]*entity.Role, int64, error) {
	return s.repo.ListRoles(ctx, page, pageSize)
}

// AssignRoleToAdmin 为管理员分配角色。
// ctx: 上下文。
// adminID: 管理员用户ID。
// roleID: 角色ID。
func (s *AdminService) AssignRoleToAdmin(ctx context.Context, adminID, roleID uint64) error {
	return s.repo.AssignRoleToAdmin(ctx, adminID, roleID)
}

// --- Permission methods ---

// CreatePermission 创建一个新的权限项。
// ctx: 上下文。
// name, code, permType, path, method: 权限的名称、编码、类型、路径和HTTP方法。
// parentID: 父权限ID，用于构建权限树。
func (s *AdminService) CreatePermission(ctx context.Context, name, code, permType, path, method string, parentID uint64) (*entity.Permission, error) {
	// 检查权限编码是否已存在。
	if _, err := s.repo.GetPermissionByCode(ctx, code); err == nil {
		return nil, entity.ErrPermCodeExists
	}

	// 创建权限实体。
	permission := &entity.Permission{
		Name:     name,
		Code:     code,
		Type:     permType,
		Path:     path,
		Method:   method,
		ParentID: parentID,
		Status:   1, // 默认状态为激活。
	}

	// 通过仓储接口保存权限。
	if err := s.repo.CreatePermission(ctx, permission); err != nil {
		s.logger.ErrorContext(ctx, "failed to create permission", "perm_code", code, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "permission created successfully", "perm_id", permission.ID, "perm_code", code)

	return permission, nil
}

// ListPermissions 列出所有权限项。
// ctx: 上下文。
func (s *AdminService) ListPermissions(ctx context.Context) ([]*entity.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

// AssignPermissionToRole 为角色分配权限。
// ctx: 上下文。
// roleID: 角色ID。
// permissionID: 权限ID。
func (s *AdminService) AssignPermissionToRole(ctx context.Context, roleID, permissionID uint64) error {
	return s.repo.AssignPermissionToRole(ctx, roleID, permissionID)
}

// CheckPermission 检查管理员是否具有访问指定路径和HTTP方法的权限。
// ctx: 上下文。
// adminID: 管理员用户ID。
// path, method: 待检查的资源路径和HTTP方法。
func (s *AdminService) CheckPermission(ctx context.Context, adminID uint64, path, method string) (bool, error) {
	// 获取管理员信息，包括其关联的角色。
	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return false, err
	}

	// 1. 检查直接权限（如果 Admin 实体中存在直接权限列表）。
	// TODO: 根据实际业务需求，可以在此添加管理员直接权限的检查逻辑。

	// 2. 检查角色权限。
	// 遍历管理员拥有的所有角色。
	for _, role := range admin.Roles {
		// 获取该角色拥有的所有权限。
		permissions, err := s.repo.GetPermissionsByRoleID(ctx, uint64(role.ID))
		if err != nil {
			// 如果获取权限失败，则跳过该角色，尝试下一个。
			continue
		}
		// 遍历角色的所有权限，检查是否匹配请求的路径和方法。
		for _, perm := range permissions {
			if perm.Path == path && perm.Method == method {
				return true, nil // 找到匹配权限，返回 true。
			}
			// TODO: 可以添加通配符支持，例如 "/users/*" 匹配 "/users/1" 和 "/users/profile"。
		}
	}

	return false, nil // 未找到匹配权限。
}
