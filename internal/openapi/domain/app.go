package domain

import (
	"gorm.io/gorm"
)

type AppStatus string

const (
	StatusActive   AppStatus = "ACTIVE"
	StatusDisabled AppStatus = "DISABLED"
)

// OpenApiApp 开放平台应用实体
type OpenApiApp struct {
	gorm.Model
	AppID       string    `gorm:"column:app_id;type:varchar(32);unique_index;not null"`
	UserID      string    `gorm:"column:user_id;type:varchar(32);index;not null"`
	AppName     string    `gorm:"column:app_name;type:varchar(100);not null"`
	Description string    `gorm:"column:description;type:text"`
	APIKey      string    `gorm:"column:api_key;type:varchar(64);unique_index;not null"`
	APISecret   string    `gorm:"column:api_secret;type:varchar(128);not null"`
	Status      AppStatus `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE'"`
	Scopes      string    `gorm:"column:scopes;type:text"` // 逗号分隔的权限权限范围
}

func (OpenApiApp) TableName() string { return "openapi_apps" }
