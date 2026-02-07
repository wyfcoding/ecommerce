package domain

import "time"

// Permission 实体代表系统中的一个权限。
type Permission struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
}

// Role 实体代表系统中的一个角色。
type Role struct {
	ID          uint          `json:"id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Permissions []*Permission `json:"permissions"`
}

// UserRole 实体代表用户与角色之间的关联关系。
type UserRole struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uint64    `json:"user_id"`
	RoleID    uint64    `json:"role_id"`
	Role      Role      `json:"role"`
}
