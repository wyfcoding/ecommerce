package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"gorm.io/gorm"
)

// AdminUserModel 管理员用户写模型。
type AdminUserModel struct {
	gorm.Model
	Username     string            `gorm:"column:username;type:varchar(50);uniqueIndex;not null;comment:用户名"`
	PasswordHash string            `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希"`
	Email        string            `gorm:"column:email;type:varchar(100);uniqueIndex;not null;comment:邮箱"`
	FullName     string            `gorm:"column:full_name;type:varchar(100);comment:全名"`
	Status       domain.UserStatus `gorm:"column:status;type:tinyint;default:1;comment:状态 1:启用 2:禁用"`
	LastLoginAt  *time.Time        `gorm:"column:last_login_at;comment:最后登录时间"`

	Roles []RoleModel `gorm:"many2many:admin_user_roles;"`
}

// RoleModel 角色写模型。
type RoleModel struct {
	gorm.Model
	Name        string `gorm:"column:name;type:varchar(50);uniqueIndex;not null;comment:角色名称"`
	Code        string `gorm:"column:code;type:varchar(50);uniqueIndex;not null;comment:角色编码(如 SUPER_ADMIN)"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`

	Permissions []PermissionModel `gorm:"many2many:role_permissions;"`
}

// PermissionModel 权限写模型。
type PermissionModel struct {
	gorm.Model
	Name        string `gorm:"column:name;type:varchar(100);not null;comment:权限名称"`
	Code        string `gorm:"column:code;type:varchar(100);uniqueIndex;not null;comment:权限编码(resource:action)"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`
	Resource    string `gorm:"column:resource;type:varchar(50);index;comment:资源类型"`
	Action      string `gorm:"column:action;type:varchar(50);comment:操作类型"`
	Type        string `gorm:"column:type;type:varchar(20);default:'api';comment:权限类型(menu/api/button)"`
	ParentID    uint   `gorm:"column:parent_id;default:0;comment:父权限ID"`
}

// ApprovalRequestModel 审批申请写模型。
type ApprovalRequestModel struct {
	gorm.Model
	RequesterID uint   `gorm:"column:requester_id;index;not null;comment:申请人ID"`
	ActionType  string `gorm:"column:action_type;type:varchar(50);index;not null;comment:申请动作类型"`
	Description string `gorm:"column:description;type:varchar(255);comment:申请描述/理由"`
	Payload     string `gorm:"column:payload;type:text;comment:操作数据快照(JSON)"`

	Status      domain.ApprovalStatus `gorm:"column:status;type:tinyint;default:1;comment:审批状态"`
	CurrentStep int                   `gorm:"column:current_step;type:int;default:1;comment:当前审批步骤"`
	TotalSteps  int                   `gorm:"column:total_steps;type:int;default:1;comment:总步骤数"`

	ApproverRole string `gorm:"column:approver_role;type:varchar(50);comment:当前需要的审批角色Code"`

	FinalizedAt   *time.Time `gorm:"column:finalized_at;comment:流程结束时间"`
	FailureReason string     `gorm:"column:failure_reason;type:varchar(255);comment:执行失败原因"`
	RetryCount    int        `gorm:"column:retry_count;type:int;default:0;comment:重试次数"`

	Logs []ApprovalLogModel `gorm:"foreignKey:RequestID"`
}

// ApprovalLogModel 审批日志写模型。
type ApprovalLogModel struct {
	gorm.Model
	RequestID    uint                  `gorm:"column:request_id;index;not null;comment:关联申请ID"`
	ApproverID   uint                  `gorm:"column:approver_id;not null;comment:审批人ID"`
	ApproverName string                `gorm:"column:approver_name;type:varchar(50);comment:审批人姓名"`
	Action       domain.ApprovalAction `gorm:"column:action;type:tinyint;not null;comment:动作 1:通过 2:拒绝"`
	Comment      string                `gorm:"column:comment;type:varchar(255);comment:审批意见"`
}

