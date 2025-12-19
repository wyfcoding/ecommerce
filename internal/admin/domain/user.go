package domain

import (
	"time"

	"gorm.io/gorm"
)

// AdminUser 代表后台管理员用户
// 拥有角色，通过角色获得权限
type AdminUser struct {
	gorm.Model
	Username     string     `gorm:"column:username;type:varchar(50);uniqueIndex;not null;comment:用户名"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希"`
	Email        string     `gorm:"column:email;type:varchar(100);uniqueIndex;not null;comment:邮箱"`
	FullName     string     `gorm:"column:full_name;type:varchar(100);comment:全名"`
	Status       UserStatus `gorm:"column:status;type:tinyint;default:1;comment:状态 1:启用 2:禁用"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at;comment:最后登录时间"`

	// 多对多关联角色
	Roles []Role `gorm:"many2many:admin_user_roles;"`
}

type UserStatus int

const (
	UserStatusActive   UserStatus = 1
	UserStatusDisabled UserStatus = 2
)

// Role 代表角色
// 角色是一组权限的集合
type Role struct {
	gorm.Model
	Name        string `gorm:"column:name;type:varchar(50);uniqueIndex;not null;comment:角色名称"`
	Code        string `gorm:"column:code;type:varchar(50);uniqueIndex;not null;comment:角色编码(如 SUPER_ADMIN)"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`

	// 多对多关联权限
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

// Permission 代表具体的权限点
// 通常对应某个资源的某个操作，如 order:view, product:edit
type Permission struct {
	gorm.Model
	Name        string `gorm:"column:name;type:varchar(100);not null;comment:权限名称"`
	Code        string `gorm:"column:code;type:varchar(100);uniqueIndex;not null;comment:权限编码(resource:action)"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`
	Resource    string `gorm:"column:resource;type:varchar(50);index;comment:资源类型"`
	Action      string `gorm:"column:action;type:varchar(50);comment:操作类型"`
	Type        string `gorm:"column:type;type:varchar(20);default:'api';comment:权限类型(menu/api/button)"`
	ParentID    uint   `gorm:"column:parent_id;default:0;comment:父权限ID"`
}
