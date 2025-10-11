package data

import (
	"gorm.io/gorm" // Even if not directly used, common to have
)

// Data 结构体持有所有数据库连接实例。
type Data struct {
	db *gorm.DB // Placeholder, not used for ClickHouse
}

// NewData 创建一个新的 Data 实例。
func NewData(db *gorm.DB) (*Data, func()) {
	cleanup := func() {
		// 这里可以放置 Data 实例的清理逻辑
	}
	return &Data{db: db}, cleanup
}
