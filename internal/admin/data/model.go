package data

import (
	"time"

	"gorm.io/gorm"
)

// AdminUser 是管理员用户的数据库模型。
type AdminUser struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null;type:varchar(64);comment:用户名" json:"username"`
	Password string `gorm:"not null;type:varchar(255);comment:密码" json:"password"`
	Name     string `gorm:"type:varchar(64);comment:姓名" json:"name"`
	Status   int32  `gorm:"not null;default:1;type:tinyint;comment:状态 1:正常 2:禁用" json:"status"`
}

// Role 是角色模型。
type Role struct {
	gorm.Model
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// Permission 是权限模型。
type Permission struct {
	gorm.Model
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// AdminUserRole 是管理员用户和角色关联表。
type AdminUserRole struct {
	AdminUserID uint32 `gorm:"primaryKey;autoIncrement:false" json:"adminUserId"`
	RoleID      uint32 `gorm:"primaryKey;autoIncrement:false" json:"roleId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (AdminUser) TableName() string     { return "admin_users" }
func (Role) TableName() string          { return "roles" }
func (Permission) TableName() string    { return "permissions" }
func (AdminUserRole) TableName() string { return "admin_user_roles" }