// AuditLogModel 审计日志写模型。
type AuditLogModel struct {
	gorm.Model
	UserID    uint   `gorm:"column:user_id;index;not null;comment:操作人ID"`
	Username  string `gorm:"column:username;type:varchar(50);not null;comment:操作人用户名(冗余)"`
	Action    string `gorm:"column:action;type:varchar(50);index;not null;comment:操作动作"`
	Resource  string `gorm:"column:resource;type:varchar(50);index;not null;comment:资源类型"`
	TargetID  string `gorm:"column:target_id;type:varchar(50);index;comment:目标资源ID"`
	ClientIP  string `gorm:"column:client_ip;type:varchar(50);comment:客户端IP"`
	Payload   string `gorm:"column:payload;type:text;comment:请求参数(JSON)"`
	Result    string `gorm:"column:result;type:text;comment:操作结果/错误信息"`
	Status    int    `gorm:"column:status;type:tinyint;default:1;comment:结果状态 1:成功 0:失败"`
	UserAgent string `gorm:"column:user_agent;type:varchar(255);comment:UserAgent"`
}

// SystemSettingModel 系统配置写模型。
type SystemSettingModel struct {
	gorm.Model
	Key         string `gorm:"column:key;type:varchar(100);uniqueIndex;not null;comment:配置键"`
	Value       string `gorm:"column:value;type:text;comment:配置值"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`
}

func (AdminUserModel) TableName() string       { return "admin_users" }
func (RoleModel) TableName() string            { return "roles" }
func (PermissionModel) TableName() string      { return "permissions" }
func (ApprovalRequestModel) TableName() string { return "approval_requests" }
func (ApprovalLogModel) TableName() string     { return "approval_logs" }
func (AuditLogModel) TableName() string        { return "audit_logs" }
func (SystemSettingModel) TableName() string   { return "system_settings" }

