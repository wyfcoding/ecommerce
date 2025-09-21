package data

import (
	"time"

	"gorm.io/gorm"
)

// AdminUser 是管理员用户的数据库模型。
type AdminUser struct {
	ID        uint32 `gorm:"primarykey"`
	Username  string `gorm:"uniqueIndex;not null"`
	Password  string `gorm:"not null"`
	Name      string
	Status    int32
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Role 是角色模型。
type Role struct {
	ID        uint32 `gorm:"primarykey"`
	Name      string `gorm:"uniqueIndex;not null"`
	Slug      string `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Permission 是权限模型。
type Permission struct {
	ID        uint32 `gorm:"primarykey"`
	Name      string `gorm:"uniqueIndex;not null"`
	Slug      string `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// AdminUserRole 是管理员用户和角色关联表。
type AdminUserRole struct {
	AdminUserID uint32 `gorm:"primaryKey;autoIncrement:false"`
	RoleID      uint32 `gorm:"primaryKey;autoIncrement:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (AdminUser) TableName() string     { return "admin_users" }
func (Role) TableName() string          { return "roles" }
func (Permission) TableName() string    { return "permissions" }
func (AdminUserRole) TableName() string { return "admin_user_roles" }
