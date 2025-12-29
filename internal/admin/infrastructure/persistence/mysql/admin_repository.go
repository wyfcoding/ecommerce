package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository 定义了数据持久层接口。
func NewAdminRepository(db *gorm.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) Create(ctx context.Context, user *domain.AdminUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *adminRepository) GetByID(ctx context.Context, id uint) (*domain.AdminUser, error) {
	var user domain.AdminUser
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到
		}
		return nil, err
	}
	return &user, nil
}

func (r *adminRepository) GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	var user domain.AdminUser
	if err := r.db.WithContext(ctx).Preload("Roles.Permissions").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *adminRepository) Update(ctx context.Context, user *domain.AdminUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *adminRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.AdminUser{}, id).Error
}

func (r *adminRepository) List(ctx context.Context, page, pageSize int) ([]*domain.AdminUser, int64, error) {
	var users []*domain.AdminUser
	var total int64
	offset := (page - 1) * pageSize

	db := r.db.WithContext(ctx).Model(&domain.AdminUser{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *adminRepository) AssignRole(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user domain.AdminUser
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		var roles []domain.Role
		if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return err
		}

		// GORM Replace 会替换掉当前的关联
		return tx.Model(&user).Association("Roles").Replace(roles)
	})
}

func (r *adminRepository) GetUserRoles(ctx context.Context, userID uint) ([]domain.Role, error) {
	var user domain.AdminUser
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return user.Roles, nil
}

// GetUserPermissions 获取用户的所有权限Code（去重）
func (r *adminRepository) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	var user domain.AdminUser
	// 深度预加载：User -> Roles -> Permissions
	if err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		First(&user, userID).Error; err != nil {
		return nil, err
	}

	permMap := make(map[string]bool)
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			permMap[perm.Code] = true
		}
	}

	perms := make([]string, 0, len(permMap))
	for code := range permMap {
		perms = append(perms, code)
	}
	return perms, nil
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 定义了数据持久层接口。
func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

// 增删改查操作

func (r *roleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) GetRoleByID(ctx context.Context, id uint) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetRoleByCode(ctx context.Context, code string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	var roles []*domain.Role
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) DeleteRole(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, id).Error
}

// 增删改查操作

func (r *roleRepository) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

func (r *roleRepository) GetPermissionByID(ctx context.Context, id uint) (*domain.Permission, error) {
	var perm domain.Permission
	if err := r.db.WithContext(ctx).First(&perm, id).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *roleRepository) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	var perms []*domain.Permission
	if err := r.db.WithContext(ctx).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *roleRepository) AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role domain.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			return err
		}

		var perms []domain.Permission
		if err := tx.Where("id IN ?", permIDs).Find(&perms).Error; err != nil {
			return err
		}

		return tx.Model(&role).Association("Permissions").Replace(perms)
	})
}

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 定义了数据持久层接口。
func NewAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Save(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditRepository) Find(ctx context.Context, filter map[string]any, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64
	offset := (page - 1) * pageSize

	db := r.db.WithContext(ctx).Model(&domain.AuditLog{})

	if uid, ok := filter["user_id"]; ok {
		db = db.Where("user_id = ?", uid)
	}
	if action, ok := filter["action"]; ok {
		db = db.Where("action = ?", action)
	}
	if res, ok := filter["resource"]; ok {
		db = db.Where("resource = ?", res)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

type approvalRepository struct {
	db *gorm.DB
}

// NewApprovalRepository 定义了数据持久层接口。
func NewApprovalRepository(db *gorm.DB) domain.ApprovalRepository {
	return &approvalRepository{db: db}
}

func (r *approvalRepository) CreateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *approvalRepository) GetRequestByID(ctx context.Context, id uint) (*domain.ApprovalRequest, error) {
	var req domain.ApprovalRequest
	if err := r.db.WithContext(ctx).Preload("Logs").First(&req, id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *approvalRepository) UpdateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *approvalRepository) ListPendingRequests(ctx context.Context, roleLimit string) ([]*domain.ApprovalRequest, error) {
	var reqs []*domain.ApprovalRequest
	db := r.db.WithContext(ctx).Where("status = ?", domain.ApprovalStatusPending)

	if roleLimit != "" {
		db = db.Where("approver_role = ?", roleLimit)
	}

	if err := db.Order("created_at asc").Find(&reqs).Error; err != nil {
		return nil, err
	}
	return reqs, nil
}

func (r *approvalRepository) AddLog(ctx context.Context, log *domain.ApprovalLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

type settingRepository struct {
	db *gorm.DB
}

// NewSettingRepository 定义了数据持久层接口。
func NewSettingRepository(db *gorm.DB) domain.SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error) {
	var setting domain.SystemSetting
	if err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *settingRepository) Save(ctx context.Context, setting *domain.SystemSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}
