package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrAdminNotFound      = errors.New("admin not found")
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionNotFound = errors.New("permission not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrRoleCodeExists     = errors.New("role code already exists")
	ErrPermCodeExists     = errors.New("permission code already exists")
)

// AdminStatus 管理员状态
type AdminStatus int8

const (
	AdminStatusActive   AdminStatus = 1 // 激活
	AdminStatusInactive AdminStatus = 2 // 停用
	AdminStatusLocked   AdminStatus = 3 // 锁定
)

// Admin 管理员实体
type Admin struct {
	gorm.Model
	Username         string          `gorm:"type:varchar(64);uniqueIndex;not null;comment:用户名" json:"username"`
	Email            string          `gorm:"type:varchar(128);uniqueIndex;not null;comment:邮箱" json:"email"`
	Password         string          `gorm:"type:varchar(128);not null;comment:密码" json:"-"`
	RealName         string          `gorm:"type:varchar(64);comment:真实姓名" json:"real_name"`
	Phone            string          `gorm:"type:varchar(20);comment:手机号" json:"phone"`
	Avatar           string          `gorm:"type:varchar(255);comment:头像" json:"avatar"`
	Status           AdminStatus     `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`
	LoginAttempts    int             `gorm:"not null;default:0;comment:登录失败次数" json:"login_attempts"`
	LastLoginAt      *time.Time      `gorm:"comment:最后登录时间" json:"last_login_at"`
	LastLoginIP      string          `gorm:"type:varchar(64);comment:最后登录IP" json:"last_login_ip"`
	PasswordExpiry   *time.Time      `gorm:"comment:密码过期时间" json:"password_expiry"`
	MustChangePass   bool            `gorm:"not null;default:true;comment:是否必须修改密码" json:"must_change_pass"`
	TwoFactorSecret  string          `gorm:"type:varchar(64);comment:2FA密钥" json:"-"`
	TwoFactorEnabled bool            `gorm:"not null;default:false;comment:是否启用2FA" json:"two_factor_enabled"`
	Roles            []*Role         `gorm:"many2many:admin_roles;" json:"roles"`
	Permissions      []*Permission   `gorm:"many2many:admin_permissions;" json:"permissions"`
	LoginLogs        []*LoginLog     `gorm:"foreignKey:AdminID" json:"login_logs"`
	OperationLogs    []*OperationLog `gorm:"foreignKey:AdminID" json:"operation_logs"`
}

// Role 角色实体
type Role struct {
	gorm.Model
	Name        string        `gorm:"type:varchar(64);not null;comment:角色名称" json:"name"`
	Code        string        `gorm:"type:varchar(64);uniqueIndex;not null;comment:角色编码" json:"code"`
	Description string        `gorm:"type:varchar(255);comment:描述" json:"description"`
	Status      int           `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`
	Sort        int           `gorm:"not null;default:0;comment:排序" json:"sort"`
	Permissions []*Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	Admins      []*Admin      `gorm:"many2many:admin_roles;" json:"-"`
}

// Permission 权限实体
type Permission struct {
	gorm.Model
	Name     string        `gorm:"type:varchar(64);not null;comment:权限名称" json:"name"`
	Code     string        `gorm:"type:varchar(64);uniqueIndex;not null;comment:权限编码" json:"code"`
	Type     string        `gorm:"type:varchar(20);not null;comment:类型(menu/button/api)" json:"type"`
	ParentID uint64        `gorm:"not null;default:0;comment:父级ID" json:"parent_id"`
	Path     string        `gorm:"type:varchar(255);comment:路径" json:"path"`
	Method   string        `gorm:"type:varchar(20);comment:方法" json:"method"`
	Icon     string        `gorm:"type:varchar(64);comment:图标" json:"icon"`
	Sort     int           `gorm:"not null;default:0;comment:排序" json:"sort"`
	Status   int           `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`
	Roles    []*Role       `gorm:"many2many:role_permissions;" json:"-"`
	Admins   []*Admin      `gorm:"many2many:admin_permissions;" json:"-"`
	Children []*Permission `gorm:"-" json:"children"`
}

// LoginLog 登录日志
type LoginLog struct {
	gorm.Model
	AdminID   uint64 `gorm:"index;not null;comment:管理员ID" json:"admin_id"`
	IP        string `gorm:"type:varchar(64);not null;comment:IP地址" json:"ip"`
	UserAgent string `gorm:"type:varchar(255);comment:UserAgent" json:"user_agent"`
	Location  string `gorm:"type:varchar(128);comment:地理位置" json:"location"`
	Success   bool   `gorm:"not null;comment:是否成功" json:"success"`
	Reason    string `gorm:"type:varchar(255);comment:失败原因" json:"reason"`
}

// OperationLog 操作日志
type OperationLog struct {
	gorm.Model
	AdminID   uint64 `gorm:"index;not null;comment:管理员ID" json:"admin_id"`
	Module    string `gorm:"type:varchar(64);not null;comment:模块" json:"module"`
	Action    string `gorm:"type:varchar(64);not null;comment:动作" json:"action"`
	Method    string `gorm:"type:varchar(20);not null;comment:请求方法" json:"method"`
	Path      string `gorm:"type:varchar(255);not null;comment:请求路径" json:"path"`
	IP        string `gorm:"type:varchar(64);not null;comment:IP地址" json:"ip"`
	UserAgent string `gorm:"type:varchar(255);comment:UserAgent" json:"user_agent"`
	Request   string `gorm:"type:text;comment:请求参数" json:"request"`
	Response  string `gorm:"type:text;comment:响应结果" json:"response"`
	Duration  int64  `gorm:"not null;comment:耗时(ms)" json:"duration"`
	Status    int    `gorm:"not null;comment:响应状态码" json:"status"`
	ErrorMsg  string `gorm:"type:text;comment:错误信息" json:"error_msg"`
}

// NewAdmin 创建管理员
func NewAdmin(username, email, password, realName, phone string) *Admin {
	now := time.Now()
	expiry := now.AddDate(0, 3, 0)
	return &Admin{
		Username:       username,
		Email:          email,
		Password:       password,
		RealName:       realName,
		Phone:          phone,
		Status:         AdminStatusActive,
		MustChangePass: true,
		PasswordExpiry: &expiry,
	}
}

// IsActive 检查是否激活
func (a *Admin) IsActive() bool {
	return a.Status == AdminStatusActive
}

// IsLocked 检查是否锁定
func (a *Admin) IsLocked() bool {
	return a.Status == AdminStatusLocked
}

// RecordLoginSuccess 记录登录成功
func (a *Admin) RecordLoginSuccess(ip string) {
	now := time.Now()
	a.LastLoginAt = &now
	a.LastLoginIP = ip
	a.LoginAttempts = 0
}

// RecordLoginFailure 记录登录失败
func (a *Admin) RecordLoginFailure() {
	a.LoginAttempts++
	if a.LoginAttempts >= 5 {
		a.Status = AdminStatusLocked
	}
}
