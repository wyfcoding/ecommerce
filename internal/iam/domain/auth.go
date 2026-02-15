// Package domain 统一身份认证领域模型
package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// SubjectType 主体类型
type SubjectType string

const (
	SubjectTypeUser     SubjectType = "USER"
	SubjectTypeApp      SubjectType = "APP"
	SubjectTypeService  SubjectType = "SERVICE"
	SubjectTypeMerchant SubjectType = "MERCHANT"
)

// Credential 认证凭证聚合根
type Credential struct {
	gorm.Model
	SubjectID    string      `gorm:"column:subject_id;type:varchar(64);uniqueIndex;not null"`
	SubjectType  SubjectType `gorm:"column:subject_type;type:varchar(32);not null"`
	Username     string      `gorm:"column:username;type:varchar(128);uniqueIndex;not null"`
	PasswordHash string      `gorm:"column:password_hash;type:varchar(255);not null"`
	Email        string      `gorm:"column:email;type:varchar(128);index"`
	Phone        string      `gorm:"column:phone;type:varchar(20);index"`
	Status       int         `gorm:"column:status;default:1"` // 1:Active 0:Disabled
	LastLoginAt  *time.Time  `gorm:"column:last_login_at"`
	TenantID     string      `gorm:"column:tenant_id;type:varchar(64)"` // 租户隔离
}

func (Credential) TableName() string { return "iam_credentials" }

// Role 角色
type Role struct {
	gorm.Model
	Code        string `gorm:"column:code;type:varchar(64);uniqueIndex;not null"`
	Name        string `gorm:"column:name;type:varchar(128);not null"`
	Description string `gorm:"column:description;type:varchar(255)"`
	TenantID    string `gorm:"column:tenant_id;type:varchar(64);index"` // 角色也可以归属租户
}

func (Role) TableName() string { return "iam_roles" }

// Permission 权限 (RBAC)
type Permission struct {
	gorm.Model
	Code     string `gorm:"column:code;type:varchar(128);uniqueIndex;not null"` // e.g. "order:view"
	Service  string `gorm:"column:service;type:varchar(64);not null"`           // "ecommerce-order"
	Resource string `gorm:"column:resource;type:varchar(64)"`
	Action   string `gorm:"column:action;type:varchar(32)"`
}

func (Permission) TableName() string { return "iam_permissions" }

// AuthLog 登录日志
type AuthLog struct {
	ID        uint64    `gorm:"primaryKey"`
	SubjectID string    `gorm:"index"`
	ClientIP  string    `gorm:"type:varchar(45)"`
	UserAgent string    `gorm:"type:varchar(255)"`
	Action    string    `gorm:"type:varchar(32)"` // LOGIN, LOGOUT, REFRESH
	Result    string    `gorm:"type:varchar(32)"` // SUCCESS, FAILED
	Reason    string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (AuthLog) TableName() string { return "iam_auth_logs" }

// Repository Interfaces
type CredentialRepository interface {
	Save(ctx context.Context, c *Credential) error
	GetByUsername(ctx context.Context, username string) (*Credential, error)
	GetBySubjectID(ctx context.Context, subjectID string) (*Credential, error)
	GetByPhone(ctx context.Context, phone string) (*Credential, error)
	Update(ctx context.Context, c *Credential) error
}

type RoleRepository interface {
	GetBySubjectID(ctx context.Context, subjectID string) ([]string, error) // Return List of Role Codes
	AssignRole(ctx context.Context, subjectID, roleCode string) error
}

type AuthLogRepository interface {
	Save(ctx context.Context, log *AuthLog) error
}
