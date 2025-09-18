package data

import (
	"time"

	"gorm.io/gorm"
)

// UserBasic 用户基础信息模型
type UserBasic struct {
	ID        uint64 `gorm:"primarykey"`
	UserID    uint64 `gorm:"uniqueIndex:uk_user_id;not null"`
	Username  string `gorm:"uniqueIndex:uk_username;size:64;not null;default:''"`
	Nickname  string `gorm:"size:64;not null;default:''"`
	Avatar    string `gorm:"size:255;not null;default:''"`
	Gender    int8   `gorm:"not null;default:0"`
	Birthday  *time.Time
	Status    int `gorm:"not null;default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (UserBasic) TableName() string {
	return "user_basic"
}

// UserAuth 用户认证模型
type UserAuth struct {
	ID         uint64 `gorm:"primarykey"`
	UserID     uint64 `gorm:"index:idx_user_id;not null"`
	AuthType   string `gorm:"uniqueIndex:uk_auth_type_identifier;size:32;not null"`
	Identifier string `gorm:"uniqueIndex:uk_auth_type_identifier;size:128;not null"`
	Credential string `gorm:"size:255;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (UserAuth) TableName() string {
	return "user_auth"
}
