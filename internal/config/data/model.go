package data

import (
	"time"

	"gorm.io/gorm"
)

// ConfigEntry represents a configuration entry.
type ConfigEntry struct {
	gorm.Model
	Key         string `gorm:"uniqueIndex;not null;size:255;comment:配置键" json:"key"`
	Value       string `gorm:"type:text;not null;comment:配置值" json:"value"`
	Description string `gorm:"type:text;comment:配置描述" json:"description"`
}

// TableName specifies the table name for ConfigEntry.
func (ConfigEntry) TableName() string {
	return "config_entries"
}
