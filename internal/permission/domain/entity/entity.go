package entity

import (
	"gorm.io/gorm" // 导入GORM库。
)

// Permission 实体代表系统中的一个权限。
// 例如，“user:read”、“product:create”。
type Permission struct {
	gorm.Model         // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Code        string `gorm:"type:varchar(64);uniqueIndex;not null;comment:权限代码" json:"code"` // 权限代码，唯一索引，不允许为空。
	Description string `gorm:"type:varchar(255);comment:描述" json:"description"`                // 权限的简要描述。
}

// Role 实体代表系统中的一个角色。
// 角色是一组权限的集合。
type Role struct {
	gorm.Model                // 嵌入gorm.Model。
	Name        string        `gorm:"type:varchar(64);uniqueIndex;not null;comment:角色名称" json:"name"` // 角色名称，唯一索引，不允许为空。
	Description string        `gorm:"type:varchar(255);comment:描述" json:"description"`                // 角色的简要描述。
	Permissions []*Permission `gorm:"many2many:role_permissions;" json:"permissions"`                 // 角色与权限是多对多关系，通过 role_permissions 中间表关联。
}

// UserRole 实体代表用户与角色之间的关联关系。
// 这是一个连接实体，用于实现用户与角色的多对多关系。
type UserRole struct {
	gorm.Model        // 嵌入gorm.Model。
	UserID     uint64 `gorm:"uniqueIndex:idx_user_role;not null;comment:用户ID" json:"user_id"` // 用户ID，与RoleID共同构成唯一索引。
	RoleID     uint64 `gorm:"uniqueIndex:idx_user_role;not null;comment:角色ID" json:"role_id"` // 角色ID，与UserID共同构成唯一索引。
	Role       Role   `gorm:"foreignKey:RoleID" json:"role"`                                  // 关联的角色实体，方便GORM加载角色信息。
}
