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

// NewAdminRepository 创建并返回一个新的 AdminRepository 实例。
func NewAdminRepository(db *gorm.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *adminRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *adminRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *adminRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *adminRepository) Create(ctx context.Context, user *domain.AdminUser) error {
	return r.CreateInTx(ctx, r.db, user)
}

func (r *adminRepository) CreateInTx(ctx context.Context, tx any, user *domain.AdminUser) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toAdminUserModel(user)
	if err := gormTx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *adminRepository) GetByID(ctx context.Context, id uint) (*domain.AdminUser, error) {
	var model AdminUserModel
	if err := r.db.WithContext(ctx).Preload("Roles").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAdminUser(&model), nil
}

func (r *adminRepository) GetByUsername(ctx context.Context, username string) (*domain.AdminUser, error) {
	var model AdminUserModel
	if err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Where("username = ?", username).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAdminUser(&model), nil
}

func (r *adminRepository) Update(ctx context.Context, user *domain.AdminUser) error {
	return r.UpdateInTx(ctx, r.db, user)
}

func (r *adminRepository) UpdateInTx(ctx context.Context, tx any, user *domain.AdminUser) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toAdminUserModel(user)
	if err := gormTx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	user.ID = model.ID
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *adminRepository) Delete(ctx context.Context, id uint) error {
	return r.DeleteInTx(ctx, r.db, id)
}

func (r *adminRepository) DeleteInTx(ctx context.Context, tx any, id uint) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.WithContext(ctx).Delete(&AdminUserModel{}, id).Error
}

func (r *adminRepository) List(ctx context.Context, page, pageSize int) ([]*domain.AdminUser, int64, error) {
	var users []*AdminUserModel
	var total int64
	offset := (page - 1) * pageSize

	db := r.db.WithContext(ctx).Model(&AdminUserModel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.AdminUser, len(users))
	for i, u := range users {
		items[i] = toAdminUser(u)
	}

	return items, total, nil
}

func (r *adminRepository) AssignRole(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.AssignRoleInTx(ctx, r.db, userID, roleIDs)
}

func (r *adminRepository) AssignRoleInTx(ctx context.Context, tx any, userID uint, roleIDs []uint) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	var user AdminUserModel
	if err := gormTx.WithContext(ctx).First(&user, userID).Error; err != nil {
		return err
	}
	var roles []RoleModel
	if err := gormTx.WithContext(ctx).Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return err
	}
	return gormTx.Model(&user).Association("Roles").Replace(roles)
}

func (r *adminRepository) GetUserRoles(ctx context.Context, userID uint) ([]domain.Role, error) {
	var user AdminUserModel
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, err
	}
	if len(user.Roles) == 0 {
		return nil, nil
	}
	roles := make([]domain.Role, len(user.Roles))
	for i, role := range user.Roles {
		rp := toRole(&role)
		if rp != nil {
			roles[i] = *rp
		}
	}
	return roles, nil
}

func (r *adminRepository) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
	var user AdminUserModel
	if err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		First(&user, userID).Error; err != nil {
		return nil, err
	}

	permMap := make(map[string]struct{})
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			permMap[perm.Code] = struct{}{}
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

// NewRoleRepository 创建并返回一个新的 RoleRepository 实例。
func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *roleRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *roleRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *roleRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *roleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	return r.CreateRoleInTx(ctx, r.db, role)
}

func (r *roleRepository) CreateRoleInTx(ctx context.Context, tx any, role *domain.Role) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toRoleModel(role)
	if err := gormTx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	role.ID = model.ID
	role.CreatedAt = model.CreatedAt
	role.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *roleRepository) GetRoleByID(ctx context.Context, id uint) (*domain.Role, error) {
	var role RoleModel
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, err
	}
	return toRole(&role), nil
}

func (r *roleRepository) GetRoleByCode(ctx context.Context, code string) (*domain.Role, error) {
	var role RoleModel
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return toRole(&role), nil
}

func (r *roleRepository) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	var roles []*RoleModel
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.Role, len(roles))
	for i, role := range roles {
		items[i] = toRole(role)
	}
	return items, nil
}

func (r *roleRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	return r.UpdateRoleInTx(ctx, r.db, role)
}

func (r *roleRepository) UpdateRoleInTx(ctx context.Context, tx any, role *domain.Role) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toRoleModel(role)
	if err := gormTx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	role.ID = model.ID
	role.CreatedAt = model.CreatedAt
	role.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *roleRepository) DeleteRole(ctx context.Context, id uint) error {
	return r.DeleteRoleInTx(ctx, r.db, id)
}

func (r *roleRepository) DeleteRoleInTx(ctx context.Context, tx any, id uint) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.WithContext(ctx).Delete(&RoleModel{}, id).Error
}

func (r *roleRepository) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	return r.CreatePermissionInTx(ctx, r.db, perm)
}

func (r *roleRepository) CreatePermissionInTx(ctx context.Context, tx any, perm *domain.Permission) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toPermissionModel(perm)
	if err := gormTx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	perm.ID = model.ID
	perm.CreatedAt = model.CreatedAt
	perm.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *roleRepository) GetPermissionByID(ctx context.Context, id uint) (*domain.Permission, error) {
	var perm PermissionModel
	if err := r.db.WithContext(ctx).First(&perm, id).Error; err != nil {
		return nil, err
	}
	return toPermission(&perm), nil
}

