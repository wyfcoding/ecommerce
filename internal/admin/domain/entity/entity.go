package entity

import (
	"errors"
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// 定义Admin模块的业务错误。
var (
	ErrAdminNotFound      = errors.New("admin not found")                // 管理员未找到。
	ErrRoleNotFound       = errors.New("role not found")                 // 角色未找到。
	ErrPermissionNotFound = errors.New("permission not found")           // 权限未找到。
	ErrUsernameExists     = errors.New("username already exists")        // 用户名已存在。
	ErrEmailExists        = errors.New("email already exists")           // 邮箱已存在。
	ErrRoleCodeExists     = errors.New("role code already exists")       // 角色编码已存在。
	ErrPermCodeExists     = errors.New("permission code already exists") // 权限编码已存在。
)

// AdminStatus 定义了管理员账户的状态。
type AdminStatus int8

const (
	AdminStatusActive   AdminStatus = 1 // 激活状态，可以正常登录和操作。
	AdminStatusInactive AdminStatus = 2 // 停用状态，账户被禁用。
	AdminStatusLocked   AdminStatus = 3 // 锁定状态，通常由于多次登录失败导致。
)

// Admin 实体代表一个后台管理员用户。
// 它包含了管理员的个人信息、认证信息、账户状态以及关联的角色和权限。
type Admin struct {
	gorm.Model                       // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Username         string          `gorm:"type:varchar(64);uniqueIndex;not null;comment:用户名" json:"username"` // 用户名，唯一索引，不允许为空。
	Email            string          `gorm:"type:varchar(128);uniqueIndex;not null;comment:邮箱" json:"email"`    // 邮箱，唯一索引，不允许为空。
	Password         string          `gorm:"type:varchar(128);not null;comment:密码" json:"-"`                    // 密码哈希值，不允许为空，JSON序列化时忽略。
	RealName         string          `gorm:"type:varchar(64);comment:真实姓名" json:"real_name"`                    // 真实姓名。
	Phone            string          `gorm:"type:varchar(20);comment:手机号" json:"phone"`                         // 手机号。
	Avatar           string          `gorm:"type:varchar(255);comment:头像" json:"avatar"`                        // 头像URL。
	Status           AdminStatus     `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`          // 账户状态，默认为激活。
	LoginAttempts    int             `gorm:"not null;default:0;comment:登录失败次数" json:"login_attempts"`           // 连续登录失败次数。
	LastLoginAt      *time.Time      `gorm:"comment:最后登录时间" json:"last_login_at"`                               // 最后一次登录的时间。
	LastLoginIP      string          `gorm:"type:varchar(64);comment:最后登录IP" json:"last_login_ip"`              // 最后一次登录的IP地址。
	PasswordExpiry   *time.Time      `gorm:"comment:密码过期时间" json:"password_expiry"`                             // 密码的过期时间。
	MustChangePass   bool            `gorm:"not null;default:true;comment:是否必须修改密码" json:"must_change_pass"`    // 首次登录或密码过期后是否强制修改密码。
	TwoFactorSecret  string          `gorm:"type:varchar(64);comment:2FA密钥" json:"-"`                           // 两步验证（2FA）的密钥，JSON序列化时忽略。
	TwoFactorEnabled bool            `gorm:"not null;default:false;comment:是否启用2FA" json:"two_factor_enabled"`  // 是否启用了两步验证。
	Roles            []*Role         `gorm:"many2many:admin_roles;" json:"roles"`                               // 管理员拥有的角色列表，多对多关系。
	Permissions      []*Permission   `gorm:"many2many:admin_permissions;" json:"permissions"`                   // 管理员拥有的直接权限列表，多对多关系。
	LoginLogs        []*LoginLog     `gorm:"foreignKey:AdminID" json:"login_logs"`                              // 管理员的登录日志列表，一对多关系。
	OperationLogs    []*OperationLog `gorm:"foreignKey:AdminID" json:"operation_logs"`                          // 管理员的操作日志列表，一对多关系。
}

// Role 实体代表一个后台管理系统中的角色。
// 角色是一组权限的集合，可以分配给多个管理员用户。
type Role struct {
	gorm.Model                // 嵌入gorm.Model。
	Name        string        `gorm:"type:varchar(64);not null;comment:角色名称" json:"name"`             // 角色名称。
	Code        string        `gorm:"type:varchar(64);uniqueIndex;not null;comment:角色编码" json:"code"` // 角色编码，唯一索引，不允许为空。
	Description string        `gorm:"type:varchar(255);comment:描述" json:"description"`                // 角色描述。
	Status      int           `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`       // 角色状态，例如1为启用，0为禁用。
	Sort        int           `gorm:"not null;default:0;comment:排序" json:"sort"`                      // 排序字段。
	Permissions []*Permission `gorm:"many2many:role_permissions;" json:"permissions"`                 // 角色拥有的权限列表，多对多关系。
	Admins      []*Admin      `gorm:"many2many:admin_roles;" json:"-"`                                // 拥有该角色的管理员列表，多对多关系，JSON序列化时忽略。
}

// Permission 实体代表一个后台管理系统中的权限项。
// 权限定义了用户可以执行的特定操作或访问的资源。
type Permission struct {
	gorm.Model               // 嵌入gorm.Model。
	Name       string        `gorm:"type:varchar(64);not null;comment:权限名称" json:"name"`                // 权限名称。
	Code       string        `gorm:"type:varchar(64);uniqueIndex;not null;comment:权限编码" json:"code"`    // 权限编码，唯一索引，不允许为空。
	Type       string        `gorm:"type:varchar(20);not null;comment:类型(menu/button/api)" json:"type"` // 权限类型，例如菜单、按钮、API等。
	ParentID   uint64        `gorm:"not null;default:0;comment:父级ID" json:"parent_id"`                  // 父权限的ID，用于构建权限树结构。
	Path       string        `gorm:"type:varchar(255);comment:路径" json:"path"`                          // 权限对应的资源路径（例如API路径）。
	Method     string        `gorm:"type:varchar(20);comment:方法" json:"method"`                         // 权限对应的HTTP方法（例如GET, POST）。
	Icon       string        `gorm:"type:varchar(64);comment:图标" json:"icon"`                           // 权限对应的图标。
	Sort       int           `gorm:"not null;default:0;comment:排序" json:"sort"`                         // 排序字段。
	Status     int           `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`          // 权限状态，例如1为启用，0为禁用。
	Roles      []*Role       `gorm:"many2many:role_permissions;" json:"-"`                              // 拥有该权限的角色列表，多对多关系，JSON序列化时忽略。
	Admins     []*Admin      `gorm:"many2many:admin_permissions;" json:"-"`                             // 拥有该权限的管理员列表，多对多关系，JSON序列化时忽略。
	Children   []*Permission `gorm:"-" json:"children"`                                                 // 子权限列表，用于构建树状结构，GORM忽略此字段。
}

// LoginLog 实体代表管理员的登录日志记录。
type LoginLog struct {
	gorm.Model        // 嵌入gorm.Model。
	AdminID    uint64 `gorm:"index;not null;comment:管理员ID" json:"admin_id"`          // 管理员ID，索引字段，不允许为空。
	IP         string `gorm:"type:varchar(64);not null;comment:IP地址" json:"ip"`      // 登录IP地址。
	UserAgent  string `gorm:"type:varchar(255);comment:UserAgent" json:"user_agent"` // 用户代理字符串。
	Location   string `gorm:"type:varchar(128);comment:地理位置" json:"location"`        // 登录的地理位置信息。
	Success    bool   `gorm:"not null;comment:是否成功" json:"success"`                  // 登录是否成功。
	Reason     string `gorm:"type:varchar(255);comment:失败原因" json:"reason"`          // 登录失败的原因。
}

// OperationLog 实体代表管理员的操作日志记录。
type OperationLog struct {
	gorm.Model        // 嵌入gorm.Model。
	AdminID    uint64 `gorm:"index;not null;comment:管理员ID" json:"admin_id"`          // 管理员ID，索引字段，不允许为空。
	Module     string `gorm:"type:varchar(64);not null;comment:模块" json:"module"`    // 操作所属的模块。
	Action     string `gorm:"type:varchar(64);not null;comment:动作" json:"action"`    // 执行的具体操作。
	Method     string `gorm:"type:varchar(20);not null;comment:请求方法" json:"method"`  // 请求的HTTP方法。
	Path       string `gorm:"type:varchar(255);not null;comment:请求路径" json:"path"`   // 请求的路径。
	IP         string `gorm:"type:varchar(64);not null;comment:IP地址" json:"ip"`      // 操作发生的IP地址。
	UserAgent  string `gorm:"type:varchar(255);comment:UserAgent" json:"user_agent"` // 用户代理字符串。
	Request    string `gorm:"type:text;comment:请求参数" json:"request"`                 // 请求的参数（JSON字符串或其他）。
	Response   string `gorm:"type:text;comment:响应结果" json:"response"`                // 响应的结果（JSON字符串或其他）。
	Duration   int64  `gorm:"not null;comment:耗时(ms)" json:"duration"`               // 操作的总耗时（毫秒）。
	Status     int    `gorm:"not null;comment:响应状态码" json:"status"`                  // HTTP响应状态码。
	ErrorMsg   string `gorm:"type:text;comment:错误信息" json:"error_msg"`               // 错误信息（如果操作失败）。
}

// NewAdmin 创建并返回一个新的 Admin 实体实例。
// username, email, password, realName, phone: 管理员的基本信息。
func NewAdmin(username, email, password, realName, phone string) *Admin {
	now := time.Now()
	// 设置初始密码过期时间为当前日期起三个月后。
	expiry := now.AddDate(0, 3, 0)
	return &Admin{
		Username:       username,
		Email:          email,
		Password:       password,
		RealName:       realName,
		Phone:          phone,
		Status:         AdminStatusActive, // 默认为激活状态。
		MustChangePass: true,              // 首次创建强制修改密码。
		PasswordExpiry: &expiry,           // 设置密码过期时间。
	}
}

// IsActive 检查管理员账户是否处于激活状态。
func (a *Admin) IsActive() bool {
	return a.Status == AdminStatusActive
}

// IsLocked 检查管理员账户是否处于锁定状态。
func (a *Admin) IsLocked() bool {
	return a.Status == AdminStatusLocked
}

// RecordLoginSuccess 记录管理员成功登录后的状态变更。
// ip: 本次登录的IP地址。
func (a *Admin) RecordLoginSuccess(ip string) {
	now := time.Now()
	a.LastLoginAt = &now // 更新最后登录时间。
	a.LastLoginIP = ip   // 更新最后登录IP。
	a.LoginAttempts = 0  // 重置登录失败次数。
}

// RecordLoginFailure 记录管理员登录失败后的状态变更。
// 如果连续失败次数达到阈值（例如5次），则锁定账户。
func (a *Admin) RecordLoginFailure() {
	a.LoginAttempts++         // 增加登录失败次数。
	if a.LoginAttempts >= 5 { // 假设连续失败5次则锁定账户。
		a.Status = AdminStatusLocked // 将账户状态设置为锁定。
	}
}
