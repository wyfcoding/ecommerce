package domain

import (
	"gorm.io/gorm"
)

// SystemSetting 系统配置
type SystemSetting struct {
	gorm.Model
	Key         string `gorm:"column:key;type:varchar(100);uniqueIndex;not null;comment:配置键"`
	Value       string `gorm:"column:value;type:text;comment:配置值"`
	Description string `gorm:"column:description;type:varchar(255);comment:描述"`
}
