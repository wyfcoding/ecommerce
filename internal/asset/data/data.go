package data

import (
	"gorm.io/gorm"
)

// Data 结构体持有所有数据库连接实例。
type Data struct {
	db *gorm.DB
}

// NewData 创建一个新的 Data 实例。
func NewData(db *gorm.DB) (*Data, func()) {
	cleanup := func() {
		// 这里可以放置 Data 实例的清理逻辑，例如关闭非 GORM 管理的连接
	}
	return &Data{db: db}, cleanup
}