func toAdminUserModel(user *domain.AdminUser) *AdminUserModel {
	if user == nil {
		return nil
	}
	return &AdminUserModel{
		Model: gorm.Model{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Email:        user.Email,
		FullName:     user.FullName,
		Status:       user.Status,
		LastLoginAt:  user.LastLoginAt,
	}
}

func toAdminUser(model *AdminUserModel) *domain.AdminUser {
	if model == nil {
		return nil
	}
	user := &domain.AdminUser{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		Username:     model.Username,
		PasswordHash: model.PasswordHash,
		Email:        model.Email,
		FullName:     model.FullName,
		Status:       model.Status,
		LastLoginAt:  model.LastLoginAt,
	}
	if len(model.Roles) > 0 {
		roles := make([]domain.Role, len(model.Roles))
		for i, r := range model.Roles {
			role := toRole(&r)
			if role != nil {
				roles[i] = *role
			}
		}
		user.Roles = roles
	}
	return user
}

func toRoleModel(role *domain.Role) *RoleModel {
	if role == nil {
		return nil
	}
	return &RoleModel{
		Model: gorm.Model{
			ID:        role.ID,
			CreatedAt: role.CreatedAt,
			UpdatedAt: role.UpdatedAt,
		},
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
	}
}

func toRole(model *RoleModel) *domain.Role {
	if model == nil {
		return nil
	}
	role := &domain.Role{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Name:        model.Name,
		Code:        model.Code,
		Description: model.Description,
	}
	if len(model.Permissions) > 0 {
		perms := make([]domain.Permission, len(model.Permissions))
		for i, p := range model.Permissions {
			perm := toPermission(&p)
			if perm != nil {
				perms[i] = *perm
			}
		}
		role.Permissions = perms
	}
	return role
}

func toPermissionModel(perm *domain.Permission) *PermissionModel {
	if perm == nil {
		return nil
	}
	return &PermissionModel{
		Model: gorm.Model{
			ID:        perm.ID,
			CreatedAt: perm.CreatedAt,
			UpdatedAt: perm.UpdatedAt,
		},
		Name:        perm.Name,
		Code:        perm.Code,
		Description: perm.Description,
		Resource:    perm.Resource,
		Action:      perm.Action,
		Type:        perm.Type,
		ParentID:    perm.ParentID,
	}
}

func toPermission(model *PermissionModel) *domain.Permission {
	if model == nil {
		return nil
	}
	return &domain.Permission{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Name:        model.Name,
		Code:        model.Code,
		Description: model.Description,
		Resource:    model.Resource,
		Action:      model.Action,
		Type:        model.Type,
		ParentID:    model.ParentID,
	}
}

func toApprovalRequestModel(req *domain.ApprovalRequest) *ApprovalRequestModel {
	if req == nil {
		return nil
	}
	return &ApprovalRequestModel{
		Model: gorm.Model{
			ID:        req.ID,
			CreatedAt: req.CreatedAt,
			UpdatedAt: req.UpdatedAt,
		},
		RequesterID:   req.RequesterID,
		ActionType:    req.ActionType,
		Description:   req.Description,
		Payload:       req.Payload,
		Status:        req.Status,
		CurrentStep:   req.CurrentStep,
		TotalSteps:    req.TotalSteps,
		ApproverRole:  req.ApproverRole,
		FinalizedAt:   req.FinalizedAt,
		FailureReason: req.FailureReason,
		RetryCount:    req.RetryCount,
	}
}

func toApprovalRequest(model *ApprovalRequestModel) *domain.ApprovalRequest {
	if model == nil {
		return nil
	}
	req := &domain.ApprovalRequest{
		ID:            model.ID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		RequesterID:   model.RequesterID,
		ActionType:    model.ActionType,
		Description:   model.Description,
		Payload:       model.Payload,
		Status:        model.Status,
		CurrentStep:   model.CurrentStep,
		TotalSteps:    model.TotalSteps,
		ApproverRole:  model.ApproverRole,
		FinalizedAt:   model.FinalizedAt,
		FailureReason: model.FailureReason,
		RetryCount:    model.RetryCount,
	}
	if len(model.Logs) > 0 {
		logs := make([]domain.ApprovalLog, len(model.Logs))
		for i, l := range model.Logs {
			logItem := toApprovalLog(&l)
			if logItem != nil {
				logs[i] = *logItem
			}
		}
		req.Logs = logs
	}
	return req
}

func toApprovalLogModel(log *domain.ApprovalLog) *ApprovalLogModel {
	if log == nil {
		return nil
	}
	return &ApprovalLogModel{
		Model: gorm.Model{
			ID:        log.ID,
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		RequestID:    log.RequestID,
		ApproverID:   log.ApproverID,
		ApproverName: log.ApproverName,
		Action:       log.Action,
		Comment:      log.Comment,
	}
}

func toApprovalLog(model *ApprovalLogModel) *domain.ApprovalLog {
	if model == nil {
		return nil
	}
	return &domain.ApprovalLog{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		RequestID:    model.RequestID,
		ApproverID:   model.ApproverID,
		ApproverName: model.ApproverName,
		Action:       model.Action,
		Comment:      model.Comment,
	}
}

func toAuditLogModel(log *domain.AuditLog) *AuditLogModel {
	if log == nil {
		return nil
	}
	return &AuditLogModel{
		Model: gorm.Model{
			ID:        log.ID,
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		UserID:    log.UserID,
		Username:  log.Username,
		Action:    log.Action,
		Resource:  log.Resource,
		TargetID:  log.TargetID,
		ClientIP:  log.ClientIP,
		Payload:   log.Payload,
		Result:    log.Result,
		Status:    log.Status,
		UserAgent: log.UserAgent,
	}
}

func toAuditLog(model *AuditLogModel) *domain.AuditLog {
	if model == nil {
		return nil
	}
	return &domain.AuditLog{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		UserID:    model.UserID,
		Username:  model.Username,
		Action:    model.Action,
		Resource:  model.Resource,
		TargetID:  model.TargetID,
		ClientIP:  model.ClientIP,
		Payload:   model.Payload,
		Result:    model.Result,
		Status:    model.Status,
		UserAgent: model.UserAgent,
	}
}

func toSystemSettingModel(setting *domain.SystemSetting) *SystemSettingModel {
	if setting == nil {
		return nil
	}
	return &SystemSettingModel{
		Model: gorm.Model{
			ID:        setting.ID,
			CreatedAt: setting.CreatedAt,
			UpdatedAt: setting.UpdatedAt,
		},
		Key:         setting.Key,
		Value:       setting.Value,
		Description: setting.Description,
	}
}

func toSystemSetting(model *SystemSettingModel) *domain.SystemSetting {
	if model == nil {
		return nil
	}
	return &domain.SystemSetting{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Key:         model.Key,
		Value:       model.Value,
		Description: model.Description,
	}
}
