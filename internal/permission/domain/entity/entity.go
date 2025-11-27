package entity

import (
	"gorm.io/gorm"
)

// Permission 权限实体
type Permission struct {
	gorm.Model
	Code        string `gorm:"type:varchar(64);uniqueIndex;not null;comment:权限代码" json:"code"`
	Description string `gorm:"type:varchar(255);comment:描述" json:"description"`
}

// Role 角色实体
type Role struct {
	gorm.Model
	Name        string        `gorm:"type:varchar(64);uniqueIndex;not null;comment:角色名称" json:"name"`
	Description string        `gorm:"type:varchar(255);comment:描述" json:"description"`
	Permissions []*Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// UserRole 用户角色关联
type UserRole struct {
	gorm.Model
	UserID uint64 `gorm:"uniqueIndex:idx_user_role;not null;comment:用户ID" json:"user_id"`
	RoleID uint64 `gorm:"uniqueIndex:idx_user_role;not null;comment:角色ID" json:"role_id"`
	Role   Role   `gorm:"foreignKey:RoleID" json:"role"`
}