func (r *roleRepository) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	var perms []*PermissionModel
	if err := r.db.WithContext(ctx).Find(&perms).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.Permission, len(perms))
	for i, perm := range perms {
		items[i] = toPermission(perm)
	}
	return items, nil
}

func (r *roleRepository) AssignPermissions(ctx context.Context, roleID uint, permIDs []uint) error {
	return r.AssignPermissionsInTx(ctx, r.db, roleID, permIDs)
}

func (r *roleRepository) AssignPermissionsInTx(ctx context.Context, tx any, roleID uint, permIDs []uint) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	var role RoleModel
	if err := gormTx.WithContext(ctx).First(&role, roleID).Error; err != nil {
		return err
	}
	var perms []PermissionModel
	if err := gormTx.WithContext(ctx).Where("id IN ?", permIDs).Find(&perms).Error; err != nil {
		return err
	}
	return gormTx.Model(&role).Association("Permissions").Replace(perms)
}

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 创建并返回一个新的 AuditRepository 实例。
func NewAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Save(ctx context.Context, log *domain.AuditLog) error {
	return r.SaveInTx(ctx, r.db, log)
}

func (r *auditRepository) SaveInTx(ctx context.Context, tx any, log *domain.AuditLog) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toAuditLogModel(log)
	if err := gormTx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	log.ID = model.ID
	log.CreatedAt = model.CreatedAt
	log.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *auditRepository) GetByID(ctx context.Context, id uint) (*domain.AuditLog, error) {
	var model AuditLogModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAuditLog(&model), nil
}

func (r *auditRepository) Find(ctx context.Context, filter map[string]any, page, pageSize int) ([]*domain.AuditLog, int64, error) {
	var logs []*AuditLogModel
	var total int64
	offset := (page - 1) * pageSize

	db := r.db.WithContext(ctx).Model(&AuditLogModel{})

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

	items := make([]*domain.AuditLog, len(logs))
	for i, log := range logs {
		items[i] = toAuditLog(log)
	}

	return items, total, nil
}

func (r *auditRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

type approvalRepository struct {
	db *gorm.DB
}

// NewApprovalRepository 创建并返回一个新的 ApprovalRepository 实例。
func NewApprovalRepository(db *gorm.DB) domain.ApprovalRepository {
	return &approvalRepository{db: db}
}

func (r *approvalRepository) CreateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	return r.CreateRequestInTx(ctx, r.db, req)
}

func (r *approvalRepository) CreateRequestInTx(ctx context.Context, tx any, req *domain.ApprovalRequest) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toApprovalRequestModel(req)
	if err := gormTx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	req.ID = model.ID
	req.CreatedAt = model.CreatedAt
	req.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *approvalRepository) GetRequestByID(ctx context.Context, id uint) (*domain.ApprovalRequest, error) {
	var req ApprovalRequestModel
	if err := r.db.WithContext(ctx).Preload("Logs").First(&req, id).Error; err != nil {
		return nil, err
	}
	return toApprovalRequest(&req), nil
}

func (r *approvalRepository) UpdateRequest(ctx context.Context, req *domain.ApprovalRequest) error {
	return r.UpdateRequestInTx(ctx, r.db, req)
}

func (r *approvalRepository) UpdateRequestInTx(ctx context.Context, tx any, req *domain.ApprovalRequest) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toApprovalRequestModel(req)
	if err := gormTx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	req.ID = model.ID
	req.CreatedAt = model.CreatedAt
	req.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *approvalRepository) ListPendingRequests(ctx context.Context, roleLimit string) ([]*domain.ApprovalRequest, error) {
	var reqs []*ApprovalRequestModel
	db := r.db.WithContext(ctx).Where("status = ?", domain.ApprovalStatusPending)

	if roleLimit != "" {
		db = db.Where("approver_role = ?", roleLimit)
	}

	if err := db.Order("created_at asc").Find(&reqs).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.ApprovalRequest, len(reqs))
	for i, req := range reqs {
		items[i] = toApprovalRequest(req)
	}
	return items, nil
}

func (r *approvalRepository) AddLog(ctx context.Context, log *domain.ApprovalLog) error {
	return r.AddLogInTx(ctx, r.db, log)
}

func (r *approvalRepository) AddLogInTx(ctx context.Context, tx any, log *domain.ApprovalLog) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toApprovalLogModel(log)
	if err := gormTx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	log.ID = model.ID
	log.CreatedAt = model.CreatedAt
	log.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *approvalRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

type settingRepository struct {
	db *gorm.DB
}

// NewSettingRepository 创建并返回一个新的 SettingRepository 实例。
func NewSettingRepository(db *gorm.DB) domain.SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error) {
	var setting SystemSettingModel
	if err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return toSystemSetting(&setting), nil
}

func (r *settingRepository) Save(ctx context.Context, setting *domain.SystemSetting) error {
	return r.SaveInTx(ctx, r.db, setting)
}

func (r *settingRepository) SaveInTx(ctx context.Context, tx any, setting *domain.SystemSetting) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	model := toSystemSettingModel(setting)
	if err := gormTx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	setting.ID = model.ID
	setting.CreatedAt = model.CreatedAt
	setting.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *settingRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
